/*
File:		structs.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Datagram and device structs, and helpful functions
*/

package main

import(
	"bufio"
	"os"
)

type Network struct {
	ID		string `json:"id"`
	Name		string `json:"name"`
	Author		string `json:"auth"`
	Class		string `json:"clas"`
	Router		Router `json:"rtr"`
	Switches	[]Switch `json:"swts"`
	Hosts		[]Host `json:"hsts"`
	DebugLevel	int	`json:"dbug"`
}

var snet Network //selected network, essentially the loaded save file
var listenSync = make(chan string)
var scanner = bufio.NewScanner(os.Stdin)

type Router struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"` // LAN-facing interface
	Hostname	string `json:"hnme"`
	Gateway		string `json:"gway"`
	DHCPPool	int `json:"dpol"` //maximum, not just available
	//Downports	int `json:"dpts"`
	//Ports		[]string `json:"prtt"`
	VSwitch		Switch	`json:"vsid"` // Virtual built-in switch to router
	//MACTable	map[string]int `json:"mact"`
	DHCPTable	map[string]string `json:"dhct"` //maps IP address to MAC address
}

type Switch struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	Hostname	string `json:"hnme"`
	MgmtIP		string `json:"mgip"`
	MACTable	map[string]int `json:"mact"`
	Maxports	int `json:"mxpt"`
	Ports		[]string `json:"prts"` // maps port # to downlink ID
	PortIDs		[]string `json:"pids"` // maps port # to Port ID
	PortMACs	[]string `json:"pmcs"` // maps port # to interface MAC address
}

type Host struct {
	ID		string `json:"id"`
	Model		string `json:"modl"`
	MACAddr		string `json:"maca"`
	Hostname	string `json:"hnme"`
	IPAddr		string `json:"ipa"`
	SubnetMask	string `json:"mask"`
	DefaultGateway	string `json:"gway"`
	UplinkID	string `json:"upid"`
}

/* DATAGRAMS */

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
