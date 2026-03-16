package camtac

import (
	"maps-agent/util"
	"os"
)

type Counts struct {
	NumUnits      int
	NumFlights    int
	NumBattalions int
	NumBrigades   int
	NumPackages   int
	NumSquadrons  int
	NumTaskForces int
}

type UnitReader struct {
	classTable []*CT
	numUnits   int
	counters   Counts
	c          *Cursor
	log        *util.Logger
}

func NewUnitReader(classTable []*CT) *UnitReader {
	return &UnitReader{
		classTable: classTable,
		log:        util.NewLogger("UNI-Reader", os.Stdout, util.Info, true),
	}
}

func (ur *UnitReader) ReadUniFile(data []byte) []any {
	ur.log.Infof("Reading units")

	ur.counters = Counts{
		NumUnits:      0,
		NumFlights:    0,
		NumBattalions: 0,
		NumBrigades:   0,
		NumPackages:   0,
		NumSquadrons:  0,
		NumTaskForces: 0,
	}

	hdrCur := NewCursor(data)
	_ = hdrCur.Int32() // compSize
	ur.numUnits = int(hdrCur.Int16())
	expSize := int(hdrCur.Int32())

	expanded, err := Expand(data[10:], expSize)
	if err != nil {
		ur.log.Errorf("Error expanding data: %v", err)
		return nil
	}

	ur.log.Debugf("Compressed size: %d", len(data))
	ur.log.Debugf("Uncompressed size: %d", len(expanded))
	ur.log.Infof("Units: %d", ur.numUnits)

	return ur.readUnits(expanded)
}

func (ur *UnitReader) Counts() Counts {
	return ur.counters
}

func (ur *UnitReader) readUnits(data []byte) []any {
	ur.c = NewCursor(data)
	var units []any
	for i := 0; i < ur.numUnits; i++ {
		unit := ur.createUnit()
		units = append(units, unit)
	}
	return units
}

func (ur *UnitReader) createUnit() any {
	unitType := int(ur.c.Uint16())
	if unitType <= len(ur.classTable)+100 {
		if unitType >= 100 {
			ur.log.Debugf("UnitType: %d", unitType)
			start := ur.c.pos
			unit := ur.createUnitByClassType(ur.classTable[unitType-100])
			size := ur.c.pos - start
			numWp := -1
			hu, ok := unit.(HasUnit)
			if ok {
				numWp = len(hu.GetUnit().Waypoints)
			}
			if numWp > 20 || size > 2000 {
				ur.log.Warnf("Unit size: %d, waypoints: %d", size, numWp)
			}
			return unit
		} else if unitType > 3707 { // TODO: ergibt wenig Sinn, ist aber im Original so
			ur.log.Warnf("UnitType strange:", unitType)
			return ur.createUnitByClassType(ur.classTable[3607])
		} else {
			ur.log.Warnf("UnitType 430")
			return ur.createUnitByClassType(ur.classTable[430])
		}
	}
	return nil
}

func (ur *UnitReader) createUnitByClassType(classType *CT) any {
	switch classType.Domain {
	case DomainAir:
		switch classType.Type {
		case TypeFlight:
			ur.log.Debugf("Reading flight")
			ur.counters.NumUnits++
			ur.counters.NumFlights++
			return readFlight(ur.c)
		case TypePackage:
			ur.log.Debugf("Reading package")
			ur.counters.NumUnits++
			ur.counters.NumPackages++
			return readPackage(ur.c)
		case TypeSquadron:
			ur.log.Debugf("Reading squadron")
			ur.counters.NumUnits++
			ur.counters.NumSquadrons++
			return readSquadron(ur.c)
		}

	case DomainLand:
		switch classType.Type {
		case TypeBattalion:
			ur.log.Debugf("Reading battalion")
			ur.counters.NumUnits++
			ur.counters.NumBattalions++
			return readBattalion(ur.c)
		case TypeBrigade:
			ur.log.Debugf("Reading brigade")
			ur.counters.NumUnits++
			ur.counters.NumBrigades++
			return readBrigade(ur.c)
		}

	case DomainSea:
		switch classType.Type {
		case TypeTaskForce:
			ur.log.Debugf("Reading task-force")
			ur.counters.NumUnits++
			ur.counters.NumTaskForces++
			return readTaskForce(ur.c)
		}
	}
	ur.log.Warnf("Unknown unit type: %d", classType.Type)
	return nil
}

