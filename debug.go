/*
File:		debug.co
Author: 	https://bitbucket.org/vincebel
Purpose:	Functions related to debugging and testing
*/

package main

import(
	"fmt"
	"time"
	"strconv"
)

/* DEBUG LEVELS
0 - No debugging
1 - Errors
2 - General network traffic
3 - All network traffic
4 - Garbage
*/

func setDebug(val string) {
	intval, _ := strconv.Atoi(val)
	snet.DebugLevel = intval
	fmt.Printf("Debug level set to %d\n", snet.DebugLevel)
}

func getDebug() int {
	return snet.DebugLevel
}

func debug(level int, generatingFunc string, generatingID string, message string) {
	if(snet.DebugLevel >= level) {
		hostname := ""
		if(generatingID == "Listener") {
			hostname = "Listener"
		} else {
			deviceType := getDeviceType(generatingID)
			if(deviceType == "host"){
				hostname = snet.Hosts[getHostIndexFromID(generatingID)].Hostname
			} else if(deviceType == "router") {
				hostname = snet.Router.Hostname
			} else {
				hostname = generatingID
			}
		}
		fmt.Printf("\n[%s] %s\n", hostname, message)
	}
}

func inspectFrame(f Frame) {
	p := f.Data
	s := p.Data

	fmt.Printf("\n ========== FRAME ========== \n")
	fmt.Printf("Source MAC:\t%s\n", f.SrcMAC)
	fmt.Printf("Dest MAC:\t%s\n", f.DstMAC)

	fmt.Printf("\n ========== PACKET ========== \n")
	fmt.Printf("Source IP:\t%s\n", p.SrcIP)
	fmt.Printf("Dest IP:\t%s\n", p.DstIP)

	fmt.Printf("\n ========== SEGMENT ========== \n")
	fmt.Printf("Protocol: %s\n", s.Protocol)
	fmt.Printf("Source port: %d, Destination port: %d\n", s.SrcPort, s.DstPort)
	fmt.Printf("Data: %s\n", s.Data)

	fmt.Printf("\n")
}

func sleepDiv() {
	time.Sleep(time.Second)
	fmt.Println("---------------------")
	time.Sleep(time.Second)
}
