/*
File:		switch.go
Author: 	https://github.com/vincebel7
Purpose:	Switch-specific functions
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Switch struct {
	ID              string              `json:"id"`
	Model           string              `json:"model"`
	Hostname        string              `json:"hostname"`
	MACTable        map[string]MACEntry `json:"mactable"`
	Maxports        int                 `json:"maxports"`
	PortLinksRemote []string            `json:"links_remote"` // maps port # to remote link ID
	PortLinksLocal  []string            `json:"links_local"`  // maps port # to local link ID
	ARPTable        map[string]ARPEntry `json:"arptable"`
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

	s.PortLinksLocal = make([]string, s.Maxports)
	for i := range s.PortLinksLocal {
		s.PortLinksLocal[i] = idgen(8)
	}

	s.PortLinksRemote = make([]string, s.Maxports)
	for i := range s.PortLinksRemote {
		s.PortLinksRemote[i] = ""
	}

	s.MACTable = make(map[string]MACEntry)
	snet.Switches = append(snet.Switches, s)

	generateSwitchChannels(getSwitchIndexFromID(s.ID))
	for j := 0; j < getActivePorts(s); j++ {
		channels[s.PortLinksLocal[j]] = make(chan json.RawMessage)
		socketMaps[s.PortLinksLocal[j]] = make(map[string]chan Frame)
		actionsync[s.PortLinksLocal[j]] = make(chan int)

		go listenSwitchportChannel(s.ID, s.PortLinksLocal[j])
	}
}

func addVirtualSwitch(maxports int) Switch {
	v := Switch{}
	v.ID = idgen(8)
	v.Model = "virtual"
	v.Hostname = "V-" + v.ID
	v.Maxports = maxports

	v.PortLinksLocal = make([]string, v.Maxports)
	for i := range v.PortLinksLocal {
		v.PortLinksLocal[i] = idgen(8)
	}

	v.PortLinksRemote = make([]string, v.Maxports)
	for i := range v.PortLinksRemote {
		v.PortLinksRemote[i] = ""
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
			for j := range snet.Switches[i].PortLinksLocal {
				if snet.Switches[i].PortLinksRemote[j] != "" {
					// Unlink if host
					for h := range snet.Hosts {
						if snet.Hosts[h].Interfaces["eth0"].RemoteL1ID == snet.Switches[i].PortLinksLocal[j] {
							iface := snet.Hosts[h].Interfaces["eth0"]
							iface.RemoteL1ID = ""
							snet.Hosts[h].Interfaces["eth0"] = iface
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
				for k := range snet.Router.VSwitch.PortLinksLocal {
					if (snet.Router.VSwitch.PortLinksRemote[k] == "") && (uplinkID == "") {
						uplinkID = snet.Router.VSwitch.PortLinksLocal[k]
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
						for k := range snet.Switches[j].PortLinksLocal {
							if (snet.Switches[j].PortLinksRemote[k] == "") && (uplinkID == "") {
								uplinkID = snet.Switches[j].PortLinksLocal[k]

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
	result := -1
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
				debug(4, "checkMACTable", id, "Source address found in MAC table, but wrong - removing old.")
				delMACEntry(macaddr, id, port)
			}
		}
	}

	if result == -1 {
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
	for i := range sw.PortLinksLocal {
		if sw.PortLinksLocal[i] == id {
			return true
		}
	}

	return false
}

func getActivePorts(sw Switch) int {
	count := 0

	for i := range sw.PortLinksRemote {
		if sw.PortLinksRemote[i] != "" {
			count++
		}
	}

	return count
}

func assignSwitchport(sw Switch, id string) int {
	portIndex := -1
	for i := range sw.PortLinksRemote {
		if sw.PortLinksRemote[i] == "" {
			sw.PortLinksRemote[i] = id
			portIndex = i
			break
		}
	}

	channels[sw.PortLinksLocal[portIndex]] = make(chan json.RawMessage)
	debug(4, "assignSwitchport", sw.PortLinksLocal[portIndex], "listening for id")
	go listenSwitchportChannel(sw.ID, sw.PortLinksLocal[portIndex])

	return portIndex
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
		debug(4, "switchforward", switchID, "Destination address "+dstMAC+" not found in MAC table. Flooding frame on all ports")
	} else {
		if isSwitchportID(snet.Router.VSwitch, switchportID) { // VSwitch
			debug(4, "switchforward", switchID, "Destination address found in MAC table.")
			linkID = snet.Router.VSwitch.PortLinksRemote[outboundPort]
		} else { // Regular switch
			for i := range snet.Switches {
				if isSwitchportID(snet.Switches[i], switchportID) {
					debug(4, "switchforward", switchID, "Destination address found in MAC table.")
					linkID = snet.Switches[i].PortLinksRemote[outboundPort]
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
			for port := range snet.Router.VSwitch.PortLinksRemote {
				linkID = snet.Router.VSwitch.PortLinksRemote[port]
				// Don't send out source interface, or unplugged ports
				if (snet.Router.VSwitch.PortLinksLocal[port] != switchportID) && (linkID != "") {
					channels[linkID] <- outFrame
				}
			}
		} else { // Regular switch
			switchIndex := getSwitchIndexFromID(switchID)
			for port := range snet.Switches[switchIndex].PortLinksRemote {
				linkID = snet.Switches[switchIndex].PortLinksRemote[port]
				// Don't send out source interface, or unplugged ports
				if (snet.Switches[switchIndex].PortLinksLocal[port] != switchportID) && (linkID != "") {
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
		snet.Router.VSwitch.PortLinksRemote[switchport] = ""
	} else {
		i := getSwitchIndexFromID(switchID)
		snet.Switches[i].PortLinksRemote[switchport] = ""
	}

}
