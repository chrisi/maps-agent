# Maps Agent

A small web server that exposes a couple of local Falcon BMS mission/campaign info for the
Interactive Maps-app to connect to in order to simplify the mission planning.

The agent can directly communicate with the Falcon BMS IMCS server need for the Intarctive Maps-app to be running.
This iis useful, for example, when the Maps-app is not being used, but position data should still be provided to
the other members of the IMCS session.

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
  -path string
    	Falcon BMS directory (default "C:\\Progam Files\\Falcon BMS 4.38")
  -pos-update-freq string
    	Milliseconds between sending position updates (default "250")
```

### Start

```
maps-agent.exe --path="C:\apps\Falcon BMS 4.38" --callsign="Joker"
```
