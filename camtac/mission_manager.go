package camtac

import (
	"log"
	"maps-agent/util"
	"os"
	"path/filepath"
)

var logger = util.NewLogger("Mission-Manager", os.Stdout, util.Info, true)

func logChanges(obj *Objective, delta *ObjectiveDeltas) {
	if obj.LastRepair != delta.LastRepair {
		logger.Debugf("Objective %d: LastRepair %d -> %d", obj.CampaignBase.ID.Num, obj.LastRepair, delta.LastRepair)
	}
	if obj.FirstOwner != delta.Owner {
		logger.Debugf("Objective %d: FirstOwner %d -> %d", obj.CampaignBase.ID.Num, obj.FirstOwner, delta.Owner)
	}
	if obj.Supply != delta.Supply {
		logger.Debugf("Objective %d: Supply %d -> %d", obj.CampaignBase.ID.Num, obj.Supply, delta.Supply)
	}
	if obj.Fuel != delta.Fuel {
		logger.Debugf("Objective %d: Fuel %d -> %d", obj.CampaignBase.ID.Num, obj.Fuel, delta.Fuel)
	}
	if obj.Losses != delta.Losses {
		logger.Debugf("Objective %d: Losses %d -> %d", obj.CampaignBase.ID.Num, obj.Losses, delta.Losses)
	}
}

type MissionManager struct {
	falconBase      string
	deltaReader     *ObjectiveDeltaReader
	objectiveReader *ObjectiveReader
	campaignReader  *CampaignReader
	unitReader      *UnitReader
	campaign        *Campaign
	classTable      []*SCT
	objectives      []*Objective
	deltas          []*ObjectiveDeltas
	units           []*any
}

func NewMissionManager(falconBase string) *MissionManager {
	return &MissionManager{
		falconBase: falconBase,
	}
}

type Theater string

const (
	Korea   Theater = ""
	Balkans Theater = "/Add-On Balkans"
	Israel  Theater = "/Add-On Israel"
	Hellas  Theater = "/Add-On Hellas"
)

func (m *MissionManager) ReadMission(theater Theater, missionFilename string) {
	logger.Infof("Reading mission %s", missionFilename)

	campaignBase := m.falconBase + "/Data" + string(theater) + "/Campaign"

	m.loadClassTable()
	m.initializeReaders()
	missionBundle := m.loadBundle(campaignBase + "/" + missionFilename)

	deltaData, err := missionBundle.GetEmbeddedFileContentsByType(ObjectiveDeltaType)
	if err != nil {
		log.Fatal(err)
	}
	m.deltas = m.deltaReader.ReadObdFile(deltaData)
	logger.Infof("Num Deltas: %d", len(m.deltas))

	campaignData, err := missionBundle.GetEmbeddedFileContentsByType(CampaignType)
	if err != nil {
		log.Fatal(err)
	}
	m.campaign, err = m.campaignReader.ReadCmpFile(campaignData)
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("Theater: %s", m.campaign.TheaterName)
	logger.Infof("Scenario: %s", m.campaign.Scenario)
	logger.Infof("UI-name: %s", m.campaign.UiName)

	unitData, err := missionBundle.GetEmbeddedFileContentsByType(UnitType)
	if err != nil {
		log.Fatal(err)
	}
	m.units = m.unitReader.ReadUniFile(unitData)
	unitCounts := m.unitReader.Counts()
	logger.Debugf("Num Units:       %d", unitCounts.NumUnits)
	logger.Debugf("Num Squadrons:   %d", unitCounts.NumSquadrons)
	logger.Debugf("Num Packages:    %d", unitCounts.NumPackages)
	logger.Debugf("Num Flights:     %d", unitCounts.NumFlights)
	logger.Debugf("Num Brigades:    %d", unitCounts.NumBrigades)
	logger.Debugf("Num Battalions:  %d", unitCounts.NumBattalions)
	logger.Debugf("Num Task Forces: %d", unitCounts.NumTaskForces)

	baseMission := m.campaign.Scenario + filepath.Ext(missionFilename)
	logger.Infof("Base Mission: %s", baseMission)
	objectiveBundle := m.loadBundle(campaignBase + "/" + baseMission)
	objectiveData, err := objectiveBundle.GetEmbeddedFileContentsByType(ObjectiveType)
	if err != nil {
		log.Fatal(err)
	}
	m.objectives = m.objectiveReader.ReadObjFile(objectiveData)
	logger.Infof("Num Objectives: %d", len(m.objectives))

	m.applyDeltas()
}

func (m *MissionManager) OutputJson(outputBase string, outputFilePrefix string) {
	logger.Infof("Writing json data to %s/%s_*.json", outputBase, outputFilePrefix)
	err := WriteToJSON(m.units, outputBase+"/"+outputFilePrefix+"_units.json")
	if err != nil {
		logger.Errorf("error writing units to JSON: %v", err)
	}
	err = WriteToJSON(m.objectives, outputBase+"/"+outputFilePrefix+"_objectives.json")
	if err != nil {
		logger.Errorf("error writing objectives to JSON: %v", err)
	}
	err = WriteToJSON(m.deltas, outputBase+"/"+outputFilePrefix+"_deltas.json")
	if err != nil {
		logger.Errorf("error writing deltas to JSON: %v", err)
	}

	objectiveTree := buildObjectiveTree(m.objectives)
	err = WriteToJSON(objectiveTree, outputBase+"/"+outputFilePrefix+"_objective_tree.json")
	if err != nil {
		logger.Errorf("error writing deltas to JSON: %v", err)
	}
}

func (m *MissionManager) loadClassTable() {
	logger.Infof("Loading ClassTable")
	if m.classTable != nil {
		return
	}
	records, err := LoadCTRecords(m.falconBase + "/Data/TerrData/Objects/Falcon4_CT.xml")
	if err != nil {
		log.Fatal(err)
	}
	m.classTable = CreateStrippedClassTable(records)
	logger.Infof("ClassTypes: %d", len(m.classTable))
}

func (m *MissionManager) initializeReaders() {
	m.unitReader = NewUnitReader(m.classTable)
	m.objectiveReader = NewObjectiveReader(m.classTable)
	m.deltaReader = NewObjectiveDeltaReader()
	m.campaignReader = NewCampaignReader()
}

func (m *MissionManager) loadBundle(filename string) *FileBundleReader {
	logger.Infof("Reading table of content of %s", filename)
	reader, err := NewFileBundleReaderFromFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	files, err := reader.GetEmbeddedFileDirectory()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		logger.Debugf("Name: %s, Offset: %d,  Size: %d", f.FileName, f.FileOffset, f.FileSizeBytes)
	}
	return reader
}

func (m *MissionManager) applyDeltas() {
	logger.Infof("Applying deltas to objectives")
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

func (m *MissionManager) GetObjectivesByTypes(objectTypes []int) []*Objective {
	result := make([]*Objective, 0)
	for _, obj := range m.objectives {
		if obj == nil || obj.ClassType == nil {
			continue
		}
		if Contains(objectTypes, obj.ClassType.Type) {
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
