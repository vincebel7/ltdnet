/*
File:		datagram.go
Author: 	https://github.com/vincebel7
Purpose:	Datagram structs, and associated functions
*/

package main

import (
	"encoding/json"
	"fmt"
	"net"
)

/** Datagram Structs - L5 **/
type DHCPMessage struct {
	Op      byte            // Message type: 1 = Request, 2 = Reply
	HType   byte            // Hardware address type (e.g., 1 for Ethernet)
	HLen    byte            // Length of hardware address
	Hops    byte            // Hops
	XID     uint32          // Transaction ID
	Flags   uint16          // Flags (e.g., broadcast)
	CIAddr  net.IP          // Client IP address
	YIAddr  net.IP          // 'Your' IP address (server's offer)
	SIAddr  net.IP          // Server IP address
	GIAddr  net.IP          // Gateway IP address
	CHAddr  string          // Client MAC address
	Options map[byte][]byte // DHCP options
}

/** Datagram Structs - L4 **/
type UDPSegment struct {
	SrcPort  int             `json:"src_port"`
	DstPort  int             `json:"dst_port"`
	Length   string          `json:"length"`
	Checksum string          `json:"checksum"`
	Data     json.RawMessage `json:"data"`
}

/** Datagram Structs - L3 **/
type IPv4Packet struct {
	Header json.RawMessage `json:"header"`
	Data   json.RawMessage `json:"data"`
}

type PacketHeader struct {
	Protocol int    `json:"protocol"`
	SrcIP    string `json:"src_ip"`
	DstIP    string `json:"dst_ip"`
}

type ICMPEchoPacket struct {
	ControlType int             `json:"control_type"` // 8 for Request, 0 for Reply
	ControlCode int             `json:"control_code"` // Often 0
	Checksum    string          `json:"checksum"`
	Identifier  int             `json:"identifier"`
	SeqNumber   int             `json:"seq"`
	Data        json.RawMessage `json:"data"`
}

/** Datagram Structs - L2 **/
type Frame struct {
	SrcMAC    string          `json:"src_mac"`
	DstMAC    string          `json:"dst_mac"`
	EtherType string          `json:"ether_type"`
	Data      json.RawMessage `json:"data"`
}

type ArpMessage struct {
	HTYPE     int    `json:"HTYPE"`
	PTYPE     string `json:"PTYPE"`
	HLEN      int    `json:"HLEN"`
	PLEN      int    `json:"PLEN"`
	Opcode    int    `json:"OPER"` // 1 for request, 2 for reply
	SenderMAC string `json:"SHA"`
	SenderIP  string `json:"SPA"`
	TargetMAC string `json:"THA"`
	TargetIP  string `json:"TPA"`
}

func ConstructDHCPMessage(
	op byte, htype byte, hlen byte, xid uint32,
	ciaddr, yiaddr, siaddr, giaddr net.IP,
	chaddr string, options map[byte][]byte,
) json.RawMessage {
	dhcpMessage := DHCPMessage{
		Op:      op,
		HType:   htype,
		HLen:    hlen,
		XID:     xid,
		CIAddr:  ciaddr,
		YIAddr:  yiaddr,
		SIAddr:  siaddr,
		GIAddr:  giaddr,
		CHAddr:  chaddr,
		Options: options,
	}

	messageBytes, _ := json.Marshal(dhcpMessage)
	return messageBytes
}

func constructUDPSegment(srcPort int, dstPort int, data json.RawMessage) json.RawMessage {
	segment := UDPSegment{
		SrcPort: srcPort,
		DstPort: dstPort,
		Data:    data,
	}

	segmentBytes, _ := json.Marshal(segment)
	return segmentBytes
}

func constructIPv4Packet(srcIP string, dstIP string, protocolName string, data json.RawMessage) json.RawMessage {
	protocolNumber := -1
	switch protocolName {
	case "UDP":
		protocolNumber = 17
	case "ICMP":
		protocolNumber = 1
	}

	header := PacketHeader{
		Protocol: protocolNumber,
		SrcIP:    srcIP,
		DstIP:    dstIP,
	}

	packetHeaderBytes, _ := json.Marshal(header)

	packet := IPv4Packet{
		Header: json.RawMessage(packetHeaderBytes),
		Data:   data,
	}

	packetBytes, _ := json.Marshal(packet)
	return json.RawMessage(packetBytes)
}

func constructFrame(srcMAC string, dstMAC string, protocolName string, data json.RawMessage) json.RawMessage {
	etherType := "0x0"
	switch protocolName {
	case "IPv4":
		etherType = "0x0800"
	case "ARP":
		etherType = "0x0806"
	}

	frame := Frame{
		SrcMAC:    srcMAC,
		DstMAC:    dstMAC,
		EtherType: etherType,
		Data:      data,
	}

	frameBytes, _ := json.Marshal(frame)
	return frameBytes
}

// Turns DHCPMessage into an accessible object
func ReadDHCPMessage(rawDHCPMessage json.RawMessage) DHCPMessage {
	var dhcpMessage DHCPMessage
	err := json.Unmarshal(rawDHCPMessage, &dhcpMessage)
	if err != nil {
		fmt.Println("[DHCP] Error unmarshalling JSON:", err)
		return DHCPMessage{}
	}

	return dhcpMessage
}

// Turns segment into an accessible object
func readUDPSegment(rawUDPSegment json.RawMessage) UDPSegment {
	var segment UDPSegment
	err := json.Unmarshal(rawUDPSegment, &segment)
	if err != nil {
		fmt.Println("[UDP] Error unmarshalling JSON:", err)
		return UDPSegment{}
	}

	return segment
}

// Turns packet into an accessible object
func readIPv4Packet(rawPacket json.RawMessage) IPv4Packet {
	var packet IPv4Packet
	err := json.Unmarshal(rawPacket, &packet)
	if err != nil {
		fmt.Println("[IPv4 Packet] Error unmarshalling JSON:", err)
		return IPv4Packet{}
	}

	return packet
}

// Turns IPv4 packet into an accessible object
func readIPv4PacketHeader(rawPacketHeader json.RawMessage) PacketHeader {
	var packetHeader PacketHeader
	err := json.Unmarshal(rawPacketHeader, &packetHeader)
	if err != nil {
		fmt.Println("[IPv4 Header] Error unmarshalling JSON:", err)
		return PacketHeader{}
	}

	return packetHeader
}

// Turns ICMP echo packet into an accessible object
func readICMPEchoPacket(rawPacket json.RawMessage) ICMPEchoPacket {
	var packet ICMPEchoPacket
	err := json.Unmarshal(rawPacket, &packet)
	if err != nil {
		fmt.Println("[ICMP Packet] Error unmarshalling JSON:", err)
		return ICMPEchoPacket{}
	}

	return packet
}

// Turns frame into an accessible object
func readFrame(rawFrame json.RawMessage) Frame {
	var frame Frame
	err := json.Unmarshal(rawFrame, &frame)
	if err != nil {
		fmt.Println("[Frame] Error unmarshalling JSON:", err)
		return Frame{}
	}

	return frame
}

// Turns ArpMessage into an accessible object
func readArpMessage(rawMessage json.RawMessage) ArpMessage {
	var arpMessage ArpMessage
	err := json.Unmarshal(rawMessage, &arpMessage)
	if err != nil {
		fmt.Println("[ARP] Error unmarshalling JSON:", err)
		return ArpMessage{}
	}

	return arpMessage
}
