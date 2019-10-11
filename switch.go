/*
File:		switch.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Switch-specific functions
*/

package main

import(
	"fmt"
	"strings"
)

func NewSumerian2100(hostname string) Switch {
	s := Switch{}
	s.ID = idgen(8)
	s.Model = "Sumerian 2100"
	s.Hostname = hostname
	s.Maxports = 4

	return s
}

func addSwitch() {
	fmt.Println("What model?")
	fmt.Println("Available: Sumerian")
	fmt.Print("Model: ")
	scanner.Scan()
	switchModel := scanner.Text()
	switchModel = strings.ToUpper(switchModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	switchHostname := scanner.Text()

	// input validation
	if switchHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

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

func delSwitch() {
	//TODO For all linked devices, unlink. then delete
}

func lookupMACTable(macaddr string, id string) int { // For looking up addresses
	resultPort := 0
	table := make(map[string]int)
	if(isSwitchportID(snet.Router.VSwitch, id)) {
		table = snet.Router.VSwitch.MACTable
	} else {
		for i := range snet.Switches {
			if(isSwitchportID(snet.Switches[i], id)) {
				table = snet.Switches[i].MACTable
			}
		}
	}

	//if switch
	for k, v := range table {
		if(k == macaddr) {
			resultPort = v
		}
	}

	return resultPort
}

func checkMACTable(macaddr string, id string, port int) { // For checking table on incoming frames
	result := -1
	table := make(map[string]int)
	if(isSwitchportID(snet.Router.VSwitch, id)) {
		table = snet.Router.VSwitch.MACTable
	} else {
		for i := range snet.Switches {
			if(isSwitchportID(snet.Switches[i], id)) {
			table = snet.Switches[i].MACTable
			}
		}
	}

	for k, v := range table {
		if(k == macaddr) {
			if(v == port) {
				debug(4, "lookupMACTable", id, "Address found in MAC table")
				result = v
			} else {
				delMACEntry(macaddr, id, port)
			}
		}
	}

	if(result == 0) {
		debug(3, "learnMACTable", id, "Address not found in MAC table. Adding")
		addMACEntry(macaddr, id, port)
	}
}

func addMACEntry(macaddr string, id string, port int) {
	if(isSwitchportID(snet.Router.VSwitch, id)) {
		snet.Router.VSwitch.MACTable[macaddr] = port
	} else {
		for i := range snet.Switches {
			if(isSwitchportID(snet.Switches[i], id)) {
				snet.Switches[i].MACTable[macaddr] = port
			}
		}
	}


}

func delMACEntry(macaddr string, id string, port int) {
}

func isSwitchportID(sw Switch, id string) bool {
	for i := range sw.PortIDs {
		if(sw.PortIDs[i] == id) { return true }
	}

	return false
}

func getActivePorts(sw Switch) int {
	count := 0

	for i := range sw.Ports {
		if(sw.Ports[i] != "") {
			count++
		}
	}

	return count
}

func assignSwitchport(sw Switch, id string) Switch {
	index := getActivePorts(sw)
	sw.Ports[index] = id

	channels[sw.PortIDs[index]] = make(chan Frame)
	internal[sw.PortIDs[index]] = make(chan Frame)
	debug(4, "generateRouterChannels", sw.PortIDs[index], "listening for id")
	go switchportlisten(sw.PortIDs[index])

	return sw
}

func switchforward(frame Frame, id string) {
	srcIP := frame.Data.SrcIP
	dstIP := frame.Data.DstIP
	srcMAC := frame.SrcMAC
	dstMAC := frame.DstMAC
	linkID := ""

	outboundPort := lookupMACTable(dstMAC, id)
	if(outboundPort == -1) {
		debug(1, "switchforward", id, "Warning: Not found in MAC table, using bypass") //TODO implement flooding
		linkID = getIDfromMAC(dstMAC)
	} else {
		if(isSwitchportID(snet.Router.VSwitch, id)) {
			linkID = snet.Router.VSwitch.Ports[outboundPort]
			fmt.Println("linkID: ", linkID)
		} else {
			for i := range snet.Switches {
				fmt.Println("Should never print this yet")
				if(isSwitchportID(snet.Switches[i], id)){
					linkID = snet.Switches[i].Ports[outboundPort]
				}
			}
		}
	}

	s := frame.Data.Data
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
}
