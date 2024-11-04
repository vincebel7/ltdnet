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
	"strconv"
	"strings"
	"time"

	"github.com/vincebel7/ltdnet/iphelper"
)

func ping(srcID string, dst string, count int) {
	debug(4, "ping", srcID, "About to ping")

	identifier := idgen_int(5)
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srcHostname := ""
	srcHost := Host{}
	dnsTable := make(map[string]DNSRecord)

	sendCount := 0
	recvCount := 0
	lossCount := 0

	// Get DNS table
	if snet.Router.ID == srcID {
		dnsTable = snet.Router.DNSTable
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				dnsTable = snet.Hosts[h].DNSTable
			}
		}
	}

	// Hostname lookup, if needed
	var dstIP string
	if ip := net.ParseIP(dst); ip != nil {
		dstIP = dst
	} else {
		dstIP = resolveHostname(srcID, dst, dnsTable).RData

		if dstIP == "" {
			debug(1, "ping", srcID, "[Error] Hostname could not be resolved")
			actionsync[srcID] <- 1
			return
		}
	}

	iface := Interface{}
	if snet.Router.ID == srcID {
		iface = snet.Router.routeToInterface(dstIP)
		srcHostname = snet.Router.Hostname
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				iface = snet.Hosts[h].routeToInterface(dstIP)
				srcHost = snet.Hosts[h]
				srcHostname = snet.Router.Hostname
			}
		}
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	fmt.Printf("\nPinging %s from %s\n", dstIP, srcHostname)

	for i := 0; i < count; i++ {
		// Get destination MAC address
		if snet.Router.ID == srcID {
			dstMAC = routerDetermineDstMAC(snet.Router, dstIP, iface.Name, true)
		} else {
			dstMAC = hostDetermineDstMAC(srcHost, dstIP, iface.Name, true)
		}

		if dstMAC == "TIMEOUT" {
			lossCount++
			sendCount++
			continue
		}

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

		debug(4, "ping", srcID, "Awaiting ping send")
		sendFrame(frameBytes, iface, srcID)
		debug(3, "ping", srcID, "Ping request sent")

		sendCount++

		sockets := socketMaps[srcID]
		socketID := "icmp_" + strconv.Itoa(identifier)
		sockets[socketID] = make(chan Frame)
		socketMaps[srcID] = sockets // Write updated map back to the collection

		debug(4, "ping", srcID, "Awaiting ping reply on "+srcID)
		select {
		case pongFrame := <-sockets[socketID]:
			pongIpv4Packet := readIPv4Packet(pongFrame.Data)
			pongIcmpPacket := readICMPEchoPacket(pongIpv4Packet.Data)

			if pongIcmpPacket.ControlType == 0 {
				recvCount++
				fmt.Printf("Reply from %s: seq=%d\n", dstIP, i)

				if srcIP != dstIP {
					achievementTester(UNITED_PINGDOM)
				}

				if dstIP == "127.0.0.1" {
					achievementTester(SNIFF_FRAMES)
				}
			} else {
				debug(1, "ping", srcID, "Error: Out-of-order channel")
			}
		case <-time.After(time.Second * 4):
			lossCount++
			fmt.Printf("Request timed out.\n")
		}

		if i < count-1 { //Only wait a second if not the last ping.
			time.Sleep(time.Second)
		}
	}

	// Ping stats
	fmt.Printf("\nPing statistics for %s:\n", dstIP)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))
	fmt.Printf("\tSource address: %s\n\n", srcIP)

	actionsync[srcID] <- lossCount
}

