/*
File:		client.go
Author: 	https://github.com/vincebel7
Purpose:	General network configuration, main menu+general program functions
*/

package main

import(
	"fmt"
	"encoding/json"
	"os"
	"log"
	"strings"
	"strconv"
	"path/filepath"
)

func mainmenu() {
	fmt.Println("ltdnet v0.2.6")
	fmt.Println("by vincebel\n")
	fmt.Println("Please note that switch functionality is limited and in development\n")
	selection := false
	fmt.Println("Please create or select a network:")
	fmt.Println(" 1) Create new network")
	fmt.Println(" 2) Select saved network")
	for selection == false {
		fmt.Print("\nAction: ")

		scanner.Scan()
		option := scanner.Text()

		if option == "1" || strings.ToUpper(option) == "C" || strings.ToUpper(option) == "NEW" {
			selection = true
			newnetwork()
		} else if option == "2" || strings.ToUpper(option) == "S" || strings.ToUpper(option) == "select" {
			selection = true
			selectnetwork()
		} else {
			fmt.Println("Not a valid option.")
		}
	}
}

func newnetwork() {
	fmt.Println("Creating a new network")
	fmt.Print("Your new network's name: ")
	scanner.Scan()
	netname := scanner.Text()

	fmt.Print("\nYour name: ")
	scanner.Scan()
	user_name := scanner.Text()

	class_valid := false
	network_class := "C"
	for class_valid == false {
		fmt.Print("\nNetwork class (A, B, or C): ")
		scanner.Scan()
		network_class = scanner.Text()
		network_class = strings.ToUpper(network_class)

		if network_class == "A" ||
		network_class == "B" ||
		network_class == "C" {
			class_valid = true
		}
	}

	netid := idgen(8)
	net := Network{
		ID: netid,
		Name: netname,
		Author: user_name,
		Class: network_class,
		DebugLevel: 1,
	}

	marshString, err := json.Marshal(net)
	if err != nil {
		log.Println(err)
	}

	// Write to file
	filename := "saves/" + netname + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("\nNetwork created!")
	loadnetwork(netname)
}

func selectnetwork() {
	fmt.Println("\nPlease select a saved network")

	//display files
	searchDir := "saves/"
	fileList := []string{}
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	i := 0
	option_map := make(map[int]string)
	for _, file := range fileList{
		if(i >= 1){
			file = file[6:] //strip saves
			fmt.Printf(" %d) %s\n", i, file)

			//map i to file somehow for select
			option_map[i] = file
		}
		i = i+1
	}

	fmt.Print("\nLoad: ")
	scanner.Scan()
	network_selection := scanner.Text()
	int_select, err := strconv.Atoi(network_selection)

	for ((network_selection == "") || (int_select >= i) || (int_select < 1)) {
		fmt.Println("Not a valid option.")
		fmt.Print("\nLoad: ")
		scanner.Scan()
		network_selection = scanner.Text()
		int_select, err = strconv.Atoi(network_selection)
	}
	netname := option_map[int_select]
	netname = netname[:len(netname)-len(".json")]

	loadnetwork(netname)
}

func loadnetwork(netname string) {
	//open file
	filename := "saves/" + netname + ".json"
	f, err := os.Open(filename)
	b1 := make([]byte, 1000000) //TODO: secure this
	n1, err := f.Read(b1)

	if err != nil {
		fmt.Printf("File not found: %s", filename)
	}

	//unmarshal
	var net Network
	err2 := json.Unmarshal(b1[:n1], &net)
	if err2 != nil {
		fmt.Printf("err: %v", err2)
	}

	//save global
	snet = net
	fmt.Printf("Loaded %s\n", snet.Name)
}

