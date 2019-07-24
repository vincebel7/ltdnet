package main

import(
	"fmt"
	//"encoding/json"
	"os"
	//"log"
	//"math/rand"
	//"time"
	"bufio"
	"strings"
	//"strconv"
	//"path/filepath"
	//"io/ioutil"
)

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
	//connection
	//TODO loop
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n")
	action := ""
	for strings.ToUpper(action) != "EXIT" {
		fmt.Printf("%s> ", host.Hostname)
		scanner.Scan()
		action = scanner.Text()
		if action != "" {
		switch action {
			case "ping":
				go ping(host.Hostname, host.Hostname)
			case "help":
				fmt.Println("",
				"ping <dest_hostname> [seconds]\t\tPings host\n")

			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}
		}
	}

}

func ping(srchost string, dsthost string) {
	fmt.Printf("\nPinging %s from %s\n", dsthost, srchost)
}
