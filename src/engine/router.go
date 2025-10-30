package engine

import "github.com/vincebel7/ltdnet/src/model"

func (e *Engine) RouteToRouterInterface(deviceID, dstIP string) (model.Interface, bool) {
	if e.Net.Router.ID == deviceID {
		return e.Net.Router.Interfaces["eth0"], true
	}
	for i := range e.Net.Hosts {
		if e.Net.Hosts[i].ID == deviceID {
			return e.Net.Hosts[i].Interfaces["eth0"], true
		}
	}
	return model.Interface{}, false
}
