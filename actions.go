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

func ping(srcID string, dstIP string, secs int) {
	debug(4, "ping", srcID, "About to ping")

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

	for i := 0; i < secs; i++ {
		// Get MAC addresses
		if snet.Router.ID == srcID {
			srcIP = snet.Router.Gateway.String()
			srcMAC = snet.Router.MACAddr

			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, "router", dstIP)

			//get link to send ping to
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}

		} else { // Assumed to be host source.
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, "host", dstIP)
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
		protocol := "ICMP"
		payload, _ := json.Marshal("101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")

		icmpRequestPacket := ICMPPacket{
			ControlType: 8,
			Data:        json.RawMessage(payload),
		}
		icmpRequestPacketBytes, _ := json.Marshal(icmpRequestPacket)

		ipv4PacketBytes := constructIPv4Packet(srcIP, dstIP, protocol, icmpRequestPacketBytes)

		frameBytes := constructFrame(string(ipv4PacketBytes), srcMAC, dstMAC, "IPv4")

		debug(4, "ping", srcID, "Sending the ping now")
		channels[linkID] <- frameBytes
		debug(2, "ping", srcID, "Ping sent")

		sendCount++

		pong := make(chan bool, 1)

		select {
		case pongFrame := <-internal[srcID]:
			pongIpv4Packet := readIPv4Packet(string(pongFrame.Data))
			pongIcmpPacket := readICMPPacket(string(pongIpv4Packet.Data))

			if pongIcmpPacket.ControlType == 0 {
				recvCount++
				pong <- true
			} else {
				debug(1, "ping", srcID, "Error: Out-of-order channel")
			}
		case <-time.After(time.Second * 4):
			lossCount++
			pong <- false
		}

		if <-pong {
			fmt.Printf("Reply from %s: seq=%d\n", dstIP, i)
			timeoutCounter = 0
		} else {
			fmt.Printf("Request timed out.\n")
			timeoutCounter++
			i--
		}

		if timeoutCounter == 4 { //Skip rest of pings if timeout
			i = secs
		}

		if i < secs-1 { //Only wait a second if
			time.Sleep(time.Second)
		}
	}
	actionsync[srcID] <- 1

	// Ping stats
	fmt.Printf("\nPing statistics for %s:\n", dstIP)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))
}

func pong(srcID string, frame Frame) {
	receivedIpv4Packet := readIPv4Packet(string(frame.Data))
	receivedIcmpPacket := readICMPPacket(string(receivedIpv4Packet.Data))

	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstIP := readIPv4PacketHeader(string(receivedIpv4Packet.Header)).SrcIP
	dstMAC := frame.SrcMAC // Get MAC myself via ARP/MAC table, or use request's source MAC?
	if snet.Router.ID == srcID {
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway.String()

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcID, "router", dstIP)
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

	protocol := "ICMP"
	payload := receivedIcmpPacket.Data

	icmpReplyPacket := ICMPPacket{
		ControlType: 0,
		Data:        payload,
	}
	icmpReplyPacketBytes, _ := json.Marshal(icmpReplyPacket)

	ipv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, icmpReplyPacketBytes)
	ipv4PacketBytes, _ := json.Marshal(ipv4Packet)

	frameBytes := constructFrame(string(ipv4PacketBytes), srcMAC, dstMAC, "IPv4")

	debug(4, "pong", srcID, "Awaiting pong send")
	channels[linkID] <- frameBytes
	debug(2, "pong", srcID, "Pong sent")
}

func arp_request(srcID string, device_type string, targetIP string) string {
	debug(4, "arp_request", srcID, "About to ARP request")
	//Construct frame
	linkID := "FFFFFFFF"
	srcIP := ""
	srcMAC := ""
	dstMAC := "00:00:00:00:00:00"

	if device_type == "router" {
		srcIP = snet.Router.Gateway.String()
		srcMAC = snet.Router.MACAddr
	} else {
		index := getHostIndexFromID(srcID)
		srcIP = snet.Hosts[index].IPAddr.String()
		srcMAC = snet.Hosts[index].MACAddr
	}

	arpRequest := ArpMessage{
		Opcode:    1,
		SenderMAC: srcMAC,
		SenderIP:  srcIP,
		TargetMAC: dstMAC,
		TargetIP:  targetIP,
	}

	arpRequestBytes, _ := json.Marshal(arpRequest)
	frameBytes := constructFrame(string(arpRequestBytes), srcMAC, dstMAC, "ARP")

	channels[linkID] <- frameBytes
	debug(2, "arp_request", srcID, "ARPREQUEST sent")

	arpReplyFrame := <-internal[srcID]
	arpReply := readArpMessage(string(arpReplyFrame.Data))

	return arpReply.SenderMAC
}

