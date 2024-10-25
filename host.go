/*
File:		host.go
Author: 	https://github.com/vincebel7
Purpose:	Host-specific functions
*/

package main

import (
	"fmt"
	"net"
	"strings"
)

type Host struct {
	ID             string `json:"id"`
	Model          string `json:"model"`
	MACAddr        string `json:"macaddr"`
	Hostname       string `json:"hostname"`
	IPAddr         net.IP `json:"ipaddr"`
	SubnetMask     string `json:"mask"`
	DefaultGateway net.IP `json:"gateway"`
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

	h.IPAddr = net.ParseIP("0.0.0.0")

	snet.Hosts = append(snet.Hosts, h)

	generateHostChannels(getHostIndexFromID(h.ID))
	<-listenSync
}

func linkHost(localDevice string, remoteDevice string) {
	localDevice = strings.ToUpper(localDevice)
	remoteDevice = strings.ToUpper(remoteDevice)

	//Make sure there's enough ports - if uplink device is a router
	if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
		if getActivePorts(snet.Router.VSwitch) >= snet.Router.VSwitch.Maxports {
			fmt.Printf("No available ports - %s only has %d ports\n", snet.Router.Model, snet.Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range snet.Switches {
		if remoteDevice == strings.ToUpper(snet.Switches[s].Hostname) {
			if getActivePorts(snet.Switches[s]) >= snet.Switches[s].Maxports {
				fmt.Printf("No available ports - %s only has %d ports\n", snet.Switches[s].Model, snet.Switches[s].Maxports)
				return
			}
		}
	}

	//find host with that hostname
	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == localDevice {
			uplinkID := ""
			//Remote device on new link is the Router
			if remoteDevice == strings.ToUpper(snet.Router.Hostname) {
				//find next free port
				for k := range snet.Router.VSwitch.Ports {
					if (snet.Router.VSwitch.Ports[k] == "") && (uplinkID == "") {
						uplinkID = snet.Router.VSwitch.PortIDs[k]
					}
				}
				//uplinkID = snet.Router.VSwitch.ID

				assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].ID)
			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range snet.Switches {
					if remoteDevice == strings.ToUpper(snet.Switches[j].Hostname) {

						//find next free port
						for k := range snet.Switches[j].Ports {
							if (snet.Switches[j].Ports[k] == "") && (uplinkID == "") {
								uplinkID = snet.Switches[j].PortIDs[k]
								k = len(snet.Switches[j].Ports)
								fmt.Println("DEBUG TEST")
							}
						}
					}
				}
			}

			snet.Hosts[i].UplinkID = uplinkID
			return
		}
	}
}

func unlinkHost(hostname string) {
	hostname = strings.ToUpper(hostname)

	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == hostname {
			//first, unplug from switch (switch-end unlink). TODO try/catch this whole block.
			freeSwitchport(snet.Hosts[i].UplinkID)

			//next, remove the host's uplink (host-end unlink)
			uplinkID := ""
			snet.Hosts[i].UplinkID = uplinkID
			//snet.Router.Ports = removeStringFromSlice(snet.Router.Ports, i)

			return
		}
	}
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

func ipclear(id string) {
	index := getHostIndexFromID(id)
	snet.Hosts[index].IPAddr = nil
	snet.Hosts[index].SubnetMask = ""
	snet.Hosts[index].DefaultGateway = nil
	fmt.Println("Network configuration cleared")
}
