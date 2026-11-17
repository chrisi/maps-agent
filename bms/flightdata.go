package bms

// FlightDataVersion is the FlightData version these definitions were ported from (FLIGHTDATA_VERSION).
const FlightDataVersion = 118

// MaxRwrObjects is the size of the RWR arrays (MAX_RWR_OBJECTS).
const MaxRwrObjects = 40

// FlightData mirrors the "FalconSharedMemoryArea" layout.
//
// Field order and types must match FlightData.h exactly; see layout_test.go.
// C "long" is 4 bytes on Windows and therefore mapped to (u)int32.
type FlightData struct {
	X            float32 // Ownship North (Ft)
	Y            float32 // Ownship East (Ft)
	Z            float32 // Ownship Down (Ft); use FlightData2.AAUZ for barometric altitude
	XDot         float32 // Ownship North Rate (ft/sec)
	YDot         float32 // Ownship East Rate (ft/sec)
	ZDot         float32 // Ownship Down Rate (ft/sec)
	Alpha        float32 // Ownship AOA (Degrees)
	Beta         float32 // Ownship Beta (Degrees)
	Gamma        float32 // Ownship Gamma (Radians)
	Pitch        float32 // Ownship Pitch (Radians)
	Roll         float32 // Ownship Roll (Radians)
	Yaw          float32 // Ownship Yaw (Radians)
	Mach         float32 // Ownship Mach number
	Kias         float32 // Ownship Indicated Airspeed (Knots)
	Vt           float32 // Ownship True Airspeed (Ft/Sec)
	Gs           float32 // Ownship Normal Gs
	WindOffset   float32 // Wind delta to FPM (Radians)
	NozzlePos    float32 // Ownship engine nozzle percent open (0-100)
	InternalFuel float32 // Ownship internal fuel (Lbs)
	ExternalFuel float32 // Ownship external fuel (Lbs)
	FuelFlow     float32 // Ownship fuel flow (Lbs/Hour)
	Rpm          float32 // Ownship engine rpm (Percent 0-103)
	Ftit         float32 // Ownship Forward Turbine Inlet Temp (Degrees C)
	GearPos      float32 // Ownship Gear position 0 = up, 1 = down
	SpeedBrake   float32 // Ownship speed brake position 0 = closed, 1 = 60 Degrees open
	EpuFuel      float32 // Ownship EPU fuel (Percent 0-100)
	OilPressure  float32 // Ownship Oil Pressure (Percent 0-100)
	LightBits    LightBits

	// Inputs; only active when BMS is started with "-head" and TrackIR is disabled.
	HeadPitch float32 // Head pitch offset from design eye (radians)
	HeadRoll  float32 // Head roll offset from design eye (radians)
	HeadYaw   float32 // Head yaw offset from design eye (radians)

	LightBits2 LightBits2
	LightBits3 LightBits3

	ChaffCount float32 // Number of Chaff left
	FlareCount float32 // Number of Flare left

	NoseGearPos  float32 // Position of the nose landing gear; full down values defined in dat files
	LeftGearPos  float32 // Position of the left landing gear
	RightGearPos float32 // Position of the right landing gear

	AdiIlsHorPos float32 // Position of horizontal ILS bar
	AdiIlsVerPos float32 // Position of vertical ILS bar

	CourseState  int32 // HSI_STA_CRS_STATE
	HeadingState int32 // HSI_STA_HDG_STATE
	TotalStates  int32 // HSI_STA_TOTAL_STATES; never set

	CourseDeviation    float32 // HSI_VAL_CRS_DEVIATION
	DesiredCourse      float32 // HSI_VAL_DESIRED_CRS
	DistanceToBeacon   float32 // HSI_VAL_DISTANCE_TO_BEACON
	BearingToBeacon    float32 // HSI_VAL_BEARING_TO_BEACON
	CurrentHeading     float32 // HSI_VAL_CURRENT_HEADING
	DesiredHeading     float32 // HSI_VAL_DESIRED_HEADING
	DeviationLimit     float32 // HSI_VAL_DEV_LIMIT
	HalfDeviationLimit float32 // HSI_VAL_HALF_DEV_LIMIT
	LocalizerCourse    float32 // HSI_VAL_LOCALIZER_CRS
	AirbaseX           float32 // HSI_VAL_AIRBASE_X
	AirbaseY           float32 // HSI_VAL_AIRBASE_Y
	TotalValues        float32 // HSI_VAL_TOTAL_VALUES; never set

	TrimPitch float32 // -0.5 to +0.5
	TrimRoll  float32 // -0.5 to +0.5
	TrimYaw   float32 // -0.5 to +0.5

	HsiBits HsiBits

	DEDLines  [5][26]byte // 25 usable chars, NUL terminated; see DEDLine
	Invert    [5][26]byte
	PFLLines  [5][26]byte // see PFLLine
	PFLInvert [5][26]byte

	UFCTChan int32 // TACAN channels
	AUXTChan int32

	RwrObjectCount  int32
	RWRSymbol       [MaxRwrObjects]int32
	Bearing         [MaxRwrObjects]float32
	MissileActivity [MaxRwrObjects]uint32
	MissileLaunch   [MaxRwrObjects]uint32
	Selected        [MaxRwrObjects]uint32
	Lethality       [MaxRwrObjects]float32
	NewDetection    [MaxRwrObjects]uint32

	Fwd   float32 // fuel values
	Aft   float32
	Total float32

	VersionNum int32 // Version of the FlightData area

	HeadX float32 // Head X offset from design eye (feet)
	HeadY float32 // Head Y offset from design eye (feet)
	HeadZ float32 // Head Z offset from design eye (feet)

	MainPower int32 // Main Power switch state, 0=down, 1=middle, 2=up
}

