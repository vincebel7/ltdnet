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
}

var snet Network //selected network, essentially the loaded save file
var listenSync = make(chan string)
var scanner = bufio.NewScanner(os.Stdin)
