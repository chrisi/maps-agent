package camtac

import (
	"log"
	"maps-agent/util"
	"os"
	"path/filepath"
	"strings"
)

var logBdRed = util.NewLogger("FileBundleReader", os.Stdout, util.Info, true)

type MissionManager struct {
	falconBase string
	classTable []*CT
}

func (m *MissionManager) loadClassTable() {
	logBdRed.Infof("Loading ClassTable")
	if m.classTable != nil {
		return
	}
	records, err := LoadCTRecords(m.falconBase + "/Data/TerrData/Objects/Falcon4_CT.xml")
	if err != nil {
		log.Fatal(err)
	}
	m.classTable = CreateClassTable(records)
	logBdRed.Debugf("ClassTypes: %d", len(m.classTable))
}

func (m *MissionManager) loadBundle(filename string) *FileBundleReader {
	logBdRed.Infof("Reading table of content of %s", filename)
	reader, err := NewFileBundleReaderFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	files, err := reader.GetEmbeddedFileDirectory()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		logBdRed.Debugf("Name: %s, Offset: %d,  Size: %d", f.FileName, f.FileOffset, f.FileSizeBytes)
	}
	return reader
}

func NewMissionManager(falconBase string) *MissionManager {
	return &MissionManager{
		falconBase: falconBase,
	}
}

func (m *MissionManager) ReadMission(missionFilename string, outputBase string) {
	logBdRed.Infof("Reading mission %s", missionFilename)

	campaignBase := m.falconBase + "/Data/Campaign"

	m.loadClassTable()
	objectiveBundle := m.loadBundle(campaignBase + "/Te_New.tac") // TODO: read from campaign scenario
	missionBundle := m.loadBundle(campaignBase + "/" + missionFilename)

	objectiveData, err := objectiveBundle.GetEmbeddedFileContentsByType(ObjectiveType)
	if err != nil {
		log.Fatal(err)
	}
	objectiveReader := NewObjectiveReader()
	objectives := objectiveReader.ReadObjFile(objectiveData)
	logBdRed.Infof("Num Objectives: %d", len(objectives))

	deltaData, err := missionBundle.GetEmbeddedFileContentsByType(ObjectiveDeltaType)
	if err != nil {
		log.Fatal(err)
	}
	deltaReader := NewObjectiveDeltaReader()
	deltas := deltaReader.ReadObdFile(deltaData)
	logBdRed.Infof("Num Deltas: %d", len(deltas))

	unitData, err := missionBundle.GetEmbeddedFileContentsByType(UnitType)
	if err != nil {
		log.Fatal(err)
	}
	unitReader := NewUnitReader(m.classTable)
	units := unitReader.ReadUniFile(unitData)

	unitCounts := unitReader.Counts()
	logBdRed.Infof("Num Units:       %d", unitCounts.NumUnits)
	logBdRed.Infof("Num Squadrons:   %d", unitCounts.NumSquadrons)
	logBdRed.Infof("Num Packages:    %d", unitCounts.NumPackages)
	logBdRed.Infof("Num Flights:     %d", unitCounts.NumFlights)
	logBdRed.Infof("Num Brigades:    %d", unitCounts.NumBrigades)
	logBdRed.Infof("Num Battalions:  %d", unitCounts.NumBattalions)
	logBdRed.Infof("Num Task Forces: %d", unitCounts.NumTaskForces)

	fileNoExt := strings.TrimSuffix(missionFilename, filepath.Ext(missionFilename))
	err = WriteToJSON(units, outputBase+"/"+fileNoExt+"_units.json")
	if err != nil {
		logBdRed.Errorf("error writing units to JSON: %v", err)
	}
	err = WriteToJSON(objectives, outputBase+"/"+fileNoExt+"_objectives.json")
	if err != nil {
		logBdRed.Errorf("error writing objectives to JSON: %v", err)
	}
	err = WriteToJSON(deltas, outputBase+"/"+fileNoExt+"_deltas.json")
	if err != nil {
		logBdRed.Errorf("error writing deltas to JSON: %v", err)
	}
}
