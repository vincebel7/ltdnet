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
)

func ping(srcID string, dstIP string, count int) {
	debug(4, "ping", srcID, "About to ping")

	identifier := idgen_int(5)
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""

	if snet.Router.ID == srcID {
		srchost = snet.Router.Hostname
		srcIP = snet.Router.Gateway.String()
		srcMAC = snet.Router.MACAddr

		// Get linkID (which link to send ping over)
		dstID := getIDfromMAC(dstMAC)
		if getHostIndexFromID(dstID) != -1 {
			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		} else if snet.Router.ID == dstID {
			linkID = snet.Router.ID
		}
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				srchost = snet.Hosts[h].Hostname
				srcIP = snet.Hosts[h].IPAddr.String()
				srcMAC = snet.Hosts[h].MACAddr
				linkID = snet.Hosts[h].UplinkID
			}
		}
	}

	sendCount := 0
	recvCount := 0
	lossCount := 0

	fmt.Printf("\nPinging %s from %s\n", dstIP, srchost)

	for i := 0; i < count; i++ {
		// Get MAC addresses
		if snet.Router.ID == srcID {
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("Request timed out.\n")
				lossCount++
				sendCount++
				continue
			}

		} else { // Assumed to be host source.
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, dstIP)
			if dstMAC == "TIMEOUT" { // ARP did not return a MAC
				fmt.Printf("Request timed out.\n")
				lossCount++
				sendCount++
				continue
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

		debug(4, "ping", srcID, "Awaiting ping send")
		channels[linkID] <- frameBytes
		debug(2, "ping", srcID, "Ping sent")

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
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))

	actionsync[srcID] <- 1
}

func pong(srcID string, frame Frame) {
	receivedIpv4Packet := readIPv4Packet(frame.Data)
	receivedIcmpPacket := readICMPEchoPacket(receivedIpv4Packet.Data)

	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstIP := readIPv4PacketHeader(receivedIpv4Packet.Header).SrcIP
	dstMAC := frame.SrcMAC // TODO: get MAC myself via ARP/MAC table

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
	channels[linkID] <- frameData
	debug(2, "dhcp_discover", host.ID, "DHCPDISCOVER sent")

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

		// Send DHCPREQUEST, await DHCPACKNOWLEDGEMENT
		channels[linkID] <- dhcpRequestFrame
		debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent")
		dhcpAckFrame := <-sockets[socketID]

		// De-encapsulate DHCPACKNOWLEDGEMENT
		dhcpAckIpv4Packet := readIPv4Packet(dhcpAckFrame.Data)
		dhcpAckUDPSegment := readUDPSegment(dhcpAckIpv4Packet.Data)
		dhcpAckMessage := ReadDHCPMessage(dhcpAckUDPSegment.Data)

		if int(dhcpAckMessage.Options[53][0]) == 5 {
			debug(2, "dhcp_discover", srcID, "DHCPACKNOWLEDGEMENT received - "+dhcpAckMessage.YIAddr.String())

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

	srcIP := snet.Router.Gateway.String()
	dstIP := "255.255.255.255"
	srcMAC := snet.Router.MACAddr
	dstMAC := dhcpDiscoverFrame.SrcMAC
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid

	// Find open address
	addr_to_give := snet.Router.NextFreePoolAddress()
	gateway := snet.Router.Gateway.String()
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
	channels[linkID] <- dhcpOfferFrame
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - "+addr_to_give.String())

	socketID := "udp_" + strconv.Itoa(67)
	sockets := socketMaps[snet.Router.ID]
	sockets[socketID] = make(chan Frame)
	socketMaps[snet.Router.ID] = sockets // Write updated map back to the collection
	requestFrame := <-sockets[socketID]

	// De-encapsulate DHCPREQUEST
	dhcpRequestIPv4Packet := readIPv4Packet(requestFrame.Data)
	dhcpRequestIPv4PacketHeader := readIPv4PacketHeader(dhcpRequestIPv4Packet.Header)
	dhcpRequestUDPSegment := readUDPSegment(dhcpRequestIPv4Packet.Data)
	dhcpRequestMessage := ReadDHCPMessage(dhcpRequestUDPSegment.Data)

	messageType = 6
	if dhcpRequestUDPSegment.Data != nil {
		if int(dhcpRequestMessage.Options[53][0]) == 3 { // 3 = DHCPREQUEST
			if dhcpRequestMessage.YIAddr.Equal(addr_to_give) {
				messageType = 5
			} else {
				debug(1, "dhcp_offer", snet.Router.ID, "Error 4: DHCP address requested is not same as offer")
			}
		} else {
			debug(1, "dhcp_offer", snet.Router.ID, "Error 3: Empty DHCP request")
		}
	}

	dstIP = dhcpRequestIPv4PacketHeader.SrcIP

	messageType = 6
	if addr_to_give != nil {
		messageType = 5
	}

	// Construct DHCPACKNOWLEDGEMENT
	options = map[byte][]byte{
		53: {byte(messageType)}, // Option 53: DHCPACKNOWLEDGEMENT
		1:  []byte(subnetmask),  // Subnet mask
		3:  []byte(gateway),     // Gateway
		51: {0, 0, 10, 0},       // Lease time
		54: []byte(gateway),     // DHCP server
	}
	dhcpAckMessage := DHCPMessage{
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

	// Encapsulate DHCPACKNOWLEDGEMENT
	dhcpAckMessageBytes, _ := json.Marshal(dhcpAckMessage)
	dhcpAckSegment := constructUDPSegment(67, 68, dhcpAckMessageBytes)
	dhcpAckIPv4Packet := constructIPv4Packet(srcIP, dstIP, protocol, dhcpAckSegment)
	dhcpAckFrame := constructFrame(srcMAC, dstMAC, "IPv4", dhcpAckIPv4Packet)

	// Send DHCPACKNOWLEDGEMENT
	channels[linkID] <- dhcpAckFrame
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPACKNOWLEDGEMENT sent - "+addr_to_give.String())

	// Setting leasee's MAC in pool (new)
	pool := snet.Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].Equal(addr_to_give) {
			debug(4, "dhcp_offer", snet.Router.ID, "Assigning and removing address "+addr_to_give.String()+" from pool")
			snet.Router.DHCPPool.DHCPPoolLeases[addr_to_give.String()] = getMACfromID(dstid) //NI TODO have client pass their MAC in DHCPREQUEST instead of relying on this NI
		}
	}
}

func ipset(hostname string, ipaddr string) {
	prefixLength, _ := strconv.Atoi(snet.Netsize)
	subnetMask := prefixLengthToSubnetMask(prefixLength)
	defaultGateway := snet.Router.Gateway.String()

	fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetMask, defaultGateway)
	fmt.Print("\nIs this correct? [Y/n/exit]")
	scanner.Scan()
	affirmation := scanner.Text()

	if strings.ToUpper(affirmation) == "Y" {
		// error checking
		if net.ParseIP(ipaddr).To4() == nil {
			fmt.Printf("Error: '%s' is not a valid IP address\n", ipaddr)

			return
		}

	} else if strings.ToUpper(affirmation) == "EXIT" {
		fmt.Println("Network changes reverted")

		return
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			snet.Hosts[h].IPAddr = net.ParseIP(ipaddr)
			snet.Hosts[h].SubnetMask = subnetMask
			snet.Hosts[h].DefaultGateway = net.ParseIP(defaultGateway)
			fmt.Println("Network configuration updated")
		}
	}
}

// Run an ARP request, but synchronize with client
func arpSynchronized(id string, targetIP string) {
	arp_request(id, targetIP)
	actionsync[id] <- 1
}
