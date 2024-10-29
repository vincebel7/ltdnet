/*
File:		display.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to drawing network diagrams and displaying info
*/

package main

import (
	"fmt"
)

/* DIAGRAMMING */

func drawDiagram(rootID string) {
	drawDiagramAction(rootID, "")

	//Unlinked switches
	// drawDiagramConnected(switch ID)

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

	if rootType == "" {
		for i := range snet.Switches {
			if rootID == snet.Switches[i].ID {
				rootHostname = snet.Switches[i].Hostname
				rootType = "switch"
				//rootIndex = i
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

		for i := range snet.Router.VSwitch.Ports {
			if snet.Router.VSwitch.Ports[i] != "" && i != 0 {
				drawConnectedHost(snet.Router.VSwitch.Ports[i], i)
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
	space2 := 14 - len(snet.Router.Gateway.String())
	space3 := 16 - len(snet.Router.Model)

	fmt.Println("|------------------------|")
	fmt.Println("|         Router         |")
	fmt.Printf("| Hostname: %s", snet.Router.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Gateway: %s", snet.Router.Gateway.String())
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Model: %s", snet.Router.Model)
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Println("|\n|------------------------|")
}

func drawHost(id string) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.IPAddr.String())
	space3 := 16 - len(h.Model)

	fmt.Println("")
	fmt.Println("|------------------------|")
	fmt.Println("|          Host          |")
	fmt.Printf("| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| IP Addr: %s", h.IPAddr)
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

func drawConnectedHost(id string, iter int) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.IPAddr.String())
	space3 := 16 - len(h.Model)

	fmt.Println("            ||")
	fmt.Println("            ||      |------------------------|")
	fmt.Println("            ||      |          Host          |")
	fmt.Printf("            ||------| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("            ||------| IP Addr: %s", h.IPAddr)
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")

	if iter == getActivePorts(snet.Router.VSwitch)-1 {
		fmt.Printf("                    | Model: %s", h.Model)
	} else {
		fmt.Printf("            ||      | Model: %s", h.Model)
	}
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	if iter == getActivePorts(snet.Router.VSwitch)-1 {
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
		fmt.Printf("\tMgmt IP:\t%s\n", snet.Switches[i].MgmtIP)
		switchCount = i + 1
	}

	//hosts
	hostCount := 0
	for i := 0; i < len(snet.Hosts); i++ {
		fmt.Printf("\nHost %v\n", snet.Hosts[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[i].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[i].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[i].IPAddr.String())
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[i].DefaultGateway.String())
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[i].SubnetMask)
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
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[id].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[id].IPAddr.String())
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[id].DefaultGateway.String())
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[id].SubnetMask)
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
		fmt.Printf("\tMgmt IP:\t%s\n\n", snet.Switches[id].MgmtIP)
	} else if device_type == "vswitch" {
		fmt.Printf("\nSwitch %s\n", snet.Router.VSwitch.Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Router.VSwitch.ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Router.VSwitch.Model)
		fmt.Printf("\tMgmt IP:\t%s\n\n", snet.Router.VSwitch.MgmtIP)
	} else if device_type == "router" {
		fmt.Printf("\nRouter %s\n", snet.Router.Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Router.ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Router.Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Router.MACAddr)
		fmt.Printf("\tGateway:\t%s\n", snet.Router.Gateway.String())
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
