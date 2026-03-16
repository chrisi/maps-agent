package camtac

import (
	"maps-agent/util"
	"os"
)

type ObjectiveDeltaReader struct {
	numDeltas int
	log       *util.Logger
}

func NewObjectiveDeltaReader() *ObjectiveDeltaReader {
	return &ObjectiveDeltaReader{
		log: util.NewLogger("OBD-Reader", os.Stdout, util.Info, true),
	}
}

func (or *ObjectiveDeltaReader) ReadObdFile(data []byte) []ObjectiveDeltas {
	or.log.Infof("Reading objective deltas")

	hdrCur := NewCursor(data)
	_ = hdrCur.Int32() // newSize
	or.numDeltas = int(hdrCur.Int16())
	expSize := int(hdrCur.Int32())

	expanded, err := Expand(data[10:], expSize)
	if err != nil {
		or.log.Errorf("Error expanding data: %v", err)
		return nil
	}

	or.log.Infof("Compressed size: %d", len(data))
	or.log.Infof("Uncompressed size: %d", len(expanded))
	or.log.Infof("Deltas: %d", or.numDeltas)

	return or.readAllObjectiveDeltas(expanded)
}

func (or *ObjectiveDeltaReader) readAllObjectiveDeltas(data []byte) []ObjectiveDeltas {
	c := NewCursor(data)
	var deltas []ObjectiveDeltas
	for range or.numDeltas {
		d := readObjectiveDeltas(c)
		deltas = append(deltas, d)
	}
	return deltas
}

func readObjectiveDeltas(c *Cursor) ObjectiveDeltas {
	d := ObjectiveDeltas{
		ID:          readVU_ID(c),
		LastRepair:  c.Uint32(),
		Owner:       c.Uint8(),
		Supply:      c.Uint8(),
		Fuel:        c.Uint8(),
		Losses:      c.Uint8(),
		NumStatuses: c.Uint8(),
	}
	d.Statuses = make([]uint8, d.NumStatuses)
	for i := range d.NumStatuses {
		d.Statuses[i] = c.Uint8()
	}
	return d
}
