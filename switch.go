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
	if(id == snet.Router.VSwitch.ID) {
		table = snet.Router.VSwitch.MACTable
	} else {
		table = snet.Switches[getSwitchIndexFromID(id)].MACTable
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
	result := 0
	table := make(map[string]int)
	if(id == snet.Router.VSwitch.ID) {
		table = snet.Router.VSwitch.MACTable
	} else {
		table = snet.Switches[getSwitchIndexFromID(id)].MACTable
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
	if(id == snet.Router.VSwitch.ID) {
		snet.Router.VSwitch.MACTable[macaddr] = port
	} else {
		snet.Switches[getSwitchIndexFromID(id)].MACTable[macaddr] = port
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

func switchforward(frame Frame, id string) {
	srcIP := frame.Data.SrcIP
	dstIP := frame.Data.DstIP
	srcMAC := frame.SrcMAC
	dstMAC := frame.DstMAC
	linkID := ""

	outboundPort := lookupMACTable(dstMAC, id)
	if(outboundPort == 0) {
		debug(1, "switchforward", id, "Warning: Not found in MAC table, using bypass") //TODO implement flooding
		linkID = getIDfromMAC(dstMAC)
	} else {
		linkID = snet.Switches[getSwitchIndexFromID(id)].Ports[outboundPort]
	}
	//linkID := getIDfromMAC(dstMAC) //TODO fix

	s := frame.Data.Data
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
}
