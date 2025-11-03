package engine

import (
	"github.com/vincebel7/ltdnet/src/model"
)

func NewSumerian2100(hostname string) model.Switch {
	s := model.Switch{}
	s.ID = IDgen(8)
	s.Model = "Sumerian 2100"
	s.Hostname = hostname
	s.Maxports = 4
	s.ARPTable = make(map[string]model.ARPEntry)

	return s
}

func AddVirtualSwitch(maxports int) model.Switch {
	v := model.Switch{}
	v.ID = IDgen(8)
	v.Model = "virtual"
	v.Hostname = "V-" + v.ID
	v.Maxports = maxports

	v.PortLinksLocal = make([]string, v.Maxports)
	for i := range v.PortLinksLocal {
		v.PortLinksLocal[i] = IDgen(8)
	}

	v.PortLinksRemote = make([]string, v.Maxports)
	for i := range v.PortLinksRemote {
		v.PortLinksRemote[i] = ""
	}

	v.MACTable = make(map[string]model.MACEntry)

	return v
}
