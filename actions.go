/*
File:		actions.go
Author: 	https://github.com/vincebel7
Purpose:	Defines network functions such as ARP, DHCP, etc.
*/

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

func ping(srcID string, dstIP string, count int) {
	debug(4, "ping", srcID, "About to ping")

	identifier := idgen_int(5)
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""

	// Get srchost
	if snet.Router.ID == srcID {
		srchost = snet.Router.Hostname
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				srchost = snet.Hosts[h].Hostname
			}
		}
	}

	fmt.Printf("\nPinging %s from %s\n", dstIP, srchost)
	timeoutCounter := 0
	sendCount := 0
	recvCount := 0
	lossCount := 0

	for i := 0; i < count; i++ {
		// Get MAC addresses
		if snet.Router.ID == srcID {
			srcIP = snet.Router.Gateway.String()
			srcMAC = snet.Router.MACAddr

			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("Request timed out.\n")
				lossCount++
				timeoutCounter++
				sendCount++
				continue
			}

			//get link to send ping to
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}

		} else { // Assumed to be host source.
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("Request timed out.\n")
				lossCount++
				timeoutCounter++
				sendCount++
				continue
			}
		}

		if srcMAC == "" {
			for h := range snet.Hosts {
				if snet.Hosts[h].ID == srcID {
					linkID = snet.Hosts[h].UplinkID
					srcIP = snet.Hosts[h].IPAddr.String()
					srcMAC = snet.Hosts[h].MACAddr
				}
			}
		}
		debug(4, "ping", srcID, "Constructing ping")
		payload, _ := json.Marshal("101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")

		icmpRequestPacket := ICMPEchoPacket{
			ControlType: 8,
			ControlCode: 0,
			Checksum:    "checksum",
			Identifier:  identifier,
			SeqNumber:   i,
			Data:        json.RawMessage(payload),
		}
		icmpRequestPacketBytes, _ := json.Marshal(icmpRequestPacket)

		ipv4PacketBytes := constructIPv4Packet(srcIP, dstIP, "ICMP", icmpRequestPacketBytes)

		frameBytes := constructFrame(srcMAC, dstMAC, "IPv4", ipv4PacketBytes)

		debug(4, "ping", srcID, "Sending the ping now")
		channels[linkID] <- frameBytes
		debug(2, "ping", srcID, "Ping sent")

		sendCount++

		sockets := socketMaps[srcID]
		socketID := "icmp_" + string(identifier)
		sockets[socketID] = make(chan Frame)
		socketMaps[srcID] = sockets // Write updated map back to the collection

		debug(4, "ping", srcID, "Expecting ping reply on "+srcID)
		select {
		case pongFrame := <-sockets[socketID]:
			pongIpv4Packet := readIPv4Packet(pongFrame.Data)
			pongIcmpPacket := readICMPEchoPacket(pongIpv4Packet.Data)

			if pongIcmpPacket.ControlType == 0 {
				recvCount++
				fmt.Printf("Reply from %s: seq=%d\n", dstIP, i)
				timeoutCounter = 0
			} else {
				debug(1, "ping", srcID, "Error: Out-of-order channel")
			}
		case <-time.After(time.Second * 4):
			lossCount++
			fmt.Printf("Request timed out.\n")
			timeoutCounter++
		}

		if timeoutCounter == 4 { //Skip rest of pings if timeout
			i = count
		}

		if i < count-1 { //Only wait a second if
			time.Sleep(time.Second)
		}
	}
	actionsync[srcID] <- 1

	// Ping stats
	fmt.Printf("\nPing statistics for %s:\n", dstIP)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))
}

