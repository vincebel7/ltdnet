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
	"time"
)

type Host struct {
	ID             string              `json:"id"`
	Model          string              `json:"model"`
	MACAddr        string              `json:"macaddr"`
	Hostname       string              `json:"hostname"`
	IPAddr         net.IP              `json:"ipaddr"`
	SubnetMask     string              `json:"mask"`
	DefaultGateway net.IP              `json:"gateway"`
	UplinkID       string              `json:"uplinkid"`
	ARPTable       map[string]ARPEntry `json:"arptable"`
}

type ARPEntry struct {
	MACAddr    string    `json:"macaddr"`
	ExpireTime time.Time `json:"expireTime"`
	Interface  string    `json:"interface"`
	State      string    `json:"state"`
}

func NewProbox(hostname string) Host {
	p := Host{}
	p.ID = idgen(8)
	p.Model = "ProBox 1"
	p.MACAddr = macgen()
	p.Hostname = hostname
	p.ARPTable = make(map[string]ARPEntry)

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
	go listenHostChannel(h.ID)
	<-listenSync
}

func linkHostTo(localDevice string, remoteDevice string) {
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
				portIndex := assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].ID)
				uplinkID = snet.Router.VSwitch.PortLinksLocal[portIndex]

			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range snet.Switches {
					if remoteDevice == strings.ToUpper(snet.Switches[j].Hostname) {
						//find next free port
						portIndex := assignSwitchport(snet.Switches[j], snet.Hosts[i].ID)
						uplinkID = snet.Switches[j].PortLinksLocal[portIndex]

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

			return
		}
	}
}

func delHost(hostname string) {
	hostname = strings.ToUpper(hostname)
	//search for host
	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == hostname {
			//unlink, Vswitch
			for j := range snet.Router.VSwitch.PortLinksRemote {
				if snet.Router.VSwitch.PortLinksLocal[j] == snet.Hosts[i].UplinkID {
					snet.Router.VSwitch.PortLinksRemote[j] = ""

					snet.Hosts = removeHostFromSlice(snet.Hosts, i)
					fmt.Printf("\nHost deleted\n")
					return
				}
			}

			//unlink, other switches
			for sw := range snet.Switches {
				for p := range snet.Switches[sw].PortLinksRemote {
					if snet.Switches[sw].PortLinksLocal[p] == snet.Hosts[i].UplinkID {
						snet.Switches[sw].PortLinksRemote[p] = ""

						snet.Hosts = removeHostFromSlice(snet.Hosts, i)
						fmt.Printf("\nHost deleted\n")
						return
					}
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
