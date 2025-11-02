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

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

// Populate fields specific to the Probox 1
func NewProbox(h model.Host) model.Host {
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

	h := model.Host{}
	if hostModel == "PROBOX" {
		h = NewProbox(h)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	h.ID = idgen(8)
	h.Hostname = hostHostname
	h.ARPTable = make(map[string]model.ARPEntry)

	// Interfaces
	h.Interfaces = make(map[string]model.Interface)

	loopbackIPConfig := model.IPConfig{
		IPAddress:      net.ParseIP("127.0.0.1"),
		SubnetMask:     "255.0.0.0",
		DefaultGateway: nil,
		DNSServer:      nil,
		ConfigType:     "static",
	}
	eth0IPConfig := model.IPConfig{
		IPAddress:      nil,
		SubnetMask:     "",
		DefaultGateway: nil,
		DNSServer:      nil,
		ConfigType:     "",
	}

	h.Interfaces["lo"] = model.Interface{
		Name:     "lo",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: loopbackIPConfig,
	}
	h.Interfaces["eth0"] = model.Interface{
		Name:     "eth0",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: eth0IPConfig,
	}

	// DNS table
	h.DNSTable = make(map[string]model.DNSRecord)

	h.DNSTable[h.Hostname] = model.DNSRecord{
		Name:  h.Hostname,
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}
	h.DNSTable["localhost"] = model.DNSRecord{
		Name:  "localhost",
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}

	Net().Hosts = append(Net().Hosts, h)

	generateHostChannels(getHostIndexFromID(h.ID))
	go listenHostChannel(h, "lo")
	<-engine.Instance().ListenSync
	go listenHostChannel(h, "eth0")
	<-engine.Instance().ListenSync
}

func linkHostTo(localDevice string, remoteDevice string) {
	localDevice = strings.ToUpper(localDevice)
	remoteDevice = strings.ToUpper(remoteDevice)

	//Make sure there's enough ports - if uplink device is a router
	if remoteDevice == strings.ToUpper(Net().Router.Hostname) {
		if getActivePorts(Net().Router.VSwitch) >= Net().Router.VSwitch.Maxports {
			fmt.Printf("No available ports - %s only has %d ports\n", Net().Router.Model, Net().Router.VSwitch.Maxports)
			return
		}
	}

	//Make sure there's enough ports - if uplink device is a switch
	for s := range Net().Switches {
		if remoteDevice == strings.ToUpper(Net().Switches[s].Hostname) {
			if getActivePorts(Net().Switches[s]) >= Net().Switches[s].Maxports {
				fmt.Printf("No available ports - %s only has %d ports\n", Net().Switches[s].Model, Net().Switches[s].Maxports)
				return
			}
		}
	}

	//find host with that hostname
	for i := range Net().Hosts {
		if strings.ToUpper(Net().Hosts[i].Hostname) == localDevice {
			uplinkID := ""
			//Remote device on new link is the Router
			if remoteDevice == strings.ToUpper(Net().Router.Hostname) {
				//find next free port
				portIndex := assignSwitchport(Net().Router.VSwitch, Net().Hosts[i].Interfaces["eth0"].L1ID)
				uplinkID = Net().Router.VSwitch.PortLinksLocal[portIndex]

			} else {
				//Remote device on the new link is not the Router. Search switches
				for j := range Net().Switches {
					if remoteDevice == strings.ToUpper(Net().Switches[j].Hostname) {
						//find next free port
						portIndex := assignSwitchport(Net().Switches[j], Net().Hosts[i].Interfaces["eth0"].L1ID)
						uplinkID = Net().Switches[j].PortLinksLocal[portIndex]

					}
				}
			}

			// Assign uplink ID to host
			iface := Net().Hosts[i].Interfaces["eth0"]
			iface.RemoteL1ID = uplinkID
			Net().Hosts[i].Interfaces["eth0"] = iface

			return
		}
	}
}

func unlinkHost(hostname string) {
	hostname = strings.ToUpper(hostname)

	for i := range Net().Hosts {
		if strings.ToUpper(Net().Hosts[i].Hostname) == hostname {
			//first, unplug from switch (switch-end unlink). TODO try/catch this whole block.
			freeSwitchport(Net().Hosts[i].Interfaces["eth0"].RemoteL1ID)

			//next, remove the host's uplink (host-end unlink)
			uplinkID := ""
			iface := Net().Hosts[i].Interfaces["eth0"]
			iface.RemoteL1ID = uplinkID
			Net().Router.Interfaces["eth0"] = iface

			return
		}
	}
}

func delHost(hostname string) {
	hostname = strings.ToUpper(hostname)
	//search for host
	for i := range Net().Hosts {
		if strings.ToUpper(Net().Hosts[i].Hostname) == hostname {
			//unlink, Vswitch
			for j := range Net().Router.VSwitch.PortLinksRemote {
				if Net().Router.VSwitch.PortLinksLocal[j] == Net().Hosts[i].Interfaces["eth0"].RemoteL1ID {
					Net().Router.VSwitch.PortLinksRemote[j] = ""

					Net().Hosts = removeHostFromSlice(Net().Hosts, i)
					deviceLog(2, "delHost", hostname, "Host deleted")
					return
				}
			}

			//unlink, other switches
			for sw := range Net().Switches {
				for p := range Net().Switches[sw].PortLinksRemote {
					if Net().Switches[sw].PortLinksLocal[p] == Net().Hosts[i].Interfaces["eth0"].RemoteL1ID {
						Net().Switches[sw].PortLinksRemote[p] = ""

						Net().Hosts = removeHostFromSlice(Net().Hosts, i)
						deviceLog(2, "delHost", hostname, "Host deleted")
						return
					}
				}
			}

			Net().Hosts = removeHostFromSlice(Net().Hosts, i)
			deviceLog(2, "delHost", hostname, "Host deleted")
			return
		}
	}
	systemLog(1, "delHost", fmt.Sprintf("Host %s not found - deletion failed", hostname))
}

func ipclear(id string) {
	index := getHostIndexFromID(id)

	iface := Net().Hosts[index].Interfaces["eth0"]

	iface.IPConfig.IPAddress = nil
	iface.IPConfig.SubnetMask = ""
	iface.IPConfig.DefaultGateway = nil

	Net().Hosts[index].Interfaces["eth0"] = iface

	deviceLog(2, "ipclear", Net().Hosts[index].Hostname, "Cleared IP configuration")
}

func printResolveHostname(srcID string, hostname string, dnsTable map[string]model.DNSRecord) {
	dnsRecord := resolveHostname(srcID, hostname, dnsTable)
	fmt.Println("Name: " + hostname)
	fmt.Println("Address: " + dnsRecord.RData + "\n")

	engine.Instance().ActionSync[srcID] <- 1
}
