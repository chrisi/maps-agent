package main

import (
	"flag"
	"strconv"
)

type Config struct {
	FalconBase      string
	Callsign        string
	ServerAddr      string
	Imcs            bool
	ImcsServer      string
	ImcsSession     string
	ReadSharedMem   bool
	PosUpdateFreq   int
	FsCheckFreq     int
	MissionFilename string
	ExportJsonDir   string
}

func parseConfig() Config {
	defaultFalconBase := "C:/Progam Files/Falcon BMS 4.38"
	defaultCallsign := "Viper"
	defaultSession := "47DF"
	defaultServerAddr := ":8080"
	defaultImcsServer := "wss://collab.falcon-bms.com:443"

	falconBase := flag.String("path", defaultFalconBase, "Falcon BMS directory")
	imcs := flag.Bool("imcs", false, "Establish IMCS connection")
	imcsServer := flag.String("imcs-server", defaultImcsServer, "IMCS Server the client should connect to")
	imcsSession := flag.String("imcs-session", defaultSession, "IMCS Session the client should connect to")
	callsign := flag.String("callsign", defaultCallsign, "Callsign")
	readSharedMem := flag.Bool("shared-mem", false, "Read Falcon BMS Shared memory and broadcast them")
	fsCheckFreqStr := flag.String("fs-check-freq", "1000", "Milliseconds between checking file system for changes")
	posUpdateFreqStr := flag.String("pos-update-freq", "250", "Milliseconds between sending position updates")
	serverAddr := flag.String("addr", defaultServerAddr, "HTTP service address")
	missionFilename := flag.String("mission-load", "", "Read CAM/TAC file")
	exportJsonDir := flag.String("mission-export", "", "Export mission data to directory")

	fsCheckFreq, _ := strconv.Atoi(*fsCheckFreqStr)
	posUpdateFreq, _ := strconv.Atoi(*posUpdateFreqStr)

	flag.Parse()

	return Config{
		FalconBase:      *falconBase,
		Callsign:        *callsign,
		ServerAddr:      *serverAddr,
		Imcs:            *imcs,
		ImcsServer:      *imcsServer,
		ImcsSession:     *imcsSession,
		ReadSharedMem:   *readSharedMem,
		PosUpdateFreq:   posUpdateFreq,
		FsCheckFreq:     fsCheckFreq,
		MissionFilename: *missionFilename,
		ExportJsonDir:   *exportJsonDir,
	}
}
