# Maps Agent

A small web server that exposes a couple of local Falcon BMS mission/campaign info for the
Interactive Maps-app to connect to in order to simplify the mission planning.

The agent can directly communicate with the Falcon BMS IMCS server without the need for the
Intarctive Maps to be running. This iis useful, for example, when the Maps-app is not being used,
but position data should still be provided to the other members of the IMCS session.

## Build

Download the source, goto the main project directory and execute the following command

```
go build -o c:\apps\maps-agent.exe maps-agent
```

The agent is pure Go (no cgo), so no C compiler is needed.

### Cross-compile on Linux (e.g. to run it in Wine/Proton next to Falcon BMS)

```
make            # builds bin/maps-agent.exe
```

or without make: `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o maps-agent.exe .`
Other targets: `make test`, `make vet`, `make layout` (after updating `FlightData.h`), `make clean`.

The result is a static binary that only uses standard Windows system DLLs, which Wine provides.

To read the shared memory, the agent has to run in the **same Wine prefix** and with the **same
Wine/Proton build** as Falcon BMS. Named shared memory only exists inside the wineserver of a
prefix; a different prefix or Wine version starts its own wineserver and won't find it. For example:

```
# plain Wine / Lutris
WINEPREFIX=/path/to/bms-prefix /path/to/the/same/wine maps-agent.exe --shared-mem ...

# Proton (BMS added to Steam as a non-Steam game)
protontricks-launch --appid <appid> maps-agent.exe --shared-mem ...
```

### Fallback: build with cgo and MinGW

Older versions of the agent read the BMS structs through cgo. If cgo is ever needed again, a cgo
build for Windows needs a Windows C toolchain, because cross-compiling disables cgo by default.
On Linux, install MinGW-w64 (e.g. `apt install gcc-mingw-w64-x86-64`) and run:

```
CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 \
  go build -ldflags '-extldflags "-static"' -o maps-agent.exe .
```

`-extldflags "-static"` links the MinGW runtime (`libgcc`, `libwinpthread`) into the exe, so no
extra MinGW DLLs need to be next to it in Wine. Check with
`x86_64-w64-mingw32-objdump -p maps-agent.exe | grep 'DLL Name'`; only Windows system DLLs should be
listed.

## Usage

```
Usage of maps-agent.exe:
  -addr string
    	HTTP service address (default ":8080")
  -callsign string
    	Callsign (default "Viper")
  -fs-check-freq string
    	Milliseconds between checking file system for changes (default "1000")
  -imcs
    	Establish IMCS connection
  -imcs-server string
    	IMCS Server the client should connect to (default "wss://collab.falcon-bms.com:443")
  -imcs-session string
    	IMCS Session the client should connect to (default "47DF")
  -mission-export string
    	Export mission data to directory
  -mission-load string
    	Read CAM/TAC file
  -path string
    	Falcon BMS directory (default "C:/Progam Files/Falcon BMS 4.38")
  -pos-update-freq string
    	Milliseconds between sending position updates (default "250")
  -shared-mem
    	Read Falcon BMS Shared memory and broadcast them
```

### Start

```
maps-agent.exe --path="C:\apps\Falcon BMS 4.38" --callsign="Joker" --mission-load="mc-test.cam"
```

## Flight data

The agent watches two files that Falcon BMS writes while you plan a flight:

- `User\Config\<callsign>.ini`: the flight configuration (steerpoints, loadout, etc.)
- `User\Briefings\briefing.txt`: the mission briefing

BMS writes these files at different times, so reading them separately can give you a new INI and an old
briefing, or the other way round. The agent pairs them first and serves them together from one endpoint.

### Endpoint

`GET /flight` returns the current flight as JSON. If no paired flight is available yet, it returns
`404`.

```json
{
  "callsign": "Joker",
  "ini": "<content of Joker.ini>",
  "briefing": "<content of briefing.txt>",
  "updatedAt": "2026-09-03T12:54:53Z"
}
```

When a new flight becomes available, the agent sends this message to all websocket clients:

```json
{
  "type": "update",
  "payload": {
    "status": "ready",
    "endpoint": "/flight",
    "message": "Current mission is ready for download"
  }
}
```

### How INI and briefing are paired

The file watcher debounces changes to each file separately, using `-fs-check-freq`. Once a file has
stopped changing, it is handled as follows.

**INI saved**

- No briefing yet: store the INI and wait for the briefing.
- Briefing present: pair the INI with the briefing, update the current flight and send `ready`.
  This also covers saving the INI again for the same mission.

**Briefing saved**

- First briefing, no INI yet: store the briefing and wait for the INI.
- First briefing, INI present: pair them, update the current flight and send `ready`.
- Same as the stored briefing: ignore it. The second line of `briefing.txt` holds the time the
  briefing was generated, so the comparison skips that line and also ignores line-ending differences.
  This way, generating the same briefing again does not count as a new mission.
- Different from the stored briefing: this means a new mission. Store the new briefing, discard the
  stored INI and the current flight, and wait for the matching INI. `/flight` returns `404` until it
  arrives.

This means `/flight` only ever returns an INI and a briefing that belong to the same mission.

## Falcon BMS shared memory

The package `bms` reads the shared memory areas of Falcon BMS (enabled with `-shared-mem`):

| Area                            | Go type           | Content                                            |
|---------------------------------|-------------------|----------------------------------------------------|
| `FalconSharedMemoryArea`        | `bms.FlightData`  | position, flight values, lamps, DED/PFL, RWR, ...  |
| `FalconSharedMemoryArea2`       | `bms.FlightData2` | second engine, altimeter, MP pilots, lat/lon, ...  |
| `FalconSharedOsbMemoryArea`     | `bms.OSBData`     | MFD button labels                                  |
| `FalconSharedMemoryAreaString`  | `bms.StringData`  | BMS/theater directories, aircraft, nav points, ... |
| `FalconSharedMemoryAreaDrawing` | `bms.DrawingData` | HUD/RWR/HMS drawing commands                       |

The Go structs mirror the C++ classes in `include/FlightData.h` byte by byte. Every getter returns
a copy, so the data can't change while you're using it. If an older BMS version provides a shorter
area, the fields it doesn't have stay zero.

### Updating to a new FlightData.h

1. Replace `include/FlightData.h` with the new version from Falcon BMS.
2. Regenerate the reference layout (needs `g++`): `python3 bms/testdata/gen_layout.py`
3. Run `go test ./bms`. The test compares every field's name, offset and size, and the total struct
   size, with the header and reports any differences.
4. Add or adjust the fields in `bms/flightdata.go` / `bms/flightdata2.go` until the test passes, and
   update the bit constants and version numbers.
