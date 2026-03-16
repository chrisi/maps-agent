package camtac

type HasUnit interface {
	GetUnit() Unit
}

func (s Squadron) GetUnit() Unit {
	return s.Unit
}

func (f Flight) GetUnit() Unit {
	return f.Unit
}

func (b Battalion) GetUnit() Unit {
	return b.GroundUnit.Unit
}

func (b Brigade) GetUnit() Unit {
	return b.GroundUnit.Unit
}

func (p Package) GetUnit() Unit {
	return p.Unit
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
