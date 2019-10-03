package main

import(
	"fmt"
)

var channels = map[string]chan Frame{}
var internal = map[string]chan Frame{} //for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	channels["FFFFFFFF"] = make(chan Frame)
	go broadcastlisten()

	for i := range snet.Hosts {
		//create channel
		channels[snet.Hosts[i].ID] = make(chan Frame)
		internal[snet.Hosts[i].ID] = make(chan Frame)
		actionsync[snet.Hosts[i].ID] = make(chan int)
		go hostlisten(i)
	}
	if snet.Router.Hostname != "" {
		channels[snet.Router.ID] = make(chan Frame)
		internal[snet.Router.ID] = make(chan Frame)
		go routerlisten()
	}
}

func broadcastlisten() {
	for true {
		frame := <-channels["FFFFFFFF"]

		for i := range snet.Hosts {
			go hostactionhandler(frame, i)
		}
	}
}

func hostlisten(index int) {
	id := snet.Hosts[index].ID
	//hostname := snet.Hosts[index].Hostname

	//fmt.Printf("\n%s listening", snet.Hosts[index].Hostname)
	listenSync<-index //synchronizing with client.go

	for true {
		frame := <-channels[id]
		//fmt.Printf("%s just got: %s\n", hostname, frame.Data.Data.Data)
		go hostactionhandler(frame, index)
	}
}

func hostactionhandler(frame Frame, index int) {
	data := frame.Data.Data.Data
	if data == "ping!" {
		srcid := snet.Hosts[index].ID
		dstIP := frame.Data.SrcIP
		pong(srcid, dstIP)
	}

	if data == "pong!" {
		internal[snet.Hosts[index].ID]<-frame
	}

	if(len(data) > 7){
		if data[0:8] == "ARPREPLY" {
			//fmt.Printf("[Host %s] ARPREPLY received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].ID]<-frame
		}

	}

	if(len(data) > 8){
		if data[0:9] == "DHCPOFFER" {
			fmt.Printf("\n[Host %s] DHCPOFFER received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].ID]<-frame
		}
	}

	if(len(data) > 8){
		if data[0:10] == "ARPREQUEST" {
			//fmt.Printf("\n[Host %s] ARPREQUEST received\n", snet.Hosts[index].Hostname)
			arp_reply(index, frame)
		}
	}

	if(len(data) > 17){
		if data[0:19] == "DHCPACKNOWLEDGEMENT" {
			fmt.Printf("\n[Host %s] DHCPACKNOWLEDGEMENT received\n", snet.Hosts[index].Hostname)
			internal[snet.Hosts[index].ID]<-frame
		}
	}
}

func routerlisten() {
	//fmt.Println("router listening")
	for true {
		frame := <-channels[snet.Router.ID]
		go routeractionhandler(frame)
	}
}

func routeractionhandler(frame Frame) {
	if((frame.Data.DstIP == snet.Router.Gateway) || (frame.Data.DstIP == "255.255.255.255")) {
		//fmt.Println("\n[Router] My packet") // debug
		data := frame.Data.Data.Data
		srcid := snet.Router.Gateway
		dstIP := frame.Data.SrcIP

		if data == "ping!" {
			pong(srcid, dstIP)
		}

		if data == "pong!" {
			internal[snet.Router.ID]<-frame
		}

		if data == "DHCPDISCOVER" {
			fmt.Println("\n[Router] DHCPDISCOVER received")
			dhcp_offer(frame)
		}

		if(len(data) > 9){
			if data[0:11] == "DHCPREQUEST" {
				fmt.Println("\n[Router] DHCPREQUEST received")
				internal[snet.Router.ID]<-frame
			}
		}

	} else {
		fmt.Println("\n[Router] Not my packet, I will forward")
	}
}
