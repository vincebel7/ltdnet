package main

import(
	"fmt"
)

var channels = map[string]chan Frame{}
var internal = map[string]chan Frame{} //for internal device communication
var actionsync = map[string]chan int{}

func Listener() {
	for i := range snet.Hosts {
		//create channel
		channels[snet.Hosts[i].ID] = make(chan Frame)
		internal[snet.Hosts[i].ID] = make(chan Frame)
		actionsync[snet.Hosts[i].ID] = make(chan int)
		go hostlisten(i)
	}
	if snet.Router.Hostname != "" {
		go routerlisten()
	}
}

func hostlisten(index int) {
	//declarations to make things easier
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
	srcIP := snet.Hosts[index].IPAddr
	dstIP := frame.Data.SrcIP
	if data == "ping!" {
		//fmt.Printf("(%s) Time to respond to this ping\n", srcIP)
		pong(srcIP, dstIP)
	}

	if data == "pong!" {
		internal[snet.Hosts[index].ID]<-frame
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
	if(frame.Data.DstIP != snet.Router.Gateway) { //packet not for me
		fmt.Println("This packet is not for me. I will forward")
	} else {
		fmt.Println("my packet")
	}
}