func pong(srcID string, frame Frame) {
	receivedIpv4Packet := readIPv4Packet(frame.Data)
	receivedIcmpPacket := readICMPEchoPacket(receivedIpv4Packet.Data)

	srcIP := ""
	srcMAC := ""
	dstIP := readIPv4PacketHeader(receivedIpv4Packet.Header).SrcIP
	dstMAC := ""

	iface := Interface{}
	if snet.Router.ID == srcID {
		iface = snet.Router.routeToInterface(dstIP)
		dstMAC = routerDetermineDstMAC(snet.Router, dstIP, iface.Name, true)
	} else {
		index := getHostIndexFromID(srcID)
		iface = snet.Hosts[index].routeToInterface(dstIP)
		dstMAC = hostDetermineDstMAC(snet.Hosts[index], dstIP, iface.Name, true)
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

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
	sendFrame(frameBytes, iface, srcID)
	debug(3, "pong", srcID, "Ping reply sent")
}

func arp_request(srcID string, targetIP string) string {
	debug(4, "arp_request", srcID, "About to ARP request")

	// Construct frame
	srcMAC := ""
	srcIP := ""
	dstMAC := "ff:ff:ff:ff:ff:ff"

	iface := Interface{}
	if srcID == snet.Router.ID {
		iface = snet.Router.Interfaces["eth0"]
	} else {
		index := getHostIndexFromID(srcID)
		iface = snet.Hosts[index].Interfaces["eth0"]
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	// First, check if it is trying to ARP itself.
	if targetIP == srcIP {
		debug(4, "arp_request", srcID, "Destination IP is source IP! Canceling ARP request.")
		return srcMAC
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
	sendFrame(arpRequestFrameBytes, iface, srcID)
	debug(3, "arp_request", srcID, "ARPREQUEST sent")

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
	srcID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := arpRequestMessage.SenderMAC // This usage of SenderMAC is according to ARP protocol.
	dstIP := arpRequestMessage.SenderIP

	// Network listener decided to reply to this request - no checking needed.
	iface := Interface{}
	if id == snet.Router.ID {
		iface = snet.Router.Interfaces["eth0"]
		srcID = snet.Router.ID
	} else {
		index := getHostIndexFromID(id)
		iface = snet.Hosts[index].Interfaces["eth0"]
		srcID = snet.Hosts[index].ID
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

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
	sendFrame(arpReplyFrameBytes, iface, srcID)
	debug(3, "arp_reply", srcID, "ARPREPLY sent")
}

func dhcp_discover(host Host) {
	debug(4, "dhcp_discover", host.ID, "Starting DHCPDISCOVER")
	//get info
	iface := host.Interfaces["eth0"]
	srcIP := host.GetIP(iface.Name)
	srcMAC := iface.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := "ff:ff:ff:ff:ff:ff"

	// Construct DHCPDISCOVER
	options := map[byte][]byte{
		53: {1},                   // Option 53: DHCPDISCOVER
		12: []byte(host.Hostname), // Option 12: Hostname
	}
	dhcpDiscoverMessage := DHCPMessage{
		Op:      1,                      // Message type: 1 = Request, 2 = Reply
		HType:   1,                      // Hardware address type (e.g., 1 for Ethernet)
		HLen:    6,                      // Length of hardware address
		Hops:    0,                      // Hops
		XID:     uint32(idgen_int(5)),   // Transaction ID
		Flags:   0,                      // Flags (e.g., broadcast)
		CIAddr:  net.ParseIP("0.0.0.0"), // Client IP address
		YIAddr:  net.ParseIP("0.0.0.0"), // 'Your' IP address (server's offer)
		SIAddr:  net.ParseIP("0.0.0.0"), // Server IP address
		GIAddr:  net.ParseIP("0.0.0.0"), // Gateway IP address
		CHAddr:  srcMAC,                 // Client MAC address
		Options: options,                // DHCP options
	}

	// Encapsulate DHCPDISCOVER
	protocol := "UDP"
	dhcpDiscoverMessageBytes, _ := json.Marshal(dhcpDiscoverMessage)
	segmentData := constructUDPSegment(68, 67, dhcpDiscoverMessageBytes)
	packetData := constructIPv4Packet(srcIP, dstIP, protocol, segmentData)
	frameData := constructFrame(srcMAC, dstMAC, "IPv4", packetData)

	// Send DHCPDISCOVER, await DHCPOFFER
	//need to give it to uplink
	sendFrame(frameData, iface, srcID)
	debug(3, "dhcp_discover", host.ID, "DHCPDISCOVER sent")

	sockets := socketMaps[srcID]
	socketID := "udp_" + strconv.Itoa(68)
	sockets[socketID] = make(chan Frame)
	socketMaps[srcID] = sockets // Write updated map back to the collection
	dhcpOfferFrame := <-sockets[socketID]

	// De-encapsulate DHCPOFFER
	dhcpOfferIPv4Packet := readIPv4Packet(dhcpOfferFrame.Data)
	dhcpOfferIPv4PacketHeader := readIPv4PacketHeader(dhcpOfferIPv4Packet.Header)
	dhcpOfferUDPSegment := readUDPSegment(dhcpOfferIPv4Packet.Data)
	dhcpOfferMessage := ReadDHCPMessage(dhcpOfferUDPSegment.Data)

	if int(dhcpOfferMessage.Options[53][0]) == 6 { // 6 is DHCPNAK
		debug(1, "dhcp_discover", srcID, "Failed to obtain IP address: No free addresses available")
	} else {
		dstIP = dhcpOfferIPv4PacketHeader.SrcIP

		// Construct DHCPREQUEST
		options = map[byte][]byte{
			53: {3},                   // Option 53: DHCPREQUEST
			12: []byte(host.Hostname), // Option 12: Hostname
		}
		dhcpRequestMessage := DHCPMessage{
			Op:      1,                       // Message type: 1 = Request, 2 = Reply
			HType:   1,                       // Hardware address type (e.g., 1 for Ethernet)
			HLen:    6,                       // Length of hardware address
			Hops:    0,                       // Hops
			XID:     dhcpOfferMessage.XID,    // Transaction ID
			Flags:   0,                       // Flags (e.g., broadcast)
			CIAddr:  net.ParseIP("0.0.0.0"),  // Client IP address
			YIAddr:  dhcpOfferMessage.YIAddr, // 'Your' IP address (server's offer)
			SIAddr:  net.ParseIP("0.0.0.0"),  // Server IP address
			GIAddr:  net.ParseIP("0.0.0.0"),  // Gateway IP address
			CHAddr:  srcMAC,                  // Client MAC address
			Options: options,                 // DHCP options
		}

		// Encapsulate DHCPREQUEST
		protocol := "UDP"
		dhcpRequestMessageBytes, _ := json.Marshal(dhcpRequestMessage)
		dhcpRequestUDPSegment := constructUDPSegment(68, 67, dhcpRequestMessageBytes)
		dhcpRequestIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, dhcpRequestUDPSegment)
		dhcpRequestFrame := constructFrame(srcMAC, dstMAC, "IPv4", dhcpRequestIPv4Packet)

		// Send DHCPREQUEST, await DHCPACK
		sendFrame(dhcpRequestFrame, iface, srcID)
		debug(3, "dhcp_discover", srcID, "DHCPREQUEST sent")
		dhcpAckFrame := <-sockets[socketID]

		// De-encapsulate DHCPACK
		dhcpAckIpv4Packet := readIPv4Packet(dhcpAckFrame.Data)
		dhcpAckUDPSegment := readUDPSegment(dhcpAckIpv4Packet.Data)
		dhcpAckMessage := ReadDHCPMessage(dhcpAckUDPSegment.Data)

		if int(dhcpAckMessage.Options[53][0]) == 5 {
			debug(3, "dhcp_discover", srcID, "DHCPACK assigned a lease - "+dhcpAckMessage.YIAddr.String())

			assignedAddress := dhcpAckMessage.YIAddr
			defaultGateway := net.IP(dhcpAckMessage.Options[3]).To4()
			subnetMask := net.IP(dhcpAckMessage.Options[1]).To4()

			dynamic_assign(srcID, assignedAddress, defaultGateway, subnetMask.String())

		} else { // 5 is DHCPACK
			debug(1, "dhcp_discover", srcID, "Failed to obtain IP address")
		}
	}
	actionsync[srcID] <- 1
}

func dhcp_offer(dhcpDiscoverFrame Frame) {
	// De-encapsulate DHCPDISCOVER
	dhcpDiscoverIPv4Packet := readIPv4Packet(dhcpDiscoverFrame.Data)
	//dhcpDiscoverIpv4PacketHeader := readIPv4PacketHeader(dhcpDiscoverIPv4Packet.Header)
	dhcpDiscoverUDPSegment := readUDPSegment(dhcpDiscoverIPv4Packet.Data)
	dhcpDiscoverMessage := ReadDHCPMessage(dhcpDiscoverUDPSegment.Data)

	iface := snet.Router.Interfaces["eth0"]
	srcIP := snet.Router.GetIP(iface.Name)
	dstIP := "255.255.255.255"
	srcMAC := iface.MACAddr
	dstMAC := dhcpDiscoverFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.

	// Find open address
	addr_to_give := snet.Router.NextFreePoolAddress()
	gateway := snet.Router.GetIP(iface.Name)
	netSize, _ := strconv.Atoi(snet.Netsize)
	subnetmask := prefixLengthToSubnetMask(netSize)

	messageType := 6
	if addr_to_give != nil {
		messageType = 2
	}

	// Construct DHCPOFFER
	options := map[byte][]byte{
		53: {byte(messageType)},           // Option 53: DHCPOFFER
		1:  net.ParseIP(subnetmask).To4(), // Subnet mask
		3:  net.ParseIP(gateway).To4(),    // Gateway
		51: {0, 0, 10, 0},                 // Lease time
		54: net.ParseIP(gateway).To4(),    // DHCP server
	}
	dhcpOfferMessage := DHCPMessage{
		Op:      2,                        // Message type: 1 = Request, 2 = Reply
		HType:   1,                        // Hardware address type (e.g., 1 for Ethernet)
		HLen:    6,                        // Length of hardware address
		Hops:    0,                        // Hops
		XID:     dhcpDiscoverMessage.XID,  // Transaction ID
		Flags:   0,                        // Flags (e.g., broadcast)
		CIAddr:  net.ParseIP("0.0.0.0"),   // Client IP address
		YIAddr:  addr_to_give,             // 'Your' IP address (server's offer)
		SIAddr:  net.ParseIP("0.0.0.0"),   // Server IP address
		GIAddr:  net.ParseIP("0.0.0.0"),   // Gateway IP address
		CHAddr:  dhcpDiscoverFrame.SrcMAC, // Client MAC address
		Options: options,                  // DHCP options
	}

	// Encapsulate DHCPOFFER
	protocol := "UDP"
	dhcpOfferMessageBytes, _ := json.Marshal(dhcpOfferMessage)
	dhcpOfferSegment := constructUDPSegment(67, 68, dhcpOfferMessageBytes)
	dhcpOfferPacket := constructIPv4Packet(srcIP, dstIP, protocol, dhcpOfferSegment)
	dhcpOfferFrame := constructFrame(srcMAC, dstMAC, "IPv4", dhcpOfferPacket)

	// Send DHCPOFFER, await DHCPREQUEST
	sendFrame(dhcpOfferFrame, iface, snet.Router.ID)
	debug(3, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - "+addr_to_give.String())
}

func dhcp_ack(dhcpRequestFrame Frame) {
	// De-encapsulate DHCPREQUEST
	dhcpRequestIPv4Packet := readIPv4Packet(dhcpRequestFrame.Data)
	dhcpRequestIPv4PacketHeader := readIPv4PacketHeader(dhcpRequestIPv4Packet.Header)
	dhcpRequestUDPSegment := readUDPSegment(dhcpRequestIPv4Packet.Data)
	dhcpRequestMessage := ReadDHCPMessage(dhcpRequestUDPSegment.Data)

	iface := snet.Router.Interfaces["eth0"]
	srcIP := snet.Router.GetIP(iface.Name)
	dstIP := dhcpRequestIPv4PacketHeader.SrcIP
	srcMAC := iface.MACAddr
	dstMAC := dhcpRequestFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.

	messageType := 6
	if dhcpRequestUDPSegment.Data != nil {
		if int(dhcpRequestMessage.Options[53][0]) == 3 { // 3 = DHCPREQUEST
			if snet.Router.IsAvailableAddress(dhcpRequestMessage.YIAddr) {
				messageType = 5
			} else {
				debug(1, "dhcp_offer", snet.Router.ID, "Error: DHCP address requested is not available")
			}
		} else {
			debug(1, "dhcp_offer", snet.Router.ID, "Error: Empty DHCP request")
		}
	}

	gateway := snet.Router.GetIP(iface.Name)
	netSize, _ := strconv.Atoi(snet.Netsize)
	subnetmask := prefixLengthToSubnetMask(netSize)

	// Construct DHCPACK
	options := map[byte][]byte{
		53: {byte(messageType)},           // Option 53: DHCPACK
		1:  net.ParseIP(subnetmask).To4(), // Subnet mask
		3:  net.ParseIP(gateway).To4(),    // Gateway
		51: {0, 0, 10, 0},                 // Lease time
		54: net.ParseIP(gateway).To4(),    // DHCP server
	}
	dhcpAckMessage := DHCPMessage{
		Op:      2,                         // Message type: 1 = Request, 2 = Reply
		HType:   1,                         // Hardware address type (e.g., 1 for Ethernet)
		HLen:    6,                         // Length of hardware address
		Hops:    0,                         // Hops
		XID:     dhcpRequestMessage.XID,    // Transaction ID
		Flags:   0,                         // Flags (e.g., broadcast)
		CIAddr:  net.ParseIP("0.0.0.0"),    // Client IP address
		YIAddr:  dhcpRequestMessage.YIAddr, // 'Your' IP address (server's offer)
		SIAddr:  net.ParseIP("0.0.0.0"),    // Server IP address
		GIAddr:  net.ParseIP("0.0.0.0"),    // Gateway IP address
		CHAddr:  dhcpRequestFrame.SrcMAC,   // Client MAC address
		Options: options,                   // DHCP options
	}

	// Encapsulate DHCPACK
	protocol := "UDP"
	dhcpAckMessageBytes, _ := json.Marshal(dhcpAckMessage)
	dhcpAckSegment := constructUDPSegment(67, 68, dhcpAckMessageBytes)
	dhcpAckIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, dhcpAckSegment)
	dhcpAckFrame := constructFrame(srcMAC, dstMAC, "IPv4", dhcpAckIPv4Packet)

	// Send DHCPACK
	sendFrame(dhcpAckFrame, iface, snet.Router.ID)
	debug(3, "dhcp_offer", snet.Router.ID, "DHCPACK sent - "+dhcpAckMessage.YIAddr.String())

	// Setting leasee's MAC in pool (new)
	pool := snet.Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].Equal(dhcpAckMessage.YIAddr) {
			debug(4, "dhcp_offer", snet.Router.ID, "Assigning and removing address "+dhcpAckMessage.YIAddr.String()+" from pool")
			snet.Router.DHCPPool.DHCPPoolLeases[dhcpAckMessage.YIAddr.String()] = dhcpAckMessage.CHAddr
		}
	}
}

func dns_query(srcID string, hostname string, reqType uint16) DNSMessage {
	srcIP := ""
	dstMAC := ""
	dstIP := ""
	iface := Interface{}

	if snet.Router.ID == srcID {
		iface = snet.Router.Interfaces["lo"]
		srcIP = snet.Router.GetIP(iface.Name)
		//dstIP := snet.Router.Interfaces["eth0"].IPConfig.DNSServer
		dstIP = snet.Router.Interfaces["lo"].IPConfig.DNSServer.String() // temporary
		dstMAC = routerDetermineDstMAC(snet.Router, dstIP, iface.Name, true)
	} else {
		hostIndex := getHostIndexFromID(srcID)
		host := snet.Hosts[hostIndex]
		iface = host.Interfaces["eth0"]
		srcIP = host.GetIP(iface.Name)
		//dstIP := host.Interfaces["eth0"].IPConfig.DNSServer
		dstIP = host.GetGateway(iface.Name) // temporary
		dstMAC = hostDetermineDstMAC(host, dstIP, iface.Name, true)
	}

	srcMAC := iface.MACAddr

	var dnsQueryMessage = DNSMessage{}
	switch reqType {
	case 'A':
		dnsQuestionMessage := DNSQuestion{
			QName:  hostname,
			QType:  reqType,
			QClass: 1,
		}
		dnsQuestions := make([]DNSQuestion, 1)
		dnsQuestions[0] = dnsQuestionMessage

		dnsQueryMessage = DNSMessage{
			QR:        false, // false = query
			Opcode:    0,
			QDCount:   1,
			Questions: dnsQuestions,
		}

	default:
		debug(1, "dns_query", srcID, "[Error] DNS query type not implemented yet")
		return DNSMessage{}
	}

	protocol := "UDP"
	srcPort := ephemeralPortGen()
	dnsQueryMessageBytes, _ := json.Marshal(dnsQueryMessage)
	dnsQuerySegment := constructUDPSegment(srcPort, 53, dnsQueryMessageBytes)
	dnsQueryIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, dnsQuerySegment)
	dnsQueryFrame := constructFrame(srcMAC, dstMAC, "IPv4", dnsQueryIPv4Packet)

	sendFrame(dnsQueryFrame, iface, srcID)
	debug(3, "dns_query", srcID, "DNS query sent - "+hostname)

	sockets := socketMaps[srcID]
	socketID := "udp_" + strconv.Itoa(srcPort)
	sockets[socketID] = make(chan Frame)
	socketMaps[srcID] = sockets // Write updated map back to the collection

	select {
	case dnsResponseFrame := <-sockets[socketID]:
		dnsResponsePacket := readIPv4Packet(dnsResponseFrame.Data)
		dnsResponseSegment := readUDPSegment(dnsResponsePacket.Data)
		dnsResponseMessage := ReadDNSMessage(dnsResponseSegment.Data)

		switch dnsResponseMessage.Rcode {
		case 2:
			fmt.Printf("server can't find %s: SERVFAIL\n", hostname)

		case 3:
			fmt.Printf("server can't find %s: NXDOMAIN\n", hostname)

		case 0:
			achievementTester(MY_NAME)
			return dnsResponseMessage
		}

	case <-time.After(time.Second * 4):
		fmt.Printf("DNS request timed out.\n")
	}
	return DNSMessage{}
}

