/*
File:		router.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Router-specific functions
*/

package main

import (
	"fmt"
	"strconv"
	"strings"
)

const BOBCAT_PORTS = 4
const OSIRIS_PORTS = 2

func NewBobcat(hostname string) Router {
	b := Router{}
	b.ID = idgen(8)
	b.Model = "Bobcat 100"
	b.MACAddr = macgen()
	b.Hostname = hostname
	b.DHCPPool = 253

	v := addVirtualSwitch(BOBCAT_PORTS)

	b.VSwitch = v
	return b
}

func NewOsiris(hostname string) Router {
	o := Router{}
	o.ID = idgen(8)
	o.Model = "Osiris 2-I"
	o.MACAddr = macgen()
	o.Hostname = hostname
	o.DHCPPool = 2

	v := addVirtualSwitch(OSIRIS_PORTS)

	o.VSwitch = v
	return o
}

func addVirtualSwitch(maxports int) Switch {
	v := Switch{}
	v.ID = idgen(8)
	v.Model = "virtual"
	v.Hostname = "V-" + v.ID
	v.Maxports = maxports

	v.PortIDs = make([]string, v.Maxports)
	for i := range v.PortIDs {
		v.PortIDs[i] = idgen(8)
	}

	v.Ports = make([]string, v.Maxports)
	for i := range v.Ports {
		v.Ports[i] = ""
	}

	v.MACTable = make(map[string]int)

	return v
}

func addRouter(routerHostname string, routerModel string) {
	routerModel = strings.ToUpper(routerModel)

	// input validation
	if hostname_exists(routerHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	r := Router{}

	if routerModel == "BOBCAT" {
		r = NewBobcat(routerHostname)
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(routerHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	if snet.Netsize == "8" {
		r.Gateway = "10.0.0.1"
	} else if snet.Netsize == "16" {
		r.Gateway = "172.16.0.1"
	} else if snet.Netsize == "24" {
		r.Gateway = "192.168.0.1"
	}
	addrconstruct := ""

	network_portion := strings.TrimSuffix(r.Gateway, "1")

	r.DHCPTable = make(map[string]string)
	r.DHCPTableOrder = make([]string, r.DHCPPool)

	for i := 2; i < (r.DHCPPool + 2); i++ {
		addrconstruct = network_portion + strconv.Itoa(i)
		r.DHCPTable[addrconstruct] = ""
		r.DHCPTableOrder[i-2] = addrconstruct
	}

	snet.Router = r

	generateRouterChannels()

	assignSwitchport(snet.Router.VSwitch, snet.Router.ID)
}

func delRouter(hostname string) {
	r := Router{}

	r.ID = ""
	r.Model = ""
	r.MACAddr = ""
	r.Hostname = ""
	r.DHCPPool = 0
	//r.Downports = 0
	//r.Ports = nil
	r.VSwitch = addVirtualSwitch(0)

	snet.Router = r
	fmt.Printf("\nRouter deleted\n")
}

func next_free_addr() string {
	for _, v := range snet.Router.DHCPTableOrder {
		if snet.Router.DHCPTable[v] == "" {
			return v
		}
	}

	return ""
}
