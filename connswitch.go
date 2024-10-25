/*
File:		connswitch.go
Author: 	https://github.com/vincebel7
Purpose:	Handles connection and interface for switches
*/

package main

import (
	"fmt"
	//"time"

	"strings"
)

func controlSwitch(hostname string) {
	fmt.Printf("Attempting to control switch %s...\n", hostname)
	for i := range snet.Switches {
		if snet.Switches[i].Hostname == hostname {
			SwitchConn(snet.Switches[i].ID)
			return
		}
	}

	if snet.Router.VSwitch.Hostname == hostname {
		SwitchConn(snet.Router.VSwitch.ID)
		return
	}

	fmt.Println("Switch not found")
}

func SwitchConn(id string) {
	sw := Switch{}

	//interface
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		for i := range snet.Switches {
			if snet.Switches[i].ID == id {
				sw = snet.Switches[i]
			}
		}
		if snet.Router.VSwitch.ID == id {
			sw = snet.Router.VSwitch
		}
		if sw.ID == "" {
			fmt.Println("Error: ID cannot be located. Please try again")
			return
		}

		fmt.Printf("%s> ", sw.Hostname)
		scanner.Scan()
		action_selection := scanner.Text()
		actionword1 := ""
		if action_selection != "" {
			action := strings.Fields(action_selection)
			if len(action) > 0 {
				actionword1 = action[0]
			}

			switch actionword1 {
			case "":
			case "ping":
				fmt.Println("Not yet implemented on switches")
			case "arprequest":
				fmt.Println("Not yet implemented on switches")
			case "ipset":
				fmt.Println("Not yet implemented on switches")
			case "ipclear":
				fmt.Println("Not yet implemented on switches")
				//ipclear(sw.ID)
				//save()
			case "arptable":
				fmt.Println("Not yet implemented on switches")
				//displayARPTable(sw.ID)
			case "mactable":
				displayMACTable(sw.ID)
			case "exit", "quit", "q":
				return
			case "help", "?":
				fmt.Println("",
					//"ping <dest_ip> [seconds]\tPings an IP address\n",
					//"ipset\t\t\t\tStarts dialogue for statically assigning an IP configuration\n",
					//"ipclear\t\t\tClears an IP configuration\n",
					//"arptable\t\t\tShows the switch's ARP table (IP address : MAC address)\n",
					"mactable\t\t\tShows the switch's MAC address table (MAC address : Switchport #)\n",
					"exit\t\t\t\tReturns to main menu",
				)
			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
			}
		}
	}
}
