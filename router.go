package main

import(
	//"fmt"
	//"strings"
	//"strconv"
)

func routerforward(frame Frame) {
	srcIP := frame.Data.SrcIP
	dstIP := frame.Data.DstIP
	srcMAC := snet.Router.MACAddr
	dstMAC := arp_request(snet.Router.ID, "router", dstIP)

	s := frame.Data.Data
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[dstMAC]<-f
}