// DEDLine returns DED line i (0-4) as string.
func (fd *FlightData) DEDLine(i int) string { return cString(fd.DEDLines[i][:]) }

// PFLLine returns PFL line i (0-4) as string.
func (fd *FlightData) PFLLine(i int) string { return cString(fd.PFLLines[i][:]) }

// IsFlying reports whether the player is attached to an aircraft (i.e. not in the UI).
func (fd *FlightData) IsFlying() bool { return fd.HsiBits.Has(HsiFlying) }

type LightBits uint32

func (b LightBits) Has(f LightBits) bool { return b&f != 0 }

const (
	LightMasterCaution LightBits = 0x1 // Left eyebrow

	// Brow lights
	LightTF       LightBits = 0x2   // Left eyebrow
	LightOxyBrow  LightBits = 0x4   // eyebrow OXY LOW
	LightEquipHot LightBits = 0x8   // caution light; cooling fault
	LightOnGround LightBits = 0x10  // true if on ground; not a lamp bit
	LightEngFire  LightBits = 0x20  // right eyebrow; upper half of split face lamp
	LightConfig   LightBits = 0x40  // stores config, caution panel
	LightHyd      LightBits = 0x80  // right eyebrow
	LightFlcsABCD LightBits = 0x100 // TEST panel FLCS channel lamps
	LightFlcs     LightBits = 0x200 // right eyebrow
	LightCan      LightBits = 0x400 // right eyebrow
	LightTLCfg    LightBits = 0x800 // right eyebrow

	// AOA indexers
	LightAOAAbove LightBits = 0x1000
	LightAOAOn    LightBits = 0x2000
	LightAOABelow LightBits = 0x4000

	// Refuel/NWS
	LightRefuelRDY LightBits = 0x8000
	LightRefuelAR  LightBits = 0x10000
	LightRefuelDSC LightBits = 0x20000

	// Caution lights
	LightFltControlSys LightBits = 0x40000
	LightLEFlaps       LightBits = 0x80000
	LightEngineFault   LightBits = 0x100000
	LightOverheat      LightBits = 0x200000
	LightFuelLow       LightBits = 0x400000
	LightAvionics      LightBits = 0x800000
	LightRadarAlt      LightBits = 0x1000000
	LightIFF           LightBits = 0x2000000
	LightECM           LightBits = 0x4000000
	LightHook          LightBits = 0x8000000
	LightNWSFail       LightBits = 0x10000000
	LightCabinPress    LightBits = 0x20000000

	LightAutoPilotOn LightBits = 0x40000000 // not a lamp bit
	LightTFRStby     LightBits = 0x80000000 // MISC panel; lower half of split face TFR lamp

	LightAllLampBitsOn LightBits = 0xBFFFFFEF
)

type LightBits2 uint32

func (b LightBits2) Has(f LightBits2) bool { return b&f != 0 }

const (
	// Threat Warning Prime
	Light2HandOff LightBits2 = 0x1
	Light2Launch  LightBits2 = 0x2
	Light2PriMode LightBits2 = 0x4
	Light2Naval   LightBits2 = 0x8
	Light2Unk     LightBits2 = 0x10
	Light2TgtSep  LightBits2 = 0x20

	// EWS
	Light2Go      LightBits2 = 0x40  // on and operating normally
	Light2NoGo    LightBits2 = 0x80  // on but malfunction present
	Light2Degr    LightBits2 = 0x100 // AUTO DEGR
	Light2Rdy     LightBits2 = 0x200 // DISPENSE RDY
	Light2ChaffLo LightBits2 = 0x400 // bingo chaff quantity reached
	Light2FlareLo LightBits2 = 0x800 // bingo flare quantity reached

	// Aux Threat Warning
	Light2AuxSrch LightBits2 = 0x1000
	Light2AuxAct  LightBits2 = 0x2000
	Light2AuxLow  LightBits2 = 0x4000
	Light2AuxPwr  LightBits2 = 0x8000

	// ECM
	Light2EcmPwr  LightBits2 = 0x10000
	Light2EcmFail LightBits2 = 0x20000

	// Caution lights
	Light2FwdFuelLow LightBits2 = 0x40000
	Light2AftFuelLow LightBits2 = 0x80000

	Light2EPUOn LightBits2 = 0x100000 // EPU panel; run light
	Light2JFSOn LightBits2 = 0x200000 // Eng Jet Start panel; run light

	// Caution panel
	Light2Sec        LightBits2 = 0x400000
	Light2OxyLow     LightBits2 = 0x800000
	Light2ProbeHeat  LightBits2 = 0x1000000
	Light2SeatArm    LightBits2 = 0x2000000
	Light2BUC        LightBits2 = 0x4000000
	Light2FuelOilHot LightBits2 = 0x8000000
	Light2AntiSkid   LightBits2 = 0x10000000

	Light2TFREngaged LightBits2 = 0x20000000 // MISC panel; upper half of split face TFR lamp
	Light2GearHandle LightBits2 = 0x40000000 // gear handle lamp
	Light2Engine     LightBits2 = 0x80000000 // lower half of right eyebrow ENG FIRE/ENGINE lamp

	Light2AllLampBitsOn               LightBits2 = 0xFFFFF03F
	Light2AllLampBitsOnExceptCarapace            = Light2AllLampBitsOn ^ Light2HandOff ^ Light2Launch ^ Light2PriMode ^
		Light2Naval ^ Light2Unk ^ Light2TgtSep ^ Light2AuxSrch ^ Light2AuxAct ^ Light2AuxLow ^ Light2AuxPwr
)

