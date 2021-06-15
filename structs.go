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
	Author		string `json:"author"`
	Netsize		string `json:"netsize"`
	Router		Router `json:"router"`
	Switches	[]Switch `json:"switches"`
	Hosts		[]Host `json:"hosts"`
	DebugLevel	int	`json:"debug_level"`
}

var snet Network //selected network, essentially the loaded save file
var listenSync = make(chan string)
var scanner = bufio.NewScanner(os.Stdin)

type Router struct {
	ID		string `json:"id"`
	Model		string `json:"model"`
	MACAddr		string `json:"macaddr"` // LAN-facing interface
	Hostname	string `json:"hostname"`
	Gateway		string `json:"gateway"`
	DHCPPool	int `json:"dhcppool"` //maximum, not just available
	//Downports	int `json:"dpts"`
	//Ports		[]string `json:"prtt"`
	VSwitch		Switch	`json:"vswitchid"` // Virtual built-in switch to router
	//MACTable	map[string]int `json:"mact"`
	DHCPTable	map[string]string `json:"dhcptable"` //maps IP address to MAC address
	DHCPTableOrder	[]string `json:"dhcptableorder"` //reference for proper sorting of IP addresses
}

type Switch struct {
	ID		string `json:"id"`
	Model		string `json:"model"`
	Hostname	string `json:"hostname"`
	MgmtIP		string `json:"mgmtip"`
	MACTable	map[string]int `json:"mactable"`
	Maxports	int `json:"maxports"`
	Ports		[]string `json:"ports"` // maps port # to downlink ID
	PortIDs		[]string `json:"portids"` // maps port # to Port ID
	PortMACs	[]string `json:"portmacs"` // maps port # to interface MAC address
}

type Host struct {
	ID		string `json:"id"`
	Model		string `json:"model"`
	MACAddr		string `json:"macaddr"`
	Hostname	string `json:"hostname"`
	IPAddr		string `json:"ipaddr"`
	SubnetMask	string `json:"mask"`
	DefaultGateway	string `json:"gateway"`
	UplinkID	string `json:"uplinkid"`
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