func dns_response(dnsQueryFrame Frame) {
	// De-encapsulate DNS Query
	dnsQueryIPv4Packet := readIPv4Packet(dnsQueryFrame.Data)
	dnsQueryIPv4PacketHeader := readIPv4PacketHeader(dnsQueryIPv4Packet.Header)
	dnsQueryUDPSegment := readUDPSegment(dnsQueryIPv4Packet.Data)
	dnsQueryMessage := ReadDNSMessage(dnsQueryUDPSegment.Data)

	dstIP := dnsQueryIPv4PacketHeader.SrcIP
	iface := snet.Router.routeToInterface(dstIP)
	srcIP := snet.Router.GetIP(iface.Name)
	dstPort := dnsQueryUDPSegment.SrcPort
	srcMAC := iface.MACAddr
	dstMAC := dnsQueryFrame.SrcMAC

	var dnsResponseMessage = DNSMessage{}
	switch dnsQueryMessage.Questions[0].QType {
	case 'A':
		dnsResponseMessage = DNSMessage{
			QR:     true, // false = query
			Opcode: 0,
			Rcode:  2,
		}

		dnsAnswerRecord := snet.Router.DNSServer.aRecordLookup(dnsQueryMessage.Questions[0].QName)

		if dnsAnswerRecord.Name != "" {
			dnsResponseMessage.Rcode = 0
			dnsResponseMessage.ANCount = 1

			dnsAnswers := make([]DNSRecord, 1)
			dnsAnswers[0] = dnsAnswerRecord

			dnsResponseMessage.Answers = dnsAnswers
		} else {
			dnsResponseMessage.Rcode = 3
			dnsResponseMessage.ANCount = 0
		}

	default:
		debug(1, "dns_query", snet.Router.ID, "[Warning] DNS query type not implemented yet - returning SERVFAIL message")
		return
	}

	protocol := "UDP"
	dnsResponseMessageBytes, _ := json.Marshal(dnsResponseMessage)
	dnsResponseSegment := constructUDPSegment(53, dstPort, dnsResponseMessageBytes)
	dnsResponseIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, dnsResponseSegment)
	dnsResponseFrame := constructFrame(srcMAC, dstMAC, "IPv4", dnsResponseIPv4Packet)

	sendFrame(dnsResponseFrame, iface, snet.Router.ID)
	debug(3, "dns_query", snet.Router.ID, "DNS response sent")

}