func pong(srcID string, frame Frame) {
	receivedIpv4Packet := readIPv4Packet(frame.Data)
	receivedIcmpPacket := readICMPEchoPacket(receivedIpv4Packet.Data)

	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstIP := readIPv4PacketHeader(receivedIpv4Packet.Header).SrcIP
	dstMAC := frame.SrcMAC // Get MAC myself via ARP/MAC table, or use request's source MAC?
	if snet.Router.ID == srcID {
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway.String()

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcID, dstIP)
		debug(4, "pong", srcID, "ARP completed. Dstmac acquired. dstMAC: "+dstMAC)

		//Get link to send ping to
		dstID := getIDfromMAC(dstMAC)
		if getHostIndexFromID(dstID) != -1 {
			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		} else if snet.Router.ID == dstID {
			linkID = snet.Router.ID
		}

	} else {
		index := getHostIndexFromID(srcID)
		srcMAC = snet.Hosts[index].MACAddr
		srcIP = snet.Hosts[index].IPAddr.String()

		linkID = snet.Hosts[index].UplinkID
	}

	icmpReplyPacket := ICMPEchoPacket{
		ControlType: 0,
		ControlCode: 0,
		Checksum:    "checksum",
		Identifier:  receivedIcmpPacket.Identifier,
		SeqNumber:   receivedIcmpPacket.SeqNumber,
		Data:        receivedIcmpPacket.Data,
	}
	icmpReplyPacketBytes, _ := json.Marshal(icmpReplyPacket)

	ipv4PacketBytes := constructIPv4Packet(srcIP, dstIP, "ICMP", icmpReplyPacketBytes)

	frameBytes := constructFrame(srcMAC, dstMAC, "IPv4", ipv4PacketBytes)

	debug(4, "pong", srcID, "Awaiting pong send")
	channels[linkID] <- frameBytes
	debug(2, "pong", srcID, "Pong sent")
}

func arp_request(srcID string, targetIP string) string {
	debug(4, "arp_request", srcID, "About to ARP request")

	// Construct frame
	linkID := "FFFFFFFF"
	srcMAC := ""
	srcIP := ""
	dstMAC := "00:00:00:00:00:00"

	if srcID == snet.Router.ID {
		srcIP = snet.Router.Gateway.String()
		srcMAC = snet.Router.MACAddr
	} else {
		index := getHostIndexFromID(srcID)
		srcIP = snet.Hosts[index].IPAddr.String()
		srcMAC = snet.Hosts[index].MACAddr
	}

	arpRequestMessage := ArpMessage{
		HTYPE:     1,
		PTYPE:     "0x800",
		HLEN:      6,
		PLEN:      4,
		Opcode:    1,
		SenderMAC: srcMAC,
		SenderIP:  srcIP,
		TargetMAC: dstMAC,
		TargetIP:  targetIP,
	}

	arpRequestMessageBytes, _ := json.Marshal(arpRequestMessage)
	arpRequestFrameBytes := constructFrame(srcMAC, dstMAC, "ARP", arpRequestMessageBytes)

	// Send frame and wait for ARPREPLY
	channels[linkID] <- arpRequestFrameBytes
	debug(2, "arp_request", srcID, "ARPREQUEST sent")

	sockets := socketMaps[srcID]
	socketID := "arp_" + string(targetIP)
	sockets[socketID] = make(chan Frame)
	socketMaps[srcID] = sockets // Write updated map back to the collection

	select {
	case arpReplyFrameBytes := <-sockets[socketID]:
		arpReplyMessage := readArpMessage(arpReplyFrameBytes.Data)
		return arpReplyMessage.SenderMAC

	case <-time.After(time.Second * 4):
		debug(1, "arp_request", srcID, "ARP request timed out.")
		return "TIMEOUT"
	}
}

func arp_reply(id string, arpRequestFrame Frame) {
	arpRequestMessage := readArpMessage(arpRequestFrame.Data)

	// Construct frame
	linkID := ""
	srcID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := arpRequestMessage.SenderMAC
	dstIP := arpRequestMessage.SenderIP

	// Network listener decided to reply to this request - no checking needed.
	if id == snet.Router.ID {
		srcID = snet.Router.ID
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway.String()

		// Determine linkID
		dstID := getIDfromMAC(dstMAC)
		if getHostIndexFromID(dstID) != -1 {
			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		} else if snet.Router.ID == dstID {
			linkID = snet.Router.ID
		}

	} else {
		i := getHostIndexFromID(id)
		linkID = snet.Hosts[i].UplinkID
		srcID = snet.Hosts[i].ID
		srcMAC = snet.Hosts[i].MACAddr
		srcIP = snet.Hosts[i].IPAddr.String()
	}

	arpReplyMessage := ArpMessage{
		HTYPE:     1,
		PTYPE:     "0x800",
		HLEN:      6,
		PLEN:      4,
		Opcode:    2,
		SenderMAC: srcMAC,
		SenderIP:  srcIP,
		TargetMAC: dstMAC,
		TargetIP:  dstIP,
	}

	arpReplyMessageBytes, _ := json.Marshal(arpReplyMessage)
	arpReplyFrameBytes := constructFrame(srcMAC, dstMAC, "ARP", arpReplyMessageBytes)

	// Send frame
	channels[linkID] <- arpReplyFrameBytes
	debug(2, "arp_reply", srcID, "ARPREPLY sent")
}

