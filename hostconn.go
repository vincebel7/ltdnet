/*
File:		hostconn.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Handles connection and interface for hosts
*/

package main

import(
	"fmt"
	//"time"
	"strings"
	"strconv"
)

func Conn(device string, id string) {
	//find host
	host := Host{}
	hostindex := -1
	for i := range snet.Hosts {
		if(snet.Hosts[i].ID == id){
			hostindex = i
			host = snet.Hosts[i]
		}
	}
	if host.ID == "" {
		fmt.Println("Error: ID cannot be located. Please try again")
	}

	//interface
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		host = snet.Hosts[hostindex]

		fmt.Printf("%s> ", host.Hostname)
		scanner.Scan()
		action_selection := scanner.Text()
		actionword1 := ""
		if action_selection != "" {
		action := strings.Fields(action_selection)
		if(len(action) > 0){
			actionword1 = action[0]
		}

		switch actionword1 {
			case "":
			case "ping":
				if(host.UplinkID == "") {
					fmt.Println("Device is not connected. Please set an uplink")
				} else if(host.IPAddr == "0.0.0.0") {
					fmt.Println("Device does not have IP configuration. Please use DHCP or statically assign an IP configuration")
				}else {
					if len(action) > 1 {
						if len(action) > 2 { //if seconds is specified
							seconds, _ := strconv.Atoi(action[2])
							go ping(host.ID, action[1], seconds)
						} else {
							go ping(host.ID, action[1], 1)
						}
						<-actionsync[id]
					}
				}
			case "dhcp":
				if(host.UplinkID == "") {
					fmt.Println("Device is not connected. Please set an uplink")
				} else {

					go dhcp_discover(host)
					<-actionsync[id]
					save()
				}
			case "ipset":
				if(host.UplinkID == "") {
					fmt.Println("Device is not connected. Please set an uplink")
				} else {
					ipset(host.Hostname)
					save()
				}
			case "ipclear":
				ipclear(host.ID)
				save()
			case "exit":
				return
			case "help":
				fmt.Println("",
				"ping <dest_ip> [seconds]\tPings an IP address\n",
				"dhcp\t\t\t\tGets IP configuration via DHCP\n",
				"ipset\t\t\t\tStarts dialogue for statically assigning an IP configuration\n",
				"ipclear\t\t\tClears an IP configuration (WARNING: This does not release any DHCP leases)",
			)
			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}
		}
	}
}
