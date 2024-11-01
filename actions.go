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

func ping(srcID string, dstIP string, count int) {
	debug(4, "ping", srcID, "About to ping")

	identifier := idgen_int(5)
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srcHostname := ""
	srcHost := Host{}

	sendCount := 0
	recvCount := 0
	lossCount := 0

	if snet.Router.ID == srcID {
		srcHostname = snet.Router.Hostname
		srcIP = snet.Router.GetIP()
		srcMAC = snet.Router.Interface.MACAddr
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				srcHost = snet.Hosts[h]
				srcHostname = snet.Hosts[h].Hostname
				srcIP = snet.Hosts[h].GetIP()
				srcMAC = snet.Hosts[h].Interface.MACAddr
				linkID = snet.Hosts[h].UplinkID
			}
		}
	}

	fmt.Printf("\nPinging %s from %s\n", dstIP, srcHostname)

	for i := 0; i < count; i++ {
		// Get destination MAC address
		if snet.Router.ID == srcID {
			dstMAC = routerDetermineDstMAC(snet.Router, dstIP, true)

			// Get linkID (which link to send ping over)
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID

			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}
		} else {
			dstMAC = hostDetermineDstMAC(srcHost, dstIP, true)
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
				achievementTester(UNITED_PINGDOM)
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
	dstMAC := frame.SrcMAC // Value only used if it cannot be determined below

	if snet.Router.ID == srcID {
		srcMAC = snet.Router.Interface.MACAddr
		srcIP = snet.Router.GetIP()
		dstMAC := routerDetermineDstMAC(snet.Router, dstIP, true)

		//Get link to send ping to
		dstID := getIDfromMAC(dstMAC)
		if getHostIndexFromID(dstID) != -1 {
			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		} else if snet.Router.ID == dstID {
			linkID = snet.Router.ID
		}

	} else {
		index := getHostIndexFromID(srcID)
		srcMAC = snet.Hosts[index].Interface.MACAddr
		srcIP = snet.Hosts[index].GetIP()
		dstMAC = hostDetermineDstMAC(snet.Hosts[index], dstIP, true)

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
	linkID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := "ff:ff:ff:ff:ff:ff"

	if srcID == snet.Router.ID {
		srcIP = snet.Router.GetIP()
		srcMAC = snet.Router.Interface.MACAddr
		linkID = snet.Router.LANLinkID
	} else {
		index := getHostIndexFromID(srcID)
		srcIP = snet.Hosts[index].GetIP()
		srcMAC = snet.Hosts[index].Interface.MACAddr
		linkID = snet.Hosts[index].UplinkID
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
	debug(4, "arp_request", srcID, "ARPREQUEST sent - linkid: "+linkID)

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
	dstMAC := arpRequestMessage.SenderMAC // This usage of SenderMAC is according to ARP protocol.
	dstIP := arpRequestMessage.SenderIP

	// Network listener decided to reply to this request - no checking needed.
	if id == snet.Router.ID {
		srcID = snet.Router.ID
		srcMAC = snet.Router.Interface.MACAddr
		srcIP = snet.Router.GetIP()

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
		srcMAC = snet.Hosts[i].Interface.MACAddr
		srcIP = snet.Hosts[i].GetIP()
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
	srcIP := host.GetIP()
	srcMAC := host.Interface.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := "ff:ff:ff:ff:ff:ff"
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

		// Send DHCPREQUEST, await DHCPACK
		channels[linkID] <- dhcpRequestFrame
		debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent")
		dhcpAckFrame := <-sockets[socketID]

		// De-encapsulate DHCPACK
		dhcpAckIpv4Packet := readIPv4Packet(dhcpAckFrame.Data)
		dhcpAckUDPSegment := readUDPSegment(dhcpAckIpv4Packet.Data)
		dhcpAckMessage := ReadDHCPMessage(dhcpAckUDPSegment.Data)

		if int(dhcpAckMessage.Options[53][0]) == 5 {
			debug(2, "dhcp_discover", srcID, "DHCPACK received - "+dhcpAckMessage.YIAddr.String())

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

	srcIP := snet.Router.GetIP()
	dstIP := "255.255.255.255"
	srcMAC := snet.Router.Interface.MACAddr
	dstMAC := dhcpDiscoverFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid

	// Find open address
	addr_to_give := snet.Router.NextFreePoolAddress()
	gateway := snet.Router.GetIP()
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
}

func dhcp_ack(dhcpRequestFrame Frame) {
	// De-encapsulate DHCPREQUEST
	dhcpRequestIPv4Packet := readIPv4Packet(dhcpRequestFrame.Data)
	dhcpRequestIPv4PacketHeader := readIPv4PacketHeader(dhcpRequestIPv4Packet.Header)
	dhcpRequestUDPSegment := readUDPSegment(dhcpRequestIPv4Packet.Data)
	dhcpRequestMessage := ReadDHCPMessage(dhcpRequestUDPSegment.Data)

	srcIP := snet.Router.GetIP()
	dstIP := dhcpRequestIPv4PacketHeader.SrcIP
	srcMAC := snet.Router.Interface.MACAddr
	dstMAC := dhcpRequestFrame.SrcMAC // This usage of SrcMAC is according to DHCP protocol.
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid

	messageType := 6
	if dhcpRequestUDPSegment.Data != nil {
		if int(dhcpRequestMessage.Options[53][0]) == 3 { // 3 = DHCPREQUEST
			if snet.Router.IsAvailableAddress(dhcpRequestMessage.YIAddr) {
				messageType = 5
			} else {
				debug(1, "dhcp_offer", snet.Router.ID, "Error 4: DHCP address requested is not available")
			}
		} else {
			debug(1, "dhcp_offer", snet.Router.ID, "Error 3: Empty DHCP request")
		}
	}

	gateway := snet.Router.GetIP()
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
	channels[linkID] <- dhcpAckFrame
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPACK sent - "+dhcpAckMessage.YIAddr.String())

	// Setting leasee's MAC in pool (new)
	pool := snet.Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].Equal(dhcpAckMessage.YIAddr) {
			debug(4, "dhcp_offer", snet.Router.ID, "Assigning and removing address "+dhcpAckMessage.YIAddr.String()+" from pool")
			snet.Router.DHCPPool.DHCPPoolLeases[dhcpAckMessage.YIAddr.String()] = dhcpAckMessage.CHAddr
		}
	}
}

func ipset(hostname string, ipaddr string) {
	prefixLength, _ := strconv.Atoi(snet.Netsize)
	subnetMask := prefixLengthToSubnetMask(prefixLength)
	defaultGateway := snet.Router.GetIP()

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
			snet.Hosts[h].Interface.IPConfig.IPAddress = net.ParseIP(ipaddr)
			snet.Hosts[h].Interface.IPConfig.SubnetMask = subnetMask
			snet.Hosts[h].Interface.IPConfig.DefaultGateway = net.ParseIP(defaultGateway)
			fmt.Println("Network configuration updated")
		}
	}
}

