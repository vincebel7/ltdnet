package main

import( "fmt"
	"encoding/json"
//	"io/ioutil"
	"os"
	"log"
)

type network struct {
	name []string `json:"name"`
	author []string `json:"author"`
}

func main() {
	fmt.Println("ltdnet v0.0.1")
	var net []network
	net = append(net, network{
		name: []string{"Test Name"},
		author: []string{"Vince"},
	})

	// Print to demonstrate
	marshString, _ := json.MarshalIndent(net, "", " ")
	fmt.Println(string(marshString))

	// Write to file
	f, err := os.OpenFile("file.json", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	f.Write(marshString)
}
