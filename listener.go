package main

import(
	"fmt"
)

var channels = map[string]chan Frame{}
var internal = map[string]chan Frame{} //for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	channels["FF:FF:FF:FF:FF:FF"] = make(chan Frame)
	go broadcastlisten()

	for i := range snet.Hosts {
		channels[snet.Hosts[i].MACAddr] = make(chan Frame)
		internal[snet.Hosts[i].MACAddr] = make(chan Frame)
		actionsync[snet.Hosts[i].MACAddr] = make(chan int)
		go hostlisten(i)
	}
	if snet.Router.Hostname != "" {
		channels[snet.Router.MACAddr] = make(chan Frame)
		internal[snet.Router.MACAddr] = make(chan Frame)
		go routerlisten()
	}
}

func broadcastlisten() { //Listens for broadcast frames on FF:.. and broadcasts
	for true {
		frame := <-channels["FF:FF:FF:FF:FF:FF"]

		for i := range snet.Hosts {
			go hostactionhandler(frame, i)
		}
	}
}

func hostlisten(index int) {
	mac := snet.Hosts[index].MACAddr

	listenSync<-index //synchronizing with client.go

	for true {
		frame := <-channels[mac]
		go hostactionhandler(frame, index)
	}
}

func hostactionhandler(frame Frame, index int) {
	data := frame.Data.Data.Data
	if data == "ping!" {
		srcid := snet.Hosts[index].ID
		dstIP := frame.Data.SrcIP
		pong(srcid, dstIP, frame)
	}

	if data == "pong!" {
		internal[snet.Hosts[index].MACAddr]<-frame
	}

	if(len(data) > 7){
		if data[0:8] == "ARPREPLY" {
			//fmt.Printf("[Host %s] ARPREPLY received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].MACAddr]<-frame
		}

	}

	if(len(data) > 8){
		if data[0:9] == "DHCPOFFER" {
			fmt.Printf("\n[Host %s] DHCPOFFER received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].MACAddr]<-frame
		}
	}

	if(len(data) > 9){
		if data[0:10] == "ARPREQUEST" {
			//fmt.Printf("\n[Host %s] ARPREQUEST received\n", snet.Hosts[index].Hostname)
			arp_reply(index, "host", frame)
		}
	}

	if(len(data) > 17){
		if data[0:19] == "DHCPACKNOWLEDGEMENT" {
			fmt.Printf("\n[Host %s] DHCPACKNOWLEDGEMENT received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].MACAddr]<-frame
		}
	}
}

func routerlisten() {
	for true {
		frame := <-channels[snet.Router.MACAddr]
		go routeractionhandler(frame)
	}
}

func routeractionhandler(frame Frame) {
	if((frame.Data.DstIP == snet.Router.Gateway) || (frame.DstMAC == "FF:FF:FF:FF:FF:FF")) {
		//fmt.Println("\n[Router] My packet") // debug
		data := frame.Data.Data.Data
		srcid := snet.Router.ID
		dstIP := frame.Data.SrcIP

		if data == "ping!" {
			pong(srcid, dstIP, frame)
		}

		if data == "pong!" {
			internal[snet.Router.MACAddr]<-frame
		}

		if(len(data) > 7){
			if data[0:8] == "ARPREPLY" {
				//fmt.Printf("[Host %s] ARPREPLY received\n", snet.Hosts[index].Hostname)
				internal[snet.Router.MACAddr]<-frame
			}

		}

		if(len(data) > 9){
			if data[0:10] == "ARPREQUEST" {
				fmt.Printf("\n[Router] ARPREQUEST received\n")
				index := 0
				arp_reply(index, "router", frame)
			}
		}


		if data == "DHCPDISCOVER" {
			fmt.Println("\n[Router] DHCPDISCOVER received")
			dhcp_offer(frame)
		}

		if(len(data) > 9){
			if data[0:11] == "DHCPREQUEST" {
				fmt.Println("\n[Router] DHCPREQUEST received")
				internal[snet.Router.MACAddr]<-frame
			}
		}

	} else {
		routerforward(frame)
	}
}
