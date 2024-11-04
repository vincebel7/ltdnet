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

	"github.com/vincebel7/ltdnet/iphelper"
)

type Host struct {
	ID         string               `json:"id"`
	Model      string               `json:"model"`
	Hostname   string               `json:"hostname"`
	ARPTable   map[string]ARPEntry  `json:"arptable"`
	DNSTable   map[string]DNSRecord `json:"dnstable"`
	Interfaces map[string]Interface `json:"interfaces"`
}

type ARPEntry struct {
	MACAddr    string    `json:"macaddr"`
	ExpireTime time.Time `json:"expireTime"`
	Interface  string    `json:"interface"`
	State      string    `json:"state"`
}

// Populate fields specific to the Probox 1
func NewProbox(h Host) Host {
	h.Model = "ProBox 1"
	return h
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
		h = NewProbox(h)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	h.ID = idgen(8)
	h.Hostname = hostHostname
	h.ARPTable = make(map[string]ARPEntry)

	// Interfaces
	h.Interfaces = make(map[string]Interface)

	loopbackIPConfig := IPConfig{
		IPAddress:      net.ParseIP("127.0.0.1"),
		SubnetMask:     "255.0.0.0",
		DefaultGateway: nil,
		DNSServer:      nil,
		ConfigType:     "static",
	}
	eth0IPConfig := IPConfig{
		IPAddress:      nil,
		SubnetMask:     "",
		DefaultGateway: nil,
		DNSServer:      nil,
		ConfigType:     "",
	}

	h.Interfaces["lo"] = Interface{
		Name:     "lo",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: loopbackIPConfig,
	}
	h.Interfaces["eth0"] = Interface{
		Name:     "eth0",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: eth0IPConfig,
	}

	// DNS table
	h.DNSTable = make(map[string]DNSRecord)

	h.DNSTable[h.Hostname] = DNSRecord{
		Name:  h.Hostname,
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}
	h.DNSTable["localhost"] = DNSRecord{
		Name:  "localhost",
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}

	snet.Hosts = append(snet.Hosts, h)

	generateHostChannels(getHostIndexFromID(h.ID))
	go listenHostChannel(h, "lo")
	<-listenSync
	go listenHostChannel(h, "eth0")
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
				portIndex := assignSwitchport(snet.Router.VSwitch, snet.Hosts[i].Interfaces["eth0"].L1ID)
				uplinkID = snet.Router.VSwitch.PortLinksLocal[portIndex]

			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range snet.Switches {
					if remoteDevice == strings.ToUpper(snet.Switches[j].Hostname) {
						//find next free port
						portIndex := assignSwitchport(snet.Switches[j], snet.Hosts[i].Interfaces["eth0"].L1ID)
						uplinkID = snet.Switches[j].PortLinksLocal[portIndex]

					}
				}
			}

			// Assign uplink ID to host
			iface := snet.Hosts[i].Interfaces["eth0"]
			iface.RemoteL1ID = uplinkID
			snet.Hosts[i].Interfaces["eth0"] = iface

			return
		}
	}
}

func unlinkHost(hostname string) {
	hostname = strings.ToUpper(hostname)

	for i := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[i].Hostname) == hostname {
			//first, unplug from switch (switch-end unlink). TODO try/catch this whole block.
			freeSwitchport(snet.Hosts[i].Interfaces["eth0"].RemoteL1ID)

			//next, remove the host's uplink (host-end unlink)
			uplinkID := ""
			iface := snet.Hosts[i].Interfaces["eth0"]
			iface.RemoteL1ID = uplinkID
			snet.Router.Interfaces["eth0"] = iface

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
				if snet.Router.VSwitch.PortLinksLocal[j] == snet.Hosts[i].Interfaces["eth0"].RemoteL1ID {
					snet.Router.VSwitch.PortLinksRemote[j] = ""

					snet.Hosts = removeHostFromSlice(snet.Hosts, i)
					fmt.Printf("\nHost deleted\n")
					return
				}
			}

			//unlink, other switches
			for sw := range snet.Switches {
				for p := range snet.Switches[sw].PortLinksRemote {
					if snet.Switches[sw].PortLinksLocal[p] == snet.Hosts[i].Interfaces["eth0"].RemoteL1ID {
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

	iface := snet.Hosts[index].Interfaces["eth0"]

	iface.IPConfig.IPAddress = nil
	iface.IPConfig.SubnetMask = ""
	iface.IPConfig.DefaultGateway = nil

	snet.Hosts[index].Interfaces["eth0"] = iface

	fmt.Println("Network configuration cleared")
}

func (host Host) routeToInterface(dstIP string) Interface {
	for iface := range host.Interfaces {
		devIP := host.GetIP(iface)
		devMask := host.GetMask(iface)

		if iphelper.IPInSameSubnet(devIP, dstIP, devMask) {
			return host.Interfaces[iface]
		}
	}

	// Default gateway
	debug(4, "routeToInterface", host.Hostname, "Route not found. Sending to default gateway")
	return host.Interfaces["eth0"]
}

func printResolveHostname(srcID string, hostname string, dnsTable map[string]DNSRecord) {
	dnsRecord := resolveHostname(srcID, hostname, dnsTable)
	fmt.Println("Name: " + hostname)
	fmt.Println("Address: " + dnsRecord.RData + "\n")

	actionsync[srcID] <- 1
}
