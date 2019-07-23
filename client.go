package main

import(
	"fmt"
	"encoding/json"
	"os"
	"log"
	"math/rand"
	"time"
	"bufio"
	"strings"
	"strconv"
	"path/filepath"
	"io/ioutil"
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

type Router struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"`
	Hostname	string `json:"hnme"`
	Gateway		string `json:"gway"`
	DHCPPool	int `json:"dpol"` //maximum, not just available
	Downports	int `json:"dpts"`
	MACTable	map[string]string `json:"mact"`
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
	fmt.Println("ltdnet v0.0.9")

	selection := false
		for selection == false {
		fmt.Println("Please create or select a network:")
		fmt.Println("1) Create new network")
		fmt.Println("2) Select saved network")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		option := scanner.Text()

		if option == "1" || option == "C" || option == "c" || option == "create" || option == "new" {
			selection = true
			newnetwork()
		} else if option == "2" || option == "S" || option == "s" || option == "select" {
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
	scanner := bufio.NewScanner(os.Stdin)

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

	scanner := bufio.NewScanner(os.Stdin)
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
	fmt.Printf("Loaded %s", snet.Name)
}

func idgen(n int) string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, n)

	rand.Seed(time.Now().UnixNano())
	for i := range id {
		id[i] = idchars[rand.Intn(len(idchars))]
	}

	return string(id)
}

func macgen() string {
	mac := idgen(2)
	for n := 0; n < 5; n++ {
		mac = mac + ":" + idgen(2)
	}
	return mac
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
	scanner := bufio.NewScanner(os.Stdin)
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
	snet.Router = r
}

func addHost() {
	scanner := bufio.NewScanner(os.Stdin)
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
	scanner := bufio.NewScanner(os.Stdin)
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

func overview() {
	fmt.Printf("Network name:\t\t%s\n", snet.Name)
	fmt.Printf("Network ID:\t\t%s\n", snet.ID)
	fmt.Printf("Network class:\t\tClass %s\n", snet.Class)

	fmt.Printf("\nRouter %s\n", snet.Router.Hostname)
	fmt.Printf("\tID:\t\t%s\n", snet.Router.ID)
	fmt.Printf("\tModel:\t\t%s\n", snet.Router.Model)
	fmt.Printf("\tMAC:\t\t%s\n", snet.Router.MACAddr)
	fmt.Printf("\tGateway:\t%s\n", snet.Router.Gateway)
	fmt.Printf("\tDHCP pool:\t%d addresses\n", snet.Router.DHCPPool)
	fmt.Printf("\tUser ports:\t%d ports\n", snet.Router.Downports)

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

func actions() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	scanner.Scan()
	action_selection := scanner.Text()
	actionword1 := ""
	if action_selection != "" {
		actionword0 := strings.Fields(action_selection)
		actionword1 = actionword0[0]
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
	case "show":
		switch action_selection {
		case "show network overview":
			overview()
		default:
			fmt.Println(" Usage: show network overview")
		}
	case "help":
		fmt.Println("",
		"show <args>	Displays information\n",
		"add <args>	Adds device to network\n",
		"del <args>	Removes device from network\n",
		"link <args>	Links two devices\n",
		"unlink <args>	Unlinks two devices\n")
	default:
		fmt.Println(" Invalid command. Type 'help' for a list of commands.")
	}
}

func save() {
	//are you sure prompt

	marshString, err := json.Marshal(snet)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Saving", string(marshString))
	// Write to file
	filename := "saves/" + snet.Name + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("Network saved")
}

func main() {
	mainmenu()

	fmt.Println("\nPlease type an action:")
	for true {
		actions()
	}
}