// Run an ARP request, but synchronize with client
func arpSynchronized(id string, targetIP string) {
	dstMAC := ""
	if snet.Router.ID == id {
		dstMAC = routerDetermineDstMAC(snet.Router, targetIP, false)
	} else {
		dstMAC = hostDetermineDstMAC(snet.Hosts[getHostIndexFromID(id)], targetIP, false)
	}

	if dstMAC != "" {
		achievementTester(ARP_HOT)
	}

	actionsync[id] <- 1
}

// A host determines the destination MAC to send to... Either by ARP, sending to GW, or reading ARP table
func hostDetermineDstMAC(srcHost Host, dstIP string, useTable bool) string {
	srcID := srcHost.ID
	dstMAC := ""

	// Same subnet - ARP table, or ARP request.
	if iphelper.IPInSameSubnet(srcHost.GetIP(), dstIP, srcHost.GetMask()) {
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
					Interface: snet.Hosts[getHostIndexFromID(srcID)].UplinkID,
				}
				snet.Hosts[getHostIndexFromID(srcID)].ARPTable[dstIP] = arpEntry // Add to ARP table
			}
		}

	} else { // Different subnet - GW.
		debug(4, "hostDetermineDstMAC", srcID, "Sending to different subnet, sending to GW")
		gateway := srcHost.GetGateway()

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
					Interface: snet.Hosts[getHostIndexFromID(srcID)].UplinkID,
				}
				snet.Hosts[getHostIndexFromID(srcID)].ARPTable[gateway] = arpEntry // Add to ARP table
			}
		}
	}

	return dstMAC
}

// A router determines the destination MAC to send to... Either by ARP, or reading ARP table
func routerDetermineDstMAC(router Router, dstIP string, useTable bool) string {
	dstMAC := ""

	// Same subnet - ARP table, or ARP request.
	netsizeInt, _ := strconv.Atoi(snet.Netsize)
	subnetMask := prefixLengthToSubnetMask(netsizeInt)
	if iphelper.IPInSameSubnet(router.GetIP(), dstIP, subnetMask) {
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
					Interface: router.LANLinkID,
				}
				snet.Router.ARPTable[dstIP] = arpEntry // Add to ARP table
			}
		}

	} else {
		fmt.Printf("Error: Routing not implemented yet.")
	}

	return dstMAC
}
