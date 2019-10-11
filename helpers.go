/*
File:		helpers.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Various misc helper functions
*/

package main

import(
	"fmt"
	"math/rand"
	"time"
)

func idgen(n int) string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, n)

	rand.Seed(time.Now().UnixNano())
	for i := range id {
		id[i] = idchars[rand.Intn(len(idchars))]
	}

	return string(id)
}

func macgen() string {
	mac := idgen(2)
	for n := 0; n < 5; n++ {
		mac = mac + ":" + idgen(2)
	}

	return mac
}

func getDeviceType(id string) string {
	if(snet.Router.ID == id) { return "router" }
	if(snet.Router.VSwitch.ID == id) { return "vswitch" }

	for s := range snet.Switches {
		if(snet.Switches[s].ID == id) { return "switch" }
	}

	return "host"
}

func getHostIndexFromID(id string) int {
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id { return h }
	}
	return -1
}

func getSwitchIndexFromID(id string) int {
	for s := range snet.Switches {
		if snet.Switches[s].ID == id { return s }
	}

	return -1
}

func getMACfromID(id string) string {
	//Router
	if id == snet.Router.ID { return snet.Router.MACAddr }

	//Hosts
	return snet.Hosts[getHostIndexFromID(id)].MACAddr
}

func getIDfromMAC(mac string) string {
	//Router
	if mac == snet.Router.MACAddr { return snet.Router.ID }

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].MACAddr == mac { return snet.Hosts[h].ID }
	}

	return ""
}

func dynamic_assign(id string, ipaddr string, defaultgateway string, subnetmask string) {
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id {
			snet.Hosts[h].IPAddr = ipaddr
			snet.Hosts[h].SubnetMask = subnetmask
			snet.Hosts[h].DefaultGateway = defaultgateway
			fmt.Println("Network configuration updated")
		}
	}

}

func hostname_exists(hostname string) bool {
	if snet.Router.Hostname == hostname { return true }
	if snet.Router.VSwitch.Hostname == hostname { return true }

	for s := range snet.Switches {
		if snet.Switches[s].Hostname == hostname { return true }
	}

	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname { return true }
	}

	return false
}

func removeHostFromSlice(s []Host, i int) []Host {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeStringFromSlice(s []string, i int) []string {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