func readTarget(c *Cursor) Target {
	return Target{
		ID:       readVU_ID(c),
		Building: c.Uint8(),
	}
}

func readWaypoint(c *Cursor) Waypoint {
	wp := Waypoint{
		Haves:       c.Uint8(),
		GridX:       c.Uint16(),
		GridY:       c.Uint16(),
		GridZ:       c.Uint16(),
		Arrive:      c.Uint32(),
		Action:      c.Uint8(),
		RouteAction: c.Uint8(),
		Formation:   c.Uint8(),
		Flags:       c.Uint32(),
	}
	if (wp.Haves & 2) != 0 {
		td := &WaypointTargetData{
			Target: readTarget(c),
		}
		for i := range 4 {
			td.DesignatedTarget[i] = readTarget(c)
		}
		wp.TargetData = td
	}
	if (wp.Haves & 1) != 0 {
		depart := c.Uint32()
		wp.Depart = &depart
	}
	return wp
}

func readUnit(c *Cursor) Unit {
	u := Unit{
		CampaignBase:  readCampaignBase(c),
		LastCheck:     c.Uint32(),
		Roster:        c.Int32(),
		UnitFlags:     c.Int32(),
		DestX:         c.Int16(),
		DestY:         c.Int16(),
		TargetID:      readVU_ID(c),
		CargoID:       readVU_ID(c),
		Moved:         c.Uint8(),
		Losses:        c.Uint8(),
		Tactic:        c.Uint8(),
		CurrentWP:     c.Uint16(),
		NameID:        c.Uint16(),
		Reinforcement: c.Uint16(),
		NumWaypoints:  c.Uint16(),
	}
	u.Waypoints = make([]Waypoint, u.NumWaypoints)
	for i := range u.NumWaypoints {
		u.Waypoints[i] = readWaypoint(c)
	}
	return u
}

func readPilot(c *Cursor) Pilot {
	return Pilot{
		PilotID:       c.Uint16(),
		Skill:         c.Uint8(),
		Status:        c.Uint8(),
		AAKills:       c.Uint8(),
		AGKills:       c.Uint8(),
		ASKills:       c.Uint8(),
		ANKills:       c.Uint8(),
		MissionsFlown: c.Uint16(),
	}
}

func readSquadron(c *Cursor) Squadron {
	s := Squadron{
		Unit:       readUnit(c),
		Fuel:       c.Uint32(),
		Speciality: c.Uint8(),
	}
	for i := range 16 {
		s.CampSpecRoleRating[i] = c.Uint8()
	}
	for i := range 1000 {
		s.Stores[i] = c.Uint8()
	}
	for i := range 48 {
		s.Pilots[i] = readPilot(c)
	}
	for i := range 16 {
		s.Schedule[i] = c.Uint32()
	}
	s.AirbaseID = readVU_ID(c)
	s.Hotspot = readVU_ID(c)
	for i := range 16 {
		s.Rating[i] = c.Uint8()
	}
	s.AAKills = c.Uint16()
	s.AGKills = c.Uint16()
	s.ASKills = c.Uint16()
	s.ANKills = c.Uint16()
	s.MissionsFlown = c.Uint16()
	s.MissionScore = c.Uint16()
	s.TotalLosses = c.Uint8()
	s.PilotLosses = c.Uint8()
	s.SquadronPatch = c.Uint16()
	s.SquadronRetaskAt = c.Uint32()
	s.Relocate = c.Uint8()
	s.TexSet = c.Uint32()
	return s
}

func readLoadout(c *Cursor) Loadout {
	l := Loadout{}
	for i := range 16 {
		l.WeaponID[i] = c.Uint16()
	}
	for i := range 16 {
		l.WeaponCount[i] = c.Uint8()
	}
	return l
}

