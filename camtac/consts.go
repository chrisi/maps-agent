package camtac

const (
	DomainAbstract    = 1
	DomainAir         = 2
	DomainLand        = 3
	DomainSea         = 4
	DomainSpace       = 5
	DomainUnderground = 6
	DomainUndersea    = 7
)

const (
	ClassAbstract  = 0
	ClassAnimal    = 1
	ClassFeature   = 2
	ClassManager   = 3
	ClassObjective = 4
	ClassSfx       = 5
	ClassUnit      = 6
	ClassVehicle   = 7
	ClassWeapon    = 8
	ClassWeather   = 9
	ClassSession   = 10
	ClassGame      = 11
	ClassGroup     = 12
	ClassDialog    = 13
)

const (
	//Unit/Air
	TypeFlight   = 1
	TypePackage  = 2
	TypeSquadron = 3

	//Unit/Land
	TypeBattalion = 1
	TypeBrigade   = 2

	//Unit/Sea
	TypeTaskForce = 1

	//Objectives/Land
	TypeAirbase       = 1
	TypeAirstrip      = 2
	TypeArmyBase      = 3
	TypeBeach         = 4
	TypeBorder        = 5
	TypeBridge        = 6
	TypeChemical      = 7
	TypeCity          = 8
	TypeComCon        = 9
	TypeDepot         = 10
	TypeFactory       = 11
	TypeFord          = 12
	TypeFortification = 13
	TypeScenery       = 14
	TypeIntersect     = 15
	TypeNavBeacon     = 16
	TypeNuclear       = 17
	TypePass          = 18
	TypePort          = 19
	TypePowerPlant    = 20
	TypeRadar         = 21
	TypeRadioTower    = 22
	TypeRailTerminal  = 23
	TypeRailroad      = 24
	TypeRefinery      = 25
	TypeRailroad2     = 26
	TypeSea           = 27
	TypeTown          = 28
	TypeVillage       = 29
	TypeHARTS         = 30
	TypeSAMSite       = 31
	TypeAAASite       = 32
)
