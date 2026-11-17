//go:build windows

package bms

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OpenFileMappingW is not wrapped by x/sys/windows.
var procOpenFileMappingW = windows.NewLazySystemDLL("kernel32.dll").NewProc("OpenFileMappingW")

// area is a read-only view of a named shared memory area.
type area struct {
	handle windows.Handle
	addr   uintptr
	view   []byte // the mapped region (page aligned size)
	size   uintptr
}

func openArea(name string) (*area, error) {
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	// HANDLE OpenFileMappingW(DWORD dwDesiredAccess, BOOL bInheritHandle, LPCWSTR lpName)
	h, _, callErr := procOpenFileMappingW.Call(windows.FILE_MAP_READ, 0, uintptr(unsafe.Pointer(namePtr)))
	if h == 0 {
		return nil, fmt.Errorf("open %s: %w", name, callErr)
	}
	handle := windows.Handle(h)

	addr, err := windows.MapViewOfFile(handle, windows.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("map %s: %w", name, err)
	}

	var info windows.MemoryBasicInformation
	if err := windows.VirtualQuery(addr, &info, unsafe.Sizeof(info)); err != nil {
		_ = windows.UnmapViewOfFile(addr)
		_ = windows.CloseHandle(handle)
		return nil, fmt.Errorf("query %s: %w", name, err)
	}

	// The view is memory owned by Windows, not by the Go heap, so converting the
	// address is safe. Going through a pointer avoids a false positive in go vet.
	base := *(*unsafe.Pointer)(unsafe.Pointer(&addr))
	view := unsafe.Slice((*byte)(base), info.RegionSize)

	return &area{handle: handle, addr: addr, view: view, size: info.RegionSize}, nil
}

// read copies the start of the area into dst, at most the mapped size.
func (a *area) read(dst []byte) {
	copy(dst, a.view)
}

func (a *area) close() {
	_ = windows.UnmapViewOfFile(a.addr)
	_ = windows.CloseHandle(a.handle)
}
