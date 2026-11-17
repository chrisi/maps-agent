package bms

// FlightData2Version is the FlightData2 version these definitions were ported from (FLIGHTDATA2_VERSION).
const FlightData2Version = 24

const (
	RwrInfoSize    = 512 // RWRINFO_SIZE
	CallsignLen    = 12  // CALLSIGN_LEN
	MaxCallsigns   = 32  // MAX_CALLSIGNS
	MaxEcmPrograms = 5   // MAX_ECM_PROGRAMS
	rttNoOfAreas   = 7
)

// FlightData2 mirrors the "FalconSharedMemoryArea2" layout.
//
// Field order and types must match FlightData.h exactly; see layout_test.go.
// Fields are appended per version; older BMS versions provide a shorter area,
// in which case the trailing fields are zero.
type FlightData2 struct {
	// Version 1
	NozzlePos2   float32      // Ownship engine nozzle2 percent open (0-100)
	Rpm2         float32      // Ownship engine rpm2 (Percent 0-103)
	Ftit2        float32      // Ownship Forward Turbine Inlet Temp2 (Degrees C)
	OilPressure2 float32      // Ownship Oil Pressure2 (Percent 0-100)
	NavMode      NavMode      // current HSI/eHSI mode
	AAUZ         float32      // barometric altitude given by AAU (depends on calibration)
	TacanInfo    [2]TacanBits // indexed by TacanUFC / TacanAUX

	// Version 2 / 7
	AltCalReading  int32 // barometric altitude calibration (depends on AltCalType)
	AltBits        AltBits
	PowerBits      PowerBits
	BlinkBits      BlinkBits // only *if* a lamp is blinking; the blinking itself is up to the reader
	CmdsMode       CmdsMode
	UHFPanelPreset int32 // BUP UHF channel preset (F16), radio 1 preset (other aircraft)

	// Version 3
	UHFPanelFrequency int32   // BUP UHF channel frequency, radio 1 frequency (other aircraft)
	CabinAlt          float32 // cabin altitude
	HydPressureA      float32
	HydPressureB      float32
	CurrentTime       int32 // seconds (max 60*60*24)
	VehicleACD        int16 // ownship ACD index number (aircraft type)
	VersionNum        int32 // version of the FlightData2 area

	// Version 4
	FuelFlow2 float32 // Lbs/Hour

	// Version 5 / 8
	RwrInfo [RwrInfoSize]byte
	LefPos  float32 // LEF position
	TefPos  float32 // TEF position

	// Version 6
	VtolPos float32 // VTOL exhaust angle

	// Version 9
	PilotsOnline   int8                            // number of pilots in an MP session
	PilotsCallsign [MaxCallsigns][CallsignLen]byte // see PilotCallsign
	PilotsStatus   [MaxCallsigns]FlyState

	// Version 10
	BumpIntensity float32 // 0..1 while taxiing/rolling

	// Version 11
	Latitude  float32 // degrees, as known by avionics
	Longitude float32 // degrees, as known by avionics

	// Version 12
	RTTSize [2]uint16               // RTT overall width and height
	RTTArea [rttNoOfAreas][4]uint16 // per RTTArea index: left/top/right/bottom

	// Version 13
	IffBackupMode1Digit1  int8
	IffBackupMode1Digit2  int8
	IffBackupMode3ADigit1 int8
	IffBackupMode3ADigit2 int8

	// Version 14
	InstrLight InstrLight

	// Version 15
	BettyBits       BettyBits
	MiscBits        MiscBits
	RALT            float32 // radar altitude; only valid if MiscRALTValid is set
	BingoFuel       float32
	CaraAlow        float32
	BullseyeX       float32 // sim coordinates, North (Ft)
	BullseyeY       float32 // sim coordinates, East (Ft)
	BMSVersionMajor int32
	BMSVersionMinor int32
	BMSVersionMicro int32
	BMSBuildNumber  int32
	StringAreaSize  uint32 // size of the "FalconSharedMemoryAreaString" area
	StringAreaTime  uint32 // last change of the string area; only re-read it when this changes
	DrawingAreaSize uint32 // size of the "FalconSharedMemoryAreaDrawing" area

	// Version 16
	TurnRate float32 // degrees/second

	// Version 18
	FloodConsole FloodConsole

	// Version 19
	MagDeviationSystem float32
	MagDeviationReal   float32
	EcmBits            [MaxEcmPrograms]EcmBits // mutually exclusive states, not combinable bits
	EcmOper            EcmOperState
	RWRJammingStatus   [MaxRwrObjects]JammingState

	// Version 20
	Radio2Preset               int32
	Radio2Frequency            int32
	IffTransponderActiveCode1  int8 // negative for OFF or n/a
	IffTransponderActiveCode2  int16
	IffTransponderActiveCode3A int16
	IffTransponderActiveCodeC  int16
	IffTransponderActiveCode4  int16

	// Version 21
	TacanILSFrequency int32 // 110.30 = 11030, valid interval [108.10, 111.95]

	// Version 22
	DesiredRTTFPS int32 // configured RTT export FPS

	SideSlipDeg float32 // ADI side slip

	// Version 23
	GsMax float32
	GsMin float32

	// Version 24
	SunAzimuth   float32 // radians (0 to 2*PI)
	SunElevation float32 // radians (-PI/2 to +PI/2)
}

