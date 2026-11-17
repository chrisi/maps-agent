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
