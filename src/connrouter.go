/*
File:		connrouter.go
Author: 	https://github.com/vincebel7
Purpose:	Handles connection and interface for router
*/

package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

func controlRouter(hostname string) {
	fmt.Printf("Attempting to control router %s...\n", hostname)
	if Net().Router.Hostname == hostname {
		RouterConn("router", Net().Router.ID)
		return
	}
	fmt.Println("Router not found")
}

func RouterConn(device string, id string) {
	//interface
	fmt.Printf("\n")
	action_selection := ""

	// Set up readline for actionsMenu
	rl, err := readline.New(Net().Router.Hostname + "> ")
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
			if (Net().Router.GetIP("eth0") == "0.0.0.0") || (Net().Router.GetIP("eth0") == "") {
				fmt.Println("Device does not have IP configuration. Please statically assign an IP configuration")
			} else {
				if len(commandSplit) > 1 {
					if len(commandSplit) > 2 { //if seconds is specified
						seconds, _ := strconv.Atoi(commandSplit[2])
						go ping(Net().Router.ID, commandSplit[1], seconds)
					} else {
						go ping(Net().Router.ID, commandSplit[1], 4)
					}
					<-actionsync[id]
				} else {
					fmt.Println("Usage: ping <dst_ip> [seconds]")
				}
			}

		case "dhcpserver":
			if len(commandSplit) > 1 {
				switch commandSplit[1] {
				default:
				}
			} else {
				displayDHCPServer()
			}

		case "dnsserver":
			printDNSServerHelp := func() {
				fmt.Println("",
					"dnsserver\t\t\tShow status of DNS server\n",
					"dnsserver add\t\t\tAdd DNS record to server\n",
					"dnsserver remove\t\tRemove DNS record from server",
				)
			}
			if len(commandSplit) > 1 {
				switch commandSplit[1] {
				case "add":
					if len(commandSplit) > 3 {
						Net().Router.DNSServer.addDNSRecordToServer('A', commandSplit[2], commandSplit[3])
					} else {
						fmt.Println("Usage: dnsserver add <hostname> <ip_address>")
					}
					save()

				case "remove":
					fmt.Println("DNS record removing not implemented yet")
					save()

				default:
					printDNSServerHelp()
				}
			} else {
				Net().Router.DNSServer.dnsServerMenu()
			}

		case "hosts":
			displayDNSTable(Net().Router.DNSTable)

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
					for iface := range Net().Router.Interfaces {
						fmt.Printf("Interface %s\n", Net().Router.Interfaces[iface].Name)
						fmt.Printf("\tIPv4 address: %s\n", Net().Router.GetIP(iface))
						fmt.Printf("\tSubnet mask: %s\n\n", Net().Router.GetMask(iface))
					}

				case "route":
					fmt.Println("Routing not implemented yet.")

				case "set":
					if len(commandSplit) > 3 {
						ipset(Net().Router.Hostname, commandSplit[2], commandSplit[3])
						save()
					} else {
						fmt.Println("Usage: ipset <ip_address> <subnet_mask>")
					}

				case "clear":
					ipclear(Net().Router.GetIP("eth0"))
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
					Net().Router.ARPTable = make(map[string]ARPEntry)
					fmt.Println("ARP table cleared")

				case "help", "?":
					fmt.Println("",
						"arp\t\t\tShows the device's ARP table (IP address : MAC address)\n",
						"arp request <dst_ip>\tManually ARP request an address",
						"arp clear\t\tClears the device's ARP table",
					)

				default:
					fmt.Println(" Invalid command. Type '?' for a list of commands.")
				}
			} else {
				displayARPTable(Net().Router.ID)
			}

		case "nslookup":
			if len(commandSplit) > 1 {
				go printResolveHostname(Net().Router.ID, commandSplit[1], Net().Router.DNSTable)
				<-actionsync[id]
				save()

			} else {
				fmt.Println("Usage: nslookup <hostname>")
			}

		case "exit", "quit", "q":
			return

		case "help", "?":
			fmt.Println("",
				"ping <dst_ip> [seconds]\tPings an IP address\n",
				"dhcpserver\t\t\tDisplays DHCP server and settings\n",
				"dnsserver\t\t\tDisplays DNS server and settings\n",
				"hosts\t\t\t\tDisplays local host entries (from DNS)\n",
				"ip\t\t\t\tManage IP addressing\n",
				"arp\t\t\t\tShow and manage the ARP table\n",
				"nslookup\t\t\tPerform a DNS lookup\n",
				"exit\t\t\t\tReturns to main menu",
			)
		default:
			fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}
	}
}

func displayDHCPServer() {
	pool := Net().Router.GetDHCPPoolAddresses()
	leaseCount := len(Net().Router.DHCPPool.DHCPPoolLeases)
	poolCount := len(pool)

	fmt.Printf("DHCP Server:\n")
	fmt.Printf("\tPool range:\t\t%s\n", Net().Router.DHCPPool.DHCPPoolStart.String()+" - "+Net().Router.DHCPPool.DHCPPoolEnd.String())
	fmt.Printf("\tPool utilization:\t%d/%d (%.2f%% full)\n", leaseCount, poolCount, float64(leaseCount)/float64(poolCount)*100)
	fmt.Printf("\tNext available address:\t%s\n", Net().Router.NextFreePoolAddress())
	fmt.Printf("\nActive leases:\n")

	for i := range pool {
		addr := pool[i].String()
		if Net().Router.DHCPPool.DHCPPoolLeases[addr] != "" {
			fmt.Printf("\t%s - %s\n", addr, Net().Router.DHCPPool.DHCPPoolLeases[addr])
		}
	}
}
