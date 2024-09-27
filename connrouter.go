/*
File:		connrouter.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Handles connection and interface for router
*/

package main

import (
	"fmt"
	//"time"
	"strconv"
	"strings"
)

func controlRouter(hostname string) {
	fmt.Printf("Attempting to control router %s...\n", hostname)
	RouterConn("router", snet.Router.ID)
}

func RouterConn(device string, id string) {
	//interface
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		fmt.Printf("%s> ", snet.Router.Hostname)
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
				if (snet.Router.Gateway.String() == "0.0.0.0") || (snet.Router.Gateway == nil) {
					fmt.Println("Device does not have IP configuration. Please statically assign an IP configuration")
				} else {
					if len(action) > 1 {
						if len(action) > 2 { //if seconds is specified
							seconds, _ := strconv.Atoi(action[2])
							go ping(snet.Router.ID, action[1], seconds)
						} else {
							go ping(snet.Router.ID, action[1], 4)
						}
						<-actionsync[id]
					} else {
						fmt.Println("Usage: ping <dest_ip> [seconds]")
					}
				}
			case "dhcpserver":
				dhcpserver()
				save()
			case "ipset":
				ipset(snet.Router.Hostname)
				save()
			case "ipclear":
				ipclear(snet.Router.Gateway.String())
				save()
			case "exit", "quit", "q":
				return
			case "help", "?":
				fmt.Println("",
					"ping <dest_ip> [seconds]\tPings an IP address\n",
					"dhcpserver\t\t\tDisplays DHCP server and DHCP pool settings\n",
					"ipset\t\t\t\tStarts dialogue for statically assigning an IP configuration\n",
					"ipclear\t\t\tClears an IP configuration\n",
					"exit\t\t\t\tReturns to main menu",
				)
			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
			}
		}
	}
}

func dhcpserver() {
	pool := snet.Router.GetDHCPPoolAddresses()
	leaseCount := len(snet.Router.DHCPPool.DHCPPoolLeases)
	poolCount := len(pool)

	print("DHCP Server:\n")
	fmt.Printf("\tPool range:\t\t%s\n", snet.Router.DHCPPool.DHCPPoolStart.String()+" - "+snet.Router.DHCPPool.DHCPPoolEnd.String())
	fmt.Printf("\tPool utilization:\t%d/%d (%.2f%% full)\n", leaseCount, poolCount, float64(leaseCount)/float64(poolCount)*100)
	fmt.Printf("\tNext available address:\t%s\n", snet.Router.NextFreePoolAddress())
	fmt.Printf("\nActive leases:\n")

	for i := range pool {
		addr := pool[i].String()
		if snet.Router.DHCPPool.DHCPPoolLeases[addr] != "" {
			fmt.Printf("\t%s - %s\n", addr, snet.Router.DHCPPool.DHCPPoolLeases[addr])
		}
	}
}
