/*
File:		display.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to drawing network diagrams and displaying info
*/

package main

import (
	"fmt"
	"strconv"
)

/* DIAGRAMMING */

func drawDiagram(rootID string) {
	drawDiagramAction(rootID, "")

	//Unlinked switches
	for i := range snet.Switches {
		drawDiagramAction(snet.Switches[i].ID, "switch")
	}

	// Unlinked hosts
	for i := range snet.Hosts {
		if snet.Hosts[i].UplinkID == "" {
			drawHost(snet.Hosts[i].ID)
		}
	}
}

func drawDiagramAction(rootID string, rootType string) { // TODO make recursive - in progress 10/7
	// Identify device info about rootID
	rootHostname := ""
	//rootIndex := -1
	if rootID == snet.Router.ID {
		rootHostname = snet.Router.Hostname
		rootType = "router"
	}

	if rootType == "switch" {
		for i := range snet.Switches {
			if rootID == snet.Switches[i].ID {
				rootHostname = snet.Switches[i].Hostname
				rootType = "switch"
				//rootIndex = i
				drawSwitch(snet.Switches[i].ID)

				for j := range snet.Switches[i].PortLinksRemote {
					if snet.Switches[i].PortLinksRemote[j] != "" {
						drawConnectedHost(snet.Switches[i].PortLinksRemote[j], j, snet.Switches[i])
					}
				}
			}
		}

	}

	if rootType == "" {
		for i := range snet.Hosts {
			if rootID == snet.Hosts[i].ID {
				rootHostname = snet.Hosts[i].Hostname
				rootType = "host"
				//rootIndex = i
			}
		}
	}

	// ROUTER
	if rootType == "router" {
		if rootHostname != "" {
			drawRouter()
		}

		for i := range snet.Router.VSwitch.PortLinksRemote {
			if snet.Router.VSwitch.PortLinksRemote[i] != "" && i != 0 {
				drawConnectedHost(snet.Router.VSwitch.PortLinksRemote[i], i, snet.Router.VSwitch)
			}
		}
	}

	// SWITCH

	// HOST
	if rootType == "host" {
	}
}

func drawRouter() {
	space1 := 13 - len(snet.Router.Hostname)
	space2 := 14 - len(snet.Router.GetIP())
	space3 := 16 - len(snet.Router.Model)

	fmt.Println("|------------------------|")
	fmt.Println("|         Router         |")
	fmt.Printf("| Hostname: %s", snet.Router.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Gateway: %s", snet.Router.GetIP())
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Model: %s", snet.Router.Model)
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Println("|\n|------------------------|")
}

func drawSwitch(id string) {
	sw := snet.Switches[getSwitchIndexFromID(id)]

	connectedPorts := 0
	for i := range sw.PortLinksRemote {
		if sw.PortLinksRemote[i] != "" {
			connectedPorts++
		}
	}

	space1 := 13 - len(sw.Hostname)
	space2 := 11 - len(strconv.Itoa(len(sw.PortLinksLocal)))
	space3 := 5
	space4 := 16 - len(sw.Model)

	fmt.Println("")
	fmt.Println("|------------------------|")
	fmt.Println("|          Switch        |")
	fmt.Printf("| Hostname: %s", sw.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| Port count: %d", len(sw.PortLinksLocal))
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| Connected ports: %d", connectedPorts)
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| Model: %s", sw.Model)
	for i := 0; i < space4; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Println("|------------------------|")
}

func drawHost(id string) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.GetIP())
	space3 := 16 - len(h.Model)

	fmt.Println("")
	fmt.Println("|------------------------|")
	fmt.Println("|          Host          |")
	fmt.Printf("| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| IP Addr: %s", h.GetIP())
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| Model: %s", h.Model)
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Println("|------------------------|")
}

func drawConnectedHost(id string, iter int, sw Switch) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.GetIP())
	space3 := 16 - len(h.Model)

	fmt.Println("            ||")
	fmt.Println("            ||      |------------------------|")
	fmt.Println("            ||      |          Host          |")
	fmt.Printf("            ||------| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("            ||------| IP Addr: %s", h.GetIP())
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")

	if iter == getActivePorts(sw)-1 {
		fmt.Printf("                    | Model: %s", h.Model)
	} else {
		fmt.Printf("            ||      | Model: %s", h.Model)
	}
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	if iter == getActivePorts(sw)-1 {
		fmt.Println("                    |------------------------|")
	} else {
		fmt.Println("            ||      |------------------------|")
	}

}

/* DISPLAYING */

