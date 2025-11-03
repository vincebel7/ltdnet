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
1 - ERROR - Errors
2 - INFO - General info
3 - TRAF - General network traffic (like packet capture)
4 - DEBUG - All network traffic, plus warnings, debug output
5 - TRACE - For tracing function calls, network decision making, etc
*/

func setLogLevel(val string) {
	intval, _ := strconv.Atoi(val)

	if intval < 2 || intval > 5 {
		fmt.Println("Invalid log level. Please enter a value between 2 and 5.")
		return
	}

	// Update current logger
	engine.Instance().Logger.SetLevel(logging.Level(intval))

	// Write to settings
	engine.UserSettings().LogLevel = intval
	saveUserSettings()

	fmt.Printf("Log level set to %d\n", getLogLevel())
}

func getLogLevel() int {
	return int(engine.Instance().Logger.GetLevel())
}

// Wrapper to get hostname and determine log level
func deviceLog(level int, generatingFunc string, generatingID string, message string) {
	hostname := ""
	if generatingID == "Listener" {
		hostname = "Listener"
	} else {
		hostname = getHostnameFromID(generatingID)
	}

	logger := engine.Instance().Logger
	logger.Log(logging.Level(level), generatingFunc, "[%s] %s", hostname, message)
}

func systemLog(level int, generatingFunc string, message string) {
	logger := engine.Instance().Logger
	logger.Log(logging.Level(level), generatingFunc, "%s", message)
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
