/*
File:		sender.go
Author: 	https://github.com/vincebel7
Purpose:	Handles the sending of frames
*/

package main

import (
	"encoding/json"
)

func sendFrame(frameBytes json.RawMessage, linkID string, srcID string) {
	if isToSelf(readFrame(frameBytes)) {
		debug(4, "sendFrame", srcID, "Frame destination is to itself. Mirroring back across the interface.")

		mirrorLinkID := ""

		if srcID == snet.Router.ID {
			mirrorLinkID = snet.Router.Interface.L1ID
		} else {
			index := getHostIndexFromID(srcID)
			mirrorLinkID = snet.Hosts[index].Interface.L1ID
		}

		channels[mirrorLinkID] <- frameBytes

	} else {
		channels[linkID] <- frameBytes
	}
}

func isToSelf(frame Frame) bool {
	// L2 (Reminder: ARPREQUEST is broadcast, not mirrored)
	if frame.SrcMAC == frame.DstMAC {
		return true
	}

	// L3 (optional)
	if frame.EtherType == "0x0800" { // IPv4
		packet := readIPv4Packet(frame.Data)
		packetHeader := readIPv4PacketHeader(packet.Header)
		return packetHeader.SrcIP == packetHeader.DstIP
	}

	return false
}
