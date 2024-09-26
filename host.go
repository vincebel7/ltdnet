/*
File:		host.go
Author: 	https://github.com/vincebel7
Purpose:	Host-specific functions
*/

package main

import (
	"fmt"
	"strings"
)

type Host struct {
	ID             string `json:"id"`
	Model          string `json:"model"`
	MACAddr        string `json:"macaddr"`
	Hostname       string `json:"hostname"`
	IPAddr         string `json:"ipaddr"`
	SubnetMask     string `json:"mask"`
	DefaultGateway string `json:"gateway"`
	UplinkID       string `json:"uplinkid"`
}

func NewProbox(hostname string) Host {
	p := Host{}
	p.ID = idgen(8)
	p.Model = "ProBox 1"
	p.MACAddr = macgen()
	p.Hostname = hostname

	return p
}

func addHost(hostHostname string) {
	hostModel := strings.ToUpper("ProBox")

	// input validation
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

func delHost(hostname string) {
	hostname = strings.ToUpper(hostname)
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
	fmt.Printf("\nHost %s was not deleted.\n", hostname)
}
