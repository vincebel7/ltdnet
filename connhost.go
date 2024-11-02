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
	for strings.ToUpper(action_selection) != "EXIT" {
		host = snet.Hosts[hostindex]

		fmt.Printf("%s> ", host.Hostname)
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
				if host.Interface.RemoteL1ID == "" {
					fmt.Println("Device is not connected. Please set an uplink")
				} else if (host.GetIP() == "0.0.0.0") || (host.GetIP() == "") {
					fmt.Println("Device does not have IP configuration. Please use DHCP or statically assign an IP configuration")
				} else {
					if len(action) > 1 {
						if len(action) > 2 { //if count is specified
							count, _ := strconv.Atoi(action[2])
							go ping(host.ID, action[1], count)
						} else {
							go ping(host.ID, action[1], 4)
						}
						<-actionsync[id]
					} else {
						fmt.Println("Usage: ping <dst_ip> [count]")
					}
				}

			case "dhcp":
				if host.Interface.RemoteL1ID == "" {
					fmt.Println("Device is not connected. Please set an uplink")
				} else {
					go dhcp_discover(host)
					<-actionsync[id]
					save()
				}

			case "ip":
				printIPHelp := func() {
					fmt.Println("",
						"ip address\t\t\tShow IP configuration\n",
						"ip route\t\t\tShow known routes\n",
						"ip set\t\t\t\tStarts dialogue for statically assigning an IP configuration\n",
						"ip clear\t\t\tClears an IP configuration (WARNING: This does not release DHCP leases)",
					)
				}

				if len(action) > 1 {
					switch action[1] {
					case "a", "addr", "address":
						fmt.Println("IPv4 address: " + host.GetIP())
						fmt.Println("Subnet mask: " + host.GetMask())

					case "route":
						fmt.Println("Default gateway: " + host.GetGateway())

					case "set":
						if host.Interface.RemoteL1ID == "" {
							fmt.Println("Device is not connected. Please set an uplink")
						} else {
							if len(action) > 2 {
								ipset(host.Hostname, action[2])
								save()
							} else {
								fmt.Println("Usage: ip set <ip_address>")
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
				if len(action) > 1 {
					switch action[1] {
					case "request":
						if host.Interface.RemoteL1ID == "" {
							fmt.Println("Device is not connected. Please set an uplink")
						} else if (host.GetIP() == "0.0.0.0") || (host.GetIP() == "") {
							fmt.Println("Device does not have IP configuration. Please use DHCP or statically assign an IP configuration")
						} else {
							if len(action) > 2 {
								go arpSynchronized(id, action[2])
								<-actionsync[id]
							} else {
								fmt.Println("Usage: arp request <target_ip>")
							}
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
				if len(action) > 1 {
					address := resolveHostname(action[1], host.DNSTable)
					fmt.Println("Name: " + action[1])
					fmt.Println("Address: " + address + "\n")

				} else {
					fmt.Println("Usage: nslookup <hostname>")
				}

			case "exit", "quit", "q":
				return

			case "help", "?":
				fmt.Println("",
					"ping\t\t\t\tPings an IP address\n",
					"dhcp\t\t\t\tGets IP configuration via DHCP\n",
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
}