func ipset(hostname string, ipaddr string, subnetMask string) {
	defaultGateway := snet.Router.GetIP("eth0")

	fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetMask, defaultGateway)
	fmt.Print("\nIs this correct? [Y/n]: ")
	scanner.Scan()
	affirmation := scanner.Text()

	if strings.ToUpper(affirmation) == "Y" {
		// error checking
		if net.ParseIP(ipaddr).To4() == nil {
			fmt.Printf("Error: '%s' is not a valid IP address\n", ipaddr)
			return
		}

	} else {
		fmt.Println("Network changes reverted")
		return
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			iface := snet.Hosts[h].Interfaces["eth0"]

			iface.IPConfig.IPAddress = net.ParseIP(ipaddr)
			iface.IPConfig.SubnetMask = subnetMask
			iface.IPConfig.DefaultGateway = net.ParseIP(defaultGateway)

			snet.Hosts[h].Interfaces["eth0"] = iface

			fmt.Println("Network configuration updated")
		}
	}
}

// Run an ARP request, but synchronize with client
func arpSynchronized(id string, targetIP string) {
	dstMAC := ""

	if snet.Router.ID == id {
		dstMAC = routerDetermineDstMAC(snet.Router, targetIP, "eth0", false)
	} else {
		dstMAC = hostDetermineDstMAC(snet.Hosts[getHostIndexFromID(id)], targetIP, "eth0", false)
	}

	if dstMAC != "" {
		achievementTester(ARP_HOT)
	}

	actionsync[id] <- 1
}

