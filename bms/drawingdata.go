package bms

// DrawingDataVersion is the DrawingData version these definitions were ported from (DRAWINGDATA_VERSION).
const DrawingDataVersion = 1

// DrawingData is the parsed content of the "FalconSharedMemoryAreaDrawing" area.
//
// The command strings are only populated if the matching g_bExportDrawingCommandsFor*
// option is enabled in the Falcon BMS config. See FlightData.h for the command format.
type DrawingData struct {
	VersionNum  uint32
	HUDCommands string
	RWRCommands string
	HMSCommands string
}

// ParseDrawingData parses the raw "FalconSharedMemoryAreaDrawing" area.
//
// Layout: VersionNum (uint32), then for HUD, RWR and HMS each a length (uint32)
// followed by length+1 bytes (NUL terminated).
func ParseDrawingData(buf []byte) (*DrawingData, error) {
	p := parser{buf: buf}
	dd := &DrawingData{VersionNum: p.u32()}
	dd.HUDCommands = p.str(p.u32())
	dd.RWRCommands = p.str(p.u32())
	dd.HMSCommands = p.str(p.u32())
	if p.err != nil {
		return nil, p.err
	}
	return dd, nil
}
