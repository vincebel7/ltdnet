/*
File:		listener.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Listener for network and all devices
*/

package main

import "encoding/json"

var channels = make(map[string]chan string) // the physical link
var internal = map[string]chan Frame{}      // for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	channels["FFFFFFFF"] = make(chan string)
	go broadcastlisten()

	generateRouterChannels()

	for i := range snet.Switches {
		generateSwitchChannels(i)
	}

	for i := range snet.Hosts {
		generateHostChannels(i)
	}
}

func generateHostChannels(i int) {
	channels[snet.Hosts[i].ID] = make(chan string)
	internal[snet.Hosts[i].ID] = make(chan Frame)
	actionsync[snet.Hosts[i].ID] = make(chan int)

	go hostlisten(snet.Hosts[i].ID)
}

func generateSwitchChannels(i int) {
	for j := 0; j < getActivePorts(snet.Switches[j]); j++ {
		channels[snet.Switches[i].PortIDs[j]] = make(chan string)
		internal[snet.Switches[i].PortIDs[j]] = make(chan Frame)
		actionsync[snet.Switches[i].PortIDs[j]] = make(chan int)

		go switchportlisten(snet.Switches[i].PortIDs[j])
	}
}

func generateRouterChannels() {
	if snet.Router.Hostname != "" {
		channels[snet.Router.ID] = make(chan string)
		internal[snet.Router.ID] = make(chan Frame)
		go routerlisten()

		//vswitch
		//channels[snet.Router.VSwitch.ID] = make(chan Frame)
		//internal[snet.Router.VSwitch.ID] = make(chan Frame)
		for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
			channels[snet.Router.VSwitch.PortIDs[i]] = make(chan string)
			internal[snet.Router.VSwitch.PortIDs[i]] = make(chan Frame)
			actionsync[snet.Router.ID] = make(chan int)

			go switchportlisten(snet.Router.VSwitch.PortIDs[i])
		}
	}
}

func broadcastlisten() { //Listens for broadcast frames on FF.. and broadcasts
	for {
		frameString := <-channels["FFFFFFFF"]
		debug(4, "broadcastlisten", "Listener", "detected broadcast")

		go actionHandler(frameString, snet.Router.ID)

		for i := range snet.Hosts {
			go actionHandler(frameString, snet.Hosts[i].ID)
		}
	}
}

func hostlisten(id string) {
	listenSync <- id //synchronizing with client.go

	for {
		frameString := <-channels[id]
		debug(4, "hostlisten", id, "Received unicast frame")

		go actionHandler(frameString, id)
	}
}

func routerlisten() {
	for {
		frameString := <-channels[snet.Router.ID]
		debug(4, "routerlisten", snet.Router.ID, "Received unicast frame")
		go actionHandler(frameString, snet.Router.ID)
	}
}

// TODO eventually break this down into functions so it isn't so nested
func actionHandler(frameString string, id string) {
	debug(4, "actionHandler", id, "My packet")

	frame := readFrame(json.RawMessage(frameString))

	switch frame.EtherType {
	case "0x0806": // ARP
		arpMessage := readArpMessage(frame.Data)
		if arpMessage.Opcode == 2 {
			debug(3, "actionHandler", id, "ARPREPLY received")

			if snet.Router.ID == id {
				if arpMessage.TargetIP == snet.Router.Gateway.String() {
					internal[id] <- frame
				} else {
					debug(4, "actionhandler", id, "I'm the router. Not sure why i got this ARPREPLY not involving me.")
				}
			} else {
				if arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].IPAddr.String() {
					internal[id] <- frame
				} else {
					debug(4, "actionhandler", id, "I'm an uninvolved host. Not sure why i got this ARPREPLY not involving me.")
				}
			}
		}

		if arpMessage.Opcode == 1 {
			debug(3, "actionHandler", id, "ARPREQUEST received")

			// Temporary fix until arp_reply is generalized.
			if snet.Router.ID == id {
				if arpMessage.TargetIP == snet.Router.Gateway.String() {
					index := 0
					arp_reply(index, "router", frame)
				}
			} else {
				if arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].IPAddr.String() {
					arp_reply(getHostIndexFromID(id), "host", frame)
				}
			}
		}

	case "0x0800": // IPv4
		packet := readIPv4Packet(frame.Data)

		switch readIPv4PacketHeader(packet.Header).Protocol {
		case 1: // ICMP
			icmpPacket := readICMPPacket(packet.Data)
			if icmpPacket.ControlType == 8 {
				debug(3, "actionHandler", id, "Ping received")
				pong(id, frame)
			}

			if icmpPacket.ControlType == 0 {
				debug(3, "actionHandler", id, "Pong received")
				internal[id] <- frame
			}

		case 17: // UDP
			udpSegment := readUDPSegment(packet.Data)

			switch udpSegment.DstPort {
			case 67: // Server-bound messages
				if snet.Router.ID != id { // Temporary validation
					debug(4, "actionHandler", id, "DHCP server traffic received on host. Ignoring")
					return
				}

				if string(udpSegment.Data) == "DHCPDISCOVER" { // TODO not sure if this []byte to string thing works.
					debug(2, "actionHandler", id, "DHCPDISCOVER received")
					dhcp_offer(frame)
				}

				if len(string(udpSegment.Data)) > 10 { // TODO not sure if this []byte to string thing works.
					if string(udpSegment.Data)[0:11] == "DHCPREQUEST" { // TODO not sure if this []byte to string thing works.
						debug(2, "actionHandler", id, "DHCPREQUEST received")
						internal[id] <- frame
					}
				}

			case 68: // Client-bound messages
				if len(string(udpSegment.Data)) > 8 { // TODO not sure if this []byte to string thing works.
					if string(udpSegment.Data)[0:9] == "DHCPOFFER" { // TODO not sure if this []byte to string thing works.
						debug(2, "actionHandler", id, "DHCPOFFER received")
						internal[id] <- frame
					}
				}

				if len(string(udpSegment.Data)) > 17 { // TODO not sure if this []byte to string thing works.
					if string(udpSegment.Data)[0:19] == "DHCPACKNOWLEDGEMENT" { // TODO not sure if this []byte to string thing works.
						debug(2, "actionHandler", id, "DHCPACKNOWLEDGEMENT received")
						internal[id] <- frame
					}

				}
			}
		}
	}
}

func switchportlisten(id string) {
	for {
		frameString := <-channels[id]
		debug(4, "switchlisten", id, "(Switch) Received unicast frame")

		port := getSwitchportIDFromLink(id)

		checkMACTable(readFrame(json.RawMessage(frameString)).SrcMAC, id, port)

		go switchportactionhandler(frameString, id)
	}
}

func switchportactionhandler(frameString string, id string) {
	debug(4, "switchportactionhandler", id, "My packet")
	if readFrame(json.RawMessage(frameString)).DstMAC == "FF:FF:FF:FF:FF:FF" {
		channels["FFFFFFFF"] <- frameString
	} else if 1 == 2 { //TODO how to receive mgmt frames
		//data := frame.Data.Data.Data
	} else {
		switchforward(readFrame(json.RawMessage(frameString)), id)
	}
}
