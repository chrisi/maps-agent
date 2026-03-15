package camtac

func ReadUnits(data []byte) []interface{} {
	c := NewCursor(data)
	var units []interface{}

	// Read Data root
	// u16 typeSquad;
	// Squadron squad;
	c.Uint16() // typeSquad
	units = append(units, readSquadron(c))

	// u16 typeFlight;
	// Flight flight;
	c.Uint16() // typeFlight
	units = append(units, readFlight(c))

	// u16 typePackage;
	// Package pack;
	c.Uint16() // typePackage
	units = append(units, readPackage(c))

	return units
}

func readVU_ID(c *Cursor) VU_ID {
	return VU_ID{
		Num:     c.Uint32(),
		Creator: c.Uint32(),
	}
}

func readCampaignBase(c *Cursor) CampaignBase {
	return CampaignBase{
		ID:         readVU_ID(c),
		EntityType: c.Uint16(),
		X:          c.Uint16(),
		Y:          c.Uint16(),
		Z:          c.Float32(),
		SpotTime:   c.Uint32(),
		Spotted:    c.Uint16(),
		BaseFlags:  c.Uint16(),
		Owner:      c.Uint8(),
		CampID:     c.Uint16(),
	}
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
	u.WP = make([]Waypoint, u.NumWaypoints)
	for i := range u.NumWaypoints {
		u.WP[i] = readWaypoint(c)
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