// RwrInfoString returns the RWR info blob as string.
func (fd2 *FlightData2) RwrInfoString() string { return cString(fd2.RwrInfo[:]) }

// PilotCallsign returns the callsign of MP pilot i.
func (fd2 *FlightData2) PilotCallsign(i int) string { return cString(fd2.PilotsCallsign[i][:]) }

// TACAN source indices for TacanInfo.
const (
	TacanUFC = 0
	TacanAUX = 1
)

type TacanBits uint8

func (b TacanBits) Has(f TacanBits) bool { return b&f != 0 }

const (
	TacanBandX  TacanBits = 0x01 // band is X
	TacanModeAA TacanBits = 0x02 // air to air
)

type AltBits uint32

func (b AltBits) Has(f AltBits) bool { return b&f != 0 }

const (
	AltCalType  AltBits = 0x01 // calibration in inHg (otherwise hPa)
	AltPneuFlag AltBits = 0x02 // PNEU flag visible
)

type PowerBits uint32

func (b PowerBits) Has(f PowerBits) bool { return b&f != 0 }

const (
	PowerBusBattery       PowerBits = 0x01
	PowerBusEmergency     PowerBits = 0x02
	PowerBusEssential     PowerBits = 0x04
	PowerBusNonEssential  PowerBits = 0x08
	PowerMainGenerator    PowerBits = 0x10
	PowerStandbyGenerator PowerBits = 0x20
	PowerJetFuelStarter   PowerBits = 0x40
)

type BlinkBits uint32

func (b BlinkBits) Has(f BlinkBits) bool { return b&f != 0 }

const (
	BlinkOuterMarker  BlinkBits = 0x01
	BlinkMiddleMarker BlinkBits = 0x02
	BlinkProbeHeat    BlinkBits = 0x04
	BlinkAuxSrch      BlinkBits = 0x08
	BlinkLaunch       BlinkBits = 0x10
	BlinkPriMode      BlinkBits = 0x20
	BlinkUnk          BlinkBits = 0x40
	BlinkElecFault    BlinkBits = 0x80  // not working yet
	BlinkOxyBrow      BlinkBits = 0x100 // not working yet
	BlinkEPUOn        BlinkBits = 0x200 // not working yet
	BlinkJFSOnSlow    BlinkBits = 0x400 // non-critical failure
	BlinkJFSOnFast    BlinkBits = 0x800 // critical failure
	BlinkECMOper      BlinkBits = 0x1000
)

type CmdsMode int32

