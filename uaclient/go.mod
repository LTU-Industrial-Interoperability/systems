module github.com/sdoque/systems/uaclient

go 1.25.5

require (
	github.com/gopcua/opcua v0.6.5
	github.com/pkg/errors v0.9.1
	github.com/sdoque/mbaigo v0.1.0-alpha.1.0.20251208114144-6cc6f78035ae
	github.com/sdoque/systems/certgen v0.0.0
)

replace github.com/sdoque/systems/certgen => ../certgen
