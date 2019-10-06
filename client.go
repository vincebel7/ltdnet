/*
File:		client.go
Author: 	https://bitbucket.org/vincebel
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
	fmt.Println("ltdnet v0.2.2")
	fmt.Println("by vincebel\n")

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
	}

	// Print to demonstrate
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

func NewBobcat(hostname string) Router {
	b := Router{}
	b.ID = idgen(8)
	b.Model = "Bobcat 100"
	b.MACAddr = macgen()
	b.Hostname = hostname
	b.DHCPPool = 253
	b.Downports = 4

	return b
}

func NewOsiris(hostname string) Router {
	o := Router{}
	o.ID = idgen(8)
	o.Model = "Osiris 2-I"
	o.MACAddr = macgen()
	o.Hostname = hostname
	o.DHCPPool = 2
	o.Downports = 2

	return o
}

func NewSumerian2100(hostname string) Switch {
	s := Switch{}
	s.ID = idgen(8)
	s.Model = "Sumerian 2100"
	s.MACAddr = macgen()
	s.Hostname = hostname
	return s
}

func NewProbox(hostname string) Host {
	p := Host{}
	p.ID = idgen(8)
	p.Model = "ProBox 1"
	p.MACAddr = macgen()
	p.Hostname = hostname

	return p
}

func addRouter() {
	fmt.Println("What model?")
	fmt.Println("Available: Bobcat, Osiris")
	fmt.Print("Model: ")
	scanner.Scan()
	routerModel := scanner.Text()
	routerModel = strings.ToUpper(routerModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	routerHostname := scanner.Text()

	// input validation

	if routerHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

	if hostname_exists(routerHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	r := Router{}

	if routerModel == "BOBCAT" {
		r = NewBobcat(routerHostname)
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(routerHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	if snet.Class == "A" {
		r.Gateway = "10.0.0.1"
	} else if snet.Class == "B" {
		r.Gateway = "172.16.0.1"
	} else if snet.Class == "C" {
		r.Gateway = "192.168.0.1"
	}
	addrconstruct := ""

	network_portion := strings.TrimSuffix(r.Gateway, "1")

	r.DHCPTable = make(map[string]string)

	for k := 2; k < (r.DHCPPool + 2); k++ {
		r.DHCPIndex = append(r.DHCPIndex, strconv.Itoa(k))
	}

	for i := 0; i < len(r.DHCPIndex); i++ {
		addrconstruct = network_portion + r.DHCPIndex[i]
		r.DHCPTable[addrconstruct] = ""
	}

	snet.Router = r
}

func delRouter() {
	fmt.Printf("\nAre you sure you want do delete router %s? [Y/n]: ", snet.Router.Hostname)
	scanner.Scan()
	confirmation := scanner.Text()
	confirmation = strings.ToUpper(confirmation)
	if confirmation == "Y" {
		r := Router{}

		r.ID = ""
		r.Model = ""
		r.MACAddr = ""
		r.Hostname = ""
		r.DHCPPool = 0
		r.Downports = 0

		snet.Router = r
		fmt.Printf("\nRouter deleted\n")
	} else {
		fmt.Printf("\nRouter %s was not deleted.\n", snet.Router.Hostname)
	}
}

func addSwitch() {
	fmt.Println("What model?")
	fmt.Println("Available: Sumerian 2100")
	fmt.Print("Model: ")
	scanner.Scan()
	switchModel := scanner.Text()
	switchModel = strings.ToUpper(switchModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	switchHostname := scanner.Text()

	// input validation
	if switchHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

	if hostname_exists(switchHostname) { //TODO make hostname_exists check switches
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	s := Switch{}
	if switchModel == "SUMERIAN 2100" {
		s = NewSumerian2100(switchHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	snet.Switches = append(snet.Switches, s)
}

func delSwitch() {}

func addHost() {
	fmt.Println("What model?")
	fmt.Println("Available: ProBox")
	fmt.Print("Model: ")
	scanner.Scan()
	hostModel := scanner.Text()
	hostModel = strings.ToUpper(hostModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	hostHostname := scanner.Text()

	// input validation
	if hostHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

	if hostname_exists(hostHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	h := Host{}
	if hostModel == "PROBOX" {
		h = NewProbox(hostHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	h.IPAddr = "0.0.0.0"

	snet.Hosts = append(snet.Hosts, h)

	generateHostChannels(getHostIndexFromID(h.ID))
	<-listenSync
}

func delHost() {
	fmt.Println("Not implemented yet, sorry") //TODO
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

	//find host with that hostname
	for i := range snet.Hosts {
		if(strings.ToUpper(snet.Hosts[i].Hostname) == hostname) {
			uplinkID := ""
			//Router
			if uplinkHostname == strings.ToUpper(snet.Router.Hostname) {
				uplinkID = snet.Router.ID
			} else {
				//Search switches
				for j := range snet.Switches {
					if uplinkHostname == strings.ToUpper(snet.Switches[j].Hostname) {
						uplinkID = snet.Switches[j].ID
						fmt.Println("DEBUG TEST")
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
			uplinkID := ""
			snet.Hosts[i].UplinkID = uplinkID
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

func drawDiagram(rootID string) { // TODO make recursive
	// Identify device info about rootID
	rootHostname := ""
	rootType := ""
	//rootIndex := -1
	if(rootID == snet.Router.ID) {
		rootHostname = snet.Router.Hostname
		rootType = "router"
	}

	if(rootType == "") {
		for i := range snet.Switches {
			if(rootID == snet.Switches[i].ID) {
				rootHostname = snet.Switches[i].Hostname
				rootType = "switch"
				//rootIndex = i
			}
		}
	}

	if(rootType == "") {
		for i := range snet.Hosts {
			if(rootID == snet.Hosts[i].ID) {
				rootHostname = snet.Hosts[i].Hostname
				rootType = "host"
				//rootIndex = i
			}
		}
	}

	// ROUTER
	if(rootType == "router"){
	if(rootHostname != "") {
		space1 := 13 - len(snet.Router.Hostname)
		space2 := 14 - len(snet.Router.Gateway)
		space3 := 16 - len(snet.Router.Model)

		fmt.Println("|------------------------|")
		fmt.Println("|         Router         |")
		 fmt.Printf("| Hostname: %s", snet.Router.Hostname)
		for i := 0; i < space1; i++ { fmt.Printf(" ") }
		 fmt.Printf("|\n| Gateway: %s", snet.Router.Gateway)
		for i := 0; i < space2; i++ { fmt.Printf(" ") }
		 fmt.Printf("|\n| Model: %s", snet.Router.Model)
		for i := 0; i < space3; i++ { fmt.Printf(" ") }
		fmt.Println("|\n|------------------------|")
	}
	}

	hostCount := len(snet.Hosts)
	for i := 0; i < hostCount; i++ {
		h := snet.Hosts[i]
		if(h.UplinkID == snet.Router.ID) {
			space1 := 13 - len(h.Hostname)
			space2 := 14 - len(h.IPAddr)
			space3 := 16 - len(h.Model)

			fmt.Println("           ||")
			fmt.Println("           ||      |------------------------|")
			fmt.Println("           ||      |          Host          |")
			 fmt.Printf("           ||------| Hostname: %s", h.Hostname)
			for i := 0; i < space1; i++ { fmt.Printf(" ") }
			 fmt.Printf("|\n")
			 fmt.Printf("           ||------| IP Addr: %s", h.IPAddr)
			for i := 0; i < space2; i++ { fmt.Printf(" ") }
			 fmt.Printf("|\n")

			if(i == hostCount - 1) {
			 fmt.Printf("                   | Model: %s", h.Model)
			} else  {
			 fmt.Printf("           ||      | Model: %s", h.Model)
			}
			for i := 0; i < space3; i++ { fmt.Printf(" ") }
			 fmt.Printf("|\n")
			if(i == hostCount - 1) {
			 fmt.Println("                   |------------------------|")
			} else {
			 fmt.Println("           ||      |------------------------|")
			}
		}
	}

	//Unlinked hosts
}

func overview() {
	fmt.Printf("Network name:\t\t%s\n", snet.Name)
	fmt.Printf("Network ID:\t\t%s\n", snet.ID)
	fmt.Printf("Network class:\t\tClass %s\n", snet.Class)

	// router
	routerCount := 1
	show(snet.Router.Hostname)

	//switches
	switchCount := 0
	for i := 0; i < len(snet.Switches); i++ {
		fmt.Printf("\nSwitch %v\n", snet.Switches[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Switches[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Switches[i].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Switches[i].MACAddr)
		fmt.Printf("\tMgmt IP:\t%s\n", snet.Switches[i].MgmtIP)
		switchCount = i + 1
	}

	//hosts
	hostCount := 0
	for i := 0; i < len(snet.Hosts); i++ {
		fmt.Printf("\nHost %v\n", snet.Hosts[i].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[i].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[i].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[i].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[i].IPAddr)
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[i].DefaultGateway)
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[i].SubnetMask)
		uplinkHostname := ""
		for dev := range snet.Hosts {
			//Router
			if(snet.Hosts[i].UplinkID == snet.Router.ID) {
				uplinkHostname = snet.Router.Hostname
			}
			//TODO: Switches
			//Hosts (pointless since host cant be uplink, just here to show how to do switches)
			if(snet.Hosts[i].UplinkID == snet.Hosts[dev].ID) {
				uplinkHostname = snet.Hosts[dev].Hostname
			}
		}
		fmt.Printf("\tUplink to:\t%s\n", uplinkHostname)
		hostCount = i + 1
	}

	fmt.Printf("\nTotal devices: %d (%d Router, %d Switches, %d Hosts)\n", (routerCount + switchCount + hostCount), routerCount, switchCount, hostCount)
}

func show(hostname string) {
	device_type := "host"
	//TODO search switches
	if(snet.Router.Hostname == hostname) {
		device_type = "router"
	}

	id := -1
	for i := range snet.Hosts {
		if snet.Hosts[i].Hostname == hostname {
			device_type = "host"
			id = i
		}
	}

	for i := range snet.Switches {
		if snet.Switches[i].Hostname == hostname {
			device_type = "switch"
			id = i
		}
	}

	if id == -1 {
			fmt.Printf("Hostname not found\n")
			return
		}

	if device_type == "host" {
		fmt.Printf("\nHost %v\n", snet.Hosts[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[id].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[id].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[id].IPAddr)
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[id].DefaultGateway)
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[id].SubnetMask)
		uplinkHostname := ""
		if(snet.Hosts[id].UplinkID == snet.Router.ID) {
			uplinkHostname = snet.Router.Hostname
		}
		for i := range snet.Switches {
			if(snet.Hosts[id].UplinkID == snet.Switches[i].ID) {
				uplinkHostname = snet.Switches[i].Hostname
			}
		}
		//fmt.Printf("\tUplink ID:\t%s\n", snet.Hosts[i].UplinkID)
		fmt.Printf("\tUplink to:\t%s\n\n", uplinkHostname)
	} else if device_type == "switch" {
		fmt.Printf("\nSwitch %s\n", snet.Switches[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Switches[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Switches[id].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Switches[id].MACAddr)
		fmt.Printf("\tMgmt IP:\t%s\n\n", snet.Switches[id].MgmtIP)
	} else if device_type == "router" {
		fmt.Printf("\nRouter %s\n", snet.Router.Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Router.ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Router.Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Router.MACAddr)
		fmt.Printf("\tGateway:\t%s\n", snet.Router.Gateway)
		fmt.Printf("\tDHCP pool:\t%d addresses\n", snet.Router.DHCPPool)
		fmt.Printf("\tDHCP index:\t%d addresses\n", len(snet.Router.DHCPIndex))
		fmt.Printf("\tUser ports:\t%d ports\n\n", snet.Router.Downports)
	}
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
		fmt.Println(snet)
	case "debug":
		if actionword2 != "" {
			setDebug(actionword2)
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
	case "help":
		fmt.Println("",
		"show <args>\t\tDisplays information\n",
		"add <args>\t\tAdds device to network\n",
		"del <args>\t\tRemoves device from network\n",
		"link <args>\t\tLinks two devices\n",
		"unlink <args>\t\tUnlinks two devices\n",
		"control <args>\t\tLogs in as device\n",
		"save\t\t\tManually saves network changes\n",
		"reload\t\t\tReloads the network file (May fix runtime bugs)\n",
		"debug <0-4>\t\tSets debug level. Defaults to 1 each runtime\n",
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
