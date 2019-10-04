/*
File:		datagrams.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Datagram structs and helpful functions
*/

package main

type Segment struct {
	Protocol	string
	SrcPort		int
	DstPort		int
	Data		string
}

type Packet struct {
	SrcIP		string
	DstIP		string
	Data		Segment
}

type Frame struct {
	SrcMAC		string
	DstMAC		string
	Data		Packet
}

func constructSegment(data string) Segment {
	srcport := 7
	dstport := 7
	protocol := "UDP"

	s := Segment{
		Protocol: protocol,
		SrcPort: srcport,
		DstPort: dstport,
		Data: data,
	}

	return s
}

func constructPacket(srcIP string, dstIP string, data Segment) Packet {
	p := Packet{
		SrcIP: srcIP,
		DstIP: dstIP,
		Data: data,
	}

	return p
}

func constructFrame(data Packet, srcMAC string, dstMAC string) Frame {
	f := Frame{
		SrcMAC: srcMAC,
		DstMAC: dstMAC,
		Data: data,
	}

	return f
}