// A host determines the destination MAC to send to... Either by ARP, sending to GW, or reading ARP table
func hostDetermineDstMAC(srcHost Host, dstIP string, iface string, useTable bool) string {
	srcID := srcHost.ID
	dstMAC := ""

	if dstIP == "127.0.0.1" && iface == "lo" {
		return srcHost.Interfaces["lo"].MACAddr
	}

	// Same subnet - ARP table, or ARP request.
	if iphelper.IPInSameSubnet(srcHost.GetIP(iface), dstIP, srcHost.GetMask(iface)) {
		debug(4, "hostDetermineDstMAC", srcID, "Sending to same subnet, about to ARP table lookup or ARP")

		// Check ARP table
		if useTable && snet.Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP].MACAddr != "" {
			dstMAC = snet.Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP].MACAddr
		} else {
			// ARP request
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("ARP request timed out.\n")
			} else {
				arpEntry := ARPEntry{
					MACAddr:   dstMAC,
					Interface: snet.Hosts[getHostIndexFromID(srcID)].Interfaces[iface].RemoteL1ID,
				}
				snet.Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP] = arpEntry // Add to ARP table
			}
		}

	} else { // Different subnet - GW.
		debug(4, "hostDetermineDstMAC", srcID, "Sending to different subnet, sending to GW")
		gateway := srcHost.GetGateway(iface)

		// Check ARP table
		if snet.Hosts[getHostIndexFromID(srcID)].ARPTable[gateway].MACAddr != "" {
			dstMAC = snet.Hosts[getHostIndexFromID(srcID)].ARPTable[gateway].MACAddr
		} else {
			// ARP request
			dstMAC = arp_request(srcID, gateway)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("ARP request timed out.\n")
			} else {
				arpEntry := ARPEntry{
					MACAddr:   dstMAC,
					Interface: snet.Hosts[getHostIndexFromID(srcID)].Interfaces[iface].RemoteL1ID,
				}
				snet.Hosts[getHostIndexFromID(srcID)].ARPTable[gateway] = arpEntry // Add to ARP table
			}
		}
	}

	return dstMAC
}

