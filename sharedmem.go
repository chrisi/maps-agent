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
	hMap windows.Handle
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
	return nil
}

func (r *SharedMemReader) close() {
	windows.CloseHandle(r.hMap)
}

func (r *SharedMemReader) read() (*OwnShip, error) {
	addr, err := r.mapViewRead(r.hMap)
	if err != nil {
		return nil, err
	}
	defer func() { _ = r.unmapView(addr) }()
	fd := (*C.FlightData)(unsafe.Pointer(addr))
	x := float32(fd.x) // North (ft)
	y := float32(fd.y) // East  (ft)
	z := float32(fd.z) // Down  (ft)
	p := OwnShip{x, y, z}
	return &p, nil
}
