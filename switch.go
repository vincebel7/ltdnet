/*
File:		switch.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Switch-specific functions
*/

package main

import (
	"fmt"
	"strings"
)

type Switch struct {
	ID       string         `json:"id"`
	Model    string         `json:"model"`
	Hostname string         `json:"hostname"`
	MgmtIP   string         `json:"mgmtip"`
	MACTable map[string]int `json:"mactable"`
	Maxports int            `json:"maxports"`
	Ports    []string       `json:"ports"`    // maps port # to downlink ID
	PortIDs  []string       `json:"portids"`  // maps port # to Port ID
	PortMACs []string       `json:"portmacs"` // maps port # to interface MAC address
}

func NewSumerian2100(hostname string) Switch {
	s := Switch{}
	s.ID = idgen(8)
	s.Model = "Sumerian 2100"
	s.Hostname = hostname
	s.Maxports = 4

	return s
}

func addSwitch(switchHostname string) {
	switchModel := strings.ToUpper("Sumerian")

	// input validation
	if hostname_exists(switchHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	s := Switch{}
	if switchModel == "SUMERIAN" {
		s = NewSumerian2100(switchHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	s.PortIDs = make([]string, s.Maxports)
	for i := range s.PortIDs {
		s.PortIDs[i] = idgen(8)
	}

	s.Ports = make([]string, s.Maxports)
	for i := range s.Ports {
		s.Ports[i] = ""
	}

	s.MACTable = make(map[string]int)
	snet.Switches = append(snet.Switches, s)

	generateSwitchChannels(getSwitchIndexFromID(s.ID))
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

func delSwitch() {
	//TODO For all linked devices, unlink. then delete
}

func lookupMACTable(macaddr string, id string) int { // For looking up addresses
	resultPort := 0
	table := make(map[string]int)
	if isSwitchportID(snet.Router.VSwitch, id) {
		table = snet.Router.VSwitch.MACTable
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], id) {
				table = snet.Switches[i].MACTable
			}
		}
	}

	//if switch
	for k, v := range table {
		if k == macaddr {
			resultPort = v
		}
	}

	return resultPort
}

func checkMACTable(macaddr string, id string, port int) { // For checking table on incoming frames
	result := -1
	table := make(map[string]int)
	if isSwitchportID(snet.Router.VSwitch, id) {
		table = snet.Router.VSwitch.MACTable
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], id) {
				table = snet.Switches[i].MACTable
			}
		}
	}

	for k, v := range table {
		if k == macaddr {
			if v == port {
				debug(4, "lookupMACTable", id, "Address found in MAC table")
				result = v
			} else {
				delMACEntry(macaddr, id, port)
			}
		}
	}

	if result == -1 {
		msg := "Address " + macaddr + " not found in MAC table. Adding"
		debug(3, "learnMACTable", id, msg)
		addMACEntry(macaddr, id, port)
	}
}

func addMACEntry(macaddr string, id string, port int) {
	if isSwitchportID(snet.Router.VSwitch, id) {
		snet.Router.VSwitch.MACTable[macaddr] = port
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], id) {
				snet.Switches[i].MACTable[macaddr] = port
			}
		}
	}

}

func delMACEntry(macaddr string, id string, port int) {
}

func isSwitchportID(sw Switch, id string) bool {
	for i := range sw.PortIDs {
		if sw.PortIDs[i] == id {
			return true
		}
	}

	return false
}

func getActivePorts(sw Switch) int {
	count := 0

	for i := range sw.Ports {
		if sw.Ports[i] != "" {
			count++
		}
	}

	return count
}

func assignSwitchport(sw Switch, id string) Switch {
	portIndex := -1
	for i := range sw.Ports {
		if sw.Ports[i] == "" {
			sw.Ports[i] = id
			portIndex = i
			break
		}
	}

	channels[sw.PortIDs[portIndex]] = make(chan Frame)
	internal[sw.PortIDs[portIndex]] = make(chan Frame)
	debug(4, "generateRouterChannels", sw.PortIDs[portIndex], "listening for id")
	go switchportlisten(sw.PortIDs[portIndex])

	return sw
}

func switchforward(frame Frame, id string) {
	srcIP := frame.Data.SrcIP
	dstIP := frame.Data.DstIP
	srcMAC := frame.SrcMAC
	dstMAC := frame.DstMAC
	linkID := ""

	outboundPort := lookupMACTable(dstMAC, id)

	if outboundPort == -1 {
		debug(1, "switchforward", id, "Warning: Not found in MAC table, using bypass") //TODO implement flooding
		linkID = getIDfromMAC(dstMAC)
	} else {
		if isSwitchportID(snet.Router.VSwitch, id) {
			linkID = snet.Router.VSwitch.Ports[outboundPort]
		} else {
			for i := range snet.Switches {
				fmt.Println("Should never print this yet")
				if isSwitchportID(snet.Switches[i], id) {
					linkID = snet.Switches[i].Ports[outboundPort]
				}
			}
		}
	}

	s := frame.Data.Data
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID] <- f
}

func freeSwitchport(link string) {

	switchport := getSwitchportIDFromLink(link)
	switchID := getSwitchIDFromLink(link)

	if snet.Router.VSwitch.ID == switchID {
		snet.Router.VSwitch.Ports[switchport] = ""
	} else {
		i := getSwitchIndexFromID(switchID)
		snet.Switches[i].Ports[switchport] = ""
	}
}