func linkHost() {
	fmt.Println("Link which host? Please specify by hostname")
	fmt.Print("Available hosts:")
	for availh := range snet.Hosts {
		if len(snet.Hosts[availh].UplinkID) < 1 {
			fmt.Printf(" %s", snet.Hosts[availh].Hostname)
		}
	}
	fmt.Print("\nHostname: ")
	scanner.Scan()
	hostname := scanner.Text()
	hostname = strings.ToUpper(hostname)

	fmt.Println("Uplink to which device? Please specify by hostname")
	fmt.Printf("Router: %s\n", snet.Router.Hostname)
	fmt.Printf("Switches: ")
	for i := range snet.Switches {
		fmt.Printf(snet.Switches[i].Hostname)
	}
	fmt.Printf("\n")

	fmt.Print("Hostname: ")
	scanner.Scan()
	uplinkHostname := scanner.Text()
	uplinkHostname = strings.ToUpper(uplinkHostname)

	//Make sure there's enough ports - if uplink device is a router
	if(uplinkHostname == strings.ToUpper(snet.Router.Hostname)) {
		if(getActivePorts(snet.Router.VSwitch) >= snet.Router.VSwitch.Maxports) {
			fmt.Printf("No available ports - %s only has %d ports\n", snet.Router.Model, snet.Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range snet.Switches {
		if(uplinkHostname == strings.ToUpper(snet.Switches[s].Hostname)) {
			if(getActivePorts(snet.Switches[s]) >= snet.Switches[s].Maxports) {
				fmt.Printf("No available ports - %s only has %d ports\n", snet.Switches[s].Model, snet.Switches[s].Maxports)
				return
			}
		}
	}

	//find host with that hostname
	for i := range snet.Hosts {
		if(strings.ToUpper(snet.Hosts[i].Hostname) == hostname) {
			uplinkID := ""
			//Router
			if uplinkHostname == strings.ToUpper(snet.Router.Hostname) {
				//find next free port
				for k := range snet.Router.VSwitch.Ports {
					if ((snet.Router.VSwitch.Ports[k] == "") && (uplinkID == ""))  {
						uplinkID = snet.Router.VSwitch.PortIDs[k]
					}
				}
				//uplinkID = snet.Router.VSwitch.ID

				assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].ID)
			} else {
				//Search switches
				for j := range snet.Switches {
					if uplinkHostname == strings.ToUpper(snet.Switches[j].Hostname) {

						//find next free port
						for k := range snet.Switches[j].Ports {
							if ((snet.Switches[j].Ports[k] == "") && (uplinkID == "")) {
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

func unlinkHost() {
	fmt.Println("Unlink which host? Please specify by hostname")
	fmt.Print("Linked hosts:")
	for availh := range snet.Hosts {
		if snet.Hosts[availh].UplinkID != "" {
			fmt.Printf(" %s", snet.Hosts[availh].Hostname)
		}
	}
	fmt.Print("\nHostname: ")
	scanner.Scan()
	hostname := scanner.Text()
	hostname = strings.ToUpper(hostname)

	for i := range snet.Hosts {
		if(strings.ToUpper(snet.Hosts[i].Hostname) == hostname) {
			//first, unplug from switch
			freeSwitchport(snet.Hosts[i].UplinkID)

			//next, remove host uplink
			uplinkID := ""
			snet.Hosts[i].UplinkID = uplinkID
			//snet.Router.Ports = removeStringFromSlice(snet.Router.Ports, i)

			return
		}
	}
}

func controlHost(hostname string) {
	fmt.Printf("Attempting to control host %s...\n", hostname)
	host := Host{}
	for i := range snet.Hosts {
		if(snet.Hosts[i].Hostname == hostname){
			host = snet.Hosts[i]
			Conn("host", host.ID)
		}
	}
	if host.Hostname == "" {
		fmt.Println("Host not found")
	}
	return
}

func actions() {
	fmt.Print("> ")
	scanner.Scan()
	action_selection := scanner.Text()
	actionword1 := ""
	actionword2 := ""
	actionword3 := ""
	if action_selection != "" {
		actionword0 := strings.Fields(action_selection)
		if(len(actionword0) > 0){
			actionword1 = actionword0[0]
		}

		if(len(actionword0) > 1) {
			actionword2 = actionword0[1]

			if(len(actionword0) > 2) {
				actionword3 = actionword0[2]
			}
		}
	}
	switch actionword1 {
	case "":

	case "add":
		switch action_selection {
		case "add device router":
			addRouter()
			save()
		case "add device switch":
			addSwitch()
			save()
		case "add device host":
			addHost()
			save()
		default:
			fmt.Println(" Usage: add device <host|switch|router>")
		}
	case "del":
		switch action_selection {
		case "del device router":
			delRouter()
			save()
		case "del device switch":
			delSwitch()
			save()
		case "del device host":
			delHost()
			save()
		default:
			fmt.Println(" Usage: del device <host|switch|router>")
		}
	case "link":
		switch action_selection {
		case "link device host":
			linkHost()
			save()
		default:
			fmt.Println(" Usage: link device host")
		}
	case "unlink":
		switch action_selection {
		case "unlink device host":
			unlinkHost()
			save()
		default:
			fmt.Println(" Usage: unlink device host")
		}
	case "control":
		if actionword2 != "" {
			switch actionword2 {
			case "host":
				controlHost(actionword3)
				save()
			case "router":
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
		loadnetwork(snet.Name)
	case "show":
		switch action_selection {
		case "show network overview":
			overview()
		case "show diagram":
			drawDiagram(snet.Router.ID)
		default:
			if len(action_selection) > 12{
				show(action_selection[12:])
			} else {
				fmt.Println(" Usage: show network overview\n\tshow device <hostname>\n\tshow diagram")
			}
		}
	case "filedump":
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
			"4 - All sorts of garbage (dev use)\n")
		}
	case "exit":
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
		"filedump\t\tOutputs JSON file of loaded network file (developer use)\n",
		"exit\t\t\tExits the program\n",
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
	filename := "saves/" + snet.Name + ".json"
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
	mainmenu()
	go Listener()

	for range snet.Hosts {
		<-listenSync
	}
	fmt.Printf("\n[Notice] Debug level is set to %d\n", getDebug())
	fmt.Println("\nPlease type an action:")

	for true {
		actions()
	}
}
