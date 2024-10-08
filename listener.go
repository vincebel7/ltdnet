/*
File:		listener.go
Author: 	https://github.com/vincebel7
Purpose:	Listener for network and all devices
*/

package main

import "encoding/json"

var channels = make(map[string]chan json.RawMessage)    // Physical links
var socketMaps = make(map[string]map[string]chan Frame) // For internal device communication
var actionsync = map[string]chan int{}                  // Blocks CLI prompt until action completes

func Listener() {
	generateBroadcastChannel()
	generateRouterChannels()

	for i := range snet.Switches {
		generateSwitchChannels(i)
	}

	for i := range snet.Hosts {
		generateHostChannels(i)
	}
}

func generateBroadcastChannel() {
	channels["FFFFFFFF"] = make(chan json.RawMessage)

	go listenBroadcastChannel()
}

func generateHostChannels(i int) {
	channels[snet.Hosts[i].ID] = make(chan json.RawMessage)
	socketMaps[snet.Hosts[i].ID] = make(map[string]chan Frame)
	actionsync[snet.Hosts[i].ID] = make(chan int)

	go listenHostChannel(snet.Hosts[i].ID)
}

func generateSwitchChannels(i int) {
	for j := 0; j < getActivePorts(snet.Switches[j]); j++ {
		channels[snet.Switches[i].PortIDs[j]] = make(chan json.RawMessage)
		socketMaps[snet.Switches[i].PortIDs[j]] = make(map[string]chan Frame)
		actionsync[snet.Switches[i].PortIDs[j]] = make(chan int)

		go listenSwitchportChannel(snet.Switches[i].PortIDs[j])
	}
}

func generateRouterChannels() {
	if snet.Router.Hostname != "" {
		channels[snet.Router.ID] = make(chan json.RawMessage)
		socketMaps[snet.Router.ID] = make(map[string]chan Frame)

		go listenRouterChannel()

		for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
			channels[snet.Router.VSwitch.PortIDs[i]] = make(chan json.RawMessage)
			socketMaps[snet.Router.VSwitch.PortIDs[i]] = make(map[string]chan Frame)
			actionsync[snet.Router.ID] = make(chan int)

			go listenSwitchportChannel(snet.Router.VSwitch.PortIDs[i])
		}
	}
}

func listenBroadcastChannel() { //Listens for broadcast frames on FF.. and broadcasts
	for {
		rawFrame := <-channels["FFFFFFFF"]
		debug(4, "listenBroadcastChannel", "Listener", "detected broadcast")

		go actionHandler(rawFrame, snet.Router.ID)

		for i := range snet.Hosts {
			go actionHandler(rawFrame, snet.Hosts[i].ID)
		}
	}
}

func listenHostChannel(id string) {
	listenSync <- id //synchronizing with client.go

	for {
		rawFrame := <-channels[id]
		debug(4, "listenHostChannel", id, "Received unicast frame")
		go actionHandler(rawFrame, id)
	}
}

func listenRouterChannel() {
	for {
		rawFrame := <-channels[snet.Router.ID]
		debug(4, "listenRouterChannel", snet.Router.ID, "Received unicast frame")
		go actionHandler(rawFrame, snet.Router.ID)
	}
}

// Should actions be broken into functions?
func actionHandler(rawFrame json.RawMessage, id string) {
	debug(4, "actionHandler", id, "About to handle a frame")

	frame := readFrame(rawFrame)

	switch frame.EtherType {
	case "0x0806": // ARP
		arpMessage := readArpMessage(frame.Data)
		switch arpMessage.Opcode {
		case 2:
			debug(3, "actionHandler", id, "ARPREPLY received")

			amTarget := false
			if (snet.Router.ID == id) && (arpMessage.TargetIP == snet.Router.Gateway.String()) {
				amTarget = true
			} else if (snet.Router.ID != id) && (arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].IPAddr.String()) {
				amTarget = true
			}
			if amTarget {
				sockets := socketMaps[id]
				socketID := "arp_" + string(arpMessage.SenderIP)
				sockets[socketID] <- frame
			}

		case 1:
			debug(3, "actionHandler", id, "ARPREQUEST received")

			// Check if target device at network-level
			amTarget := false
			if (snet.Router.ID == id) && (arpMessage.TargetIP == snet.Router.Gateway.String()) {
				amTarget = true
			} else if (snet.Router.ID != id) && (arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].IPAddr.String()) {
				amTarget = true
			}

			if amTarget {
				arp_reply(id, frame)
			}
		}

	case "0x0800": // IPv4
		packet := readIPv4Packet(frame.Data)

		switch readIPv4PacketHeader(packet.Header).Protocol {
		case 1: // ICMP
			icmpPacket := readICMPEchoPacket(packet.Data)

			switch icmpPacket.ControlType {
			case 8:
				debug(3, "actionHandler", id, "Ping request received")
				pong(id, frame)

			case 0:
				debug(3, "actionHandler", id, "Ping reply received")
				sockets := socketMaps[id]
				socketID := "icmp_" + string(icmpPacket.Identifier)
				sockets[socketID] <- frame
			}

		case 17: // UDP
			udpSegment := readUDPSegment(packet.Data)

			switch udpSegment.DstPort {
			case 53: // DNS
				return
			case 67: // DHCP: Server-bound
				if snet.Router.ID != id { // Temporary validation
					debug(4, "actionHandler", id, "DHCP server traffic received on host. Ignoring")
				}

				if string(udpSegment.Data) == "DHCPDISCOVER" {
					debug(3, "actionHandler", id, "DHCPDISCOVER received")
					dhcp_offer(frame)
				}

				if len(string(udpSegment.Data)) > 10 {
					if string(udpSegment.Data)[0:11] == "DHCPREQUEST" {
						debug(3, "actionHandler", id, "DHCPREQUEST received")
						sockets := socketMaps[id]
						socketID := "udp_" + string(udpSegment.DstPort)
						sockets[socketID] <- frame
					}
				}

			case 68: // DHCP: Client-bound
				if len(string(udpSegment.Data)) > 8 {
					if string(udpSegment.Data)[0:9] == "DHCPOFFER" {
						debug(3, "actionHandler", id, "DHCPOFFER received")
						sockets := socketMaps[id]
						socketID := "udp_" + string(udpSegment.DstPort)
						sockets[socketID] <- frame
					}
				}

				if len(string(udpSegment.Data)) > 17 {
					if string(udpSegment.Data)[0:19] == "DHCPACKNOWLEDGEMENT" {
						debug(3, "actionHandler", id, "DHCPACKNOWLEDGEMENT received")
						socketID := "udp_" + string(udpSegment.DstPort)
						sockets := socketMaps[id]
						sockets[socketID] <- frame
					}

				}
			}
		}
	}
}

func listenSwitchportChannel(id string) {
	for {
		rawFrame := <-channels[id]
		debug(4, "listenSwitchportChannel", id, "(Switch) Received unicast frame")

		port := getSwitchportIDFromLink(id)
		checkMACTable(readFrame(rawFrame).SrcMAC, id, port)

		go switchportActionHandler(rawFrame, id)
	}
}

func switchportActionHandler(rawFrame json.RawMessage, id string) {
	debug(4, "switchportActionHandler", id, "My packet")
	if readFrame(rawFrame).DstMAC == "FF:FF:FF:FF:FF:FF" {
		channels["FFFFFFFF"] <- rawFrame
	} else if false { //TODO how to receive mgmt frames
		//data := frame.Data.Data.Data
	} else {
		switchforward(readFrame(rawFrame), id)
	}
}
