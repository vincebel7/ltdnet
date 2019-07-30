package main

import(
	"fmt"
	"encoding/json"
	"os"
	"log"
	//"sort"
	//"math/rand"
	"time"
	"bufio"
	"strings"
	"strconv"
	"path/filepath"
	"io/ioutil"
	"net"
)

type Network struct {
	ID		string `json:"id"`
	Name		string `json:"name"`
	Author		string `json:"auth"`
	Class		string `json:"clas"`
	Router		Router `json:"rtr"`
	Hosts		[]Host `json:"hsts"`
}

var snet Network //selected network, essentially the loaded save file
var listenSync = make(chan int)
var scanner = bufio.NewScanner(os.Stdin)

type Router struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"`
	Hostname	string `json:"hnme"`
	Gateway		string `json:"gway"`
	DHCPPool	int `json:"dpol"` //maximum, not just available
	Downports	int `json:"dpts"`
	MACTable	map[string]string `json:"mact"`
	DHCPIndex	[]string `json:"dhci"`
	DHCPTable	map[string]string `json:"dhct"` //maps IP address to MAC address
}

type Host struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"`
	Hostname	string `json:"hnme"`
	IPAddr		string `json:"ipa"`
	SubnetMask	string `json:"mask"`
	DefaultGateway	string `json:"gway"`
	UplinkID	string `json:"upid"`
}

func mainmenu() {
	fmt.Println("ltdnet v0.1.5")

	selection := false
		for selection == false {
		fmt.Println("Please create or select a network:")
		fmt.Println("1) Create new network")
		fmt.Println("2) Select saved network")
		scanner.Scan()
		option := scanner.Text()

		if option == "1" || strings.ToUpper(option) == "C" || strings.ToUpper(option) == "NEW" {
			selection = true
			newnetwork()
		} else if option == "2" || strings.ToUpper(option) == "S" || strings.ToUpper(option) == "select" {
			selection = true
			selectnetwork()
		} else {
			fmt.Println("Not a valid option.\n")
			time.Sleep(500000000)
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
	fmt.Println(string(marshString))

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
	fmt.Println("Please select a saved network")

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
			fmt.Printf("%d) %s\n", i, file)

			//map i to file somehow for select
			option_map[i] = file
		}
		i = i+1
	}

	fmt.Print("\nLoad: ")
	scanner.Scan()
	network_selection := scanner.Text()
	int_select, err := strconv.Atoi(network_selection)
	netname := option_map[int_select]
	netname = netname[:len(netname)-len(".json")]

	loadnetwork(netname)
}

func loadnetwork(netname string) {
	//open file
	filename := "saves/" + netname + ".json"
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File not found: %s", filename)
	}

	//unmarshal
	var net Network
	err2 := json.Unmarshal(f, &net)
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

	fmt.Println(b)
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

	fmt.Println(o)
	return o
}

func NewProbox(hostname string) Host {
	p := Host{}
	p.ID = idgen(8)
	p.Model = "ProBox 1"
	p.MACAddr = macgen()
	p.Hostname = hostname

	fmt.Println(p)
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
	r := Router{}
	if routerModel == "BOBCAT" {
		r = NewBobcat(routerHostname)
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(routerHostname)
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
	fmt.Printf("network portion: %s", network_portion)

	r.DHCPTable = make(map[string]string)

	for k := range r.DHCPTable {
		r.DHCPIndex = append(r.DHCPIndex, k)
	}
	//sort.Ints(keys)
	//r.DHCPIndex = keys

	for i := 2; i < len(r.DHCPIndex); i++ {
		addrconstruct = network_portion + r.DHCPTable[r.DHCPIndex[i]]
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
		snet.Router = r
		fmt.Printf("\nRouter deleted\n")
	} else {
		fmt.Printf("\nRouter %s was not deleted.\n", snet.Router.Hostname)
	}
}

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
	h := Host{}
	if hostModel == "PROBOX" {
		h = NewProbox(hostHostname)
	}

	h.IPAddr = "0.0.0.0"

	snet.Hosts = append(snet.Hosts, h)
	fmt.Println(snet.Hosts)
}

func delHost() {}

func linkHost() {
	fmt.Println("Which host? Please specify by hostname")
	fmt.Print("Available hosts:")
	for availh := range snet.Hosts {
		fmt.Printf(" %s", snet.Hosts[availh].Hostname)
	}
	fmt.Print("\nHostname: ")
	scanner.Scan()
	hostname := scanner.Text()
	hostname = strings.ToUpper(hostname)

	fmt.Println("Uplink to which device? Please specify by hostname")
	fmt.Printf("Router: %s\n", snet.Router.Hostname)
	fmt.Printf("Switches: %s\n", "coming soon")
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
			}
			//TODO: Search switches

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

