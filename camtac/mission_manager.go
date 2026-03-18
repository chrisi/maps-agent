package camtac

import (
	"log"
	"maps-agent/util"
	"os"
	"path/filepath"
)

var debuglog = util.NewLogger("Debug", os.Stdout, util.Info, true)

func logChanges(obj *Objective, delta *ObjectiveDeltas) {
	if obj.LastRepair != delta.LastRepair {
		debuglog.Debugf("Objective %d: LastRepair %d -> %d", obj.CampaignBase.ID.Num, obj.LastRepair, delta.LastRepair)
	}
	if obj.FirstOwner != delta.Owner {
		debuglog.Debugf("Objective %d: FirstOwner %d -> %d", obj.CampaignBase.ID.Num, obj.FirstOwner, delta.Owner)
	}
	if obj.Supply != delta.Supply {
		debuglog.Debugf("Objective %d: Supply %d -> %d", obj.CampaignBase.ID.Num, obj.Supply, delta.Supply)
	}
	if obj.Fuel != delta.Fuel {
		debuglog.Debugf("Objective %d: Fuel %d -> %d", obj.CampaignBase.ID.Num, obj.Fuel, delta.Fuel)
	}
	if obj.Losses != delta.Losses {
		debuglog.Debugf("Objective %d: Losses %d -> %d", obj.CampaignBase.ID.Num, obj.Losses, delta.Losses)
	}
}

type MissionManager struct {
	log        *util.Logger
	falconBase string
	classTable []*SCT
	objectives []*Objective
	deltas     []*ObjectiveDeltas
	units      []*any
}

func NewMissionManager(falconBase string) *MissionManager {
	return &MissionManager{
		log:        util.NewLogger("Mission Manager", os.Stdout, util.Info, true),
		falconBase: falconBase,
	}
}

func (m *MissionManager) ReadMission(missionFilename string) {
	m.log.Infof("Reading mission %s", missionFilename)

	campaignBase := m.falconBase + "/Data/Campaign"

	m.loadClassTable()
	missionBundle := m.loadBundle(campaignBase + "/" + missionFilename)

	deltaData, err := missionBundle.GetEmbeddedFileContentsByType(ObjectiveDeltaType)
	if err != nil {
		log.Fatal(err)
	}
	deltaReader := NewObjectiveDeltaReader()
	m.deltas = deltaReader.ReadObdFile(deltaData)
	m.log.Infof("Num Deltas: %d", len(m.deltas))

	campaignData, err := missionBundle.GetEmbeddedFileContentsByType(CampaignType)
	if err != nil {
		log.Fatal(err)
	}
	campaignReader := NewCampaignReader()
	campaign, err := campaignReader.ReadCmpFile(campaignData)
	if err != nil {
		log.Fatal(err)
	}

	m.log.Infof("Theater: %s", campaign.TheaterName)
	m.log.Infof("Scenario: %s", campaign.Scenario)

	unitData, err := missionBundle.GetEmbeddedFileContentsByType(UnitType)
	if err != nil {
		log.Fatal(err)
	}
	unitReader := NewUnitReader(m.classTable)
	m.units = unitReader.ReadUniFile(unitData)

	unitCounts := unitReader.Counts()
	m.log.Debugf("Num Units:       %d", unitCounts.NumUnits)
	m.log.Debugf("Num Squadrons:   %d", unitCounts.NumSquadrons)
	m.log.Debugf("Num Packages:    %d", unitCounts.NumPackages)
	m.log.Debugf("Num Flights:     %d", unitCounts.NumFlights)
	m.log.Debugf("Num Brigades:    %d", unitCounts.NumBrigades)
	m.log.Debugf("Num Battalions:  %d", unitCounts.NumBattalions)
	m.log.Debugf("Num Task Forces: %d", unitCounts.NumTaskForces)

	baseMission := campaign.Scenario + filepath.Ext(missionFilename)
	m.log.Infof("Base Mission: %s", baseMission)

	objectiveBundle := m.loadBundle(campaignBase + "/" + baseMission)
	objectiveData, err := objectiveBundle.GetEmbeddedFileContentsByType(ObjectiveType)
	if err != nil {
		log.Fatal(err)
	}
	objectiveReader := NewObjectiveReader(m.classTable)
	m.objectives = objectiveReader.ReadObjFile(objectiveData)
	m.log.Infof("Num Objectives: %d", len(m.objectives))

	m.applyDeltas()
}

func (m *MissionManager) OutputJson(outputBase string, outputFilePrefix string) {
	m.log.Infof("Writing json data to %s/%s_*.json", outputBase, outputFilePrefix)
	err := WriteToJSON(m.units, outputBase+"/"+outputFilePrefix+"_units.json")
	if err != nil {
		m.log.Errorf("error writing units to JSON: %v", err)
	}
	err = WriteToJSON(m.objectives, outputBase+"/"+outputFilePrefix+"_objectives.json")
	if err != nil {
		m.log.Errorf("error writing objectives to JSON: %v", err)
	}
	err = WriteToJSON(m.deltas, outputBase+"/"+outputFilePrefix+"_deltas.json")
	if err != nil {
		m.log.Errorf("error writing deltas to JSON: %v", err)
	}

	objectiveTree := buildObjectiveTree(m.objectives)
	err = WriteToJSON(objectiveTree, outputBase+"/"+outputFilePrefix+"_objective_tree.json")
	if err != nil {
		m.log.Errorf("error writing deltas to JSON: %v", err)
	}
}

func (m *MissionManager) loadClassTable() {
	m.log.Infof("Loading ClassTable")
	if m.classTable != nil {
		return
	}
	records, err := LoadCTRecords(m.falconBase + "/Data/TerrData/Objects/Falcon4_CT.xml")
	if err != nil {
		log.Fatal(err)
	}
	m.classTable = CreateStrippedClassTable(records)
	m.log.Infof("ClassTypes: %d", len(m.classTable))
}

func (m *MissionManager) loadBundle(filename string) *FileBundleReader {
	m.log.Infof("Reading table of content of %s", filename)
	reader, err := NewFileBundleReaderFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	files, err := reader.GetEmbeddedFileDirectory()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		m.log.Debugf("Name: %s, Offset: %d,  Size: %d", f.FileName, f.FileOffset, f.FileSizeBytes)
	}
	return reader
}

func (m *MissionManager) applyDeltas() {
	m.log.Infof("Applying deltas to objectives")
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

func (m *MissionManager) getAllByType(objectType int) []*Objective {
	result := make([]*Objective, 0)
	for _, obj := range m.objectives {
		if obj == nil || obj.ClassType == nil {
			continue
		}
		if obj.ClassType.Type == objectType {
			result = append(result, obj)
		}
	}
	return result
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
