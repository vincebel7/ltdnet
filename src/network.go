/*
File:		network.go
Author: 	https://github.com/vincebel7
Purpose: 	Network-level functions
*/

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Network struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
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

func newNetworkPrompt() {
	fmt.Println("Creating a new network")

	var netname = ""
	for {
		fmt.Print("Your new network's name: ")
		scanner.Scan()
		netname = scanner.Text()

		// Check if file already exists
		filename := "ltdnet_saves/user_saves/" + netname + ".json"
		if _, err := os.Stat(filename); err == nil {
			// File exists
			fmt.Println("\nError: A network with this name already exists!")

		} else if !os.IsNotExist(err) {
			// Some other error occurred
			log.Fatal(err)
		} else if netname == "" {
			fmt.Println("\nError: Network name cannot be blank.")
		} else {
			break
		}
	}

	class_valid := false
	networkPrefix := "24"
	for !class_valid {
		fmt.Print("\nNetwork size (/24, /16, or /8)")
		fmt.Print("\nChoose /24 if you are unsure.")
		fmt.Print("\nNetwork size: /")
		scanner.Scan()
		networkPrefix = scanner.Text()
		networkPrefix = strings.ToUpper(networkPrefix)

		if networkPrefix == "24" ||
			networkPrefix == "16" ||
			networkPrefix == "8" {
			class_valid = true
		}
	}

	newNetwork(netname, networkPrefix, "user")
}

func newNetwork(netname string, networkPrefix string, saveType string) {
	netid := idgen(8)
	net := Network{
		ID:         netid,
		Name:       netname,
		Netsize:    networkPrefix,
		ProgramVer: currentVersion,
		DebugLevel: 1,
	}

	marshString, err := json.Marshal(net)
	if err != nil {
		log.Println(err)
	}

	// Determine the file path
	filename := ""
	if saveType == "user" {
		filename = "ltdnet_saves/user_saves/" + netname + ".json"
	} else if saveType == "test" {
		filename = "../ltdnet_saves/test_saves/" + netname + ".json"
	}

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("\nNetwork created!")
	loadNetwork(netname, saveType)
}

func selectNetwork() {
	fmt.Println("\nPlease select a saved network")

	//display files
	searchDir := "ltdnet_saves/user_saves/"
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	i := 1
	option_map := make(map[int]string)
	for _, file := range fileList {
		file = file[24:] //strip "ltdnet_saves/user_saves/"
		if (file != ".keep") && (file != "") {
			fmt.Printf(" %d) %s\n", i, file)

			//map i to file somehow for select
			option_map[i] = file

			i = i + 1
		}
	}

	if i == 1 {
		fmt.Println("No networks to load. Try creating a new one.")
		time.Sleep(1 * time.Second)
		startMenu()
		return
	}

	fmt.Print("\nLoad: ")
	scanner.Scan()
	network_selection := scanner.Text()
	int_select, _ := strconv.Atoi(network_selection)

	for (network_selection == "") || (int_select >= i) || (int_select < 1) {
		fmt.Println("Not a valid option.")
		fmt.Print("\nLoad: ")
		scanner.Scan()
		network_selection = scanner.Text()
		int_select, _ = strconv.Atoi(network_selection)
	}
	netname := option_map[int_select]
	netname = netname[:len(netname)-len(".json")]

	loadNetwork(netname, "user")
}

func loadNetwork(netname string, saveType string) {
	//open file
	filename := ""
	if saveType == "user" {
		filename = "ltdnet_saves/user_saves/" + netname + ".json"
	} else if saveType == "test" {
		filename = "../ltdnet_saves/test_saves/" + netname + ".json"
	}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("File not found: %s", filename)
	}

	b1 := make([]byte, 1000000) //TODO: secure this
	n1, err := f.Read(b1)

	if err != nil {
		fmt.Printf("File not found: %s", filename)
	}

	//unmarshal
	var net Network
	err = json.Unmarshal(b1[:n1], &net)
	if err != nil {
		fmt.Printf("err: %v", err)
	}

	// Version check
	migrate := false
	if net.ProgramVer != currentVersion {
		fmt.Print("The selected save file was created in an older version. Attempt migrating? [y/N]: ")

		scanner.Scan()
		migrateSelection := strings.ToUpper(scanner.Text())

		switch migrateSelection {
		case "Y", "YES":
			// "Migrate". Will make this more intelligent eventually
			net.ProgramVer = currentVersion
			migrate = true
		default:
			startMenu()
			return
		}
	}

	// Clear MAC tables (ARP, MAC address table) on new launch
	net.ClearMACTables()

	//save global
	snet = net

	// Save successful version migration
	if migrate {
		save()
	}

	fmt.Printf("Loaded \"%s\"\n", snet.Name)
}

func save() {
	marshString, err := json.Marshal(snet)
	if err != nil {
		log.Println(err)
	}
	//Write to file
	filename := "ltdnet_saves/user_saves/" + snet.Name + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	os.Truncate(filename, int64(len(marshString)))
	fmt.Println("Network saved")
}

func (n *Network) ClearMACTables() {
	// Host ARP tables
	for i := range n.Hosts {
		n.Hosts[i].ARPTable = make(map[string]ARPEntry)
	}

	// Router ARP table
	n.Router.ARPTable = make(map[string]ARPEntry)

	// Switch MAC address tables
	for i := range n.Switches {
		n.Switches[i].MACTable = make(map[string]MACEntry)
	}

	// VSwitch MAC address table
	n.Router.VSwitch.MACTable = make(map[string]MACEntry)
}