const (
	CmdsOff  CmdsMode = 0
	CmdsStby CmdsMode = 1
	CmdsMan  CmdsMode = 2
	CmdsSemi CmdsMode = 3
	CmdsAuto CmdsMode = 4
	CmdsByp  CmdsMode = 5
)

type NavMode uint8

const (
	NavModeILSTacan NavMode = 0
	NavModeTacan    NavMode = 1
	NavModeNav      NavMode = 2
	NavModeILSNav   NavMode = 3
)

type FlyState uint8

const (
	FlyStateInUI    FlyState = 0
	FlyStateLoading FlyState = 1
	FlyStateWaiting FlyState = 2
	FlyStateFlying  FlyState = 3
	FlyStateDead    FlyState = 4
	FlyStateUnknown FlyState = 5
)

// RTT area indices for RTTArea.
const (
	RTTHUD      = 0
	RTTPFL      = 1
	RTTDED      = 2
	RTTRWR      = 3
	RTTMFDLeft  = 4
	RTTMFDRight = 5
	RTTHMS      = 6
)

type InstrLight uint8

const (
	InstrLightOff InstrLight = 0
	InstrLightDim InstrLight = 1
	InstrLightBrt InstrLight = 2
)

// FloodConsole is the flood console brightness, 0 (off) to 6.
type FloodConsole uint8

type BettyBits uint32

func (b BettyBits) Has(f BettyBits) bool { return b&f != 0 }

const (
	BettyAllwords      BettyBits = 0x00001
	BettyPullup        BettyBits = 0x00002
	BettyAltitude      BettyBits = 0x00004
	BettyWarning       BettyBits = 0x00008
	BettyJammer        BettyBits = 0x00010
	BettyCounter       BettyBits = 0x00020
	BettyChaffFlare    BettyBits = 0x00040
	BettyChaffFlareLow BettyBits = 0x00080
	BettyChaffFlareOut BettyBits = 0x00100
	BettyLock          BettyBits = 0x00200
	BettyCaution       BettyBits = 0x00400
	BettyBingo         BettyBits = 0x00800
	BettyData          BettyBits = 0x01000
	BettyIFF           BettyBits = 0x02000
	BettyLowspeed      BettyBits = 0x04000
	BettyBeeps         BettyBits = 0x08000
	BettyAOA           BettyBits = 0x10000
	BettyMaxG          BettyBits = 0x20000
)

type MiscBits uint32

func (b MiscBits) Has(f MiscBits) bool { return b&f != 0 }

const (
	MiscRALTValid         MiscBits = 0x01
	MiscFlcsFlccA         MiscBits = 0x02
	MiscFlcsFlccB         MiscBits = 0x04
	MiscFlcsFlccC         MiscBits = 0x08
	MiscFlcsFlccD         MiscBits = 0x10
	MiscAllLampBitsFlccOn MiscBits = 0x1e // mask for all FLCC bits
	MiscSolenoidStatus    MiscBits = 0x20
)

// EcmBits are mutually exclusive states, not combinable bits.
type EcmBits uint32

const (
	EcmUnpressedNoLit  EcmBits = 0x01
	EcmUnpressedAllLit EcmBits = 0x02
	EcmPressedNoLit    EcmBits = 0x04
	EcmPressedStandby  EcmBits = 0x08
	EcmPressedActive   EcmBits = 0x10
	EcmPressedTransmit EcmBits = 0x20
	EcmPressedFail     EcmBits = 0x40
	EcmPressedAllLit   EcmBits = 0x80
)

type EcmOperState uint8

const (
	EcmOperNoLit  EcmOperState = 0
	EcmOperStdby  EcmOperState = 1
	EcmOperActive EcmOperState = 2
	EcmOperAllLit EcmOperState = 3
)

type JammingState uint8

const (
	JammedNo     JammingState = 0
	JammedYes    JammingState = 1
	JammedShould JammingState = 2
)
