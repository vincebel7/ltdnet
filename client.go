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
)

type Network struct {
	ID	string
	Name	string
	Auth	string
	Gway	string
	Clas	string
}

var idchars = []rune("abcdef1234567890")

func mainmenu() {
	fmt.Println("ltdnet v0.0.4")

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
	network_name := scanner.Text()

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

	var net []Network
	net = append(net, Network{
		ID: idgen(),
		Name: network_name,
		Auth: user_name,
		Gway: "192.168.1.1",
		Clas: network_class,
	})

	// Print to demonstrate
	marshString, err := json.MarshalIndent(net, "", " ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(marshString))

	// Write to file
	filename := "saves/" + network_name + ".json"
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))

	fmt.Println("\nNetwork created!")
}

func selectnetwork() {
	fmt.Println("Please select a saved network")
}

func idgen() string {
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
