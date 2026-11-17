//go:build !windows

package bms

import (
	"errors"
	"fmt"
)

// area is a stub; Falcon BMS shared memory only exists on Windows (or in Wine).
type area struct {
	size uintptr
}

func openArea(name string) (*area, error) {
	return nil, fmt.Errorf("open %s: %w", name, errors.ErrUnsupported)
}

func (a *area) read([]byte) {}

func (a *area) close() {}
