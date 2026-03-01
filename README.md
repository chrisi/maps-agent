# Maps Agent

A small web server that exposes a couple of local Falcon BMS mission/campaign info for the
Interactive Maps-app to connect to in order to simplify the mission planning.


## Build
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
  -path string
        Falcon BMS directory (default "C:\\Progam Files\\Falcon BMS 4.38")
  -pos-update-freq string
        Milliseconds between sending position updates (default "250")
```
### Start
```
maps-agent.exe --path="C:\apps\Falcon BMS 4.38" --callsign="Joker"
```
