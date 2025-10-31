/*
File:		log_wrapper.go
Author: 	https://github.com/vincebel7
Purpose:	Functions related to logging and testing
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/logging"
	"github.com/vincebel7/ltdnet/src/model"
)

/* LOG LEVELS
0 - No logging
1 - Errors
2 - General network traffic
3 - All network traffic and warnings
4 - Garbage
*/

func setLogLevel(val string) {
	intval, _ := strconv.Atoi(val)

	// Update current logger
	engine.Instance().Logger.SetLevel(logging.Level(intval))

	// Write to settings
	UserSettings().LogLevel = intval
	saveUserSettings()

	fmt.Printf("Log level set to %d\n", getLogLevel())
}

func getLogLevel() int {
	return int(engine.Instance().Logger.GetLevel())
}

// Wrapper to get hostname and determine log level
func writeLog(level int, generatingFunc string, generatingID string, message string) {

	hostname := ""
	if generatingID == "Listener" {
		hostname = "Listener"
	} else {
		hostname = getHostnameFromID(generatingID)
	}

	logger := engine.Instance().Logger
	logger.Log(logging.Level(level), generatingFunc, "[%s] %s", hostname, message)
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