func dhcp_discover(host Host) {
	debug(4, "dhcp_discover", host.ID, "Starting DHCPDISCOVER")
	//get info
	srcIP := host.IPAddr.String()
	srcMAC := host.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := "FF:FF:FF:FF:FF:FF"
	linkID := host.UplinkID

	protocol := "UDP"
	segmentData := constructUDPSegment(68, 67, "DHCPDISCOVER")
	packetData := constructIPv4Packet(srcIP, dstIP, protocol, segmentData)
	frameData := constructFrame(srcMAC, dstMAC, "IPv4", packetData)

	//need to give it to uplink
	channels[linkID] <- frameData
	debug(2, "dhcp_discover", host.ID, "DHCPDISCOVER sent")

	sockets := socketMaps[srcID]
	socketID := "udp_" + string(68)
	sockets[socketID] = make(chan Frame)
	socketMaps[srcID] = sockets // Write updated map back to the collection
	offerFrame := <-sockets[socketID]

	offerIpv4Packet := readIPv4Packet(offerFrame.Data)
	offerIpv4PacketHeader := readIPv4PacketHeader(offerIpv4Packet.Header)
	offerUDPSegment := readUDPSegment(offerIpv4Packet.Data)

	if offerUDPSegment.Data == "DHCPOFFER NOAVAILABLE" {
		debug(1, "dhcp_discover", srcID, "Failed to obtain IP address: No free addresses available")
	} else {
		word := strings.Fields(offerUDPSegment.Data)
		if len(word) > 0 {
			word2 := word[1]
			gateway := word[2]
			snetmask := word[3]

			message := "DHCPREQUEST " + word2
			dstIP = offerIpv4PacketHeader.SrcIP

			protocol := "UDP"
			requestUDPSegment := constructUDPSegment(68, 67, message)
			requestIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, requestUDPSegment)
			requestFrame := constructFrame(srcMAC, dstMAC, "IPv4", requestIPv4Packet)
			channels[linkID] <- requestFrame
			debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent - "+word2)
			//wait for acknowledgement

			sockets := socketMaps[srcID]
			socketID := "udp_" + string(68)
			sockets[socketID] = make(chan Frame)
			socketMaps[srcID] = sockets // Write updated map back to the collection
			ackFrame := <-sockets[socketID]

			ackIpv4Packet := readIPv4Packet(ackFrame.Data)
			ackUDPSegment := readUDPSegment(ackIpv4Packet.Data)

			if ackUDPSegment.Data != "" {
				debug(2, "dhcp_discover", srcID, "DHCPACKNOWLEDGEMENT received - "+ackUDPSegment.Data)

				word = strings.Fields(ackUDPSegment.Data)
				if len(word) > 1 {
					fmt.Println("")
					confirmed_addr := word[1]
					dynamic_assign(srcID, net.ParseIP(confirmed_addr), net.ParseIP(gateway), snetmask)
				} else {
					debug(1, "dhcp_discover", srcID, "Error 5: Empty DHCP acknowledgement\n")
				}
			}

		} else {
			debug(1, "dhcp_discover", srcID, "Error 2: Empty DHCP offer")
		}
	}
	actionsync[srcID] <- 1
}

