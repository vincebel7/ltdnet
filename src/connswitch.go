/*
File:		connswitch.go
Author: 	https://github.com/vincebel7
Purpose:	Handles connection and interface for switches
*/

package main

import (
	"fmt"
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

			case "ip":
				fmt.Println("Not yet implemented on switches")

			case "arp":
				fmt.Println("Not yet implemented on switches")
				//displayARPTable(sw.ID)

			case "mac":
				if len(action) > 1 {
					switch action[1] {
					case "clear":
						if snet.Router.VSwitch.ID == id {
							snet.Router.VSwitch.MACTable = make(map[string]MACEntry)
						} else {
							snet.Switches[getSwitchIndexFromID(id)].MACTable = make(map[string]MACEntry)
						}
						fmt.Println("MAC table cleared")

					case "help", "?":
						fmt.Println("",
							"mac\t\t\tShows the device's MAC table (MAC address : Interface)\n",
							"mac clear\t\tClears the device's MAC table",
						)

					default:
						fmt.Println(" Invalid command. Type '?' for a list of commands.")
					}
				} else {
					displayMACTable(sw.ID)
				}

			case "exit", "quit", "q":
				return

			case "help", "?":
				fmt.Println("",
					//"ping <dst_ip> [seconds]\tPings an IP address\n",
					//"ip\t\t\t\tManage IP addressing\n",
					//"arp\t\t\t\tShow and manage the ARP table\n",
					"mac\t\t\t\tShow and manage the MAC address table\n",
					"exit\t\t\t\tReturns to main menu",
				)
			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
			}
		}
	}
}
