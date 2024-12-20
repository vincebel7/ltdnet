/*
File:		listener.go
Author: 	https://github.com/vincebel7
Purpose:	Listener for network and all devices
*/

package main

import (
	"encoding/json"
	"strconv"

	"github.com/vincebel7/ltdnet/iphelper"
)

var channels = make(map[string]chan json.RawMessage)    // Physical links
var socketMaps = make(map[string]map[string]chan Frame) // For internal device communication
var actionsync = map[string]chan int{}                  // Blocks CLI prompt until action completes

func Listener() {
	// Generate channels
	generateRouterChannels()

	for i := range snet.Switches {
		generateSwitchChannels(i)
	}

	for i := range snet.Hosts {
		generateHostChannels(i)
	}

	// Listen on channels
	if snet.Router.Hostname != "" {
		for iface := range snet.Router.Interfaces {
			go listenRouterChannel(iface)
		}

		for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
			go listenSwitchportChannel(snet.Router.VSwitch.ID, snet.Router.VSwitch.PortLinksLocal[i])
		}
	}

	for i := range snet.Switches {
		for j := 0; j < getActivePorts(snet.Switches[i]); j++ {
			go listenSwitchportChannel(snet.Switches[i].ID, snet.Switches[i].PortLinksLocal[j])
		}
	}

	for i := range snet.Hosts {
		for iface := range snet.Hosts[i].Interfaces {
			go listenHostChannel(snet.Hosts[i], iface)
		}
	}

}

func generateHostChannels(i int) {
	for iface := range snet.Hosts[i].Interfaces {
		channels[snet.Hosts[i].Interfaces[iface].L1ID] = make(chan json.RawMessage)
	}
	socketMaps[snet.Hosts[i].ID] = make(map[string]chan Frame)
	actionsync[snet.Hosts[i].ID] = make(chan int)
}

func generateSwitchChannels(i int) {
	for j := 0; j < getActivePorts(snet.Switches[i]); j++ {
		channels[snet.Switches[i].PortLinksLocal[j]] = make(chan json.RawMessage)
		socketMaps[snet.Switches[i].PortLinksLocal[j]] = make(map[string]chan Frame)
		actionsync[snet.Switches[i].PortLinksLocal[j]] = make(chan int)
	}
}

func generateRouterChannels() {
	if snet.Router.Hostname != "" {
		for iface := range snet.Router.Interfaces {
			channels[snet.Router.Interfaces[iface].L1ID] = make(chan json.RawMessage)
		}
		socketMaps[snet.Router.ID] = make(map[string]chan Frame)

		for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
			channels[snet.Router.VSwitch.PortLinksLocal[i]] = make(chan json.RawMessage)
			socketMaps[snet.Router.VSwitch.PortLinksLocal[i]] = make(map[string]chan Frame)
			actionsync[snet.Router.ID] = make(chan int)
		}
	}
}

func listenHostChannel(host Host, iface string) {
	listenSync <- host.ID //synchronizing with client.go

	for {
		rawFrame := <-channels[host.Interfaces[iface].L1ID]
		debug(4, "listenHostChannel", host.Hostname, "Received unicast frame")
		go actionHandler(rawFrame, host.ID, iface)
	}
}

func listenRouterChannel(iface string) {
	for {
		rawFrame := <-channels[snet.Router.Interfaces[iface].L1ID]
		debug(4, "listenRouterChannel", snet.Router.ID, "Received unicast frame")
		go actionHandler(rawFrame, snet.Router.ID, iface)
	}
}