func dhcp_offer(inc_f Frame) {
	srcIP := snet.Router.Gateway.String()
	dstIP := "255.255.255.255"
	srcMAC := snet.Router.MACAddr
	dstMAC := inc_f.SrcMAC
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid
	//linkID := snet.Hosts[getHostIndexFromID(dstid)].UplinkID

	//find open address
	addr_to_give := snet.Router.NextFreePoolAddress().String()
	gateway := snet.Router.Gateway.String()
	subnetmask := ""
	if snet.Netsize == "8" {
		subnetmask = "255.0.0.0"
	} else if snet.Netsize == "16" {
		subnetmask = "255.255.0.0"
	} else if snet.Netsize == "24" {
		subnetmask = "255.255.255.0"
	}

	message := ""
	if addr_to_give == "" {
		message = "DHCPOFFER NOAVAILABLE"
	} else {
		message = "DHCPOFFER " + addr_to_give + " " + gateway + " " + subnetmask
	}

	protocol := "UDP"
	segmentBytes := constructUDPSegment(67, 68, message)
	packetBytes := constructIPv4Packet(srcIP, dstIP, protocol, segmentBytes)
	frameBytes := constructFrame(srcMAC, dstMAC, "IPv4", packetBytes)
	channels[linkID] <- frameBytes
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - "+addr_to_give)

	// Acknowledge
	socketID := "udp_" + string(67)
	sockets := socketMaps[snet.Router.ID]
	requestFrame := <-sockets[socketID]

	requestIpv4Packet := readIPv4Packet(requestFrame.Data)
	requestIpv4PacketHeader := readIPv4PacketHeader(requestIpv4Packet.Header)
	requestUDPSegment := readUDPSegment(requestIpv4Packet.Data)

	message = ""
	if requestUDPSegment.Data != "" {
		word := strings.Fields(requestUDPSegment.Data)
		if len(word) > 1 {
			if word[1] == addr_to_give {
				message = "DHCPACKNOWLEDGEMENT " + addr_to_give
			} else {
				debug(1, "dhcp_offer", snet.Router.ID, "Error 4: DHCP address requested is not same as offer")
			}
		} else {
			debug(1, "dhcp_offer", snet.Router.ID, "Error 3: Empty DHCP request")
		}
	}

	dstIP = requestIpv4PacketHeader.SrcIP

	ackSegment := constructUDPSegment(67, 68, message)
	ackIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, ackSegment)
	ackFrame := constructFrame(srcMAC, dstMAC, "IPv4", ackIPv4Packet)
	channels[linkID] <- ackFrame

	// Setting leasee's MAC in pool (new)
	pool := snet.Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].String() == addr_to_give {
			debug(4, "dhcp_offer", snet.Router.ID, "Assigning and removing address "+addr_to_give+" from pool")
			snet.Router.DHCPPool.DHCPPoolLeases[addr_to_give] = getMACfromID(dstid) //NI TODO have client pass their MAC in DHCPREQUEST instead of relying on this NI
		}
	}
}

func ipset(hostname string) {
	fmt.Printf(" IP configuration for %s\n", hostname)

	correct := false
	var ipaddr, subnetmask, defaultgateway string
	for !correct {
		fmt.Print("IP Address: ")
		scanner.Scan()
		ipaddr = scanner.Text()

		fmt.Print("\nSubnet mask: ")
		scanner.Scan()
		subnetmask = scanner.Text()

		fmt.Print("\nDefault gateway: ")
		scanner.Scan()
		defaultgateway = scanner.Text()

		fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetmask, defaultgateway)
		fmt.Print("\nIs this correct? [Y/n/exit]")
		scanner.Scan()
		affirmation := scanner.Text()

		if strings.ToUpper(affirmation) == "Y" {
			// error checking
			error := false
			if net.ParseIP(ipaddr).To4() == nil {
				error = true
				fmt.Printf("Error: '%s' is not a valid IP address\n", ipaddr)
			}
			if net.ParseIP(subnetmask).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid subnet mask\n", subnetmask)
			}
			if net.ParseIP(defaultgateway).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid default gateway\n", defaultgateway)
			}

			if !error {
				correct = true
			}
		} else if strings.ToUpper(affirmation) == "EXIT" {
			fmt.Println("Network changes reverted")

			return
		}
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			snet.Hosts[h].IPAddr = net.ParseIP(ipaddr)
			snet.Hosts[h].SubnetMask = subnetmask
			snet.Hosts[h].DefaultGateway = net.ParseIP(defaultgateway)
			fmt.Println("Network configuration updated")
		}
	}

}

// Run an ARP request, but synchronize with client
func arpSynchronized(id string, targetIP string) {
	arp_request(id, targetIP)
	actionsync[id] <- 1
}