// A router determines the destination MAC to send to... Either by ARP, or reading ARP table
func routerDetermineDstMAC(router Router, dstIP string, iface string, useTable bool) string {
	dstMAC := ""

	if dstIP == "127.0.0.1" && iface == "lo" {
		return router.Interfaces["lo"].MACAddr
	}

	// Same subnet - ARP table, or ARP request.
	netsizeInt, _ := strconv.Atoi(snet.Netsize)
	subnetMask := prefixLengthToSubnetMask(netsizeInt)
	if iphelper.IPInSameSubnet(router.GetIP(iface), dstIP, subnetMask) {
		debug(4, "routerDetermineDstMAC", router.ID, "Sending to same subnet, about to MAC table lookup or ARP")

		// Check ARP table
		if useTable && snet.Router.ARPTable[dstIP].MACAddr != "" {
			dstMAC = snet.Router.ARPTable[dstIP].MACAddr
		} else {
			// ARP request
			dstMAC = arp_request(router.ID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("ARP request timed out.\n")
			} else {
				arpEntry := ARPEntry{
					MACAddr:   dstMAC,
					Interface: router.Interfaces[iface].RemoteL1ID,
				}
				snet.Router.ARPTable[dstIP] = arpEntry // Add to ARP table
			}
		}

	} else {
		fmt.Printf("Error: Routing not implemented yet.")
	}

	return dstMAC
}

func resolveHostname(srcID string, hostname string, dnsTable map[string]DNSRecord) DNSRecord {
	// Check local table
	if entry, found := dnsTable[hostname]; found {
		if entry.TTL == 65535 {
			return entry
		} else {
			delete(dnsTable, hostname) // Warning: This does not get written back.
		}
	}

	// If not found, initiate DNS request
	resultMessage := dns_query(srcID, hostname, 'A')

	// TODO: Add response to local cache
	if resultMessage.Rcode == 0 {
		if srcID == snet.Router.ID {
			snet.Router.DNSTable[hostname] = resultMessage.Answers[0]
		} else {
			snet.Hosts[getHostIndexFromID(srcID)].DNSTable[hostname] = resultMessage.Answers[0]
		}
	}

	return resultMessage.Answers[0]
}
