package main

import(
	//"fmt"
	//"time"
	//"strings"
	//"strconv"
)

func getMACfromID(id string) string {
	//Router
	if id == snet.Router.ID {
		return snet.Router.MACAddr
	}

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id {
			return snet.Hosts[h].MACAddr
		}
	}
	return ""
}

func getIDfromMAC(mac string) string {
	//Router
	if mac == snet.Router.MACAddr {
		return snet.Router.ID
	}

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].MACAddr == mac {
			return snet.Hosts[h].ID
		}
	}
	return ""
}
