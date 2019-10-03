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
	Data		Segment//layers 4+ abstracted
}

type Frame struct {
	SrcMAC		string
	DstMAC		string
	Data		Packet
}

