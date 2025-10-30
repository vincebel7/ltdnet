package engine

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/vincebel7/ltdnet/src/model"
	"github.com/vincebel7/ltdnet/src/version"
)

type Engine struct {
	Net             *model.Network
	Channels        map[string]chan json.RawMessage        // Physical links
	Sockets         map[string]map[string]chan model.Frame // For internal device communication
	ActionSync      map[string]chan int                    // Blocks CLI prompt until action completes
	ListenSync      chan string                            // Synchronizes listener goroutines with main CLI (client.go)
	Scanner         *bufio.Scanner
	ProgramVer      string
	AchievementsMap map[int]model.Achievement // catalog
	Settings        *model.Settings
}

func NewEngine() *Engine {
	return &Engine{
		Net:             &model.Network{},
		Channels:        make(map[string]chan json.RawMessage),
		Sockets:         make(map[string]map[string]chan model.Frame),
		ActionSync:      make(map[string]chan int),
		ListenSync:      make(chan string),
		Scanner:         bufio.NewScanner(os.Stdin),
		ProgramVer:      version.ProgramVersion,
		AchievementsMap: make(map[int]model.Achievement),
		Settings:        &model.Settings{Achievements: make(map[int]model.Achievement), AchievementsOn: true},
	}
}

var defaultEngine = NewEngine()

func Instance() *Engine   { return defaultEngine }
func Net() *model.Network { return defaultEngine.Net }
