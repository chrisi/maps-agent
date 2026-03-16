package camtac

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
