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
	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

func Listener() {
	generateRouterChannels()

	for i := range Net().Switches {
		generateSwitchChannels(i)
	}

	for i := range Net().Hosts {
		generateHostChannels(i)
	}

	// Listen on channels
	r := Net().Router
	if r != nil { // no router added yet
		for iface := range Net().Router.Interfaces {
			go listenRouterChannel(iface)
		}

		for i := 0; i < getActivePorts(Net().Router.VSwitch); i++ {
			go listenSwitchportChannel(Net().Router.VSwitch.ID, Net().Router.VSwitch.PortLinksLocal[i])
		}
	}

	for i := range Net().Switches {
		for j := 0; j < getActivePorts(Net().Switches[i]); j++ {
			go listenSwitchportChannel(Net().Switches[i].ID, Net().Switches[i].PortLinksLocal[j])
		}
	}

	for i := range Net().Hosts {
		for iface := range Net().Hosts[i].Interfaces {
			go listenHostChannel(Net().Hosts[i], iface)
		}
	}

}

func generateHostChannels(i int) {
	for iface := range Net().Hosts[i].Interfaces {
		engine.Instance().Channels[Net().Hosts[i].Interfaces[iface].L1ID] = make(chan json.RawMessage)
	}
	engine.Instance().Sockets[Net().Hosts[i].ID] = make(map[string]chan model.Frame)
	engine.Instance().ActionSync[Net().Hosts[i].ID] = make(chan int)
}

func generateSwitchChannels(i int) {
	for j := 0; j < getActivePorts(Net().Switches[i]); j++ {
		engine.Instance().Channels[Net().Switches[i].PortLinksLocal[j]] = make(chan json.RawMessage)
		engine.Instance().Sockets[Net().Switches[i].PortLinksLocal[j]] = make(map[string]chan model.Frame)
		engine.Instance().ActionSync[Net().Switches[i].PortLinksLocal[j]] = make(chan int)
	}
}

func generateRouterChannels() {
	r := Net().Router
	if r == nil { // no router added yet
		return
	}
	eng := engine.Instance()

	for iface := range Net().Router.Interfaces {
		eng.Channels[Net().Router.Interfaces[iface].L1ID] = make(chan json.RawMessage)
	}
	eng.Sockets[Net().Router.ID] = make(map[string]chan model.Frame)

	for i := 0; i < getActivePorts(Net().Router.VSwitch); i++ {
		eng.Channels[Net().Router.VSwitch.PortLinksLocal[i]] = make(chan json.RawMessage)
		eng.Sockets[Net().Router.VSwitch.PortLinksLocal[i]] = make(map[string]chan model.Frame)
		eng.ActionSync[Net().Router.ID] = make(chan int)
	}
}

func listenHostChannel(host model.Host, iface string) {
	engine.Instance().ListenSync <- host.ID //synchronizing with client.go

	for {
		rawFrame := <-engine.Instance().Channels[host.Interfaces[iface].L1ID]
		writeLog(4, "listenHostChannel", host.Hostname, "Received unicast frame")
		go actionHandler(rawFrame, host.ID, iface)
	}
}

func listenRouterChannel(iface string) {
	for {
		rawFrame := <-engine.Instance().Channels[Net().Router.Interfaces[iface].L1ID]
		writeLog(4, "listenRouterChannel", Net().Router.ID, "Received unicast frame")
		go actionHandler(rawFrame, Net().Router.ID, iface)
	}
}