type LightBits3 uint32

func (b LightBits3) Has(f LightBits3) bool { return b&f != 0 }

const (
	// Elec panel
	Light3FlcsPmg LightBits3 = 0x1
	Light3MainGen LightBits3 = 0x2
	Light3StbyGen LightBits3 = 0x4
	Light3EpuGen  LightBits3 = 0x8
	Light3EpuPmg  LightBits3 = 0x10
	Light3ToFlcs  LightBits3 = 0x20
	Light3FlcsRly LightBits3 = 0x40
	Light3BatFail LightBits3 = 0x80

	// EPU panel
	Light3Hydrazine LightBits3 = 0x100
	Light3Air       LightBits3 = 0x200

	// Caution panel
	Light3ElecFault LightBits3 = 0x400
	Light3LefFault  LightBits3 = 0x800

	Light3OnGround      LightBits3 = 0x1000   // weight-on-wheels
	Light3FlcsBitRun    LightBits3 = 0x2000   // FLT CONTROL panel RUN light
	Light3FlcsBitFail   LightBits3 = 0x4000   // FLT CONTROL panel FAIL light
	Light3DbuWarn       LightBits3 = 0x8000   // right eyebrow DBU ON cell
	Light3NoseGearDown  LightBits3 = 0x10000  // down and locked
	Light3LeftGearDown  LightBits3 = 0x20000  // down and locked
	Light3RightGearDown LightBits3 = 0x40000  // down and locked
	Light3ParkBrakeOn   LightBits3 = 0x100000 // not a lamp bit
	Light3PowerOff      LightBits3 = 0x200000 // no electrical power; not a lamp bit

	Light3Cadc          LightBits3 = 0x400000  // caution panel
	Light3SpeedBrake    LightBits3 = 0x800000  // speed brake not stowed
	Light3SysTest       LightBits3 = 0x1000000 // Threat Warning Prime
	Light3MCAnnounced   LightBits3 = 0x2000000 // master caution will come up (3 sec delay)
	Light3MLGWOW        LightBits3 = 0x4000000 // AFM only
	Light3NLGWOW        LightBits3 = 0x8000000
	Light3ATFNotEngaged LightBits3 = 0x10000000
	Light3InletIcing    LightBits3 = 0x20000000 // caution panel

	Light3AllLampBitsOn               LightBits3 = 0x3147EFFF
	Light3AllLampBitsOnExceptCarapace            = Light3AllLampBitsOn ^ Light3SysTest
)

type HsiBits uint32

func (b HsiBits) Has(f HsiBits) bool { return b&f != 0 }

const (
	HsiToTrue        HsiBits = 0x01 // TO
	HsiIlsWarning    HsiBits = 0x02
	HsiCourseWarning HsiBits = 0x04
	HsiInit          HsiBits = 0x08
	HsiTotalFlags    HsiBits = 0x10 // never set
	HsiADIOff        HsiBits = 0x20
	HsiADIAux        HsiBits = 0x40
	HsiADIGS         HsiBits = 0x80
	HsiADILoc        HsiBits = 0x100
	HsiHSIOff        HsiBits = 0x200
	HsiBUPADIOff     HsiBits = 0x400
	HsiVVI           HsiBits = 0x800
	HsiAOA           HsiBits = 0x1000
	HsiAVTR          HsiBits = 0x2000
	HsiOuterMarker   HsiBits = 0x4000
	HsiMiddleMarker  HsiBits = 0x8000
	HsiFromTrue      HsiBits = 0x10000 // FROM

	HsiFlying HsiBits = 0x80000000 // player is attached to an aircraft (not in UI); not a lamp bit

	HsiAllLampBitsOn HsiBits = 0xE000
)
