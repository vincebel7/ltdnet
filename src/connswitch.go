/*
File:		connswitch.go
Author: 	https://github.com/vincebel7
Purpose:	Handles connection and interface for switches
*/

package main

import (
	"fmt"
	"strings"

	"github.com/chzyer/readline"
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

	//interface
	fmt.Printf("\n")
	action_selection := ""

	// Set up readline for actionsMenu
	rl, err := readline.New(sw.Hostname + "> ")
	if err != nil {
		fmt.Printf("Error setting up readline: %v\n", err)
		return
	}
	defer rl.Close()

	for strings.ToUpper(action_selection) != "EXIT" {
		line, err := rl.Readline()
		if err != nil { // Exit on Ctrl+D or any read error
			fmt.Println("\nExiting...")
			break
		}

		if line == "" {
			continue
		}

		// Split the input into action words
		commandSplit := strings.Fields(line)

		switch commandSplit[0] {
		case "":

		case "ping":
			fmt.Println("Not yet implemented on switches")

		case "ip":
			fmt.Println("Not yet implemented on switches")

		case "arp":
			fmt.Println("Not yet implemented on switches")
			//displayARPTable(sw.ID)

		case "mac":
			if len(commandSplit) > 1 {
				switch commandSplit[1] {
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