func arp_reply(i int, device_type string, frame Frame) {
	//Inspect Arp message
	arpRequest := readArpMessage(string(frame.Data))

	requested_addr := arpRequest.TargetIP
	linkID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := arpRequest.SenderMAC
	dstIP := arpRequest.SenderIP
	srcID := ""

	if device_type == "router" {
		if requested_addr == snet.Router.Gateway.String() {
			srcIP = snet.Router.Gateway.String()
			srcMAC = snet.Router.MACAddr
			srcID = snet.Router.ID

			// Determine linkID
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}
		} else { // Not me
			return
		}
	} else {
		if requested_addr == snet.Hosts[i].IPAddr.String() {
			srcIP = snet.Hosts[i].IPAddr.String()
			srcMAC = snet.Hosts[i].MACAddr
			srcID = snet.Hosts[i].ID
			linkID = snet.Hosts[i].UplinkID
		} else { // Not me
			return
		}
	}

	arpReply := ArpMessage{
		Opcode:    2,
		SenderMAC: srcMAC,
		SenderIP:  srcIP,
		TargetMAC: dstMAC,
		TargetIP:  dstIP,
	}

	arpReplyBytes, _ := json.Marshal(arpReply)
	frameBytes := constructFrame(string(arpReplyBytes), srcMAC, dstMAC, "ARP")

	channels[linkID] <- frameBytes
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
	segmentData := constructUDPSegment("DHCPDISCOVER", 68, 67)
	packetData := constructIPv4Packet(srcIP, dstIP, protocol, segmentData)
	frameData := constructFrame(string(packetData), srcMAC, dstMAC, "IPv4")

	//need to give it to uplink
	channels[linkID] <- frameData
	debug(2, "dhcp_discover", host.ID, "DHCPDISCOVER sent")
	offerFrame := <-internal[srcID]

	offerIpv4Packet := readIPv4Packet(string(offerFrame.Data))
	offerIpv4PacketHeader := readIPv4PacketHeader(string(offerIpv4Packet.Header))
	offerUDPSegment := readUDPSegment(string(offerIpv4Packet.Data))

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
			requestUDPSegment := constructUDPSegment(message, 68, 67)
			requestIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, requestUDPSegment)
			requestFrame := constructFrame(string(requestIPv4Packet), srcMAC, dstMAC, "IPv4")
			channels[linkID] <- requestFrame
			debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent - "+word2)
			//wait for acknowledgement

			ackFrame := <-internal[srcID]
			ackIpv4Packet := readIPv4Packet(string(ackFrame.Data))
			ackUDPSegment := readUDPSegment(string(ackIpv4Packet.Data))

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
	segmentBytes := constructUDPSegment(message, 67, 68)
	packetBytes := constructIPv4Packet(srcIP, dstIP, protocol, segmentBytes)
	frameBytes := constructFrame(string(packetBytes), srcMAC, dstMAC, "IPv4")
	channels[linkID] <- frameBytes
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - "+addr_to_give)

	// Acknowledge
	requestFrame := <-internal[snet.Router.ID]
	requestIpv4Packet := readIPv4Packet(string(requestFrame.Data))
	requestIpv4PacketHeader := readIPv4PacketHeader(string(requestIpv4Packet.Header))
	requestUDPSegment := readUDPSegment(string(requestIpv4Packet.Data))

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

	ackSegment := constructUDPSegment(message, 67, 68)
	ackIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, ackSegment)
	ackFrame := constructFrame(string(ackIPv4Packet), srcMAC, dstMAC, "IPv4")
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
