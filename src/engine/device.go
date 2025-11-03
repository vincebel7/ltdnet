package engine

import "github.com/vincebel7/ltdnet/src/model"

type Device interface {
	GetID() string
	GetHostname() string
}

// Compile-time assertions
var (
	_ Device = (*model.Host)(nil)
	_ Device = (*model.Router)(nil)
	_ Device = (*model.Switch)(nil)
)

// Returns all devices in the network
func (e *Engine) Devices() []Device {
	if e.Net == nil {
		return nil
	}
	devs := make([]Device, 0, len(e.Net.Hosts)+len(e.Net.Switches)+1)
	if e.Net.Router != nil {
		devs = append(devs, e.Net.Router)
	}
	for i := range e.Net.Switches {
		devs = append(devs, &e.Net.Switches[i])
	}
	for i := range e.Net.Hosts {
		devs = append(devs, &e.Net.Hosts[i])
	}
	return devs
}
