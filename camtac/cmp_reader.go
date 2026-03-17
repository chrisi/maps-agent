package camtac

import (
	"maps-agent/util"
	"os"
)

type CampaignReader struct {
	log *util.Logger
}

func NewCampaignReader() *CampaignReader {
	return &CampaignReader{
		log: util.NewLogger("CMP-Reader", os.Stdout, util.Info, true),
	}
}

func (cr *CampaignReader) ReadCmpFile(data []byte) (*Campaign, error) {
	cr.log.Infof("Reading campaign file")

	hdrCur := NewCursor(data)
	_ = hdrCur.Int32() // newSize
	expSize := int(hdrCur.Int32())

	expanded, err := Expand(data[8:], expSize)
	if err != nil {
		cr.log.Errorf("Error expanding data: %v", err)
		return nil, err
	}

	cr.log.Infof("Compressed size: %d", len(data))
	cr.log.Infof("Uncompressed size: %d", len(expanded))

	c := NewCursor(expanded)
	return readCampaign(c), nil
}

func readCampaign(c *Cursor) *Campaign {
	cmp := &Campaign{
		CurrentTime:   c.Uint32(),
		TeStartTime:   c.Uint32(),
		TeTimeLimit:   c.Uint32(),
		VictoryPoints: c.Uint32(),
		TeType:        c.Uint32(),
		TeNumberTeams: c.Uint32(),
	}

	for i := range 8 {
		cmp.TeNumberAircraft[i] = c.Uint32()
	}
	for i := range 8 {
		cmp.TeNumberF16[i] = c.Uint32()
	}
	cmp.TeTeam = c.Uint32()
	for i := range 8 {
		cmp.TeTeamPts[i] = c.Uint32()
	}
	cmp.TeFlags = c.Uint32()

	for i := range 8 {
		cmp.Info[i] = readTeamBasicInfo(c)
	}

	cmp.LastMajorEvent = c.Uint32()
	cmp.LastResupply = c.Uint32()
	cmp.LastRepair = c.Uint32()
	cmp.LastReinforcement = c.Uint32()
	cmp.Timestamp = c.Uint16()
	cmp.Group = c.Uint16()
	cmp.GroundRatio = c.Uint16()
	cmp.AirRatio = c.Uint16()
	cmp.AirDefenseRatio = c.Uint16()
	cmp.NavalRatio = c.Uint16()
	cmp.Brief = c.Uint16()
	cmp.TheaterSizeX = c.Uint16()
	cmp.TheaterSizeY = c.Uint16()
	cmp.CurrentDay = c.Uint8()
	cmp.ActiveTeams = c.Uint8()
	cmp.DayZero = c.Uint8()
	cmp.EndgameResult = c.Uint8()
	cmp.Situation = c.Uint8()
	cmp.EnemyAirExp = c.Uint8()
	cmp.EnemyADExp = c.Uint8()
	cmp.BullseyeName = c.Uint8()
	cmp.BullseyeX = c.Uint16()
	cmp.BullseyeY = c.Uint16()

	cmp.TheaterName = readFixedString(c, 40)
	cmp.Scenario = readFixedString(c, 40)
	cmp.SaveFile = readFixedString(c, 40)
	cmp.UiName = readFixedString(c, 40)

	cmp.SquadronID = readVU_ID(c)

	cmp.NumRecentEvents = c.Uint16()
	cmp.RecentEvents = make([]Event, cmp.NumRecentEvents)
	for i := range cmp.NumRecentEvents {
		cmp.RecentEvents[i] = readEvent(c)
	}

	cmp.NumPriorityEvents = c.Uint16()
	cmp.PriorityEvents = make([]Event, cmp.NumPriorityEvents)
	for i := range cmp.NumPriorityEvents {
		cmp.PriorityEvents[i] = readEvent(c)
	}

	cmp.CampMapSize = c.Uint16()
	cmp.CampMap = make([]uint8, cmp.CampMapSize)
	for i := range cmp.CampMapSize {
		cmp.CampMap[i] = c.Uint8()
	}

	cmp.LastIndexNo = c.Uint16()
	cmp.NumAvailableSquadrons = c.Uint16()
	cmp.Squadrons = make([]SquadInfo, cmp.NumAvailableSquadrons)
	for i := range cmp.NumAvailableSquadrons {
		cmp.Squadrons[i] = readSquadInfo(c)
	}

	cmp.Tempo = c.Uint8()
	cmp.CreatorIP = c.Uint32()
	cmp.CreationTime = c.Uint32()
	cmp.CreationRand = c.Uint32()
	cmp.CampPeriodStart = c.Uint16()
	cmp.CampPeriodEnd = c.Uint16()

	return cmp
}

func readTeamBasicInfo(c *Cursor) TeamBasicInfo {
	return TeamBasicInfo{
		TeamFlag:       c.Uint8(),
		TeamColor:      c.Uint8(),
		TeamName:       readFixedString(c, 20),
		TeamMottoBytes: readFixedString(c, 200),
	}
}

func readSquadInfo(c *Cursor) SquadInfo {
	return SquadInfo{
		X:                c.Float32(),
		Y:                c.Float32(),
		ID:               readVU_ID(c),
		DescriptionIndex: c.Uint16(),
		NameID:           c.Uint16(),
		AirbaseIcon:      c.Uint16(),
		SquadronPatch:    c.Uint16(),
		Speciality:       c.Uint8(),
		CurrentStrength:  c.Uint8(),
		Country:          c.Uint8(),
		AirbaseName:      readFixedString(c, 80),
		Flags:            c.Uint32(),
		CampID:           c.Uint16(),
		TexSet:           c.Uint16(),
		Padd:             c.Uint8(),
		SquadronName:     readFixedString(c, 80),
	}
}

func readEvent(c *Cursor) Event {
	e := Event{
		X:     c.Uint16(),
		Y:     c.Uint16(),
		Time:  c.Uint32(),
		Flags: c.Uint8(),
		Team:  c.Uint8(),
	}
	for i := range 10 {
		e.Padd[i] = c.Uint8()
	}
	e.TextLen = c.Uint16()
	e.Text = readFixedString(c, int(e.TextLen))
	return e
}

func readFixedString(c *Cursor, length int) string {
	if length == 0 {
		return ""
	}
	raw := c.Bytes(length)
	n := 0
	for n < len(raw) && raw[n] != 0 {
		n++
	}
	return string(raw[:n])
}
