package main

import(
	"fmt"
	"encoding/json"
	"os"
	"log"
	"math/rand"
	"time"
	"bufio"
)

type Network struct {
	ID	string
	Name	string
	Auth	string
	Gway	string
}

var idchars = []rune("abcdef1234567890")

func mainmenu() {
	fmt.Println("ltdnet v0.0.3")

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

	var net []Network
	net = append(net, Network{
		ID: idgen(),
		Name: "Test Name",
		Auth: "Vince",
		Gway: "192.168.1.1",
	})

	// Print to demonstrate
	marshString, err := json.MarshalIndent(net, "", " ")
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(marshString))

	// Write to file
	f, err := os.OpenFile("file.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
	f.Write([]byte("\n"))
}
