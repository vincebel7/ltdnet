package model

type Network struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Netsize    string   `json:"netsize"`
	Router     *Router  `json:"router"`
	Switches   []Switch `json:"switches"`
	Hosts      []Host   `json:"hosts"`
	ProgramVer string   `json:"program_ver"`
}

func (n *Network) ClearMACTables() {
	// Router ARP
	if n.Router != nil && n.Router.ARPTable != nil {
		for k := range n.Router.ARPTable {
			delete(n.Router.ARPTable, k)
		}
	}
	// Hosts ARP
	for i := range n.Hosts {
		if n.Hosts[i].ARPTable != nil {
			for k := range n.Hosts[i].ARPTable {
				delete(n.Hosts[i].ARPTable, k)
			}
		}
	}
	// Switch MAC address tables
	for i := range n.Switches {
		n.Switches[i].MACTable = make(map[string]MACEntry)
	}
	// VSwitch MAC table only if router exists
	if n.Router != nil {
		if n.Router.VSwitch.MACTable == nil {
			n.Router.VSwitch.MACTable = make(map[string]MACEntry)
		} else {
			for k := range n.Router.VSwitch.MACTable {
				delete(n.Router.VSwitch.MACTable, k)
			}
		}
	}
}
