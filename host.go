/*
File:		host.go
Author: 	https://github.com/vincebel7
Purpose:	Host-specific functions
*/

package main

import(
	"fmt"
	"strings"
)

func NewProbox(hostname string) Host {
	p := Host{}
	p.ID = idgen(8)
	p.Model = "ProBox 1"
	p.MACAddr = macgen()
	p.Hostname = hostname

	return p
}

func addHost() {
	fmt.Println("What model?")
	fmt.Println("Available: ProBox")
	fmt.Print("Model: ")
	scanner.Scan()
	hostModel := scanner.Text()
	hostModel = strings.ToUpper(hostModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	hostHostname := scanner.Text()

	// input validation
	if hostHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

	if hostname_exists(hostHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	h := Host{}
	if hostModel == "PROBOX" {
		h = NewProbox(hostHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	h.IPAddr = "0.0.0.0"

	snet.Hosts = append(snet.Hosts, h)

	generateHostChannels(getHostIndexFromID(h.ID))
	<-listenSync
}

func delHost() {
	fmt.Println("Delete which host? Please specify by hostname")
	fmt.Print("Hosts:")
	for i := range snet.Hosts {
		fmt.Printf(" %s", snet.Hosts[i].Hostname)
	}
	fmt.Print("\nHostname: ")
	scanner.Scan()
	hostname := scanner.Text()
	hostname = strings.ToUpper(hostname)

	fmt.Printf("\nAre you sure you want do delete host %s? [Y/n]: ", hostname)
	scanner.Scan()
	confirmation := scanner.Text()
	confirmation = strings.ToUpper(confirmation)
	if confirmation == "Y" {
		//search for host
		for i := range snet.Hosts {
			if strings.ToUpper(snet.Hosts[i].Hostname) == hostname {
				//unlink
				for j := range snet.Router.VSwitch.Ports {
					if snet.Router.VSwitch.Ports[j] == snet.Hosts[i].ID {
						snet.Router.VSwitch.Ports = removeStringFromSlice(snet.Router.VSwitch.Ports, j)

						snet.Hosts = removeHostFromSlice(snet.Hosts, i)
						fmt.Printf("\nHost deleted\n")
						return

					}
				}

				snet.Hosts = removeHostFromSlice(snet.Hosts, i)
				fmt.Printf("\nHost deleted\n")
				return
			}
		}
	}
	fmt.Printf("\nHost %s was not deleted.\n", hostname)
}
