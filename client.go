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
	ID	string `json:"id"`
	Name	string `json:"name"`
	Author	string `json:"auth"`
	Gateway	string `json:"gway"`
	Class	string `json:"clas"`
}


func mainmenu() {
	fmt.Println("ltdnet v0.0.5")

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

	netid := idgen()
	var net []Network
	net = append(net, Network{
		ID: netid,
		Name: netname,
		Author: user_name,
		Gateway: "192.168.1.1",
		Class: network_class,
	})

	// Print to demonstrate
	marshString, err := json.Marshal(net)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(marshString))

	// Write to file
	filename := "saves/" + netname + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
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
	fmt.Printf("Selecting \"%s\"\n", netname)

	//open file
	filename := "saves/" + netname + ".json"
	f, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("File not found: %s", filename)
	}

	var net Network
	err2 := json.Unmarshal(f, &net)
	if err2 != nil {
		fmt.Printf("err: %v", err2)
	}
	fmt.Println(net)
}

func idgen() string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, 8)

	rand.Seed(time.Now().UnixNano())
	for i := range id {
		id[i] = idchars[rand.Intn(len(idchars))]
	}

	return string(id)
}

func main() {
	mainmenu()
}
