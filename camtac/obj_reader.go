package camtac

import (
	"maps-agent/util"
	"os"
)

type ObjectiveReader struct {
	classTable    []*SCT
	numObjectives int
	log           *util.Logger
}

func NewObjectiveReader(classTable []*SCT) *ObjectiveReader {
	return &ObjectiveReader{
		classTable: classTable,
		log:        util.NewLogger("OBJ-Reader", os.Stdout, util.Info, true),
	}
}

func (or *ObjectiveReader) ReadObjFile(data []byte) []*Objective {
	or.log.Infof("Reading objectives")

	hdrCur := NewCursor(data)
	or.numObjectives = int(hdrCur.Int16())
	expSize := int(hdrCur.Int32())
	_ = hdrCur.Int32() // newSize

	expanded, err := Expand(data[10:], expSize)
	if err != nil {
		or.log.Errorf("Error expanding data: %v", err)
		return nil
	}

	or.log.Debugf("Compressed size: %d", len(data))
	or.log.Debugf("Uncompressed size: %d", len(expanded))
	or.log.Infof("Objectives: %d", or.numObjectives)

	return or.readAllObjectives(expanded)
}

func (or *ObjectiveReader) readAllObjectives(data []byte) []*Objective {
	c := NewCursor(data)
	var objectives []*Objective
	for i := 0; i < or.numObjectives; i++ {
		_ = int(c.Uint16()) // entityType, also in squadron.entityType
		obj := readObjective(c)
		obj.ClassType = or.classTable[obj.CampaignBase.EntityType-100]
		objectives = append(objectives, &obj)
	}
	return objectives
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
	}
	ns := c.Uint8()
	o.Statuses = make([]int, ns)
	for i := range ns {
		o.Statuses[i] = c.Int8()
	}
	o.Priority = c.Uint8()
	o.NameID = c.Uint16()
	o.ParentID = readVU_ID(c)
	o.FirstOwner = c.Uint8()
	nl := c.Uint8()
	o.Links = make([]CampObjectiveLinkDataType, nl)
	for i := range nl {
		o.Links[i] = readCampObjectiveLinkDataType(c)
	}
	if c.Uint8() > 0 {
		o.DetectRatios = make([]float32, 8)
		for i := range 8 {
			o.DetectRatios[i] = c.Float32()
		}
	} else {
		o.DetectRatios = make([]float32, 0)
	}
	o.PosX = c.Float64()
	o.PosY = c.Float64()
	o.PosZ = c.Float64()
	o.Heading = c.Float32()

	campNameRaw := c.Bytes(80)
	n := 0
	for n < len(campNameRaw) && campNameRaw[n] != 0 {
		n++
	}
	o.CampName = string(campNameRaw[:n])

	return o
}
