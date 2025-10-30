/*
File:		debug.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to debugging and testing
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/vincebel7/ltdnet/src/model"
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
	Net().DebugLevel = intval
	fmt.Printf("Debug level set to %d\n", Net().DebugLevel)
}

func getDebug() int {
	return Net().DebugLevel
}

func debug(level int, generatingFunc string, generatingID string, message string) {
	if Net().DebugLevel >= level {
		hostname := ""
		if generatingID == "Listener" {
			hostname = "Listener"
		} else {
			deviceType := getDeviceType(generatingID)
			if deviceType == "host" {
				if getHostIndexFromID(generatingID) != -1 {
					hostname = Net().Hosts[getHostIndexFromID(generatingID)].Hostname
				} else {
					hostname = generatingID
				}
			} else if deviceType == "switch" {
				if getSwitchIndexFromID(generatingID) != -1 {
					hostname = Net().Switches[getSwitchIndexFromID(generatingID)].Hostname
				} else {
					hostname = generatingID
				}
			} else if deviceType == "vswitch" {
				hostname = Net().Router.VSwitch.Hostname
			} else if deviceType == "router" {
				hostname = Net().Router.Hostname
			} else {
				hostname = generatingID
			}
		}
		//fmt.Printf("\n[%s] (%s), %s\n", hostname, generatingFunc, message)
		fmt.Printf("\n[%s] %s\n", hostname, message)

	}
}

func inspectFrame(frame model.Frame) {
	frameData := frame.Data

	fmt.Printf("\n ========== FRAME ========== \n")
	fmt.Printf("Source MAC:\t%s\n", frame.SrcMAC)
	fmt.Printf("Dest MAC:\t%s\n", frame.DstMAC)

	fmt.Printf("\n ========== DATA ========== \n")
	fmt.Print(string(frameData))

	fmt.Printf("\n")
}
