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
}

type Host struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"`
	Hostname	string `json:"hnme"`
	IPAddr		string `json:"ipa"`
	SubnetMask	string `json:"mask"`
	DefaultGateway	string `json:"gway"`
}

func mainmenu() {
	fmt.Println("ltdnet v0.0.6")

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
	fmt.Println("Loaded %s", snet.Name)
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

	if snet.Class == "A" {
		b.Gateway = "10.0.0.1"
	} else if snet.Class == "B" {
		b.Gateway = "172.16.0.1"
	} else if snet.Class == "C" {
		b.Gateway = "192.168.0.1"
	}

	fmt.Println(b)
	return b
}

func addRouter() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("What model?")
	fmt.Println("Available: Bobcat")
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
	}
	snet.Router = r
}

func actions() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("\nPlease type an action:")
	fmt.Println("help")
	fmt.Println("show network overview")
	fmt.Println("add device router")
	fmt.Println("del device router")
	fmt.Print("> ")
	scanner.Scan()
	action_selection := scanner.Text()
	if action_selection == "add device router" {
		addRouter()
		save()
	}
	if action_selection == "show network overview" {
		fmt.Println(snet)
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

	for true {
		actions()
	}
}
