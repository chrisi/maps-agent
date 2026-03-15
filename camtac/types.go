package camtac

type EmbeddedFileInfo struct {
	FileName      string
	FileOffset    uint32
	FileSizeBytes uint32
}

type VU_ID struct {
	Num     uint32
	Creator uint32
}

type CampaignBase struct {
	ID         VU_ID
	EntityType uint16
	X          uint16
	Y          uint16
	Z          float32
	SpotTime   uint32
	Spotted    uint16
	BaseFlags  uint16
	Owner      uint8
	CampID     uint16
}

type Target struct {
	ID       VU_ID
	Building uint8
}

type WaypointTargetData struct {
	Target           Target
	DesignatedTarget [4]Target
}

type Waypoint struct {
	Haves       uint8
	GridX       uint16
	GridY       uint16
	GridZ       uint16
	Arrive      uint32
	Action      uint8
	RouteAction uint8
	Formation   uint8
	Flags       uint32

	// Present if (Haves & 2) != 0
	TargetData *WaypointTargetData

	// Present if (Haves & 1) != 0
	Depart *uint32
}

type Unit struct {
	CampaignBase  CampaignBase
	LastCheck     uint32
	Roster        int32
	UnitFlags     int32
	DestX         int16
	DestY         int16
	TargetID      VU_ID
	CargoID       VU_ID
	Moved         uint8
	Losses        uint8
	Tactic        uint8
	CurrentWP     uint16
	NameID        uint16
	Reinforcement uint16
	NumWaypoints  uint16
	WP            []Waypoint
}

type Pilot struct {
	PilotID       uint16
	Skill         uint8
	Status        uint8
	AAKills       uint8
	AGKills       uint8
	ASKills       uint8
	ANKills       uint8
	MissionsFlown uint16
}

type Squadron struct {
	Unit               Unit
	Fuel               uint32
	Speciality         uint8
	CampSpecRoleRating [16]uint8
	Stores             [1000]uint8
	Pilots             [48]Pilot
	Schedule           [16]uint32
	AirbaseID          VU_ID
	Hotspot            VU_ID
	Rating             [16]uint8
	AAKills            uint16
	AGKills            uint16
	ASKills            uint16
	ANKills            uint16
	MissionsFlown      uint16
	MissionScore       uint16
	TotalLosses        uint8
	PilotLosses        uint8
	SquadronPatch      uint16
	SquadronRetaskAt   uint32
	Relocate           uint8
	TexSet             uint32
}

type Loadout struct {
	WeaponID    [16]uint16
	WeaponCount [16]uint8
}

type Flight struct {
	Unit            Unit
	Z               float32
	FuelBurnt       uint32
	FuelInitial     [4]uint32
	LaserCode       [4]uint16
	LastMove        uint32
	LastCombat      uint32
	TimeOnTarget    uint32
	MissionOverTime uint32
	MissionTarget   uint16
	Loadouts        uint8
	Loadout         []Loadout
	Mission         uint8
	OldMission      uint8
	LastDirection   uint8
	Priority        uint8
	MissionID       uint8
	EvalFlags       uint8
	MissionContext  uint8
	Package         VU_ID
	Squadron        VU_ID
	Requester       VU_ID
	Slots           [4]uint8
	Pilots          [4]uint8
	PlaneStats      [4]uint8
	PlayerSlots     [4]uint8
	LastPlayerSlot  uint8
	CallsignID      uint8
	CallsignNum     uint8
	RefuelQuantity  uint32
	TexSet          [4]uint32
	TacanChannel    [4]uint8
	TacanBand       [4]uint8
	LoadedCft       [4]uint8
}

type MissionRequest struct {
	Requests     uint16
	Responses    uint16
	MRMission    uint8
	MRAircraft   uint8
	MRContext    uint8
	MRRoeCheck   uint8
	Requester    VU_ID
	Target       VU_ID
	MRTot        uint32
	MRActionType uint8
	MRPriority   uint16
}

type Package struct {
	Unit        Unit
	NumElements uint8
	Elements    []VU_ID
	Interceptor VU_ID
	Awacs       VU_ID
	JStar       VU_ID
	ECM         VU_ID
	Tanker      VU_ID
	WaitCycles  uint8

	// Present if ((Unit.UnitFlags & 0x100000) != 0) && WaitCycles == 0
	MissionRequest *MissionRequest
}
