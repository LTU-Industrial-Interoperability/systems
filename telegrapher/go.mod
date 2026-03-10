module github.com/sdoque/telegrapher

go 1.25.5

require (
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/sdoque/mbaigo v0.1.0-alpha.1.0.20251208114144-6cc6f78035ae
	github.com/sdoque/systems/certgen v0.0.0
)

replace github.com/sdoque/systems/certgen => ../certgen

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
)
