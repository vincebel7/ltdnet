/*
File:		datagram.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Datagram structs, and associated functions
*/

package main

import (
	"encoding/json"
	"fmt"
)

/** L4 structs and functions **/
type UDPSegment struct {
	SrcPort int    `json:"src_port"`
	DstPort int    `json:"dst_port"`
	Data    string `json:"data"`
}

func constructUDPSegment(data string, srcPort int, dstPort int) json.RawMessage {
	segment := UDPSegment{
		SrcPort: srcPort,
		DstPort: dstPort,
		Data:    data,
	}

	segmentBytes, _ := json.Marshal(segment)
	return segmentBytes
}

// Turns segment into an accessible object
func readUDPSegment(rawUDPSegment string) UDPSegment {
	var segment UDPSegment
	err := json.Unmarshal([]byte(rawUDPSegment), &segment)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return UDPSegment{}
	}

	return segment
}

/** L3 structs and functions **/
type PacketHeader struct {
	Protocol int    `json:"protocol"`
	SrcIP    string `json:"src_ip"`
	DstIP    string `json:"dst_ip"`
}

type IPv4Packet struct {
	Header json.RawMessage `json:"header"`
	Data   json.RawMessage `json:"data"`
}

type ICMPPacket struct {
	ControlType int             `json:"control_type"` // 8 for Request, 0 for Reply
	Data        json.RawMessage `json:"data"`
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

// Turns packet into an accessible object
func readIPv4Packet(rawPacket string) IPv4Packet {
	var packet IPv4Packet
	err := json.Unmarshal([]byte(rawPacket), &packet)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return IPv4Packet{}
	}

	return packet
}

func readIPv4PacketHeader(rawPacketHeader string) PacketHeader {
	var packetHeader PacketHeader
	err := json.Unmarshal([]byte(rawPacketHeader), &packetHeader)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return PacketHeader{}
	}

	return packetHeader
}

func readICMPPacket(rawPacket string) ICMPPacket {
	var packet ICMPPacket
	err := json.Unmarshal([]byte(rawPacket), &packet)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ICMPPacket{}
	}

	return packet
}

/** L2 structs and functions **/
type Frame struct {
	SrcMAC    string          `json:"src_mac"`
	DstMAC    string          `json:"dst_mac"`
	EtherType string          `json:"ether_type"`
	Data      json.RawMessage `json:"data"`
}

type ArpMessage struct {
	Opcode    int    `json:"opcode"` // 1 for request, 2 for reply
	SenderMAC string `json:"sender_mac"`
	SenderIP  string `json:"sender_ip"`
	TargetMAC string `json:"target_mac"`
	TargetIP  string `json:"target_ip"`
}

func constructFrame(data string, srcMAC string, dstMAC string, protocolName string) string {
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
		Data:      json.RawMessage(data),
	}

	frameBytes, _ := json.Marshal(frame)
	return string(frameBytes)
}

// Turns frame into an accessible object
func readFrame(rawFrame string) Frame {
	var frame Frame
	err := json.Unmarshal([]byte(rawFrame), &frame)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return Frame{}
	}

	return frame
}

// Turns ArpMessage into an accessible object
func readArpMessage(rawPacket string) ArpMessage {
	var arpMessage ArpMessage
	err := json.Unmarshal([]byte(rawPacket), &arpMessage)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return ArpMessage{}
	}

	return arpMessage
}
