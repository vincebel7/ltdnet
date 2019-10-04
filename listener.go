/*
File:		listener.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Listener for network and all devices
*/

package main

import(
)

var channels = map[string]chan Frame{}
var internal = map[string]chan Frame{} //for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	channels["FFFFFFFF"] = make(chan Frame)
	go broadcastlisten()

	for i := range snet.Hosts {
		generateHostChannels(i)
	}

	if snet.Router.Hostname != "" {
		channels[snet.Router.ID] = make(chan Frame)
		internal[snet.Router.ID] = make(chan Frame)
		go routerlisten()
	}
}

func broadcastlisten() { //Listens for broadcast frames on FF.. and broadcasts
	for true {
		frame := <-channels["FFFFFFFF"]
		//fmt.Println("[Listener] Detected broadcast")
		go routeractionhandler(frame)

		for i := range snet.Hosts {
			go hostactionhandler(frame, i)
		}
	}
}

func generateHostChannels(i int) {
	channels[snet.Hosts[i].ID] = make(chan Frame)
	internal[snet.Hosts[i].ID] = make(chan Frame)
	actionsync[snet.Hosts[i].ID] = make(chan int)
	go hostlisten(i)
}

func hostlisten(index int) {
	id := snet.Hosts[index].ID

	listenSync<-index //synchronizing with client.go

	for true {
		frame := <-channels[id]
		//fmt.Println("[Debug] Received frame - Host ", snet.Hosts[index].Hostname)
		go hostactionhandler(frame, index)
	}
}

func hostactionhandler(frame Frame, index int) {
	//debug(3, "hostactionhandler", snet.Hosts[index].ID, "My packet")
	data := frame.Data.Data.Data
	if data == "ping!" {
		srcid := snet.Hosts[index].ID
		dstIP := frame.Data.SrcIP
		pong(srcid, dstIP, frame)
	}

	if data == "pong!" {
		internal[snet.Hosts[index].ID]<-frame
	}

	if(len(data) > 7){
		if data[0:8] == "ARPREPLY" {
			debug(3, "hostactionhandler", snet.Hosts[index].ID, "ARPREPLY received")
			internal[snet.Hosts[index].ID]<-frame
		}

	}

	if(len(data) > 8){
		if data[0:9] == "DHCPOFFER" {
			debug(2, "hostactionhandler", snet.Hosts[index].ID, "DHCPOFFER received")
			internal[snet.Hosts[index].ID]<-frame
		}
	}

	if(len(data) > 9){
		if data[0:10] == "ARPREQUEST" {
			debug(3, "hostactionhandler", snet.Hosts[index].ID, "ARPREQUEST received")
			arp_reply(index, "host", frame)
		}
	}

	if(len(data) > 17){
		if data[0:19] == "DHCPACKNOWLEDGEMENT" {
			debug(2, "hostactionhandler", snet.Hosts[index].ID, "DHCPACKNOWLEDGEMENT received")
			internal[snet.Hosts[index].ID]<-frame
		}
	}
}

func routerlisten() {
	for true {
		frame := <-channels[snet.Router.ID]
		go routeractionhandler(frame)
	}
}

func routeractionhandler(frame Frame) {
	if((frame.Data.DstIP == snet.Router.Gateway) || (frame.DstMAC == "FF:FF:FF:FF:FF:FF")) {
		//fmt.Println("\n[Debug] [Router] My packet") // debug
		data := frame.Data.Data.Data
		srcid := snet.Router.ID
		dstIP := frame.Data.SrcIP
		//fmt.Printf("\n[Debug] [Router] Data: %s DstIP: %s", data, frame.Data.DstIP)
		if data == "ping!" {
			pong(srcid, dstIP, frame)
		}

		if data == "pong!" {
			internal[snet.Router.ID]<-frame
		}

		if(len(data) > 7){
			if data[0:8] == "ARPREPLY" {
				debug(3, "routeractionhandler", snet.Router.ID, "ARPREPLY received")
				internal[snet.Router.ID]<-frame
			}

		}

		if(len(data) > 9){
			if data[0:10] == "ARPREQUEST" {
				debug(3, "routeractionhandler", snet.Router.ID, "ARPREQUEST received")
				index := 0
				arp_reply(index, "router", frame)
			}
		}


		if data == "DHCPDISCOVER" {
			debug(2, "routeractionhandler", snet.Router.ID, "DHCPDISCOVER received")
			dhcp_offer(frame)
		}

		if(len(data) > 10){
			if data[0:11] == "DHCPREQUEST" {
				debug(2, "routeractionhandler", snet.Router.ID, "DHCPREQUEST received")
				internal[snet.Router.ID]<-frame
			}
		}

	} else {
		debug(3, "routeractionhandler", snet.Router.ID, "Not my packet, will forward")
		routerforward(frame)
	}
}
