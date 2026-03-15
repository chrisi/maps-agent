package camtac

type EmbeddedFileInfo struct {
	FileName      string
	FileOffset    uint32
	FileSizeBytes uint32
}

type VU_ID struct {
	Num     uint32 `json:"num"`
	Creator uint32 `json:"creator"`
}

type CampaignBase struct {
	ID         VU_ID   `json:"id"`
	EntityType uint16  `json:"entityType"`
	X          uint16  `json:"x"`
	Y          uint16  `json:"y"`
	Z          float32 `json:"z"`
	SpotTime   uint32  `json:"spotTime"`
	Spotted    uint16  `json:"spotted"`
	BaseFlags  uint16  `json:"baseFlags"`
	Owner      uint8   `json:"owner"`
	CampID     uint16  `json:"campId"`
}

type Target struct {
	ID       VU_ID `json:"id"`
	Building uint8 `json:"building"`
}

type WaypointTargetData struct {
	Target           Target    `json:"target"`
	DesignatedTarget [4]Target `json:"designatedTarget"`
}

type Waypoint struct {
	Haves       uint8  `json:"haves"`
	GridX       uint16 `json:"gridX"`
	GridY       uint16 `json:"gridY"`
	GridZ       uint16 `json:"gridZ"`
	Arrive      uint32 `json:"arrive"`
	Action      uint8  `json:"action"`
	RouteAction uint8  `json:"routeAction"`
	Formation   uint8  `json:"formation"`
	Flags       uint32 `json:"flags"`

	// Present if (Haves & 2) != 0
	TargetData *WaypointTargetData `json:"targetData,omitzero"`

	// Present if (Haves & 1) != 0
	Depart *uint32 `json:"depart,omitzero"`
}

type Unit struct {
	CampaignBase  CampaignBase `json:"campaignBase"`
	LastCheck     uint32       `json:"lastCheck"`
	Roster        int32        `json:"roster"`
	UnitFlags     int32        `json:"unitFlags"`
	DestX         int16        `json:"destX"`
	DestY         int16        `json:"destY"`
	TargetID      VU_ID        `json:"targetId"`
	CargoID       VU_ID        `json:"cargoId"`
	Moved         uint8        `json:"moved"`
	Losses        uint8        `json:"losses"`
	Tactic        uint8        `json:"tactic"`
	CurrentWP     uint16       `json:"currentWp"`
	NameID        uint16       `json:"nameId"`
	Reinforcement uint16       `json:"reinforcement"`
	NumWaypoints  uint16       `json:"numWaypoints"`
	Waypoints     []Waypoint   `json:"Waypoints"`
}

type Pilot struct {
	PilotID       uint16 `json:"pilotId"`
	Skill         uint8  `json:"skill"`
	Status        uint8  `json:"status"`
	AAKills       uint8  `json:"aaKills"`
	AGKills       uint8  `json:"agKills"`
	ASKills       uint8  `json:"asKills"`
	ANKills       uint8  `json:"anKills"`
	MissionsFlown uint16 `json:"missionsFlown"`
}

type Squadron struct {
	Unit               Unit        `json:"unit"`
	Fuel               uint32      `json:"fuel"`
	Speciality         uint8       `json:"speciality"`
	CampSpecRoleRating [16]uint8   `json:"campSpecRoleRating"`
	Stores             [1000]uint8 `json:"stores"`
	Pilots             [48]Pilot   `json:"pilots"`
	Schedule           [16]uint32  `json:"schedule"`
	AirbaseID          VU_ID       `json:"airbaseId"`
	Hotspot            VU_ID       `json:"hotspot"`
	Rating             [16]uint8   `json:"rating"`
	AAKills            uint16      `json:"aaKills"`
	AGKills            uint16      `json:"agKills"`
	ASKills            uint16      `json:"asKills"`
	ANKills            uint16      `json:"anKills"`
	MissionsFlown      uint16      `json:"missionsFlown"`
	MissionScore       uint16      `json:"missionScore"`
	TotalLosses        uint8       `json:"totalLosses"`
	PilotLosses        uint8       `json:"pilotLosses"`
	SquadronPatch      uint16      `json:"squadronPatch"`
	SquadronRetaskAt   uint32      `json:"squadronRetaskAt"`
	Relocate           uint8       `json:"relocate"`
	TexSet             uint32      `json:"texSet"`
}

