package bms

import (
	"errors"
	"sync"
	"unsafe"
)

// Names of the shared memory areas provided by Falcon BMS.
const (
	AreaFlightData  = "FalconSharedMemoryArea"
	AreaFlightData2 = "FalconSharedMemoryArea2"
	AreaOSBData     = "FalconSharedOsbMemoryArea"
	AreaStringData  = "FalconSharedMemoryAreaString"
	AreaDrawingData = "FalconSharedMemoryAreaDrawing"
)

// ErrNotOpen is returned when reading before Open succeeded.
var ErrNotOpen = errors.New("falcon BMS shared memory not open")

// Reader provides read-only access to the Falcon BMS shared memory areas.
//
// All getters return a copy (snapshot) of the data, so the result can be used
// without holding any locks. As BMS writes concurrently, a snapshot may contain
// values from two consecutive frames.
//
// The zero value is ready to use. Reader is safe for concurrent use.
type Reader struct {
	mu    sync.RWMutex
	areas map[string]*area
}

// Open maps the main FlightData area. It fails if BMS is not running.
// All other areas are mapped on first access by their getters.
func (r *Reader) Open() error {
	_, err := r.area(AreaFlightData)
	return err
}

// IsOpen reports whether the main FlightData area is mapped.
func (r *Reader) IsOpen() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.areas[AreaFlightData] != nil
}

// Close unmaps all areas.
func (r *Reader) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range r.areas {
		a.close()
	}
	r.areas = nil
}

// FlightData returns a snapshot of the "FalconSharedMemoryArea".
func (r *Reader) FlightData() (*FlightData, error) {
	return readStruct[FlightData](r, AreaFlightData)
}

// FlightData2 returns a snapshot of the "FalconSharedMemoryArea2".
func (r *Reader) FlightData2() (*FlightData2, error) {
	return readStruct[FlightData2](r, AreaFlightData2)
}

// OSBData returns a snapshot of the "FalconSharedOsbMemoryArea".
func (r *Reader) OSBData() (*OSBData, error) {
	return readStruct[OSBData](r, AreaOSBData)
}

// StringData returns the parsed "FalconSharedMemoryAreaString".
// Use FlightData2.StringAreaTime to detect changes instead of polling this.
func (r *Reader) StringData() (*StringData, error) {
	buf, err := r.readVariable(AreaStringData, func(fd2 *FlightData2) uint32 { return fd2.StringAreaSize })
	if err != nil {
		return nil, err
	}
	return ParseStringData(buf)
}

// DrawingData returns the parsed "FalconSharedMemoryAreaDrawing".
func (r *Reader) DrawingData() (*DrawingData, error) {
	buf, err := r.readVariable(AreaDrawingData, func(fd2 *FlightData2) uint32 { return fd2.DrawingAreaSize })
	if err != nil {
		return nil, err
	}
	return ParseDrawingData(buf)
}

// readStruct copies a fixed-size area into a T. If the area is smaller than T
// (older BMS version), the remaining fields stay zero.
func readStruct[T any](r *Reader, name string) (*T, error) {
	var v T
	err := r.withArea(name, func(a *area) {
		a.read(unsafe.Slice((*byte)(unsafe.Pointer(&v)), unsafe.Sizeof(v)))
	})
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// readVariable copies a variable-size area. Its size is taken from FlightData2.
func (r *Reader) readVariable(name string, size func(*FlightData2) uint32) ([]byte, error) {
	fd2, err := r.FlightData2()
	if err != nil {
		return nil, err
	}
	var buf []byte
	err = r.withArea(name, func(a *area) {
		buf = make([]byte, min(uintptr(size(fd2)), a.size))
		a.read(buf)
	})
	return buf, err
}

// withArea maps the area if necessary and calls fn while holding the read lock,
// so that Close cannot unmap it concurrently.
func (r *Reader) withArea(name string, fn func(*area)) error {
	if _, err := r.area(name); err != nil {
		return err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	a := r.areas[name]
	if a == nil {
		return ErrNotOpen // closed in the meantime
	}
	fn(a)
	return nil
}

// area returns the mapped area with the given name, mapping it on first access.
func (r *Reader) area(name string) (*area, error) {
	r.mu.RLock()
	a := r.areas[name]
	r.mu.RUnlock()
	if a != nil {
		return a, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if a := r.areas[name]; a != nil {
		return a, nil
	}
	a, err := openArea(name)
	if err != nil {
		return nil, err
	}
	if r.areas == nil {
		r.areas = make(map[string]*area)
	}
	r.areas[name] = a
	return a, nil
}