func overview() {
	fmt.Printf("Network name:\t\t%s\n", snet.Name)
	fmt.Printf("Network ID:\t\t%s\n", snet.ID)
	fmt.Printf("Network size:\t\t/%s\n", snet.Netsize)

	// router
	routerCount := 1
	show(snet.Router.Hostname)

	//switches
	switchCount := 0
	for i := 0; i < len(snet.Switches); i++ {
		fmt.Printf("\nSwitch %v\n", snet.Switches[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Switches[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Switches[i].Model)
		switchCount = i + 1
	}

	//hosts
	hostCount := 0
	for i := 0; i < len(snet.Hosts); i++ {
		fmt.Printf("\nHost %v\n", snet.Hosts[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[i].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[i].Interface.MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[i].GetIP())
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[i].GetGateway())
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[i].GetMask())
		uplinkHostname := ""
		//Router
		if isSwitchportID(snet.Router.VSwitch, snet.Hosts[i].UplinkID) {
			uplinkHostname = snet.Router.Hostname + " (" + snet.Router.VSwitch.Hostname + ")"
		}

		//Switches
		for j := range snet.Switches {
			if isSwitchportID(snet.Switches[j], snet.Hosts[i].UplinkID) {
				uplinkHostname = snet.Switches[j].Hostname
			}
		}
		fmt.Printf("\tUplink to:\t%s\n", uplinkHostname)
		hostCount = i + 1
	}

	fmt.Printf("\nTotal devices: %d (%d Router, %d Switches, %d Hosts)\n", (routerCount + switchCount + hostCount), routerCount, switchCount, hostCount)
}

func show(hostname string) {
	device_type := "host"
	id := -1
	if snet.Router.Hostname == hostname {
		device_type = "router"
		id = 0
	}

	if snet.Router.VSwitch.Hostname == hostname {
		device_type = "vswitch"
		id = 0
	}

	for i := range snet.Hosts {
		if snet.Hosts[i].Hostname == hostname {
			device_type = "host"
			id = i
		}
	}

	for i := range snet.Switches {
		if snet.Switches[i].Hostname == hostname {
			device_type = "switch"
			id = i
		}
	}

	if id == -1 {
		fmt.Printf("Hostname not found\n")
		return
	}

	if device_type == "host" {
		fmt.Printf("\nHost %v\n", snet.Hosts[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[id].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[id].Interface.MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[id].GetIP())
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[id].GetGateway())
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[id].GetMask())
		uplinkHostname := ""

		//Router
		if isSwitchportID(snet.Router.VSwitch, snet.Hosts[id].UplinkID) {
			uplinkHostname = snet.Router.Hostname + " (" + snet.Router.VSwitch.Hostname + ")"
		}
		//Switches
		for j := range snet.Switches {
			if isSwitchportID(snet.Switches[j], snet.Hosts[id].UplinkID) {
				uplinkHostname = snet.Switches[j].Hostname
			}
		}

		fmt.Printf("\tUplink to:\t%s\n\n", uplinkHostname)
	} else if device_type == "switch" {
		fmt.Printf("\nSwitch %s\n", snet.Switches[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Switches[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Switches[id].Model)
	} else if device_type == "vswitch" {
		fmt.Printf("\nSwitch %s\n", snet.Router.VSwitch.Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Router.VSwitch.ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Router.VSwitch.Model)
	} else if device_type == "router" {
		fmt.Printf("\nRouter %s\n", snet.Router.Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Router.ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Router.Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Router.Interface.MACAddr)
		fmt.Printf("\tGateway:\t%s\n", snet.Router.GetIP())
		fmt.Printf("\tDHCP pool:\t%d addresses\n", len(snet.Router.GetDHCPPoolAddresses()))
		fmt.Printf("\tVSwitch ID: \t%s\n", snet.Router.VSwitch.ID)
	}
}

func displayARPTable(deviceID string) {
	var ARPTable map[string]ARPEntry

	if snet.Router.ID == deviceID {
		ARPTable = snet.Router.ARPTable
	} else {
		ARPTable = snet.Hosts[getHostIndexFromID(deviceID)].ARPTable
	}

	fmt.Printf("ARP Table:\n")
	fmt.Printf("IP Address\t\tMAC Address\t\tInterface\t\tExpiration\n")

	for i := range ARPTable {
		fmt.Printf("%s\t\t%s\t%s\n", i, ARPTable[i].MACAddr, ARPTable[i].Interface)
	}
	fmt.Printf("\n")
}

func displayMACTable(deviceID string) {
	var MACTable map[string]MACEntry

	if snet.Router.VSwitch.ID == deviceID {
		MACTable = snet.Router.VSwitch.MACTable
	} else {
		MACTable = snet.Switches[getSwitchIndexFromID(deviceID)].MACTable
	}

	fmt.Printf("MAC Table:\n")
	fmt.Printf("MAC Address\t\tInterface\t\tExpiration\n")

	for i := range MACTable {
		fmt.Printf("%s\t%d\n", i, MACTable[i].Interface)
	}
	fmt.Printf("\n")
}
