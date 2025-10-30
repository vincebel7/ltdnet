/*
File:		debug.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to debugging and testing
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/vincebel7/ltdnet/src/engine"
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

	hostname := ""
	if generatingID == "Listener" {
		hostname = "Listener"
	} else {
		hostname = getHostnameFromID(generatingID)
	}
	//fmt.Printf("\n[%s] (%s), %s\n", hostname, generatingFunc, message)
	//fmt.Printf("\n[%s] %s\n", hostname, message)

	logger := engine.Instance().Logger
	switch level {
	case 1:
		logger.Error(generatingFunc, "[%s] %s", hostname, message)
	case 2:
		logger.Info(generatingFunc, "[%s] %s", hostname, message)
	case 3:
		logger.Debug(generatingFunc, "[%s] %s", hostname, message)
	case 4:
		logger.Trace(generatingFunc, "[%s] %s", hostname, message)
	default:
		// Do nothing
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
