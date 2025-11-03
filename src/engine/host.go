package engine

import (
	"github.com/vincebel7/ltdnet/iphelper"
	"github.com/vincebel7/ltdnet/src/model"
)

// Populate fields specific to the Probox 1
func NewProbox(h model.Host) model.Host {
	h.Model = "ProBox 1"
	return h
}

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
