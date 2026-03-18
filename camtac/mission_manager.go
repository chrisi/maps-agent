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
	classTable []*SCT
	objectives []*Objective
	deltas     []*ObjectiveDeltas
	units      []*any
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
	m.classTable = CreateStrippedClassTable(records)
	logBdRed.Infof("ClassTypes: %d", len(m.classTable))
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

func (m *MissionManager) outputJson(outputBase string, outputFilePrefix string) {
	logBdRed.Infof("Writing json data to %s/%s_*.json", outputBase, outputFilePrefix)
	err := WriteToJSON(m.units, outputBase+"/"+outputFilePrefix+"_units.json")
	if err != nil {
		logBdRed.Errorf("error writing units to JSON: %v", err)
	}
	err = WriteToJSON(m.objectives, outputBase+"/"+outputFilePrefix+"_objectives.json")
	if err != nil {
		logBdRed.Errorf("error writing objectives to JSON: %v", err)
	}
	err = WriteToJSON(m.deltas, outputBase+"/"+outputFilePrefix+"_deltas.json")
	if err != nil {
		logBdRed.Errorf("error writing deltas to JSON: %v", err)
	}

	objectiveTree := buildObjectiveTree(m.objectives)
	err = WriteToJSON(objectiveTree, outputBase+"/"+outputFilePrefix+"_objective_tree.json")
	if err != nil {
		logBdRed.Errorf("error writing deltas to JSON: %v", err)
	}
}

func (m *MissionManager) applyDeltas() {
	logBdRed.Infof("Applying deltas to objectives")
	deltaByID := make(map[uint32]*ObjectiveDeltas, len(m.deltas))
	for _, delta := range m.deltas {
		deltaByID[delta.ID.Num] = delta
	}
	for _, obj := range m.objectives {
		if delta, ok := deltaByID[obj.CampaignBase.ID.Num]; ok {
			logChanges(obj, delta)
			obj.LastRepair = delta.LastRepair
			obj.FirstOwner = delta.Owner //TODO: rename field to better reflect that it can be changed
			obj.Supply = delta.Supply
			obj.Fuel = delta.Fuel
			obj.Losses = delta.Losses
			obj.Statuses = delta.Statuses
		}
	}
}

func logChanges(obj *Objective, delta *ObjectiveDeltas) {
	if obj.LastRepair != delta.LastRepair {
		logBdRed.Debugf("Objective %d: LastRepair %d -> %d", obj.CampaignBase.ID.Num, obj.LastRepair, delta.LastRepair)
	}
	if obj.FirstOwner != delta.Owner {
		logBdRed.Debugf("Objective %d: FirstOwner %d -> %d", obj.CampaignBase.ID.Num, obj.FirstOwner, delta.Owner)
	}
	if obj.Supply != delta.Supply {
		logBdRed.Debugf("Objective %d: Supply %d -> %d", obj.CampaignBase.ID.Num, obj.Supply, delta.Supply)
	}
	if obj.Fuel != delta.Fuel {
		logBdRed.Debugf("Objective %d: Fuel %d -> %d", obj.CampaignBase.ID.Num, obj.Fuel, delta.Fuel)
	}
	if obj.Losses != delta.Losses {
		logBdRed.Debugf("Objective %d: Losses %d -> %d", obj.CampaignBase.ID.Num, obj.Losses, delta.Losses)
	}
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
	missionBundle := m.loadBundle(campaignBase + "/" + missionFilename)

	deltaData, err := missionBundle.GetEmbeddedFileContentsByType(ObjectiveDeltaType)
	if err != nil {
		log.Fatal(err)
	}
	deltaReader := NewObjectiveDeltaReader()
	m.deltas = deltaReader.ReadObdFile(deltaData)
	logBdRed.Infof("Num Deltas: %d", len(m.deltas))

	campaignData, err := missionBundle.GetEmbeddedFileContentsByType(CampaignType)
	if err != nil {
		log.Fatal(err)
	}
	campaignReader := NewCampaignReader()
	campaign, err := campaignReader.ReadCmpFile(campaignData)
	if err != nil {
		log.Fatal(err)
	}

	logBdRed.Infof("Theater: %s", campaign.TheaterName)
	logBdRed.Infof("Scenario: %s", campaign.Scenario)

	unitData, err := missionBundle.GetEmbeddedFileContentsByType(UnitType)
	if err != nil {
		log.Fatal(err)
	}
	unitReader := NewUnitReader(m.classTable)
	m.units = unitReader.ReadUniFile(unitData)

	unitCounts := unitReader.Counts()
	logBdRed.Infof("Num Units:       %d", unitCounts.NumUnits)
	logBdRed.Infof("Num Squadrons:   %d", unitCounts.NumSquadrons)
	logBdRed.Infof("Num Packages:    %d", unitCounts.NumPackages)
	logBdRed.Infof("Num Flights:     %d", unitCounts.NumFlights)
	logBdRed.Infof("Num Brigades:    %d", unitCounts.NumBrigades)
	logBdRed.Infof("Num Battalions:  %d", unitCounts.NumBattalions)
	logBdRed.Infof("Num Task Forces: %d", unitCounts.NumTaskForces)

	baseMission := campaign.Scenario + filepath.Ext(missionFilename)
	logBdRed.Infof("Base Mission: %s", baseMission)

	objectiveBundle := m.loadBundle(campaignBase + "/" + baseMission)
	objectiveData, err := objectiveBundle.GetEmbeddedFileContentsByType(ObjectiveType)
	if err != nil {
		log.Fatal(err)
	}
	objectiveReader := NewObjectiveReader(m.classTable)
	m.objectives = objectiveReader.ReadObjFile(objectiveData)
	logBdRed.Infof("Num Objectives: %d", len(m.objectives))

	m.applyDeltas()

	fileNoExt := strings.TrimSuffix(missionFilename, filepath.Ext(missionFilename))
	m.outputJson(outputBase, fileNoExt)
}

type ObjectiveNode struct {
	Objective *Objective
	CampName  string
	Children  []*ObjectiveNode
}

func buildObjectiveTree(objectives []*Objective) []*ObjectiveNode {
	nodes := make(map[uint32]*ObjectiveNode, len(objectives))
	roots := make([]*ObjectiveNode, 0)

	for i := range objectives {
		obj := objectives[i]
		id := obj.CampaignBase.ID.Num
		nodes[id] = &ObjectiveNode{
			Objective: obj,
			CampName:  obj.CampName,
			Children:  make([]*ObjectiveNode, 0),
		}
	}

	for i := range objectives {
		obj := objectives[i]
		id := obj.CampaignBase.ID.Num
		parentID := obj.ParentID.Num
		node := nodes[id]
		parent, ok := nodes[parentID]
		if !ok || parentID == 0 || parentID == id {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}
