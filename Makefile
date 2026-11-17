# Cross-compiles the agent for Windows (also runs in Wine/Proton next to Falcon BMS).
# The agent is pure Go, so no C toolchain is needed.

BIN_DIR := bin
EXE     := $(BIN_DIR)/maps-agent.exe

.PHONY: all windows test vet layout clean

all: windows

windows:
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o $(EXE) .

test:
	go test ./...

vet:
	go vet ./...
	GOOS=windows go vet ./...

# Regenerate bms/testdata/layout.txt after updating include/FlightData.h (needs g++).
layout:
	python3 bms/testdata/gen_layout.py
	go test ./bms

clean:
	rm -rf $(BIN_DIR)
