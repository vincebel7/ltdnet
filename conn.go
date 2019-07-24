package main

import(
	"fmt"
	//"encoding/json"
	"os"
	//"log"
	//"math/rand"
	"time"
	"bufio"
	"strings"
	"strconv"
	//"path/filepath"
	//"io/ioutil"
)

type Segment struct {
	protocol	string
	srcPort		int
	dstPort		int
	data		string
}

type Packet struct {
	srcIP		string
	dstIP		string
	data		Segment//layers 4+ abstracted
}

type Frame struct {
	srcMAC		string
	dstMAC		string
	data		Packet
}

func Conn(device string, id string) {
	fmt.Println("TEST FROM CONN")
	//find host
	host := Host{}
	for i := range snet.Hosts {
		if(snet.Hosts[i].ID == id){
			host = snet.Hosts[i]
		}
	}
	if host.ID == "" {
		fmt.Println("Error: ID cannot be located. Please try again")
	}

	//interface
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		fmt.Printf("%s> ", host.Hostname)
		scanner.Scan()
		action_selection := scanner.Text()
		if action_selection != "" {
		action := strings.Fields(action_selection)
		actionword1 := action[0]

		switch actionword1 {
			case "ping":
				if len(action) > 1 {
					if len(action) > 2 { //if seconds is specified
						seconds, _ := strconv.Atoi(action[2])
						go ping(host.Hostname, action[1], seconds)
					} else {
						go ping(host.Hostname, action[1], 1)
					}
				}
			case "help":
				fmt.Println("",
				"ping <dest_hostname> [seconds]\t\tPings host\n")

			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}
		}
	}

}

func ping(srchost string, dsthost string, secs int) {
	for i := 0; i < secs; i++ {
		fmt.Printf("\nPinging %s from %s\n", dsthost, srchost)
		time.Sleep(time.Second)
	}
	fmt.Printf("done pingin")
	//check if host is found
	return
}