type Loadout struct {
	WeaponID    [16]uint16 `json:"weaponId"`
	WeaponCount [16]uint8  `json:"weaponCount"`
}

type Flight struct {
	Unit            Unit      `json:"unit"`
	Z               float32   `json:"z"`
	FuelBurnt       uint32    `json:"fuelBurnt"`
	FuelInitial     [4]uint32 `json:"fuelInitial"`
	LaserCode       [4]uint16 `json:"laserCode"`
	LastMove        uint32    `json:"lastMove"`
	LastCombat      uint32    `json:"lastCombat"`
	TimeOnTarget    uint32    `json:"timeOnTarget"`
	MissionOverTime uint32    `json:"missionOverTime"`
	MissionTarget   uint16    `json:"missionTarget"`
	Loadouts        uint8     `json:"loadouts"`
	Loadout         []Loadout `json:"loadout"`
	Mission         uint8     `json:"mission"`
	OldMission      uint8     `json:"oldMission"`
	LastDirection   uint8     `json:"lastDirection"`
	Priority        uint8     `json:"priority"`
	MissionID       uint8     `json:"missionId"`
	EvalFlags       uint8     `json:"evalFlags"`
	MissionContext  uint8     `json:"missionContext"`
	Package         VU_ID     `json:"package"`
	Squadron        VU_ID     `json:"squadron"`
	Requester       VU_ID     `json:"requester"`
	Slots           [4]uint8  `json:"slots"`
	Pilots          [4]uint8  `json:"pilots"`
	PlaneStats      [4]uint8  `json:"planeStats"`
	PlayerSlots     [4]uint8  `json:"playerSlots"`
	LastPlayerSlot  uint8     `json:"lastPlayerSlot"`
	CallsignID      uint8     `json:"callsignId"`
	CallsignNum     uint8     `json:"callsignNum"`
	RefuelQuantity  uint32    `json:"refuelQuantity"`
	TexSet          [4]uint32 `json:"texSet"`
	TacanChannel    [4]uint8  `json:"tacanChannel"`
	TacanBand       [4]uint8  `json:"tacanBand"`
	LoadedCft       [4]uint8  `json:"loadedCft"`
}

type MissionRequest struct {
	Requests     uint16 `json:"requests"`
	Responses    uint16 `json:"responses"`
	MRMission    uint8  `json:"mrMission"`
	MRAircraft   uint8  `json:"mrAircraft"`
	MRContext    uint8  `json:"mrContext"`
	MRRoeCheck   uint8  `json:"mrRoeCheck"`
	Requester    VU_ID  `json:"requester"`
	Target       VU_ID  `json:"target"`
	MRTot        uint32 `json:"mrTot"`
	MRActionType uint8  `json:"mrActionType"`
	MRPriority   uint16 `json:"mrPriority"`
}

type Package struct {
	Unit        Unit    `json:"unit"`
	NumElements uint8   `json:"numElements"`
	Elements    []VU_ID `json:"elements"`
	Interceptor VU_ID   `json:"interceptor"`
	Awacs       VU_ID   `json:"awacs"`
	JStar       VU_ID   `json:"jstar"`
	ECM         VU_ID   `json:"ecm"`
	Tanker      VU_ID   `json:"tanker"`
	WaitCycles  uint8   `json:"waitCycles"`

	// Present if ((Unit.UnitFlags & 0x100000) != 0) && WaitCycles == 0
	MissionRequest *MissionRequest `json:"missionRequest,omitzero"`
}
