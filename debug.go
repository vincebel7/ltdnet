/*
File:		debug.co
Author: 	https://bitbucket.org/vincebel
Purpose:	Functions related to debugging and testing
*/

package main

import (
	"fmt"
	"strconv"
)

/* DEBUG LEVELS
0 - No debugging
1 - Errors
2 - General network traffic
3 - All network traffic and warnings
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
	if snet.DebugLevel >= level {
		hostname := ""
		if generatingID == "Listener" {
			hostname = "Listener"
		} else {
			deviceType := getDeviceType(generatingID)
			if deviceType == "host" {
				if getHostIndexFromID(generatingID) != -1 {
					hostname = snet.Hosts[getHostIndexFromID(generatingID)].Hostname
				} else {
					hostname = generatingID
				}
			} else if deviceType == "switch" {
				if getSwitchIndexFromID(generatingID) != -1 {
					hostname = snet.Switches[getSwitchIndexFromID(generatingID)].Hostname
				} else {
					hostname = generatingID
				}
			} else if deviceType == "vswitch" {
				hostname = snet.Router.VSwitch.Hostname
			} else if deviceType == "router" {
				hostname = snet.Router.Hostname
			} else {
				hostname = generatingID
			}
		}
		//fmt.Printf("\n[%s] (%s), %s\n", hostname, generatingFunc, message)
		fmt.Printf("\n[%s] %s\n", hostname, message)

	}
}

func inspectFrame(frame Frame) {
	frameData := frame.Data

	fmt.Printf("\n ========== FRAME ========== \n")
	fmt.Printf("Source MAC:\t%s\n", frame.SrcMAC)
	fmt.Printf("Dest MAC:\t%s\n", frame.DstMAC)

	fmt.Printf("\n ========== DATA ========== \n")
	fmt.Print(string(frameData))

	fmt.Printf("\n")
}
