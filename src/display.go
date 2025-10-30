/*
File:		display.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to drawing network diagrams and displaying info
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/vincebel7/ltdnet/src/model"
)

/* DIAGRAMMING */

func drawDiagram(rootID string) {
	drawDiagramAction(rootID, "")

	//Unlinked switches
	for i := range Net().Switches {
		drawDiagramAction(Net().Switches[i].ID, "switch")
	}

	// Unlinked hosts
	for i := range Net().Hosts {
		if Net().Hosts[i].Interfaces["eth0"].RemoteL1ID == "" {
			drawHost(Net().Hosts[i].ID)
		}
	}
}

func drawDiagramAction(rootID string, rootType string) { // TODO make recursive - in progress 10/7
	// Identify device info about rootID
	rootHostname := ""
	//rootIndex := -1
	if rootID == Net().Router.ID {
		rootHostname = Net().Router.Hostname
		rootType = "router"
	}

	if rootType == "switch" {
		for i := range Net().Switches {
			if rootID == Net().Switches[i].ID {
				rootHostname = Net().Switches[i].Hostname
				rootType = "switch"
				//rootIndex = i
				drawSwitch(Net().Switches[i].ID)

				for j := range Net().Switches[i].PortLinksRemote {
					if Net().Switches[i].PortLinksRemote[j] != "" {
						drawConnectedHost(Net().Switches[i].PortLinksRemote[j], j, Net().Switches[i])
					}
				}
			}
		}

	}

	if rootType == "" {
		for i := range Net().Hosts {
			if rootID == Net().Hosts[i].ID {
				rootHostname = Net().Hosts[i].Hostname
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

		for i := range Net().Router.VSwitch.PortLinksRemote {
			if Net().Router.VSwitch.PortLinksRemote[i] != "" && i != 0 {

				hostID := ""
				for h := range Net().Hosts {
					if Net().Hosts[h].Interfaces["eth0"].L1ID == Net().Router.VSwitch.PortLinksRemote[i] {
						hostID = Net().Hosts[h].ID
					}
				}

				drawConnectedHost(hostID, i, Net().Router.VSwitch)
			}
		}
	}

	// SWITCH

	// HOST
	if rootType == "host" {
	}
}

func drawRouter() {
	space1 := 13 - len(Net().Router.Hostname)
	space2 := 14 - len(Net().Router.GetIP("eth0"))
	space3 := 16 - len(Net().Router.Model)

	fmt.Println("|------------------------|")
	fmt.Println("|         Router         |")
	fmt.Printf("| Hostname: %s", Net().Router.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Gateway: %s", Net().Router.GetIP("eth0"))
	for i := 0; i < space2; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n| Model: %s", Net().Router.Model)
	for i := 0; i < space3; i++ {
		fmt.Printf(" ")
	}
	fmt.Println("|\n|------------------------|")
}

func drawSwitch(id string) {
	sw := Net().Switches[getSwitchIndexFromID(id)]

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
	h := Net().Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.GetIP("eth0"))
	space3 := 16 - len(h.Model)

	fmt.Println("")
	fmt.Println("|------------------------|")
	fmt.Println("|          Host          |")
	fmt.Printf("| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("| IP Addr: %s", h.GetIP("eth0"))
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

func drawConnectedHost(id string, iter int, sw model.Switch) {
	h := Net().Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.GetIP("eth0"))
	space3 := 16 - len(h.Model)

	fmt.Println("            ||")
	fmt.Println("            ||      |------------------------|")
	fmt.Println("            ||      |          Host          |")
	fmt.Printf("            ||------| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ {
		fmt.Printf(" ")
	}
	fmt.Printf("|\n")
	fmt.Printf("            ||------| IP Addr: %s", h.GetIP("eth0"))
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
	fmt.Printf("Network name:\t\t%s\n", Net().Name)
	fmt.Printf("Network ID:\t\t%s\n", Net().ID)
	fmt.Printf("Network size:\t\t/%s\n", Net().Netsize)

	// router
	routerCount := 1
	show(Net().Router.Hostname)

	//switches
	switchCount := 0
	for i := 0; i < len(Net().Switches); i++ {
		fmt.Printf("\nSwitch %v\n", Net().Switches[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Switches[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Switches[i].Model)
		switchCount = i + 1
	}

	//hosts
	hostCount := 0
	for i := 0; i < len(Net().Hosts); i++ {
		fmt.Printf("\nHost %v\n", Net().Hosts[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Hosts[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Hosts[i].Model)
		fmt.Printf("\tMAC:\t\t%s\n", Net().Hosts[i].Interfaces["eth0"].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", Net().Hosts[i].GetIP("eth0"))
		fmt.Printf("\tDef. Gateway:\t%s\n", Net().Hosts[i].GetGateway("eth0"))
		fmt.Printf("\tSubnet Mask:\t%s\n", Net().Hosts[i].GetMask("eth0"))
		uplinkHostname := ""
		//Router
		if isSwitchportID(Net().Router.VSwitch, Net().Hosts[i].Interfaces["eth0"].RemoteL1ID) {
			uplinkHostname = Net().Router.Hostname + " (" + Net().Router.VSwitch.Hostname + ")"
		}

		//Switches
		for j := range Net().Switches {
			if isSwitchportID(Net().Switches[j], Net().Hosts[i].Interfaces["eth0"].RemoteL1ID) {
				uplinkHostname = Net().Switches[j].Hostname
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
	if Net().Router.Hostname == hostname {
		device_type = "router"
		id = 0
	}

	if Net().Router.VSwitch.Hostname == hostname {
		device_type = "vswitch"
		id = 0
	}

	for i := range Net().Hosts {
		if Net().Hosts[i].Hostname == hostname {
			device_type = "host"
			id = i
		}
	}

	for i := range Net().Switches {
		if Net().Switches[i].Hostname == hostname {
			device_type = "switch"
			id = i
		}
	}

	if id == -1 {
		fmt.Printf("Hostname not found\n")
		return
	}

	if device_type == "host" {
		fmt.Printf("\nHost %v\n", Net().Hosts[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Hosts[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Hosts[id].Model)
		fmt.Printf("\tMAC:\t\t%s\n", Net().Hosts[id].Interfaces["eth0"].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", Net().Hosts[id].GetIP("eth0"))
		fmt.Printf("\tDef. Gateway:\t%s\n", Net().Hosts[id].GetGateway("eth0"))
		fmt.Printf("\tSubnet Mask:\t%s\n", Net().Hosts[id].GetMask("eth0"))
		uplinkHostname := ""

		//Router
		if isSwitchportID(Net().Router.VSwitch, Net().Hosts[id].Interfaces["eth0"].RemoteL1ID) {
			uplinkHostname = Net().Router.Hostname + " (" + Net().Router.VSwitch.Hostname + ")"
		}
		//Switches
		for j := range Net().Switches {
			if isSwitchportID(Net().Switches[j], Net().Hosts[id].Interfaces["eth0"].RemoteL1ID) {
				uplinkHostname = Net().Switches[j].Hostname
			}
		}

		fmt.Printf("\tUplink to:\t%s\n\n", uplinkHostname)
	} else if device_type == "switch" {
		fmt.Printf("\nSwitch %s\n", Net().Switches[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Switches[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Switches[id].Model)
	} else if device_type == "vswitch" {
		fmt.Printf("\nSwitch %s\n", Net().Router.VSwitch.Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Router.VSwitch.ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Router.VSwitch.Model)
	} else if device_type == "router" {
		fmt.Printf("\nRouter %s\n", Net().Router.Hostname)
		fmt.Printf("\tID:\t\t%s\n", Net().Router.ID)
		fmt.Printf("\tModel:\t\t%s\n", Net().Router.Model)
		fmt.Printf("\tMAC:\t\t%s\n", Net().Router.Interfaces["eth0"].MACAddr)
		fmt.Printf("\tGateway:\t%s\n", Net().Router.GetIP("eth0"))
		fmt.Printf("\tDHCP pool:\t%d addresses\n", len(Net().Router.GetDHCPPoolAddresses()))
		fmt.Printf("\tVSwitch ID: \t%s\n", Net().Router.VSwitch.ID)
	}
}

func displayARPTable(deviceID string) {
	var ARPTable map[string]model.ARPEntry

	if Net().Router.ID == deviceID {
		ARPTable = Net().Router.ARPTable
	} else {
		ARPTable = Net().Hosts[getHostIndexFromID(deviceID)].ARPTable
	}

	fmt.Printf("ARP Table:\n")
	fmt.Printf("IP Address\t\tMAC Address\t\tInterface\t\tExpiration\n")

	for i := range ARPTable {
		fmt.Printf("%s\t\t%s\t%s\n", i, ARPTable[i].MACAddr, ARPTable[i].Interface)
	}
	fmt.Printf("\n")
}

func displayMACTable(deviceID string) {
	var MACTable map[string]model.MACEntry

	if Net().Router.VSwitch.ID == deviceID {
		MACTable = Net().Router.VSwitch.MACTable
	} else {
		MACTable = Net().Switches[getSwitchIndexFromID(deviceID)].MACTable
	}

	fmt.Printf("MAC Table:\n")
	fmt.Printf("MAC Address\t\tInterface\t\tExpiration\n")

	for i := range MACTable {
		fmt.Printf("%s\t%d\n", i, MACTable[i].Interface)
	}
	fmt.Printf("\n")
}