func readFlight(c *Cursor) Flight {
	f := Flight{
		Unit:            readUnit(c),
		Z:               c.Float32(),
		FuelBurnt:       c.Uint32(),
		FuelInitial:     [4]uint32{c.Uint32(), c.Uint32(), c.Uint32(), c.Uint32()},
		LaserCode:       [4]uint16{c.Uint16(), c.Uint16(), c.Uint16(), c.Uint16()},
		LastMove:        c.Uint32(),
		LastCombat:      c.Uint32(),
		TimeOnTarget:    c.Uint32(),
		MissionOverTime: c.Uint32(),
		MissionTarget:   c.Uint16(),
		Loadouts:        c.Uint8(),
	}
	f.Loadout = make([]Loadout, f.Loadouts)
	for i := range f.Loadouts {
		f.Loadout[i] = readLoadout(c)
	}
	f.Mission = c.Uint8()
	f.OldMission = c.Uint8()
	f.LastDirection = c.Uint8()
	f.Priority = c.Uint8()
	f.MissionID = c.Uint8()
	f.EvalFlags = c.Uint8()
	f.MissionContext = c.Uint8()
	f.Package = readVU_ID(c)
	f.Squadron = readVU_ID(c)
	f.Requester = readVU_ID(c)
	for i := range 4 {
		f.Slots[i] = c.Uint8()
	}
	for i := range 4 {
		f.Pilots[i] = c.Uint8()
	}
	for i := range 4 {
		f.PlaneStats[i] = c.Uint8()
	}
	for i := range 4 {
		f.PlayerSlots[i] = c.Uint8()
	}
	f.LastPlayerSlot = c.Uint8()
	f.CallsignID = c.Uint8()
	f.CallsignNum = c.Uint8()
	f.RefuelQuantity = c.Uint32()
	for i := range 4 {
		f.TexSet[i] = c.Uint32()
	}
	for i := range 4 {
		f.TacanChannel[i] = c.Uint8()
	}
	for i := range 4 {
		f.TacanBand[i] = c.Uint8()
	}
	for i := range 4 {
		f.LoadedCft[i] = c.Uint8()
	}
	return f
}

func readGroundUnit(c *Cursor) GroundUnit {
	return GroundUnit{
		Unit:     readUnit(c),
		Orders:   c.Uint8(),
		Division: c.Uint16(),
		AObj:     readVU_ID(c),
	}
}

func readTaskForce(c *Cursor) TaskForce {
	return TaskForce{
		Unit:   readUnit(c),
		Orders: c.Uint8(),
		Supply: c.Uint8(),
	}
}

func readBattalion(c *Cursor) Battalion {
	return Battalion{
		GroundUnit:   readGroundUnit(c),
		LastMove:     c.Uint32(),
		LastCombat:   c.Uint32(),
		ParentID:     readVU_ID(c),
		LastObj:      readVU_ID(c),
		Supply:       c.Uint8(),
		Fatigue:      c.Uint8(),
		Morale:       c.Uint8(),
		Heading:      c.Uint8(),
		FinalHeading: c.Uint8(),
		Position:     c.Uint8(),
	}
}

func readBrigade(c *Cursor) Brigade {
	b := Brigade{
		GroundUnit:  readGroundUnit(c),
		NumElements: c.Uint8(),
	}
	b.Elements = make([]VU_ID, b.NumElements)
	for i := range b.NumElements {
		b.Elements[i] = readVU_ID(c)
	}
	return b
}

func readPackage(c *Cursor) Package {
	p := Package{
		Unit:        readUnit(c),
		NumElements: c.Uint8(),
	}
	p.Elements = make([]VU_ID, p.NumElements)
	for i := range p.NumElements {
		p.Elements[i] = readVU_ID(c)
	}
	p.Interceptor = readVU_ID(c)
	p.Awacs = readVU_ID(c)
	p.JStar = readVU_ID(c)
	p.ECM = readVU_ID(c)
	p.Tanker = readVU_ID(c)
	p.WaitCycles = c.Uint8()
	if ((p.Unit.UnitFlags & 0x100000) != 0) && p.WaitCycles == 0 {
		p.MissionRequest = &MissionRequest{
			Requests:     c.Uint16(),
			Responses:    c.Uint16(),
			MRMission:    c.Uint8(),
			MRAircraft:   c.Uint8(),
			MRContext:    c.Uint8(),
			MRRoeCheck:   c.Uint8(),
			Requester:    readVU_ID(c),
			Target:       readVU_ID(c),
			MRTot:        c.Uint32(),
			MRActionType: c.Uint8(),
			MRPriority:   c.Uint16(),
		}
	}
	return p
}
