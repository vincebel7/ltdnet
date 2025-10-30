package engine

import (
	"github.com/vincebel7/ltdnet/iphelper"
	"github.com/vincebel7/ltdnet/src/model"
)

func (e *Engine) RouteToHostInterface(host model.Host, dstIP string) (model.Interface, bool) {
	for iface := range host.Interfaces {
		devIP := host.GetIP(iface)
		devMask := host.GetMask(iface)

		if iphelper.IPInSameSubnet(devIP, dstIP, devMask) {
			return host.Interfaces[iface], true
		}
	}

	// Default gateway
	return host.Interfaces["eth0"], false
}
