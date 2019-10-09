/*
File:		diagram.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Functions related to drawing network diagrams
*/


package main

import(
	"fmt"
)

func drawDiagram(rootID string) {
	drawDiagramAction(rootID, "")

	//Unlinked switches
	// drawDiagramConnected(switch ID)

	// Unlinked hosts
	for i := range snet.Hosts {
		if(snet.Hosts[i].UplinkID == "") {
			drawHost(snet.Hosts[i].ID)
		}
	}
}

func drawDiagramAction(rootID string, rootType string) { // TODO make recursive - in progress 10/7
	// Identify device info about rootID
	rootHostname := ""
	//rootIndex := -1
	if(rootID == snet.Router.ID) {
		rootHostname = snet.Router.Hostname
		rootType = "router"
	}

	if(rootType == "") {
		for i := range snet.Switches {
			if(rootID == snet.Switches[i].ID) {
				rootHostname = snet.Switches[i].Hostname
				rootType = "switch"
				//rootIndex = i
			}
		}
	}

	if(rootType == "") {
		for i := range snet.Hosts {
			if(rootID == snet.Hosts[i].ID) {
				rootHostname = snet.Hosts[i].Hostname
				rootType = "host"
				//rootIndex = i
			}
		}
	}

	// ROUTER
	if(rootType == "router"){
		if(rootHostname != "") {
			drawRouter(snet.Router.ID)
		}

		for i := range snet.Router.VSwitch.Ports {
			drawConnectedHost(snet.Router.VSwitch.Ports[i], i)
		}
	}

	// SWITCH

	// HOST
	if(rootType == "host"){
	}
}

func drawRouter(id string) {
	space1 := 13 - len(snet.Router.Hostname)
	space2 := 14 - len(snet.Router.Gateway)
	space3 := 16 - len(snet.Router.Model)

	fmt.Println("|------------------------|")
	fmt.Println("|         Router         |")
	 fmt.Printf("| Hostname: %s", snet.Router.Hostname)
	for i := 0; i < space1; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n| Gateway: %s", snet.Router.Gateway)
	for i := 0; i < space2; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n| Model: %s", snet.Router.Model)
	for i := 0; i < space3; i++ { fmt.Printf(" ") }
	fmt.Println("|\n|------------------------|")
}

func drawHost(id string) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 13 - len(h.Hostname)
	space2 := 14 - len(h.IPAddr)
	space3 := 16 - len(h.Model)

	fmt.Println("")
	fmt.Println("|------------------------|")
	fmt.Println("|          Host          |")
	 fmt.Printf("| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")
	 fmt.Printf("| IP Addr: %s", h.IPAddr)
	for i := 0; i < space2; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")
	 fmt.Printf("| Model: %s", h.Model)
	for i := 0; i < space3; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")
	 fmt.Println("|------------------------|")
}

func drawConnectedHost(id string, iter int) {
	h := snet.Hosts[getHostIndexFromID(id)]

	space1 := 0
	space2 := 0
	space3 := 0

	if(snet.Hosts[getHostIndexFromID(id)].UplinkID != "") {
		space1 = 13 - len(h.Hostname)
		space2 = 14 - len(h.IPAddr)
		space3 = 16 - len(h.Model)
	}

	fmt.Println("            ||")
	fmt.Println("            ||      |------------------------|")
	fmt.Println("            ||      |          Host          |")
	 fmt.Printf("            ||------| Hostname: %s", h.Hostname)
	for i := 0; i < space1; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")
	 fmt.Printf("            ||------| IP Addr: %s", h.IPAddr)
	for i := 0; i < space2; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")

	if(iter == len(snet.Router.VSwitch.Ports) - 1) {
	 fmt.Printf("                    | Model: %s", h.Model)
	} else {
	 fmt.Printf("            ||      | Model: %s", h.Model)
	}
	for i := 0; i < space3; i++ { fmt.Printf(" ") }
	 fmt.Printf("|\n")
	if(iter == len(snet.Router.VSwitch.Ports) - 1) {
	 fmt.Println("                    |------------------------|")
	} else {
	 fmt.Println("            ||      |------------------------|")
	}

}