func ipset(hostname string) {
	fmt.Printf(" IP configuration for %s\n", hostname)

	correct := false
	var ipaddr, subnetmask, defaultgateway string
	for !correct {
		fmt.Print("IP Address: ")
		scanner.Scan()
		ipaddr = scanner.Text()

		fmt.Print("\nSubnet mask: ")
		scanner.Scan()
		subnetmask = scanner.Text()

		fmt.Print("\nDefault gateway: ")
		scanner.Scan()
		defaultgateway = scanner.Text()

		fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetmask, defaultgateway)
		fmt.Print("\nIs this correct? [Y/n/exit]")
		scanner.Scan()
		affirmation := scanner.Text()

		if(strings.ToUpper(affirmation) == "Y") {
			// error checking
			error := false
			if net.ParseIP(ipaddr).To4() == nil {
				error = true
				fmt.Printf("Error: '%s' is not a valid IP address\n", ipaddr)
			}
			if net.ParseIP(subnetmask).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid subnet mask\n", subnetmask)
			}
			if net.ParseIP(defaultgateway).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid default gateway\n", defaultgateway)
			}

			if(!error) {
				correct = true
			}
		 } else if(strings.ToUpper(affirmation) == "EXIT") {
			fmt.Println("Network changes reverted")
			return
		 }
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			snet.Hosts[h].IPAddr = ipaddr
			snet.Hosts[h].SubnetMask = subnetmask
			snet.Hosts[h].DefaultGateway = defaultgateway
			fmt.Println("Network configuration updated")
		}
	}

}

func overview() {
	fmt.Printf("Network name:\t\t%s\n", snet.Name)
	fmt.Printf("Network ID:\t\t%s\n", snet.ID)
	fmt.Printf("Network class:\t\tClass %s\n", snet.Class)

	show(snet.Router.Hostname)

	//hosts
	j := 0
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
		//fmt.Printf("\tUplink ID:\t%s\n", snet.Hosts[i].UplinkID)
		fmt.Printf("\tUplink to:\t%s\n", uplinkHostname)
		j = i
	}

	fmt.Printf("\nTotal devices: %d (1 Router, 0 Switches, %d Hosts)\n", (j + 1 + 1), (j + 1))
}

func show(hostname string) {
	device_type := "host"
	//TODO search switches
	if(snet.Router.Hostname == hostname) {
		device_type = "router"
	}

	if device_type == "host" {
		id := -1
		for i := range snet.Hosts {
			if snet.Hosts[i].Hostname == hostname {
				id = i
			}
		}
		if id == -1 {
			fmt.Printf("Hostname not found\n")
			return
		}
		fmt.Printf("\nHost %v\n", snet.Hosts[id].Hostname)
		fmt.Printf("\tID:\t\t%s\n", snet.Hosts[id].ID)
		fmt.Printf("\tModel:\t\t%s\n", snet.Hosts[id].Model)
		fmt.Printf("\tMAC:\t\t%s\n", snet.Hosts[id].MACAddr)
		fmt.Printf("\tIP Address:\t%s\n", snet.Hosts[id].IPAddr)
		fmt.Printf("\tDef. Gateway:\t%s\n", snet.Hosts[id].DefaultGateway)
		fmt.Printf("\tSubnet Mask:\t%s\n", snet.Hosts[id].SubnetMask)
		uplinkHostname := ""
		for dev := range snet.Hosts {
			//Router
			if(snet.Hosts[id].UplinkID == snet.Router.ID) {
				uplinkHostname = snet.Router.Hostname
			}
			//TODO: Switches
			//Hosts (pointless since host cant be uplink, just here to show how to do switches)
			if(snet.Hosts[id].UplinkID == snet.Hosts[dev].ID) {
				uplinkHostname = snet.Hosts[dev].Hostname
			}
		}
		//fmt.Printf("\tUplink ID:\t%s\n", snet.Hosts[i].UplinkID)
		fmt.Printf("\tUplink to:\t%s\n\n", uplinkHostname)
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
		case "add device host":
			addHost()
			save()
		default:
			fmt.Println(" Usage: add device <host|router>")
		}
	case "del":
		switch action_selection {
		case "del device router":
			delRouter()
			save()
		case "del device host":
			delHost()
			save()
		default:
			fmt.Println(" Usage: del device <host|router>")
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
				fmt.Println(" Usage: control <host|router> <hostname>")
			}
		} else {
			fmt.Println(" Usage: control <host|router> <hostname>")
		}
	case "ipset":
		if len(action_selection) > 11{
			switch action_selection[:11] {
			case "ipset host ":
				ipset(action_selection[11:])
				save()
			default:
				fmt.Println(" Usage: ipset host <hostname>")
			}
		} else {
			fmt.Println(" Usage: ipset host <hostname>")
		}
	case "reload":
		loadnetwork(snet.Name)
	case "show":
		switch action_selection {
		case "show network overview":
			overview()
		default:
			if len(action_selection) > 5{
				show(action_selection[5:])
			} else {
				fmt.Println(" Usage: show network overview\n\tshow <hostname>")
			}
		}
	case "help":
		fmt.Println("",
		"show <args>\t\tDisplays information\n",
		"add <args>\t\tAdds device to network\n",
		"del <args>\t\tRemoves device from network\n",
		"link <args>\t\tLinks two devices\n",
		"unlink <args>\t\tUnlinks two devices\n",
		"control <args>\t\tLogs in as device\n",
		"ipset <args>\t\tStatically assigns an IP configuration\n",
		"reload\t\t\tReloads the network file (May fix runtime bugs)")
	default:
		fmt.Println(" Invalid command. Type 'help' for a list of commands.")
	}
}

func save() {
	marshString, err := json.Marshal(snet)
	if err != nil {
		log.Println(err)
	}
	//fmt.Println("Saving", string(marshString)) //DEBUG
	// Write to file
	fmt.Printf("Saving network: %s\n", snet.Name)
	filename := "saves/" + snet.Name + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("Network saved")
	loadnetwork(snet.Name)
}

func main() {
	mainmenu()
	go Listener()

	for range snet.Hosts {
		<-listenSync
	}

	fmt.Println("\nPlease type an action:")

	for true {
		actions()
	}
}
