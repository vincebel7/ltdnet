/*
File:		connhost.go
Author: 	https://github.com/vincebel7
Purpose:	Handles connection and interface for hosts
*/

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

func controlHost(hostname string) {
	fmt.Printf("Attempting to control host %s...\n", hostname)
	for i := range snet.Hosts {
		if snet.Hosts[i].Hostname == hostname {
			HostConn("host", snet.Hosts[i].ID)
			return
		}
	}
	fmt.Println("Host not found")
}

func HostConn(device string, id string) {
	//find host
	host := Host{}
	hostindex := -1
	for i := range snet.Hosts {
		if snet.Hosts[i].ID == id {
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

	// Set up readline for actionsMenu
	rl, err := readline.New(host.Hostname + "> ")
	if err != nil {
		fmt.Printf("Error setting up readline: %v\n", err)
		return
	}
	defer rl.Close()

	for strings.ToUpper(action_selection) != "EXIT" {
		host = snet.Hosts[hostindex]

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
			if len(commandSplit) > 1 {
				if len(commandSplit) > 2 { //if count is specified
					count, _ := strconv.Atoi(commandSplit[2])
					go ping(host.ID, commandSplit[1], count)
				} else {
					go ping(host.ID, commandSplit[1], 4)
				}
				<-actionsync[id]
			} else {
				fmt.Println("Usage: ping <dst_ip> [count]")
			}

		case "dhcp":
			if host.Interfaces["eth0"].RemoteL1ID == "" {
				fmt.Println("Device is not connected. Please set an uplink")
			} else {
				go dhcp_discover(host)
				<-actionsync[id]
				save()
			}

		case "hosts":
			displayDNSTable(host.DNSTable)

		case "ip":
			printIPHelp := func() {
				fmt.Println("",
					"ip address\t\t\tShow IP configuration\n",
					"ip route\t\t\tShow known routes\n",
					"ip set\t\t\t\tStarts dialogue for statically assigning an IP configuration\n",
					"ip clear\t\t\tClears an IP configuration (WARNING: This does not release DHCP leases)",
				)
			}

			if len(commandSplit) > 1 {
				switch commandSplit[1] {
				case "a", "addr", "address":
					for iface := range host.Interfaces {
						fmt.Printf("Interface %s\n", host.Interfaces[iface].Name)
						fmt.Printf("\tIPv4 address: %s\n", host.GetIP(iface))
						fmt.Printf("\tSubnet mask: %s\n\n", host.GetMask(iface))
					}
				case "route":
					fmt.Printf("default via %s dev %s\n", host.GetGateway("eth0"), "eth0")

				case "set":
					if host.Interfaces["eth0"].RemoteL1ID == "" {
						fmt.Println("Device is not connected. Please set an uplink")
					} else {
						if len(commandSplit) > 3 {
							ipset(host.Hostname, commandSplit[2], commandSplit[3])
							save()
						} else {
							fmt.Println("Usage: ip set <ip_address> <subnet_mask>")
						}
					}
				case "clear":
					ipclear(host.ID)
					save()

				case "help", "?":
					printIPHelp()

				default:
					fmt.Println(" Invalid command. Type 'ip ?' for a list of commands.")
				}
			} else {
				printIPHelp()
			}
		case "arp":
			if len(commandSplit) > 1 {
				switch commandSplit[1] {
				case "request":
					if len(commandSplit) > 2 {
						go arpSynchronized(id, commandSplit[2])
						<-actionsync[id]
					} else {
						fmt.Println("Usage: arp request <target_ip>")
					}

				case "clear":
					snet.Hosts[getHostIndexFromID(host.ID)].ARPTable = make(map[string]ARPEntry)
					fmt.Println("ARP table cleared")

				case "help", "?":
					fmt.Println("",
						"arp\t\t\tShows the device's ARP table (IP address : MAC address)\n",
						"arp request <dst_ip>\tManually ARP request an address\n",
						"arp clear\t\tClears the device's ARP table",
					)

				default:
					fmt.Println(" Invalid command. Type '?' for a list of commands.")
				}
			} else {
				displayARPTable(host.ID)
			}

		case "nslookup":
			if len(commandSplit) > 1 {
				go printResolveHostname(host.ID, commandSplit[1], host.DNSTable)
				<-actionsync[id]
				save()

			} else {
				fmt.Println("Usage: nslookup <hostname>")
			}

		case "exit", "quit", "q":
			return

		case "help", "?":
			fmt.Println("",
				"ping\t\t\t\tPings an IP address\n",
				"dhcp\t\t\t\tGets IP configuration via DHCP\n",
				"hosts\t\t\t\tDisplays local host entries (from DNS)\n",
				"ip\t\t\t\tManage IP addressing\n",
				"arp\t\t\t\tShow and manage the ARP table\n",
				"nslookup\t\t\tPerform a DNS lookup\n",
				"exit\t\t\t\tReturns to main menu",
			)

		default:
			fmt.Println(" Invalid command. Type '?' for a list of commands.")
		}
	}
}
