# Maps Agent

A small web server that exposes a couple of local Falcon BMS mission/campaign info for the
Interactive Maps-app to connect to in order to simplify the mission planning.


## Build

```
go build -o c:\apps\maps-agent.exe maps-agent
```

## Usage

```
maps-agent.exe --path="C:\apps\Falcon BMS 4.38" --callsign="Joker"
```