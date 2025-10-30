/*
File:		datagram.go
Author: 	https://github.com/vincebel7
Purpose:	Datagram structs, and associated functions
*/

package model

import (
	"encoding/json"
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

type DNSMessage struct {
	ID         uint16        `json:"id"`         // Unique identifier for the DNS transaction
	QR         bool          `json:"qr"`         // Query (0) or Response (1) flag
	Opcode     uint8         `json:"opcode"`     // Type of query (standard = 0, inverse = 1, status = 2)
	AA         bool          `json:"aa"`         // Authoritative Answer flag
	TC         bool          `json:"tc"`         // Truncation flag
	RD         bool          `json:"rd"`         // Recursion Desired flag
	RA         bool          `json:"ra"`         // Recursion Available flag
	Z          uint8         `json:"z"`          // Reserved, must be 0
	Rcode      uint8         `json:"rcode"`      // Response code
	QDCount    uint16        `json:"qd_count"`   // Number of questions
	ANCount    uint16        `json:"an_count"`   // Number of answer records
	NSCount    uint16        `json:"ns_count"`   // Number of authority records
	ARCount    uint16        `json:"ar_count"`   // Number of additional records
	Questions  []DNSQuestion `json:"questions"`  // Question section
	Answers    []DNSRecord   `json:"answers"`    // Answer section
	Authority  []DNSRecord   `json:"authority"`  // Authority section
	Additional []DNSRecord   `json:"additional"` // Additional section
}

/** Datagram Structs - L4 **/
type UDPSegment struct {
	SrcPort  int             `json:"src_port"` // Source port
	DstPort  int             `json:"dst_port"` // Destination port
	Length   string          `json:"length"`   // Payload length
	Checksum string          `json:"checksum"` // Checksum for error-checking
	Data     json.RawMessage `json:"data"`     // Payload data
}

type TCPSegment struct {
	SrcPort    int             `json:"src_port"`    // Source port
	DstPort    int             `json:"dst_port"`    // Destination port
	SeqNumber  int             `json:"seq_number"`  // Sequence number
	AckNumber  int             `json:"ack_number"`  // Acknowledgment number
	Offset     int             `json:"offset"`      // Data offset (header length)
	Reserved   int             `json:"reserved"`    // Reserved bits for future use
	Flags      TCPFlags        `json:"flags"`       // Flags for control information
	WindowSize int             `json:"window_size"` // Window size for flow control
	Checksum   string          `json:"checksum"`    // Checksum for error-checking
	UrgentPtr  int             `json:"urgent_ptr"`  // Urgent pointer for urgent data
	Data       json.RawMessage `json:"data"`        // Payload data
}

// TCPFlags struct for control flags in the TCP header
type TCPFlags struct {
	URG bool `json:"urg"` // Urgent pointer field significant
	ACK bool `json:"ack"` // Acknowledgment field significant
	PSH bool `json:"psh"` // Push function
	RST bool `json:"rst"` // Reset the connection
	SYN bool `json:"syn"` // Synchronize sequence numbers
	FIN bool `json:"fin"` // No more data from sender
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

/** Constructors **/
func NewDHCPMessage(
	op, htype, hlen byte, xid uint32,
	ciaddr, yiaddr, siaddr, giaddr net.IP,
	chaddr string, options map[byte][]byte,
) json.RawMessage {
	msgBytes, _ := json.Marshal(DHCPMessage{
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
	})
	return msgBytes
}

func NewUDPSegment(srcPort, dstPort int, data json.RawMessage) json.RawMessage {
	segBytes, _ := json.Marshal(UDPSegment{
		SrcPort: srcPort,
		DstPort: dstPort,
		Data:    data,
	})
	return segBytes
}

func NewIPv4Packet(srcIP, dstIP, protocolName string, data json.RawMessage) json.RawMessage {
	proto := -1
	switch protocolName {
	case "UDP":
		proto = 17
	case "ICMP":
		proto = 1
	}
	hdrBytes, _ := json.Marshal(PacketHeader{
		Protocol: proto,
		SrcIP:    srcIP,
		DstIP:    dstIP,
	})
	pktBytes, _ := json.Marshal(IPv4Packet{
		Header: hdrBytes,
		Data:   data,
	})
	return pktBytes
}

func NewFrame(srcMAC, dstMAC, protocolName string, data json.RawMessage) json.RawMessage {
	ether := "0x0"
	switch protocolName {
	case "IPv4":
		ether = "0x0800"
	case "ARP":
		ether = "0x0806"
	}
	frameBytes, _ := json.Marshal(Frame{
		SrcMAC:    srcMAC,
		DstMAC:    dstMAC,
		EtherType: ether,
		Data:      data})
	return frameBytes
}

/** Decoders **/
func ParseDHCPMessage(raw json.RawMessage) (DHCPMessage, error) {
	var v DHCPMessage
	if err := json.Unmarshal(raw, &v); err != nil {
		return DHCPMessage{}, err
	}
	return v, nil
}

func ParseDNSMessage(raw json.RawMessage) (DNSMessage, error) {
	var v DNSMessage
	if err := json.Unmarshal(raw, &v); err != nil {
		return DNSMessage{}, err
	}
	return v, nil
}

func ParseUDPSegment(raw json.RawMessage) (UDPSegment, error) {
	var v UDPSegment
	if err := json.Unmarshal(raw, &v); err != nil {
		return UDPSegment{}, err
	}
	return v, nil
}

func ParseTCPSegment(raw json.RawMessage) (TCPSegment, error) {
	var v TCPSegment
	if err := json.Unmarshal(raw, &v); err != nil {
		return TCPSegment{}, err
	}
	return v, nil
}

func ParseIPv4Packet(raw json.RawMessage) (IPv4Packet, error) {
	var v IPv4Packet
	if err := json.Unmarshal(raw, &v); err != nil {
		return IPv4Packet{}, err
	}
	return v, nil
}

func ParseIPv4PacketHeader(raw json.RawMessage) (PacketHeader, error) {
	var v PacketHeader
	if err := json.Unmarshal(raw, &v); err != nil {
		return PacketHeader{}, err
	}
	return v, nil
}

func ParseICMPEchoPacket(raw json.RawMessage) (ICMPEchoPacket, error) {
	var v ICMPEchoPacket
	if err := json.Unmarshal(raw, &v); err != nil {
		return ICMPEchoPacket{}, err
	}
	return v, nil
}

func ParseFrame(raw json.RawMessage) (Frame, error) {
	var v Frame
	if err := json.Unmarshal(raw, &v); err != nil {
		return Frame{}, err
	}
	return v, nil
}

func ParseARPMessage(raw json.RawMessage) (ArpMessage, error) {
	var v ArpMessage
	if err := json.Unmarshal(raw, &v); err != nil {
		return ArpMessage{}, err
	}
	return v, nil
}
