/*
File:		network.go
Author: 	https://github.com/vincebel7
Purpose: 	Network object
*/

package main

import (
	"bufio"
	"os"
)

type Network struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Author     string   `json:"author"`
	Netsize    string   `json:"netsize"`
	Router     Router   `json:"router"`
	Switches   []Switch `json:"switches"`
	Hosts      []Host   `json:"hosts"`
	DebugLevel int      `json:"debug_level"`
	ProgramVer string   `json:"program_ver"`
}

var snet Network //selected network, essentially the loaded save file
var listenSync = make(chan string)
var scanner = bufio.NewScanner(os.Stdin)
