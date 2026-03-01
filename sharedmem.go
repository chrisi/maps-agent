package main

/*
   #cgo CFLAGS: -I include
   #include "FlightData_c.h"
*/
import "C"

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

type OwnShip struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

const (
	mappingName   = "FalconSharedMemoryArea"
	FILE_MAP_READ = 0x0004
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	procOpenFileMappingW = kernel32.NewProc("OpenFileMappingW")
	procMapViewOfFile    = kernel32.NewProc("MapViewOfFile")
	procUnmapViewOfFile  = kernel32.NewProc("UnmapViewOfFile")
)

type SharedMemReader struct {
	hMap    windows.Handle
	fd1Addr uintptr
	fd1     *C.FlightData
}

func (r *SharedMemReader) openFileMappingRead(name string) (windows.Handle, error) {
	p, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return 0, err
	}

	// HANDLE OpenFileMappingW(DWORD dwDesiredAccess, BOOL bInheritHandle, LPCWSTR lpName);
	r0, _, e1 := procOpenFileMappingW.Call(
		uintptr(FILE_MAP_READ),
		uintptr(0),
		uintptr(unsafe.Pointer(p)),
	)
	if r0 == 0 {
		if e1 != windows.ERROR_SUCCESS && e1 != nil {
			return 0, e1
		}
		return 0, windows.GetLastError()
	}
	return windows.Handle(r0), nil
}

func (r *SharedMemReader) mapViewRead(h windows.Handle) (uintptr, error) {
	// LPVOID MapViewOfFile(HANDLE hFileMappingObject, DWORD dwDesiredAccess,
	//                      DWORD dwFileOffsetHigh, DWORD dwFileOffsetLow, SIZE_T dwNumberOfBytesToMap);
	addr, _, e1 := procMapViewOfFile.Call(
		uintptr(h),
		uintptr(FILE_MAP_READ),
		0, 0,
		0, // map entire section
	)
	if addr == 0 {
		if e1 != windows.ERROR_SUCCESS && e1 != nil {
			return 0, e1
		}
		return 0, windows.GetLastError()
	}
	return addr, nil
}

func (r *SharedMemReader) unmapView(addr uintptr) error {
	r0, _, e1 := procUnmapViewOfFile.Call(addr)
	if r0 == 0 {
		if e1 != windows.ERROR_SUCCESS && e1 != nil {
			return e1
		}
		return windows.GetLastError()
	}
	return nil
}

func (r *SharedMemReader) open() error {
	hnd, err := r.openFileMappingRead(mappingName)
	if err != nil {
		return err
	}
	r.hMap = hnd
	r.fd1Addr, err = r.mapViewRead(r.hMap)
	if err != nil {
		return err
	}
	r.fd1 = (*C.FlightData)(unsafe.Pointer(r.fd1Addr))
	return nil
}

func (r *SharedMemReader) close() {
	_ = r.unmapView(r.fd1Addr)
	_ = windows.CloseHandle(r.hMap)
}

func (r *SharedMemReader) getOwnShip() *OwnShip {
	x := float32(r.fd1.y) // North (ft)
	y := float32(r.fd1.x) // East (ft)
	z := float32(r.fd1.z) // Down (ft)
	return &OwnShip{x, y, z}
}

func (r *SharedMemReader) getVersion() int {
	v := int(r.fd1.VersionNum)
	return v
}

func (r *SharedMemReader) isFlying() bool {
	return r.fd1.hsiBits|0x80000000 > 0
}
