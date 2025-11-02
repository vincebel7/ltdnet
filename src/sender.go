/*
File:		sender.go
Author: 	https://github.com/vincebel7
Purpose:	Handles the sending of frames
*/

package main

import (
	"encoding/json"

	"github.com/vincebel7/ltdnet/src/engine"
	"github.com/vincebel7/ltdnet/src/model"
)

func sendFrame(frameBytes json.RawMessage, iface model.Interface, srcID string) {
	frame, _ := model.ParseFrame(frameBytes)
	if isToSelf(frame) {
		deviceLog(5, "sendFrame", srcID, "Frame destination is to itself. Mirroring back across the interface.")

		mirrorLinkID := iface.L1ID
		engine.Instance().Channels[mirrorLinkID] <- frameBytes

	} else {
		engine.Instance().Channels[iface.RemoteL1ID] <- frameBytes
	}
}

func isToSelf(frame model.Frame) bool {
	// L2 (Reminder: ARPREQUEST is broadcast, not mirrored)
	if frame.SrcMAC == frame.DstMAC {
		return true
	}

	// L3 (optional)
	if frame.EtherType == "0x0800" { // IPv4
		packet, _ := model.ParseIPv4Packet(frame.Data)
		packetHeader, _ := model.ParseIPv4PacketHeader(packet.Header)
		return packetHeader.SrcIP == packetHeader.DstIP
	}

	return false
}
