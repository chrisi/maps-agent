package main

import (
	"maps-agent/bms"
	"maps-agent/util"
	"os"
)

type OwnShip struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

// SharedMemReader adapts the Falcon BMS shared memory to what the agent needs.
// The zero value is usable.
type SharedMemReader struct {
	logger *util.Logger
	reader bms.Reader
}

func NewSharedMemReader() *SharedMemReader {
	return &SharedMemReader{
		logger: util.NewLogger("SharedMemReader", os.Stdout, util.Info, true),
	}
}

func (r *SharedMemReader) Open() error {
	if err := r.reader.Open(); err != nil {
		r.logger.Errorf("failed to open Falcon BMS shared memory: %v", err)
		return err
	}
	r.logger.Infof("successfully opened Falcon BMS shared memory, version %d", r.GetVersion())
	return nil
}

func (r *SharedMemReader) IsReady() bool {
	return r.reader.IsOpen()
}

func (r *SharedMemReader) Close() {
	r.reader.Close()
}

// FlightData returns a full snapshot of the main shared memory area, or nil if not available.
func (r *SharedMemReader) FlightData() *bms.FlightData {
	fd, err := r.reader.FlightData()
	if err != nil {
		return nil
	}
	return fd
}

// GetOwnShip returns the ownship position with x = East and y = North (swapped
// compared to the sim's x = North, y = East), z = Down, all in feet.
func (r *SharedMemReader) GetOwnShip() *OwnShip {
	fd := r.FlightData()
	if fd == nil {
		return nil
	}
	return &OwnShip{X: fd.Y, Y: fd.X, Z: fd.Z}
}

func (r *SharedMemReader) GetVersion() int {
	fd := r.FlightData()
	if fd == nil {
		return -1
	}
	return int(fd.VersionNum)
}

func (r *SharedMemReader) GetFuelState() float32 {
	fd := r.FlightData()
	if fd == nil {
		return -1
	}
	return fd.Total
}

func (r *SharedMemReader) GetFuelFlow() float32 {
	fd := r.FlightData()
	if fd == nil {
		return -1
	}
	return fd.FuelFlow
}

func (r *SharedMemReader) IsFlying() bool {
	fd := r.FlightData()
	return fd != nil && fd.IsFlying()
}
