package camtac

import (
	"maps-agent/util"
	"os"
)

type ObjectiveReader struct {
	numObjectives int
	c             *Cursor
	log           *util.Logger
}

func NewObjectiveReader() *ObjectiveReader {
	return &ObjectiveReader{
		log: util.NewLogger("OBJ-Reader", os.Stdout, util.Info, true),
	}
}

func (or *ObjectiveReader) ReadObjFile(data []byte) []Objective {
	or.log.Infof("Reading objectives")

	hdrCur := NewCursor(data)
	_ = hdrCur.Int32() // compSize
	or.numObjectives = int(hdrCur.Int16())
	expSize := int(hdrCur.Int32())

	expanded, err := Expand(data[10:], expSize)
	if err != nil {
		or.log.Errorf("Error expanding data: %v", err)
		return nil
	}

	or.log.Debugf("Compressed size: %d", len(data))
	or.log.Debugf("Uncompressed size: %d", len(expanded))
	or.log.Infof("Objectives: %d", or.numObjectives)

	return or.readObjectives(expanded)
}

func (or *ObjectiveReader) readObjectives(data []byte) []Objective {
	or.c = NewCursor(data)
	var objectives []Objective
	for i := 0; i < or.numObjectives; i++ {
		obj := or.createObjective()
		objectives = append(objectives, obj)
	}
	return objectives
}

func (or *ObjectiveReader) createObjective() Objective {
	_ = int(or.c.Uint16()) // objectiveType (analog zu unitType)
	return readObjective(or.c)
}

func readCampObjectiveLinkDataType(c *Cursor) CampObjectiveLinkDataType {
	link := CampObjectiveLinkDataType{}
	for i := range 8 {
		link.Costs[i] = c.Uint8()
	}
	link.ID = readVU_ID(c)
	return link
}

func readObjective(c *Cursor) Objective {
	o := Objective{
		CampaignBase: readCampaignBase(c),
		LastRepair:   c.Uint32(),
		ObjFlags:     c.Uint32(),
		Supply:       c.Uint8(),
		Fuel:         c.Uint8(),
		Losses:       c.Uint8(),
		NumStatuses:  c.Uint8(),
	}
	o.Statuses = make([]uint8, o.NumStatuses)
	for i := range o.NumStatuses {
		o.Statuses[i] = c.Uint8()
	}
	o.Priority = c.Uint8()
	o.NameID = c.Uint16()
	o.ParentID = readVU_ID(c)
	o.FirstOwner = c.Uint8()
	o.NumLinks = c.Uint8()
	o.Links = make([]CampObjectiveLinkDataType, o.NumLinks)
	for i := range o.NumLinks {
		o.Links[i] = readCampObjectiveLinkDataType(c)
	}
	o.HasRadarData = c.Uint8()
	o.PosX = c.Float64()
	o.PosY = c.Float64()
	o.PosZ = c.Float64()
	o.Heading = c.Float32()
	for i := range 80 {
		o.CampName[i] = c.Uint8()
	}
	return o
}
