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
	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

func ping(srcID string, dst string, count int) {
	debug(4, "ping", srcID, "About to ping")

	identifier := idgen_int(5)
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srcHostname := ""
	srcHost := model.Host{}
	dnsTable := make(map[string]model.DNSRecord)

	sendCount := 0
	recvCount := 0
	lossCount := 0

	// Get DNS table
	if Net().Router.ID == srcID {
		dnsTable = Net().Router.DNSTable
	} else {
		for h := range Net().Hosts {
			if Net().Hosts[h].ID == srcID {
				dnsTable = Net().Hosts[h].DNSTable
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
			engine.Instance().ActionSync[srcID] <- 1
			return
		}
	}

	iface := model.Interface{}
	if Net().Router.ID == srcID {
		iface, _ = engine.Instance().RouteToRouterInterface(srcID, dstIP)
		srcHostname = Net().Router.Hostname
	} else {
		for h := range Net().Hosts {
			if Net().Hosts[h].ID == srcID {
				nonDefaultRoute := false
				iface, nonDefaultRoute = engine.Instance().RouteToHostInterface(Net().Hosts[h], dstIP)
				if !nonDefaultRoute {
					debug(4, "routeToInterface", Net().Hosts[h].Hostname, "Route not found. Sending to default gateway")
				}
				srcHost = Net().Hosts[h]
				srcHostname = Net().Router.Hostname
			}
		}
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	fmt.Printf("\nPinging %s from %s\n", dstIP, srcHostname)

	for i := 0; i < count; i++ {
		// Get destination MAC address
		if Net().Router.ID == srcID {
			dstMAC = routerDetermineDstMAC(Net().Router, dstIP, iface.Name, true)
		} else {
			dstMAC = hostDetermineDstMAC(srcHost, dstIP, iface.Name, true)
		}

		if dstMAC == "TIMEOUT" {
			lossCount++
			sendCount++
			continue
		}

		payload, _ := json.Marshal("101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031323334353637")

		icmpRequestPacket := model.ICMPEchoPacket{
			ControlType: 8,
			ControlCode: 0,
			Checksum:    "checksum",
			Identifier:  identifier,
			SeqNumber:   i,
			Data:        json.RawMessage(payload),
		}
		icmpRequestPacketBytes, _ := json.Marshal(icmpRequestPacket)
		ipv4PacketBytes := model.NewIPv4Packet(srcIP, dstIP, "ICMP", icmpRequestPacketBytes)
		frameBytes := model.NewFrame(srcMAC, dstMAC, "IPv4", ipv4PacketBytes)

		debug(4, "ping", srcID, "Awaiting ping send")
		sendFrame(frameBytes, iface, srcID)
		debug(3, "ping", srcID, "Ping request sent")

		sendCount++

		sockets := engine.Instance().Sockets[srcID]
		socketID := "icmp_" + strconv.Itoa(identifier)
		sockets[socketID] = make(chan model.Frame)

		debug(4, "ping", srcID, "Awaiting ping reply on "+srcID)
		select {
		case pongFrame := <-sockets[socketID]:
			pongIpv4Packet, _ := model.ParseIPv4Packet(pongFrame.Data)
			pongIcmpPacket, _ := model.ParseICMPEchoPacket(pongIpv4Packet.Data)

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

	engine.Instance().ActionSync[srcID] <- lossCount
}

func pong(srcID string, frame model.Frame) {
	receivedIpv4Packet, _ := model.ParseIPv4Packet(frame.Data)
	receivedIcmpPacket, _ := model.ParseICMPEchoPacket(receivedIpv4Packet.Data)

	srcIP := ""
	srcMAC := ""
	header, _ := model.ParseIPv4PacketHeader(receivedIpv4Packet.Header)
	dstIP := header.SrcIP
	dstMAC := ""

	iface := model.Interface{}
	if Net().Router.ID == srcID {
		iface, _ = engine.Instance().RouteToRouterInterface(srcID, dstIP)
		dstMAC = routerDetermineDstMAC(Net().Router, dstIP, iface.Name, true)
	} else {
		index := getHostIndexFromID(srcID)
		iface, _ = engine.Instance().RouteToHostInterface(Net().Hosts[index], dstIP)
		dstMAC = hostDetermineDstMAC(Net().Hosts[index], dstIP, iface.Name, true)
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	icmpReplyPacket := model.ICMPEchoPacket{
		ControlType: 0,
		ControlCode: 0,
		Checksum:    "checksum",
		Identifier:  receivedIcmpPacket.Identifier,
		SeqNumber:   receivedIcmpPacket.SeqNumber,
		Data:        receivedIcmpPacket.Data,
	}
	icmpReplyPacketBytes, _ := json.Marshal(icmpReplyPacket)
	ipv4PacketBytes := model.NewIPv4Packet(srcIP, dstIP, "ICMP", icmpReplyPacketBytes)
	frameBytes := model.NewFrame(srcMAC, dstMAC, "IPv4", ipv4PacketBytes)

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

	iface := model.Interface{}
	if srcID == Net().Router.ID {
		iface = Net().Router.Interfaces["eth0"]
	} else {
		index := getHostIndexFromID(srcID)
		iface = Net().Hosts[index].Interfaces["eth0"]
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	// First, check if it is trying to ARP itself.
	if targetIP == srcIP {
		debug(4, "arp_request", srcID, "Destination IP is source IP! Canceling ARP request.")
		return srcMAC
	}

	arpRequestMessage := model.ArpMessage{
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
	arpRequestFrameBytes := model.NewFrame(srcMAC, dstMAC, "ARP", arpRequestMessageBytes)

	// Send frame and wait for ARPREPLY
	sendFrame(arpRequestFrameBytes, iface, srcID)
	debug(3, "arp_request", srcID, "ARPREQUEST sent")

	sockets := engine.Instance().Sockets[srcID]
	socketID := "arp_" + string(targetIP)
	sockets[socketID] = make(chan model.Frame)

	select {
	case arpReplyFrameBytes := <-sockets[socketID]:
		arpReplyMessage, _ := model.ParseARPMessage(arpReplyFrameBytes.Data)
		return arpReplyMessage.SenderMAC

	case <-time.After(time.Second * 4):
		debug(1, "arp_request", srcID, "ARP request timed out.")
		return "TIMEOUT"
	}
}

func arp_reply(id string, arpRequestFrame model.Frame) {
	arpRequestMessage, _ := model.ParseARPMessage(arpRequestFrame.Data)

	// Construct frame
	srcID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := arpRequestMessage.SenderMAC // This usage of SenderMAC is according to ARP protocol.
	dstIP := arpRequestMessage.SenderIP

	// Network listener decided to reply to this request - no checking needed.
	iface := model.Interface{}
	if id == Net().Router.ID {
		iface = Net().Router.Interfaces["eth0"]
		srcID = Net().Router.ID
	} else {
		index := getHostIndexFromID(id)
		iface = Net().Hosts[index].Interfaces["eth0"]
		srcID = Net().Hosts[index].ID
	}

	srcIP = iface.IPConfig.IPAddress.String()
	srcMAC = iface.MACAddr

	arpReplyMessage := model.ArpMessage{
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
	arpReplyFrameBytes := model.NewFrame(srcMAC, dstMAC, "ARP", arpReplyMessageBytes)

	// Send frame
	sendFrame(arpReplyFrameBytes, iface, srcID)
	debug(3, "arp_reply", srcID, "ARPREPLY sent")
}

func dhcp_discover(host model.Host) {
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
	dhcpDiscoverMessage := model.DHCPMessage{
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
	segmentData := model.NewUDPSegment(68, 67, dhcpDiscoverMessageBytes)
	packetData := model.NewIPv4Packet(srcIP, dstIP, protocol, segmentData)
	frameData := model.NewFrame(srcMAC, dstMAC, "IPv4", packetData)

	// Send DHCPDISCOVER, await DHCPOFFER
	//need to give it to uplink
	sendFrame(frameData, iface, srcID)
	debug(3, "dhcp_discover", host.ID, "DHCPDISCOVER sent")

	sockets := engine.Instance().Sockets[srcID]
	socketID := "udp_" + strconv.Itoa(68)
	sockets[socketID] = make(chan model.Frame)
	dhcpOfferFrame := <-sockets[socketID]

	// De-encapsulate DHCPOFFER
	dhcpOfferIPv4Packet, _ := model.ParseIPv4Packet(dhcpOfferFrame.Data)
	dhcpOfferIPv4PacketHeader, _ := model.ParseIPv4PacketHeader(dhcpOfferIPv4Packet.Header)
	dhcpOfferUDPSegment, _ := model.ParseUDPSegment(dhcpOfferIPv4Packet.Data)
	dhcpOfferMessage, _ := model.ParseDHCPMessage(dhcpOfferUDPSegment.Data)

	if int(dhcpOfferMessage.Options[53][0]) == 6 { // 6 is DHCPNAK
		debug(1, "dhcp_discover", srcID, "Failed to obtain IP address: No free addresses available")
	} else {
		dstIP = dhcpOfferIPv4PacketHeader.SrcIP

		// Construct DHCPREQUEST
		options = map[byte][]byte{
			53: {3},                   // Option 53: DHCPREQUEST
			12: []byte(host.Hostname), // Option 12: Hostname
		}
		dhcpRequestMessage := model.DHCPMessage{
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
		dhcpRequestUDPSegment := model.NewUDPSegment(68, 67, dhcpRequestMessageBytes)
		dhcpRequestIPv4Packet := model.NewIPv4Packet(srcIP, dstIP, protocol, dhcpRequestUDPSegment)
		dhcpRequestFrame := model.NewFrame(srcMAC, dstMAC, "IPv4", dhcpRequestIPv4Packet)

		// Send DHCPREQUEST, await DHCPACK
		sendFrame(dhcpRequestFrame, iface, srcID)
		debug(3, "dhcp_discover", srcID, "DHCPREQUEST sent")
		dhcpAckFrame := <-sockets[socketID]

		// De-encapsulate DHCPACK
		dhcpAckIpv4Packet, _ := model.ParseIPv4Packet(dhcpAckFrame.Data)
		dhcpAckUDPSegment, _ := model.ParseUDPSegment(dhcpAckIpv4Packet.Data)
		dhcpAckMessage, _ := model.ParseDHCPMessage(dhcpAckUDPSegment.Data)

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
	engine.Instance().ActionSync[srcID] <- 1
}

func dhcp_offer(dhcpDiscoverFrame model.Frame) {
	// De-encapsulate DHCPDISCOVER
	dhcpDiscoverIPv4Packet, _ := model.ParseIPv4Packet(dhcpDiscoverFrame.Data)
	//dhcpDiscoverIpv4PacketHeader, _ := model.ParseIPv4PacketHeader(dhcpDiscoverIPv4Packet.Header)
	dhcpDiscoverUDPSegment, _ := model.ParseUDPSegment(dhcpDiscoverIPv4Packet.Data)
	dhcpDiscoverMessage, _ := model.ParseDHCPMessage(dhcpDiscoverUDPSegment.Data)

	iface := Net().Router.Interfaces["eth0"]
	srcIP := Net().Router.GetIP(iface.Name)
	dstIP := "255.255.255.255"
	srcMAC := iface.MACAddr
	dstMAC := dhcpDiscoverFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.

	// Find open address
	addr_to_give := Net().Router.NextFreePoolAddress()
	gateway := Net().Router.GetIP(iface.Name)
	netSize, _ := strconv.Atoi(Net().Netsize)
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
	dhcpOfferMessage := model.DHCPMessage{
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
	dhcpOfferSegment := model.NewUDPSegment(67, 68, dhcpOfferMessageBytes)
	dhcpOfferPacket := model.NewIPv4Packet(srcIP, dstIP, protocol, dhcpOfferSegment)
	dhcpOfferFrame := model.NewFrame(srcMAC, dstMAC, "IPv4", dhcpOfferPacket)

	// Send DHCPOFFER, await DHCPREQUEST
	sendFrame(dhcpOfferFrame, iface, Net().Router.ID)
	debug(3, "dhcp_offer", Net().Router.ID, "DHCPOFFER sent - "+addr_to_give.String())
}

func dhcp_ack(dhcpRequestFrame model.Frame) {
	// De-encapsulate DHCPREQUEST
	dhcpRequestIPv4Packet, _ := model.ParseIPv4Packet(dhcpRequestFrame.Data)
	dhcpRequestIPv4PacketHeader, _ := model.ParseIPv4PacketHeader(dhcpRequestIPv4Packet.Header)
	dhcpRequestUDPSegment, _ := model.ParseUDPSegment(dhcpRequestIPv4Packet.Data)
	dhcpRequestMessage, _ := model.ParseDHCPMessage(dhcpRequestUDPSegment.Data)

	iface := Net().Router.Interfaces["eth0"]
	srcIP := Net().Router.GetIP(iface.Name)
	dstIP := dhcpRequestIPv4PacketHeader.SrcIP
	srcMAC := iface.MACAddr
	dstMAC := dhcpRequestFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.

	messageType := 6
	if dhcpRequestUDPSegment.Data != nil {
		if int(dhcpRequestMessage.Options[53][0]) == 3 { // 3 = DHCPREQUEST
			if Net().Router.IsAvailableAddress(dhcpRequestMessage.YIAddr) {
				messageType = 5
			} else {
				debug(1, "dhcp_offer", Net().Router.ID, "Error: DHCP address requested is not available")
			}
		} else {
			debug(1, "dhcp_offer", Net().Router.ID, "Error: Empty DHCP request")
		}
	}

	gateway := Net().Router.GetIP(iface.Name)
	netSize, _ := strconv.Atoi(Net().Netsize)
	subnetmask := prefixLengthToSubnetMask(netSize)

	// Construct DHCPACK
	options := map[byte][]byte{
		53: {byte(messageType)},           // Option 53: DHCPACK
		1:  net.ParseIP(subnetmask).To4(), // Subnet mask
		3:  net.ParseIP(gateway).To4(),    // Gateway
		51: {0, 0, 10, 0},                 // Lease time
		54: net.ParseIP(gateway).To4(),    // DHCP server
	}
	dhcpAckMessage := model.DHCPMessage{
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
	dhcpAckSegment := model.NewUDPSegment(67, 68, dhcpAckMessageBytes)
	dhcpAckIPv4Packet := model.NewIPv4Packet(srcIP, dstIP, protocol, dhcpAckSegment)
	dhcpAckFrame := model.NewFrame(srcMAC, dstMAC, "IPv4", dhcpAckIPv4Packet)

	// Send DHCPACK
	sendFrame(dhcpAckFrame, iface, Net().Router.ID)
	debug(3, "dhcp_offer", Net().Router.ID, "DHCPACK sent - "+dhcpAckMessage.YIAddr.String())

	// Setting leasee's MAC in pool (new)
	pool := Net().Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].Equal(dhcpAckMessage.YIAddr) {
			debug(4, "dhcp_offer", Net().Router.ID, "Assigning and removing address "+dhcpAckMessage.YIAddr.String()+" from pool")
			Net().Router.DHCPPool.DHCPPoolLeases[dhcpAckMessage.YIAddr.String()] = dhcpAckMessage.CHAddr
		}
	}
}

func dns_query(srcID string, hostname string, reqType uint16) model.DNSMessage {
	srcIP := ""
	dstMAC := ""
	dstIP := ""
	iface := model.Interface{}

	if Net().Router.ID == srcID {
		iface = Net().Router.Interfaces["lo"]
		srcIP = Net().Router.GetIP(iface.Name)
		//dstIP := Net().Router.Interfaces["eth0"].IPConfig.DNSServer
		dstIP = Net().Router.Interfaces["lo"].IPConfig.DNSServer.String() // temporary
		dstMAC = routerDetermineDstMAC(Net().Router, dstIP, iface.Name, true)
	} else {
		hostIndex := getHostIndexFromID(srcID)
		host := Net().Hosts[hostIndex]
		iface = host.Interfaces["eth0"]
		srcIP = host.GetIP(iface.Name)
		//dstIP := host.Interfaces["eth0"].IPConfig.DNSServer
		dstIP = host.GetGateway(iface.Name) // temporary
		dstMAC = hostDetermineDstMAC(host, dstIP, iface.Name, true)
	}

	srcMAC := iface.MACAddr

	var dnsQueryMessage = model.DNSMessage{}
	switch reqType {
	case 'A':
		dnsQuestionMessage := model.DNSQuestion{
			QName:  hostname,
			QType:  reqType,
			QClass: 1,
		}
		dnsQuestions := make([]model.DNSQuestion, 1)
		dnsQuestions[0] = dnsQuestionMessage

		dnsQueryMessage = model.DNSMessage{
			QR:        false, // false = query
			Opcode:    0,
			QDCount:   1,
			Questions: dnsQuestions,
		}

	default:
		debug(1, "dns_query", srcID, "[Error] DNS query type not implemented yet")
		return model.DNSMessage{}
	}

	protocol := "UDP"
	srcPort := ephemeralPortGen()
	dnsQueryMessageBytes, _ := json.Marshal(dnsQueryMessage)
	dnsQuerySegment := model.NewUDPSegment(srcPort, 53, dnsQueryMessageBytes)
	dnsQueryIPv4Packet := model.NewIPv4Packet(srcIP, dstIP, protocol, dnsQuerySegment)
	dnsQueryFrame := model.NewFrame(srcMAC, dstMAC, "IPv4", dnsQueryIPv4Packet)

	sendFrame(dnsQueryFrame, iface, srcID)
	debug(3, "dns_query", srcID, "DNS query sent - "+hostname)

	sockets := engine.Instance().Sockets[srcID]
	socketID := "udp_" + strconv.Itoa(srcPort)
	sockets[socketID] = make(chan model.Frame)

	select {
	case dnsResponseFrame := <-sockets[socketID]:
		dnsResponsePacket, _ := model.ParseIPv4Packet(dnsResponseFrame.Data)
		dnsResponseSegment, _ := model.ParseUDPSegment(dnsResponsePacket.Data)
		dnsResponseMessage, _ := model.ParseDNSMessage(dnsResponseSegment.Data)

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
	return model.DNSMessage{}
}

func dns_response(dnsQueryFrame model.Frame) {
	// De-encapsulate DNS Query
	dnsQueryIPv4Packet, _ := model.ParseIPv4Packet(dnsQueryFrame.Data)
	dnsQueryIPv4PacketHeader, _ := model.ParseIPv4PacketHeader(dnsQueryIPv4Packet.Header)
	dnsQueryUDPSegment, _ := model.ParseUDPSegment(dnsQueryIPv4Packet.Data)
	dnsQueryMessage, _ := model.ParseDNSMessage(dnsQueryUDPSegment.Data)

	dstIP := dnsQueryIPv4PacketHeader.SrcIP
	iface, _ := engine.Instance().RouteToRouterInterface(Net().Router.ID, dstIP)
	srcIP := Net().Router.GetIP(iface.Name)
	dstPort := dnsQueryUDPSegment.SrcPort
	srcMAC := iface.MACAddr
	dstMAC := dnsQueryFrame.SrcMAC

	var dnsResponseMessage = model.DNSMessage{}
	switch dnsQueryMessage.Questions[0].QType {
	case 'A':
		dnsResponseMessage = model.DNSMessage{
			QR:     true, // false = query
			Opcode: 0,
			Rcode:  2,
		}

		dnsAnswerRecord, _ := Net().Router.DNSServer.ARecordLookup(dnsQueryMessage.Questions[0].QName)

		if dnsAnswerRecord.Name != "" {
			dnsResponseMessage.Rcode = 0
			dnsResponseMessage.ANCount = 1

			dnsAnswers := make([]model.DNSRecord, 1)
			dnsAnswers[0] = dnsAnswerRecord

			dnsResponseMessage.Answers = dnsAnswers
		} else {
			dnsResponseMessage.Rcode = 3
			dnsResponseMessage.ANCount = 0
		}

	default:
		debug(1, "dns_query", Net().Router.ID, "[Warning] DNS query type not implemented yet - returning SERVFAIL message")
		return
	}

	protocol := "UDP"
	dnsResponseMessageBytes, _ := json.Marshal(dnsResponseMessage)
	dnsResponseSegment := model.NewUDPSegment(53, dstPort, dnsResponseMessageBytes)
	dnsResponseIPv4Packet := model.NewIPv4Packet(srcIP, dstIP, protocol, dnsResponseSegment)
	dnsResponseFrame := model.NewFrame(srcMAC, dstMAC, "IPv4", dnsResponseIPv4Packet)

	sendFrame(dnsResponseFrame, iface, Net().Router.ID)
	debug(3, "dns_query", Net().Router.ID, "DNS response sent")

}

func ipset(hostname string, ipaddr string, subnetMask string) {
	defaultGateway := Net().Router.GetIP("eth0")

	fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetMask, defaultGateway)
	fmt.Print("\nIs this correct? [Y/n]: ")
	inScanner := engine.Instance().Scanner
	inScanner.Scan()
	affirmation := inScanner.Text()

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
	for h := range Net().Hosts {
		if Net().Hosts[h].Hostname == hostname {
			iface := Net().Hosts[h].Interfaces["eth0"]

			iface.IPConfig.IPAddress = net.ParseIP(ipaddr)
			iface.IPConfig.SubnetMask = subnetMask
			iface.IPConfig.DefaultGateway = net.ParseIP(defaultGateway)

			Net().Hosts[h].Interfaces["eth0"] = iface

			fmt.Println("Network configuration updated")
		}
	}
}

// Run an ARP request, but synchronize with client
func arpSynchronized(id string, targetIP string) {
	dstMAC := ""

	if Net().Router.ID == id {
		dstMAC = routerDetermineDstMAC(Net().Router, targetIP, "eth0", false)
	} else {
		dstMAC = hostDetermineDstMAC(Net().Hosts[getHostIndexFromID(id)], targetIP, "eth0", false)
	}

	if dstMAC != "" {
		achievementTester(ARP_HOT)
	}

	engine.Instance().ActionSync[id] <- 1
}

// A host determines the destination MAC to send to... Either by ARP, sending to GW, or reading ARP table
func hostDetermineDstMAC(srcHost model.Host, dstIP string, iface string, useTable bool) string {
	srcID := srcHost.ID
	dstMAC := ""

	if dstIP == "127.0.0.1" && iface == "lo" {
		return srcHost.Interfaces["lo"].MACAddr
	}

	// Same subnet - ARP table, or ARP request.
	if iphelper.IPInSameSubnet(srcHost.GetIP(iface), dstIP, srcHost.GetMask(iface)) {
		debug(4, "hostDetermineDstMAC", srcID, "Sending to same subnet, about to ARP table lookup or ARP")

		// Check ARP table
		if useTable && Net().Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP].MACAddr != "" {
			dstMAC = Net().Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP].MACAddr
		} else {
			// ARP request
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("ARP request timed out.\n")
			} else {
				arpEntry := model.ARPEntry{
					MACAddr:   dstMAC,
					Interface: Net().Hosts[getHostIndexFromID(srcID)].Interfaces[iface].RemoteL1ID,
				}
				Net().Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP] = arpEntry // Add to ARP table
			}
		}

	} else { // Different subnet - GW.
		debug(4, "hostDetermineDstMAC", srcID, "Sending to different subnet, sending to GW")
		gateway := srcHost.GetGateway(iface)

		// Check ARP table
		if Net().Hosts[getHostIndexFromID(srcID)].ARPTable[gateway].MACAddr != "" {
			dstMAC = Net().Hosts[getHostIndexFromID(srcID)].ARPTable[gateway].MACAddr
		} else {
			// ARP request
			dstMAC = arp_request(srcID, gateway)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("ARP request timed out.\n")
			} else {
				arpEntry := model.ARPEntry{
					MACAddr:   dstMAC,
					Interface: Net().Hosts[getHostIndexFromID(srcID)].Interfaces[iface].RemoteL1ID,
				}
				Net().Hosts[getHostIndexFromID(srcID)].ARPTable[gateway] = arpEntry // Add to ARP table
			}
		}
	}

	return dstMAC
}

// A router determines the destination MAC to send to... Either by ARP, or reading ARP table
func routerDetermineDstMAC(router *model.Router, dstIP string, iface string, useTable bool) string {
	if router == nil {
		return ""
	}
	dstMAC := ""

	if dstIP == "127.0.0.1" && iface == "lo" {
		return router.Interfaces["lo"].MACAddr
	}

	netsizeInt, _ := strconv.Atoi(Net().Netsize)
	subnetMask := prefixLengthToSubnetMask(netsizeInt)
	if iphelper.IPInSameSubnet(router.GetIP(iface), dstIP, subnetMask) {
		debug(4, "routerDetermineDstMAC", router.ID, "Same subnet; ARP table lookup or ARP")

		if useTable && router.ARPTable[dstIP].MACAddr != "" {
			dstMAC = router.ARPTable[dstIP].MACAddr
		} else {
			dstMAC = arp_request(router.ID, dstIP)
			if dstMAC == "TIMEOUT" {
				fmt.Printf("ARP request timed out.\n")
			} else {
				router.ARPTable[dstIP] = model.ARPEntry{
					MACAddr:   dstMAC,
					Interface: router.Interfaces[iface].RemoteL1ID,
				}
			}
		}
	} else {
		fmt.Printf("Error: Routing not implemented yet.\n")
	}

	return dstMAC
}

func resolveHostname(srcID string, hostname string, dnsTable map[string]model.DNSRecord) model.DNSRecord {
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
		if srcID == Net().Router.ID {
			Net().Router.DNSTable[hostname] = resultMessage.Answers[0]
		} else {
			Net().Hosts[getHostIndexFromID(srcID)].DNSTable[hostname] = resultMessage.Answers[0]
		}
	}

	return resultMessage.Answers[0]
}