// Should actions be broken into functions?
func actionHandler(rawFrame json.RawMessage, id string, iface string) {
	frame := readFrame(rawFrame)

	switch frame.EtherType {
	case "0x0806": // ARP
		arpMessage := readArpMessage(frame.Data)
		switch arpMessage.Opcode {
		case 2:
			debug(2, "actionHandler", id, "ARPREPLY received")

			amTarget := false
			shouldRespond := false

			if (snet.Router.ID == id) && (arpMessage.TargetIP == snet.Router.GetIP(iface)) {
				amTarget = true

				if iphelper.IPInSameSubnet(arpMessage.SenderIP, snet.Router.GetIP(iface), snet.Router.GetMask(iface)) {
					shouldRespond = true
				}

			} else if (snet.Router.ID != id) && (arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].GetIP(iface)) {
				amTarget = true

				if iphelper.IPInSameSubnet(arpMessage.SenderIP, snet.Hosts[getHostIndexFromID(id)].GetIP(iface), snet.Hosts[getHostIndexFromID(id)].GetMask(iface)) {
					shouldRespond = true
				}
			}

			if amTarget && shouldRespond {
				sockets := socketMaps[id]
				socketID := "arp_" + string(arpMessage.SenderIP)
				sockets[socketID] <- frame
			}

		case 1:
			debug(2, "actionHandler", id, "ARPREQUEST received")

			// Check if target device at network-level
			amTarget := false
			if (snet.Router.ID == id) && (arpMessage.TargetIP == snet.Router.GetIP(iface)) {
				amTarget = true
			} else if (snet.Router.ID != id) && (arpMessage.TargetIP == snet.Hosts[getHostIndexFromID(id)].GetIP(iface)) {
				amTarget = true
			}

			if amTarget {
				arp_reply(id, frame)
			}
		}

	case "0x0800": // IPv4
		packet := readIPv4Packet(frame.Data)
		packetHeader := readIPv4PacketHeader(packet.Header)

		switch packetHeader.Protocol {
		case 1: // ICMP
			icmpPacket := readICMPEchoPacket(packet.Data)

			switch icmpPacket.ControlType {
			case 8:
				debug(2, "actionHandler", id, "Ping request received")

				// Check if target device at network-level
				amTarget := false
				if (snet.Router.ID == id) && (packetHeader.DstIP == snet.Router.GetIP(iface)) {
					amTarget = true
				} else if (snet.Router.ID != id) && (packetHeader.DstIP == snet.Hosts[getHostIndexFromID(id)].GetIP(iface)) {
					amTarget = true
				}

				if amTarget {
					pong(id, frame)
				}

			case 0:
				debug(2, "actionHandler", id, "Ping reply received")

				// Check if target device at network-level
				amTarget := false
				if (snet.Router.ID == id) && (packetHeader.DstIP == snet.Router.GetIP(iface)) {
					amTarget = true
				} else if (snet.Router.ID != id) && (packetHeader.DstIP == snet.Hosts[getHostIndexFromID(id)].GetIP(iface)) {
					amTarget = true
				}

				if amTarget {
					sockets := socketMaps[id]
					socketID := "icmp_" + strconv.Itoa(icmpPacket.Identifier)
					sockets[socketID] <- frame
				}
			}

		case 6: // TCP
			tcpSegment := readTCPSegment(packet.Data)

			switch tcpSegment.DstPort {
			case 23: // Telnet
			case 80: // HTTP
			}

		case 17: // UDP
			udpSegment := readUDPSegment(packet.Data)

			switch udpSegment.DstPort {
			case 53: // DNS
				dnsMessage := ReadDNSMessage(json.RawMessage(udpSegment.Data))

				if !dnsMessage.QR {
					debug(2, "actionHandler", id, "DNS query received")
					dns_response(frame)
				}

			case 67: // DHCP: Server-bound
				if snet.Router.ID == id { // I am target
					dhcpMessage := ReadDHCPMessage(json.RawMessage(udpSegment.Data))

					// 53 is DHCP message type
					if option53, ok := dhcpMessage.Options[53]; ok && len(option53) > 0 {
						switch int(option53[0]) {
						case 1: // DHCPDISCOVER
							debug(2, "actionHandler", id, "DHCPDISCOVER received")
							dhcp_offer(frame)

						case 3: // DHCPREQUEST
							debug(2, "actionHandler", id, "DHCPREQUEST received")
							dhcp_ack(frame)

						case 2, 4, 5:
							debug(4, "actionHandler", id, "DHCP server traffic received on host. Ignoring")

						default:
							debug(1, "actionHandler", id, "Unhandled DHCP message type:"+string(option53[0]))
						}
					} else {
						debug(1, "actionHandler", id, "DHCP Option 53 is missing or empty")
					}
				}
			case 68: // DHCP: Client-bound
				dhcpMessage := ReadDHCPMessage(json.RawMessage(udpSegment.Data))

				if dhcpMessage.CHAddr == snet.Hosts[getHostIndexFromID(id)].Interfaces[iface].MACAddr { // I am target
					// 53 is DHCP message type
					if option53, ok := dhcpMessage.Options[53]; ok && len(option53) > 0 {
						switch int(option53[0]) {
						case 2: // DHCPOFFER
							debug(2, "actionHandler", id, "DHCPOFFER received")
							sockets := socketMaps[id]
							socketID := "udp_" + strconv.Itoa(udpSegment.DstPort)
							sockets[socketID] <- frame

						case 5: // DHCPACK
							debug(2, "actionHandler", id, "DHCPACK received")
							socketID := "udp_" + strconv.Itoa(udpSegment.DstPort)
							sockets := socketMaps[id]
							sockets[socketID] <- frame

						default:
							debug(1, "actionHandler", id, "Unhandled DHCP message type:"+string(option53[0]))
						}
					} else {
						debug(1, "actionHandler", id, "DHCP Option 53 is missing or empty")
					}
				}
			default: // Ephemeral
				portStr := strconv.Itoa(udpSegment.DstPort)
				debug(2, "actionHandler", id, "Ephemeral port ("+portStr+") response received")
				sockets := socketMaps[id]
				socketID := "udp_" + portStr
				sockets[socketID] <- frame
			}
		}
	}
}

func listenSwitchportChannel(switchID string, switchportID string) {
	for {
		rawFrame := <-channels[switchportID]
		debug(4, "listenSwitchportChannel", switchportID, "(Switch) Received frame from port "+switchportID)
		port := getSwitchportIDFromLink(switchportID)

		checkMACTable(readFrame(rawFrame).SrcMAC, switchportID, port)

		go switchportActionHandler(rawFrame, switchID, switchportID)
	}
}

func switchportActionHandler(rawFrame json.RawMessage, switchID string, switchportID string) {
	if false { // Traffic for switch. TODO how to receive mgmt frames
		//data := frame.Data.Data.Data
	} else { // Normal frame forward
		switchforward(readFrame(rawFrame), switchID, switchportID)
	}
}
