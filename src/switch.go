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

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

func NewSumerian2100(hostname string) model.Switch {
	s := model.Switch{}
	s.ID = idgen(8)
	s.Model = "Sumerian 2100"
	s.Hostname = hostname
	s.Maxports = 4
	s.ARPTable = make(map[string]model.ARPEntry)

	return s
}

func addSwitch(switchHostname string) {
	switchModel := strings.ToUpper("Sumerian")

	// input validation
	if hostname_exists(switchHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	s := model.Switch{}
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

	s.MACTable = make(map[string]model.MACEntry)
	Net().Switches = append(Net().Switches, s)

	generateSwitchChannels(getSwitchIndexFromID(s.ID))
	for j := 0; j < getActivePorts(s); j++ {
		engine.Instance().Channels[s.PortLinksLocal[j]] = make(chan json.RawMessage)
		engine.Instance().Sockets[s.PortLinksLocal[j]] = make(map[string]chan model.Frame)
		engine.Instance().ActionSync[s.PortLinksLocal[j]] = make(chan int)

		go listenSwitchportChannel(s.ID, s.PortLinksLocal[j])
	}
}

func addVirtualSwitch(maxports int) model.Switch {
	v := model.Switch{}
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

	v.MACTable = make(map[string]model.MACEntry)

	return v
}

func delSwitch(hostname string) {
	//TODO For all linked devices, unlink. then delete
	hostname = strings.ToUpper(hostname)
	//search for switch
	for i := range Net().Switches {
		if strings.ToUpper(Net().Switches[i].Hostname) == hostname {
			// Unlink all devices connected to this switch
			for j := range Net().Switches[i].PortLinksLocal {
				if Net().Switches[i].PortLinksRemote[j] != "" {
					// Unlink if host
					for h := range Net().Hosts {
						if Net().Hosts[h].Interfaces["eth0"].RemoteL1ID == Net().Switches[i].PortLinksLocal[j] {
							iface := Net().Hosts[h].Interfaces["eth0"]
							iface.RemoteL1ID = ""
							Net().Hosts[h].Interfaces["eth0"] = iface
						}
					}

					// Unlink if switch - TODO
				}
			}

			Net().Switches = removeSwitchFromSlice(Net().Switches, i)
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
	if remoteDevice == strings.ToUpper(Net().Router.Hostname) {
		if getActivePorts(Net().Router.VSwitch) >= Net().Router.VSwitch.Maxports {
			fmt.Printf("No available ports - %s only has %d ports\n", Net().Router.Model, Net().Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range Net().Switches {
		if remoteDevice == strings.ToUpper(Net().Switches[s].Hostname) {
			if getActivePorts(Net().Switches[s]) >= Net().Switches[s].Maxports {
				fmt.Printf("No available ports - %s only has %d ports\n", Net().Switches[s].Model, Net().Switches[s].Maxports)
				return
			}
		}
	}

	//find switch with that hostname
	for i := range Net().Switches {
		if strings.ToUpper(Net().Switches[i].Hostname) == localDevice {
			uplinkID := ""
			//Remote device on new link is the Router
			if remoteDevice == strings.ToUpper(Net().Router.Hostname) {
				//find next free port
				for k := range Net().Router.VSwitch.PortLinksLocal {
					if (Net().Router.VSwitch.PortLinksRemote[k] == "") && (uplinkID == "") {
						uplinkID = Net().Router.VSwitch.PortLinksLocal[k]
					}
				}
				//uplinkID = Net().Router.VSwitch.ID

				// Assign switchport on remote device
				assignSwitchport(Net().Router.VSwitch, Net().Hosts[i].ID)
			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range Net().Switches {
					if remoteDevice == strings.ToUpper(Net().Switches[j].Hostname) {

						//find next free port
						for k := range Net().Switches[j].PortLinksLocal {
							if (Net().Switches[j].PortLinksRemote[k] == "") && (uplinkID == "") {
								uplinkID = Net().Switches[j].PortLinksLocal[k]

								// Assign switchport on remote device
								assignSwitchport(Net().Switches[j], Net().Switches[i].ID)
							}
						}

					}
				}
			}

			// Assign switchport on local switch
			assignSwitchport(Net().Switches[i], "TEST")

			return
		}
	}
}

func lookupMACTable(dstMAC string, switchportID string) int { // For looking up addresses
	resultPort := -1
	var MACTable map[string]model.MACEntry

	if isSwitchportID(Net().Router.VSwitch, switchportID) {
		MACTable = Net().Router.VSwitch.MACTable
	} else {
		for i := range Net().Switches {
			if isSwitchportID(Net().Switches[i], switchportID) {
				MACTable = Net().Switches[i].MACTable
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
	table := make(map[string]model.MACEntry)
	if isSwitchportID(Net().Router.VSwitch, id) {
		table = Net().Router.VSwitch.MACTable
	} else {
		for i := range Net().Switches {
			if isSwitchportID(Net().Switches[i], id) {
				table = Net().Switches[i].MACTable
			}
		}
	}

	for k, v := range table {
		if k == macaddr {
			if v.Interface == port {
				writeLog(4, "checkMACTable", id, "Source address found in MAC table")
				result = v.Interface
			} else {
				writeLog(4, "checkMACTable", id, "Source address found in MAC table, but wrong - removing old.")
				delMACEntry(macaddr, id, port)
			}
		}
	}

	if result == -1 {
		msg := "Source address " + macaddr + " not found in MAC table. Adding"
		writeLog(3, "learnMACTable", id, msg)
		addMACEntry(macaddr, id, port)
	}
}

func addMACEntry(macaddr string, id string, port int) {
	if isSwitchportID(Net().Router.VSwitch, id) {
		macEntry := model.MACEntry{
			Interface: port,
		}
		Net().Router.VSwitch.MACTable[macaddr] = macEntry
	} else {
		for i := range Net().Switches {
			if isSwitchportID(Net().Switches[i], id) {
				macEntry := model.MACEntry{
					Interface: port,
				}
				Net().Switches[i].MACTable[macaddr] = macEntry
			}
		}
	}

}

func delMACEntry(macaddr string, id string, port int) {
}

func isSwitchportID(sw model.Switch, id string) bool {
	for i := range sw.PortLinksLocal {
		if sw.PortLinksLocal[i] == id {
			return true
		}
	}

	return false
}

func getActivePorts(sw model.Switch) int {
	count := 0

	for i := range sw.PortLinksRemote {
		if sw.PortLinksRemote[i] != "" {
			count++
		}
	}

	return count
}

func assignSwitchport(sw model.Switch, id string) int {
	portIndex := -1
	for i := range sw.PortLinksRemote {
		if sw.PortLinksRemote[i] == "" {
			sw.PortLinksRemote[i] = id
			portIndex = i
			break
		}
	}

	engine.Instance().Channels[sw.PortLinksLocal[portIndex]] = make(chan json.RawMessage)
	writeLog(4, "assignSwitchport", sw.PortLinksLocal[portIndex], "listening for id")
	go listenSwitchportChannel(sw.ID, sw.PortLinksLocal[portIndex])

	return portIndex
}

func switchforward(frame model.Frame, switchID string, switchportID string) {
	srcMAC := frame.SrcMAC
	dstMAC := frame.DstMAC
	linkID := ""
	floodFrame := false

	outboundPort := lookupMACTable(dstMAC, switchportID)

	if dstMAC == "ff:ff:ff:ff:ff:ff" { // Broadcast
		floodFrame = true
		writeLog(4, "switchforward", switchID, "L2 Broadcast. Flooding frame on all ports")
	} else if outboundPort == -1 { // No matching port for this MAC address was found in the MAC address table
		floodFrame = true
		writeLog(4, "switchforward", switchID, "Destination address "+dstMAC+" not found in MAC table. Flooding frame on all ports")
	} else {
		if isSwitchportID(Net().Router.VSwitch, switchportID) { // VSwitch
			writeLog(4, "switchforward", switchID, "Destination address found in MAC table.")
			linkID = Net().Router.VSwitch.PortLinksRemote[outboundPort]
		} else { // Regular switch
			for i := range Net().Switches {
				if isSwitchportID(Net().Switches[i], switchportID) {
					writeLog(4, "switchforward", switchID, "Destination address found in MAC table.")
					linkID = Net().Switches[i].PortLinksRemote[outboundPort]
				}
			}
		}
	}

	p := frame.Data
	f := model.Frame{
		SrcMAC:    srcMAC,
		DstMAC:    dstMAC,
		EtherType: frame.EtherType,
		Data:      p,
	}
	outFrame, _ := json.Marshal(f)

	if floodFrame {
		if isSwitchportID(Net().Router.VSwitch, switchportID) { // VSwitch
			for port := range Net().Router.VSwitch.PortLinksRemote {
				linkID = Net().Router.VSwitch.PortLinksRemote[port]
				// Don't send out source interface, or unplugged ports
				if (Net().Router.VSwitch.PortLinksLocal[port] != switchportID) && (linkID != "") {
					engine.Instance().Channels[linkID] <- outFrame
				}
			}
		} else { // Regular switch
			switchIndex := getSwitchIndexFromID(switchID)
			for port := range Net().Switches[switchIndex].PortLinksRemote {
				linkID = Net().Switches[switchIndex].PortLinksRemote[port]
				// Don't send out source interface, or unplugged ports
				if (Net().Switches[switchIndex].PortLinksLocal[port] != switchportID) && (linkID != "") {
					engine.Instance().Channels[linkID] <- outFrame
				}
			}
		}
	} else {
		engine.Instance().Channels[linkID] <- outFrame
	}
}

func freeSwitchport(link string) {

	switchport := getSwitchportIDFromLink(link)
	switchID := getSwitchIDFromLink(link)

	if Net().Router.VSwitch.ID == switchID {
		Net().Router.VSwitch.PortLinksRemote[switchport] = ""
	} else {
		i := getSwitchIndexFromID(switchID)
		Net().Switches[i].PortLinksRemote[switchport] = ""
	}

}
