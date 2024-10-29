/*
File:		switch.go
Author: 	https://github.com/vincebel7
Purpose:	Switch-specific functions
*/

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type Switch struct {
	ID       string              `json:"id"`
	Model    string              `json:"model"`
	Hostname string              `json:"hostname"`
	MgmtIP   net.IP              `json:"mgmtip"`
	MACTable map[string]MACEntry `json:"mactable"`
	Maxports int                 `json:"maxports"`
	Ports    []string            `json:"ports"`   // maps port # to downlink ID
	PortIDs  []string            `json:"portids"` // maps port # to Port ID
	ARPTable map[string]ARPEntry `json:"arptable"`
}

type MACEntry struct {
	Interface  int       `json:"interface"`
	State      string    `json:"state"`
	ExpireTime time.Time `json:"expireTime"`
}

func NewSumerian2100(hostname string) Switch {
	s := Switch{}
	s.ID = idgen(8)
	s.Model = "Sumerian 2100"
	s.Hostname = hostname
	s.Maxports = 4
	s.ARPTable = make(map[string]ARPEntry)

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

	s.MACTable = make(map[string]MACEntry)
	snet.Switches = append(snet.Switches, s)

	generateSwitchChannels(getSwitchIndexFromID(s.ID))
	for j := 0; j < getActivePorts(s); j++ {
		channels[s.PortIDs[j]] = make(chan json.RawMessage)
		socketMaps[s.PortIDs[j]] = make(map[string]chan Frame)
		actionsync[s.PortIDs[j]] = make(chan int)

		go listenSwitchportChannel(s.ID, s.PortIDs[j])
	}
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

	v.MACTable = make(map[string]MACEntry)

	return v
}

func delSwitch(hostname string) {
	//TODO For all linked devices, unlink. then delete
	hostname = strings.ToUpper(hostname)
	//search for switch
	for i := range snet.Switches {
		if strings.ToUpper(snet.Switches[i].Hostname) == hostname {
			// Unlink all devices connected to this switch
			for j := range snet.Switches[i].Ports {
				if snet.Switches[i].Ports[j] != "" {
					// Unlink if host
					for h := range snet.Hosts {
						if snet.Hosts[h].ID == snet.Switches[i].Ports[j] {
							snet.Hosts[h].UplinkID = ""
						}
					}

					// Unlink if switch - TODO
				}
			}

			snet.Switches = removeSwitchFromSlice(snet.Switches, i)
			fmt.Printf("\nSwitch deleted\n")
			return
		}
	}
	fmt.Printf("\nSwitch %s was not deleted.\n", hostname)
}

func linkSwitchTo(localDevice string, remoteDevice string) {
	localDevice = strings.ToUpper(localDevice)
	remoteDevice = strings.ToUpper(remoteDevice)

	//Make sure there's enough ports - if uplink device is a router
	if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
		if getActivePorts(snet.Router.VSwitch) >= snet.Router.VSwitch.Maxports {
			fmt.Printf("No available ports - %s only has %d ports\n", snet.Router.Model, snet.Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range snet.Switches {
		if remoteDevice == strings.ToUpper(snet.Switches[s].Hostname) {
			if getActivePorts(snet.Switches[s]) >= snet.Switches[s].Maxports {
				fmt.Printf("No available ports - %s only has %d ports\n", snet.Switches[s].Model, snet.Switches[s].Maxports)
				return
			}
		}
	}

	//find switch with that hostname
	for i := range snet.Switches {
		if strings.ToUpper(snet.Switches[i].Hostname) == localDevice {
			uplinkID := ""
			//Remote device on new link is the Router
			if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
				//find next free port
				for k := range snet.Router.VSwitch.Ports {
					if (snet.Router.VSwitch.Ports[k] == "") && (uplinkID == "") {
						uplinkID = snet.Router.VSwitch.PortIDs[k]
					}
				}
				//uplinkID = snet.Router.VSwitch.ID

				// Assign switchport on remote device
				assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].ID)
			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range snet.Switches {
					if remoteDevice == strings.ToUpper(snet.Switches[j].Hostname) {

						//find next free port
						for k := range snet.Switches[j].Ports {
							if (snet.Switches[j].Ports[k] == "") && (uplinkID == "") {
								uplinkID = snet.Switches[j].PortIDs[k]

								// Assign switchport on remote device
								assignSwitchport(snet.Switches[j], snet.Switches[i].ID)
							}
						}

					}
				}
			}

			// Assign switchport on local switch
			assignSwitchport(snet.Switches[i], "TEST")

			return
		}
	}
}

