package main

import(
	"fmt"
	//"encoding/json"
	//"os"
	//"log"
	//"math/rand"
	//"time"
	//"bufio"
	//"strings"
	//"strconv"
	//"path/filepath"
	//"io/ioutil"
)

var channels = map[string]chan string{}

func Listener() {
	for i := range snet.Hosts {
		//create channel and map channel ID
		channels[snet.Hosts[i].ID] = make(chan string)
		go listen(i) //TODO block Client.action() till this runs
	}
}

func listen(i int) {
	//declarations to make things easier
	id := snet.Hosts[i].ID
	hostname := snet.Hosts[i].Hostname

	fmt.Printf("\n%s listening", snet.Hosts[i].Hostname)
	action := <-channels[id]
	fmt.Printf("%s just got: %s\n", hostname, action)
}

func actionHandler() {

}
