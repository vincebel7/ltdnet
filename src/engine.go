package main

import (
	"bufio"
	"encoding/json"
	"os"
)

type Engine struct {
	Net        *Network
	Channels   map[string]chan json.RawMessage  // Physical links
	Sockets    map[string]map[string]chan Frame // For internal device communication
	ActionSync map[string]chan int
	ListenSync chan string
	Scanner    *bufio.Scanner
	ProgramVer string
}

func NewEngine() *Engine {
	return &Engine{
		Net:        &Network{},
		Channels:   make(map[string]chan json.RawMessage),
		Sockets:    make(map[string]map[string]chan Frame),
		ActionSync: make(map[string]chan int),
		ListenSync: make(chan string),
		Scanner:    bufio.NewScanner(os.Stdin),
		ProgramVer: ProgramVersion,
	}
}

var defaultEngine = NewEngine()

func EngineInstance() *Engine { return defaultEngine }
func Net() *Network           { return defaultEngine.Net }