func lookupMACTable(dstMAC string, switchportID string) int { // For looking up addresses
	resultPort := -1
	var MACTable map[string]MACEntry

	if isSwitchportID(snet.Router.VSwitch, switchportID) {
		MACTable = snet.Router.VSwitch.MACTable
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], switchportID) {
				MACTable = snet.Switches[i].MACTable
			}
		}
	}

	for k := range MACTable {
		if k == dstMAC {
			resultPort = MACTable[k].Interface
		}
	}

	return resultPort
}

func checkMACTable(macaddr string, id string, port int) { // For updating MAC table on incoming frames
	result := 0
	table := make(map[string]MACEntry)
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
			if v.Interface == port {
				debug(4, "checkMACTable", id, "Source address found in MAC table")
				result = v.Interface
			} else {
				delMACEntry(macaddr, id, port)
			}
		}
	}

	if result == 0 {
		msg := "Source address " + macaddr + " not found in MAC table. Adding"
		debug(3, "learnMACTable", id, msg)
		addMACEntry(macaddr, id, port)
	}
}

func addMACEntry(macaddr string, id string, port int) {
	if isSwitchportID(snet.Router.VSwitch, id) {
		macEntry := MACEntry{
			Interface: port,
		}
		snet.Router.VSwitch.MACTable[macaddr] = macEntry
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], id) {
				macEntry := MACEntry{
					Interface: port,
				}
				snet.Switches[i].MACTable[macaddr] = macEntry
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

	channels[sw.PortIDs[portIndex]] = make(chan json.RawMessage)
	debug(4, "assignSwitchport", sw.PortIDs[portIndex], "listening for id")
	go listenSwitchportChannel(sw.ID, sw.PortIDs[portIndex])

	return sw
}

func switchforward(frame Frame, switchID string, switchportID string) {
	srcMAC := frame.SrcMAC
	dstMAC := frame.DstMAC
	linkID := ""
	floodFrame := false

	outboundPort := lookupMACTable(dstMAC, switchportID)

	if dstMAC == "ff:ff:ff:ff:ff:ff" { // Broadcast
		floodFrame = true
		debug(4, "switchforward", switchID, "L2 Broadcast. Flooding frame on all ports")
	} else if outboundPort == -1 { // No matching port for this MAC address was found in the MAC address table
		floodFrame = true
		debug(4, "switchforward", switchID, "Destination address not found in MAC table. Flooding frame on all ports")
	} else {
		if isSwitchportID(snet.Router.VSwitch, switchportID) { // VSwitch
			debug(4, "switchforward", switchID, "Destination address found in MAC table.")
			linkID = snet.Router.VSwitch.Ports[outboundPort]
		} else { // Regular switch
			for i := range snet.Switches {
				if isSwitchportID(snet.Switches[i], switchportID) {
					debug(4, "switchforward", switchID, "Destination address found in MAC table.")
					linkID = snet.Switches[i].Ports[outboundPort]
				}
			}
		}
	}

	p := frame.Data
	f := Frame{
		SrcMAC:    srcMAC,
		DstMAC:    dstMAC,
		EtherType: frame.EtherType,
		Data:      p,
	}
	outFrame, _ := json.Marshal(f)

	if floodFrame {
		if isSwitchportID(snet.Router.VSwitch, switchportID) { // VSwitch
			for port := range snet.Router.VSwitch.Ports {
				linkID = snet.Router.VSwitch.Ports[port]
				if (snet.Router.VSwitch.PortIDs[port] != switchportID) && (snet.Router.VSwitch.PortIDs[port] != "") { // Don't send out source interface
					channels[linkID] <- outFrame
				}
			}
		} else { // Regular switch
			switchIndex := getSwitchIndexFromID(switchID)
			for port := range snet.Switches[switchIndex].Ports {
				linkID = snet.Switches[switchIndex].Ports[port]
				if (snet.Switches[switchIndex].PortIDs[port] != switchportID) && (snet.Switches[switchIndex].PortIDs[port] != "") { // Don't send out source interface
					channels[linkID] <- outFrame
				}
			}
		}
	} else {
		channels[linkID] <- outFrame
	}
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
