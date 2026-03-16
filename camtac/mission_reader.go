package camtac

import (
	"log"
	"maps-agent/util"
	"os"
	"path/filepath"
	"strings"
)

var logBdRed = util.NewLogger("FileBundleReader", os.Stdout, util.Info, true)

func ReadMission(falconBase string, missionFilename string, outputBase string) {

	logBdRed.Infof("Reading mission %s", missionFilename)

	campaignBase := falconBase + "/Data/Campaign"
	ctFile := falconBase + "/Data/TerrData/Objects/Falcon4_CT.xml"

	logBdRed.Infof("Creating ClassTable")
	records, err := LoadCTRecords(ctFile)
	if err != nil {
		log.Fatal(err)
	}
	cts := CreateClassTable(records)
	logBdRed.Debugf("ClassTypes: %d", len(cts))

	logBdRed.Infof("Reading table of content")
	bundleReader, err := NewFileBundleReaderFromFile(campaignBase + "/" + missionFilename)
	if err != nil {
		log.Fatal(err)
	}

	files, err := bundleReader.GetEmbeddedFileDirectory()
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		logBdRed.Debugf("Name: %s, Offset: %d,  Size: %d", f.FileName, f.FileOffset, f.FileSizeBytes)
	}

	fileNoExt := strings.TrimSuffix(missionFilename, filepath.Ext(missionFilename))
	data, err := bundleReader.GetEmbeddedFileContents(fileNoExt + ".uni")
	if err != nil {
		log.Fatal(err)
	}

	unitReader := NewUnitReader(cts)
	units := unitReader.ReadUniFile(data)

	unitCounts := unitReader.Counts()
	logBdRed.Infof("Num Units:       %d", unitCounts.NumUnits)
	logBdRed.Infof("Num Squadrons:   %d", unitCounts.NumSquadrons)
	logBdRed.Infof("Num Packages:    %d", unitCounts.NumPackages)
	logBdRed.Infof("Num Flights:     %d", unitCounts.NumFlights)
	logBdRed.Infof("Num Brigades:    %d", unitCounts.NumBrigades)
	logBdRed.Infof("Num Battalions:  %d", unitCounts.NumBattalions)
	logBdRed.Infof("Num Task Forces: %d", unitCounts.NumTaskForces)

	err = WriteUnitsToJSON(units, outputBase+"/"+fileNoExt+"_units.json")
	if err != nil {
		logBdRed.Errorf("error writing units to JSON: %v", err)
	}
}
