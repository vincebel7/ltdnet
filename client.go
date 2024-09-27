/*
File:		client.go
Author: 	https://github.com/vincebel7
Purpose:	General network configuration, main menu+general program functions
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func intro() {
	fmt.Println("ltdnet v0.2.9")
	fmt.Println("by vincebel")
	fmt.Println("\nPlease note that switch functionality is limited and in development")
}

func startMenu() {
	selection := false
	fmt.Println("\nPlease create or select a network:")
	fmt.Println(" 1) Create new network")
	fmt.Println(" 2) Select saved network")
	for !selection {
		fmt.Print("\nAction: ")

		scanner.Scan()
		option := scanner.Text()

		switch strings.ToUpper(option) {
		case "1", "C", "NEW", "CREATE":
			selection = true
			newNetwork()
		case "2", "S", "SELECT":
			selection = true
			selectNetwork()
		default:
			fmt.Println("Not a valid option. Options: 1, 2")
		}
	}
}

func newNetwork() {
	fmt.Println("Creating a new network")
	fmt.Print("Your new network's name: ")
	scanner.Scan()
	netname := scanner.Text()

	fmt.Print("\nYour name: ")
	scanner.Scan()
	user_name := scanner.Text()

	class_valid := false
	network_snmask := "24"
	for !class_valid {
		fmt.Print("\nNetwork size (/24, /16, or /8): /")
		scanner.Scan()
		network_snmask = scanner.Text()
		network_snmask = strings.ToUpper(network_snmask)

		if network_snmask == "24" ||
			network_snmask == "16" ||
			network_snmask == "8" {
			class_valid = true
		}
	}

	netid := idgen(8)
	net := Network{
		ID:         netid,
		Name:       netname,
		Author:     user_name,
		Netsize:    network_snmask,
		DebugLevel: 1,
	}

	marshString, err := json.Marshal(net)
	if err != nil {
		log.Println(err)
	}

	// Write to file
	filename := "saves/user_saves/" + netname + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("\nNetwork created!")
	loadNetwork(netname)
}

func selectNetwork() {
	fmt.Println("\nPlease select a saved network")

	//display files
	searchDir := "saves/user_saves/"
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
		file = file[17:] //strip "saves/user_saves/"
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

	loadNetwork(netname)
}

func loadNetwork(netname string) {
	//open file
	filename := "saves/user_saves/" + netname + ".json"
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

	//save global
	snet = net
	fmt.Printf("Loaded %s\n", snet.Name)
}

func linkHost(localDevice string, remoteDevice string) {
	localDevice = strings.ToUpper(localDevice)
	remoteDevice = strings.ToUpper(remoteDevice)

	//Make sure there's enough ports - if uplink device is a router
	if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
		if getActivePorts(snet.Router.VSwitch) >= snet.Router.VSwitch.Maxports {
			fmt.Printf("No available ports - %s only has %d ports\n", snet.Router.Model, snet.Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range snet.Switches {
		if remoteDevice == strings.ToUpper(snet.Switches[s].Hostname) {
			if getActivePorts(snet.Switches[s]) >= snet.Switches[s].Maxports {
				fmt.Printf("No available ports - %s only has %d ports\n", snet.Switches[s].Model, snet.Switches[s].Maxports)
				return
			}
		}
	}

	//find host with that hostname
	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == localDevice {
			uplinkID := ""
			//Remote device on new link is the Router
			if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
				//find next free port
				for k := range snet.Router.VSwitch.Ports {
					if (snet.Router.VSwitch.Ports[k] == "") && (uplinkID == "") {
						uplinkID = snet.Router.VSwitch.PortIDs[k]
					}
				}
				//uplinkID = snet.Router.VSwitch.ID

				assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].ID)
			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range snet.Switches {
					if remoteDevice == strings.ToUpper(snet.Switches[j].Hostname) {

						//find next free port
						for k := range snet.Switches[j].Ports {
							if (snet.Switches[j].Ports[k] == "") && (uplinkID == "") {
								uplinkID = snet.Switches[j].PortIDs[k]
								k = len(snet.Switches[j].Ports)
								fmt.Println("DEBUG TEST")
							}
						}
					}
				}
			}

			snet.Hosts[i].UplinkID = uplinkID
			return
		}
	}
}

func unlinkHost(hostname string) {
	hostname = strings.ToUpper(hostname)

	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == hostname {
			//first, unplug from switch (switch-end unlink). TODO try/catch this whole block.
			freeSwitchport(snet.Hosts[i].UplinkID)

			//next, remove the host's uplink (host-end unlink)
			uplinkID := ""
			snet.Hosts[i].UplinkID = uplinkID
			//snet.Router.Ports = removeStringFromSlice(snet.Router.Ports, i)

			return
		}
	}
}

func actionsMenu() {
	fmt.Print("> ")
	scanner.Scan()
	action_selection := scanner.Text()
	actionword1 := ""
	actionword2 := ""
	actionword3 := ""
	actionword4 := ""
	if action_selection != "" {
		actionword0 := strings.Fields(action_selection)
		if len(actionword0) > 0 {
			actionword1 = actionword0[0]
		}

		if len(actionword0) > 1 {
			actionword2 = actionword0[1]

			if len(actionword0) > 2 {
				actionword3 = actionword0[2]

				if len(actionword0) > 3 {
					actionword4 = actionword0[3]
				}
			}
		}
	}
	switch actionword1 {
	case "":

	case "add":
		if actionword3 != "" {
			switch actionword2 {
			case "router":
				if actionword4 == "" {
					actionword4 = "Bobcat"
				}
				addRouter(actionword3, actionword4)
				save()

			case "switch":
				addSwitch(actionword3)
				save()

			case "host":
				addHost(actionword3)
				save()

			default:
				fmt.Println(" Usage: add <host|switch|router> <hostname>")
			}
		} else {
			fmt.Println(" Usage: add <host|switch|router> <hostname>")
		}

	case "del", "delete":
		switch actionword2 {
		case "router":
			fmt.Printf("\nAre you sure you want do delete router %s? [y/n]: ", actionword3)
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.ToUpper(confirmation)
			if confirmation == "Y" {
				delRouter(actionword3) //Placeholder. Only one router per network currently
				save()
			}

		case "switch":
			if actionword3 != "" {
				fmt.Printf("\nAre you sure you want do delete switch %s? [y/n]: ", actionword3)
				scanner.Scan()
				confirmation := scanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delSwitch()
					save()
				}
			}

		case "host":
			if actionword3 != "" {
				fmt.Printf("\nAre you sure you want do delete host %s? [y/n]: ", actionword3)
				scanner.Scan()
				confirmation := scanner.Text()
				confirmation = strings.ToUpper(confirmation)
				if confirmation == "Y" {
					delHost(actionword3)
					save()
				}
			}

		default:
			fmt.Println(" Usage: del <host|switch|router> <hostname>")
		}

	case "link":
		if (actionword2 == "host") && (actionword3 != "") && (actionword4 != "") {
			linkHost(actionword3, actionword4)
			save()
		} else {
			fmt.Println(" Usage: link host <hostname> <router_hostname>")
		}

	case "unlink":
		if (actionword2 == "host") && (actionword3 != "") {
			unlinkHost(actionword3)
			save()
		} else {
			fmt.Println(" Usage: unlink host <hostname>")
		}

	case "control":
		if actionword2 != "" {
			switch actionword2 {
			case "host":
				controlHost(actionword3)
				save()

			case "router":
				controlRouter(actionword3)
				save()

			default:
				fmt.Println(" Usage: control <host|switch|router> <hostname>")
			}
		} else {
			fmt.Println(" Usage: control <host|switch|router> <hostname>")
		}

	case "save":
		save()

	case "reload":
		loadNetwork(snet.Name)

	case "show", "sh":
		switch action_selection {
		case "show network overview", "sh network overview":
			overview()

		case "show diagram":
			drawDiagram(snet.Router.ID)

		default:
			if len(action_selection) > 12 { // show device
				show(action_selection[12:])
			} else {
				fmt.Println(" Usage: show network overview\n\tshow device <hostname>\n\tshow diagram")
			}
		}

	case "netdump":
		fmt.Println(snet, "")

	case "debug":
		if actionword2 != "" {
			setDebug(actionword2)
			save()
		} else {
			fmt.Printf("Current debug level: %d\n", getDebug())
			fmt.Println("\nAll levels:\n",
				"0 - No debugging\n",
				"1 - Errors\n",
				"2 - General network traffic\n",
				"3 - All network traffic\n",
				"4 - All sorts of garbage (dev use)")
		}

	case "manual", "man":
		launchManual()

	case "exit", "quit", "q":
		os.Exit(0)

	case "help", "?":
		fmt.Println("",
			"show <args>\t\tDisplays information\n",
			"add <args>\t\tAdds device to network\n",
			"del <args>\t\tRemoves device from network\n",
			"link <args>\t\tLinks two devices\n",
			"unlink <args>\t\tUnlinks two devices\n",
			"control <args>\t\tLogs in as device\n",
			"save\t\t\tManually saves network changes\n",
			"reload\t\t\tReloads the network file (May fix runtime bugs)\n",
			"debug <0-4>\t\tSets debug level. Default is 1\n",
			"manual\t\t\tLaunches the user manual. Great for beginners!\n",
			"exit\t\t\tExits the program",
			//"filedump\t\tPrints loaded Network object (developer use)\n", HIDDEN
		)

	default:
		fmt.Println(" Invalid command. Type 'help' for a list of commands.")
	}
}

func save() {
	marshString, err := json.Marshal(snet)
	if err != nil {
		log.Println(err)
	}
	//Write to file
	filename := "saves/user_saves/" + snet.Name + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	os.Truncate(filename, int64(len(marshString)))
	fmt.Println("Network saved")
	//loadnetwork(snet.Name)
}

func main() {
	intro()
	startMenu()
	go Listener()

	for range snet.Hosts {
		<-listenSync
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Println("\nPlease type an action:")

	for {
		actionsMenu()
	}
}