// Should actions be broken into functions?
func actionHandler(rawFrame json.RawMessage, id string, iface string) {
	frame, _ := model.ParseFrame(rawFrame)

	switch frame.EtherType {
	case "0x0806": // ARP
		arpMessage, _ := model.ParseARPMessage(frame.Data)
		switch arpMessage.Opcode {
		case 2:
			writeLog(2, "actionHandler", id, "ARPREPLY received")

			amTarget := false
			shouldRespond := false

			if (Net().Router.ID == id) && (arpMessage.TargetIP == Net().Router.GetIP(iface)) {
				amTarget = true

				if iphelper.IPInSameSubnet(arpMessage.SenderIP, Net().Router.GetIP(iface), Net().Router.GetMask(iface)) {
					shouldRespond = true
				}

			} else if (Net().Router.ID != id) && (arpMessage.TargetIP == Net().Hosts[getHostIndexFromID(id)].GetIP(iface)) {
				amTarget = true

				if iphelper.IPInSameSubnet(arpMessage.SenderIP, Net().Hosts[getHostIndexFromID(id)].GetIP(iface), Net().Hosts[getHostIndexFromID(id)].GetMask(iface)) {
					shouldRespond = true
				}
			}

			if amTarget && shouldRespond {
				sockets := engine.Instance().Sockets[id]
				socketID := "arp_" + string(arpMessage.SenderIP)
				sockets[socketID] <- frame
			}

		case 1:
			writeLog(2, "actionHandler", id, "ARPREQUEST received")

			// Check if target device at network-level
			amTarget := false
			if (Net().Router.ID == id) && (arpMessage.TargetIP == Net().Router.GetIP(iface)) {
				amTarget = true
			} else if (Net().Router.ID != id) && (arpMessage.TargetIP == Net().Hosts[getHostIndexFromID(id)].GetIP(iface)) {
				amTarget = true
			}

			if amTarget {
				arp_reply(id, frame)
			}
		}

	case "0x0800": // IPv4
		packet, _ := model.ParseIPv4Packet(frame.Data)
		packetHeader, _ := model.ParseIPv4PacketHeader(packet.Header)

		switch packetHeader.Protocol {
		case 1: // ICMP
			icmpPacket, _ := model.ParseICMPEchoPacket(packet.Data)

			switch icmpPacket.ControlType {
			case 8:
				writeLog(2, "actionHandler", id, "Ping request received")

				// Check if target device at network-level
				amTarget := false
				if (Net().Router.ID == id) && (packetHeader.DstIP == Net().Router.GetIP(iface)) {
					amTarget = true
				} else if (Net().Router.ID != id) && (packetHeader.DstIP == Net().Hosts[getHostIndexFromID(id)].GetIP(iface)) {
					amTarget = true
				}

				if amTarget {
					pong(id, frame)
				}

			case 0:
				writeLog(2, "actionHandler", id, "Ping reply received")

				// Check if target device at network-level
				amTarget := false
				if (Net().Router.ID == id) && (packetHeader.DstIP == Net().Router.GetIP(iface)) {
					amTarget = true
				} else if (Net().Router.ID != id) && (packetHeader.DstIP == Net().Hosts[getHostIndexFromID(id)].GetIP(iface)) {
					amTarget = true
				}

				if amTarget {
					sockets := engine.Instance().Sockets[id]
					socketID := "icmp_" + strconv.Itoa(icmpPacket.Identifier)
					sockets[socketID] <- frame
				}
			}

		case 6: // TCP
			tcpSegment, _ := model.ParseTCPSegment(packet.Data)

			switch tcpSegment.DstPort {
			case 23: // Telnet
			case 80: // HTTP
			}

		case 17: // UDP
			udpSegment, _ := model.ParseUDPSegment(packet.Data)

			switch udpSegment.DstPort {
			case 53: // DNS
				dnsMessage, _ := model.ParseDNSMessage(json.RawMessage(udpSegment.Data))

				if !dnsMessage.QR {
					writeLog(2, "actionHandler", id, "DNS query received")
					dns_response(frame)
				}

			case 67: // DHCP: Server-bound
				if Net().Router.ID == id { // I am target
					dhcpMessage, _ := model.ParseDHCPMessage(json.RawMessage(udpSegment.Data))

					// 53 is DHCP message type
					if option53, ok := dhcpMessage.Options[53]; ok && len(option53) > 0 {
						switch int(option53[0]) {
						case 1: // DHCPDISCOVER
							writeLog(2, "actionHandler", id, "DHCPDISCOVER received")
							dhcp_offer(frame)

						case 3: // DHCPREQUEST
							writeLog(2, "actionHandler", id, "DHCPREQUEST received")
							dhcp_ack(frame)

						case 2, 4, 5:
							writeLog(4, "actionHandler", id, "DHCP server traffic received on host. Ignoring")

						default:
							writeLog(1, "actionHandler", id, "Unhandled DHCP message type:"+string(option53[0]))
						}
					} else {
						writeLog(1, "actionHandler", id, "DHCP Option 53 is missing or empty")
					}
				}
			case 68: // DHCP: Client-bound
				dhcpMessage, _ := model.ParseDHCPMessage(json.RawMessage(udpSegment.Data))

				if dhcpMessage.CHAddr == Net().Hosts[getHostIndexFromID(id)].Interfaces[iface].MACAddr { // I am target
					// 53 is DHCP message type
					if option53, ok := dhcpMessage.Options[53]; ok && len(option53) > 0 {
						switch int(option53[0]) {
						case 2: // DHCPOFFER
							writeLog(2, "actionHandler", id, "DHCPOFFER received")
							sockets := engine.Instance().Sockets[id]
							socketID := "udp_" + strconv.Itoa(udpSegment.DstPort)
							sockets[socketID] <- frame

						case 5: // DHCPACK
							writeLog(2, "actionHandler", id, "DHCPACK received")
							socketID := "udp_" + strconv.Itoa(udpSegment.DstPort)
							sockets := engine.Instance().Sockets[id]
							sockets[socketID] <- frame

						default:
							writeLog(1, "actionHandler", id, "Unhandled DHCP message type:"+string(option53[0]))
						}
					} else {
						writeLog(1, "actionHandler", id, "DHCP Option 53 is missing or empty")
					}
				}
			default: // Ephemeral
				portStr := strconv.Itoa(udpSegment.DstPort)
				writeLog(2, "actionHandler", id, "Ephemeral port ("+portStr+") response received")
				sockets := engine.Instance().Sockets[id]
				socketID := "udp_" + portStr
				sockets[socketID] <- frame
			}
		}
	}
}

func listenSwitchportChannel(switchID string, switchportID string) {
	for {
		rawFrame := <-engine.Instance().Channels[switchportID]
		writeLog(4, "listenSwitchportChannel", switchportID, "(Switch) Received frame from port "+switchportID)
		port := getSwitchportIDFromLink(switchportID)

		frame, _ := model.ParseFrame(rawFrame)
		checkMACTable(frame.SrcMAC, switchportID, port)

		go switchportActionHandler(rawFrame, switchID, switchportID)
	}
}

func switchportActionHandler(rawFrame json.RawMessage, switchID string, switchportID string) {
	if false { // Traffic for switch. TODO how to receive mgmt frames
		//data := frame.Data.Data.Data
	} else { // Normal frame forward
		frame, _ := model.ParseFrame(rawFrame)
		switchforward(frame, switchID, switchportID)
	}
}
