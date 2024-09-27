/*
File:		listener.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Listener for network and all devices
*/

package main

var channels = map[string]chan Frame{} // the physical link
var internal = map[string]chan Frame{} // for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	channels["FFFFFFFF"] = make(chan Frame)
	go broadcastlisten()

	generateRouterChannels()

	for i := range snet.Switches {
		generateSwitchChannels(i)
	}

	for i := range snet.Hosts {
		generateHostChannels(i)
	}
}

func generateRouterChannels() {
	if snet.Router.Hostname != "" {
		channels[snet.Router.ID] = make(chan Frame)
		internal[snet.Router.ID] = make(chan Frame)
		go routerlisten()

		//vswitch
		//channels[snet.Router.VSwitch.ID] = make(chan Frame)
		//internal[snet.Router.VSwitch.ID] = make(chan Frame)
		for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
			channels[snet.Router.VSwitch.PortIDs[i]] = make(chan Frame)
			internal[snet.Router.VSwitch.PortIDs[i]] = make(chan Frame)
			actionsync[snet.Router.ID] = make(chan int)

			go switchportlisten(snet.Router.VSwitch.PortIDs[i])
		}
	}
}

func generateSwitchChannels(i int) {
	for j := 0; j < getActivePorts(snet.Switches[j]); j++ {
		channels[snet.Switches[i].PortIDs[j]] = make(chan Frame)
		internal[snet.Switches[i].PortIDs[j]] = make(chan Frame)
		actionsync[snet.Switches[i].PortIDs[j]] = make(chan int)

		go switchportlisten(snet.Switches[i].PortIDs[j])
	}
}

func generateHostChannels(i int) {
	channels[snet.Hosts[i].ID] = make(chan Frame)
	internal[snet.Hosts[i].ID] = make(chan Frame)
	actionsync[snet.Hosts[i].ID] = make(chan int)
	
	go hostlisten(snet.Hosts[i].ID)
}

func broadcastlisten() { //Listens for broadcast frames on FF.. and broadcasts
	for {
		frame := <-channels["FFFFFFFF"]
		debug(4, "broadcastlisten", "Listener", "detected broadcast")
		go routeractionhandler(frame)

		for i := range snet.Hosts {
			go hostactionhandler(frame, snet.Hosts[i].ID)
		}
	}
}

func hostlisten(id string) {
	listenSync <- id //synchronizing with client.go

	for {
		frame := <-channels[id]
		debug(4, "hostlisten", id, "Received frame")
		go hostactionhandler(frame, id)
	}
}

func hostactionhandler(frame Frame, id string) {
	debug(4, "hostactionhandler", id, "My packet")
	data := frame.Data.Data.Data
	if data == "ping!" {
		debug(4, "routeractionhandler", snet.Router.ID, "ping received")
		srcid := id
		dstIP := frame.Data.SrcIP
		pong(srcid, dstIP, frame)
	}

	if data == "pong!" {
		internal[id] <- frame
	}

	if len(data) > 7 {
		if data[0:8] == "ARPREPLY" {
			debug(3, "hostactionhandler", id, "ARPREPLY received")
			internal[id] <- frame
			print("AAAA")
		}

	}

	if len(data) > 8 {
		if data[0:9] == "DHCPOFFER" {
			debug(2, "hostactionhandler", id, "DHCPOFFER received")
			internal[id] <- frame
		}
	}

	if len(data) > 9 {
		if data[0:10] == "ARPREQUEST" {
			debug(4, "hostactionhandler", id, "ARPREQUEST received")
			arp_reply(getHostIndexFromID(id), "host", frame)
		}
	}

	if len(data) > 17 {
		if data[0:19] == "DHCPACKNOWLEDGEMENT" {
			debug(2, "hostactionhandler", id, "DHCPACKNOWLEDGEMENT received")
			internal[id] <- frame
		}
	}
}

func switchportlisten(id string) {
	for {
		frame := <-channels[id]
		debug(4, "switchlisten", id, "Received frame")

		port := getSwitchportIDFromLink(id)

		checkMACTable(frame.SrcMAC, id, port)

		go switchportactionhandler(frame, id)
	}
}

func switchportactionhandler(frame Frame, id string) {
	debug(4, "switchportactionhandler", id, "My packet")
	if frame.DstMAC == "FF:FF:FF:FF:FF:FF" {
		channels["FFFFFFFF"] <- frame
	} else if 1 == 2 { //TODO how to receive mgmt frames
		//data := frame.Data.Data.Data
	} else {
		switchforward(frame, id)
	}
}

func routerlisten() {
	for {
		frame := <-channels[snet.Router.ID]
		debug(4, "routerlisten", snet.Router.ID, "Frame received")
		go routeractionhandler(frame)
	}
}

func routeractionhandler(frame Frame) {
	if (frame.Data.DstIP == snet.Router.Gateway.String()) || (frame.DstMAC == "FF:FF:FF:FF:FF:FF") {
		debug(4, "routeractionhandler", snet.Router.ID, "My packet")
		data := frame.Data.Data.Data
		srcid := snet.Router.ID
		dstIP := frame.Data.SrcIP
		if data == "ping!" {
			debug(3, "routeractionhandler", snet.Router.ID, "ping received")
			pong(srcid, dstIP, frame)
		}
		if data == "pong!" {
			debug(3, "routeractionhandler", snet.Router.ID, "pong received")
			internal[snet.Router.ID] <- frame
		}

		if len(data) > 7 {
			if data[0:8] == "ARPREPLY" {
				debug(3, "routeractionhandler", snet.Router.ID, "ARPREPLY received")
				internal[snet.Router.ID] <- frame
			}
		}

		if len(data) > 9 {
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

		if len(data) > 10 {
			if data[0:11] == "DHCPREQUEST" {
				debug(2, "routeractionhandler", snet.Router.ID, "DHCPREQUEST received")
				internal[snet.Router.ID] <- frame
			}
		}

	} else {
		debug(1, "routeractionhandler", snet.Router.ID, "Error: Still expecting router to forward? (not vswitch)")
		inspectFrame(frame)
	}
}
