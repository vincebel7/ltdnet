package main

import(
	"fmt"
	//"encoding/json"
	"os"
	//"log"
	//"math/rand"
	"time"
	"bufio"
	"strings"
	"strconv"
	//"path/filepath"
	//"io/ioutil"
)

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

func Conn(device string, id string) {
	//find host
	host := Host{}
	for i := range snet.Hosts {
		if(snet.Hosts[i].ID == id){
			host = snet.Hosts[i]
		}
	}
	if host.ID == "" {
		fmt.Println("Error: ID cannot be located. Please try again")
	}

	//interface
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		fmt.Printf("%s> ", host.Hostname)
		scanner.Scan()
		action_selection := scanner.Text()
		if action_selection != "" {
		action := strings.Fields(action_selection)
		actionword1 := action[0]

		switch actionword1 {
			case "ping":
				if len(action) > 1 {
					if len(action) > 2 { //if seconds is specified
						seconds, _ := strconv.Atoi(action[2])
						go ping(host.IPAddr, action[1], seconds)
					} else {
						go ping(host.IPAddr, action[1], 1)
					}
				}
			case "help":
				fmt.Println("",
				"ping <dest_hostname> [seconds]\t\tPings host\n")

			default:
				fmt.Println(" Invalid command. Type 'help' for a list of commands.")
		}
		}
	}

}

func constructSegment(data string) Segment {
	srcport := 7
	dstport := 7
	protocol := "UDP"

	s := Segment{
		Protocol: protocol,
		SrcPort: srcport,
		DstPort: dstport,
		Data: data,
	}

	return s
}

func constructPacket(srcIP string, dstIP string, data Segment) Packet {
	p := Packet{
		SrcIP: srcIP,
		DstIP: dstIP,
		Data: data,
	}

	return p
}

func constructFrame(data Packet, srcMAC string, dstMAC string) Frame {
	f := Frame{
		SrcMAC: srcMAC,
		DstMAC: dstMAC,
		Data: data,
	}

	return f
}

func ping(srcIP string, dstIP string, secs int) {
	srcid := ""
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""
	dsthost := ""
	for h := range snet.Hosts {
		if snet.Hosts[h].IPAddr == dstIP { // network-independent
			dsthost = snet.Hosts[h].Hostname
			dstid = snet.Hosts[h].ID
			dstMAC = snet.Hosts[h].MACAddr
		}

		if snet.Hosts[h].IPAddr == srcIP {
			srchost = snet.Hosts[h].Hostname
			srcid = snet.Hosts[h].ID
			srcMAC = snet.Hosts[h].MACAddr
		}
	}

	for i := 0; i < secs; i++ {
		fmt.Printf("\nPinging %s from %s (dstid %s)\n", dsthost, srchost, dstid)

		s := constructSegment("ping!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f
		pong := <-internal[srcid]
		if(pong.Data.Data.Data == "pong!") {
			fmt.Println("ok we got a pong back")
		}
		time.Sleep(time.Second) //replace with block
	}
	//check if host is found
	return
}


func pong(srcIP string, dstIP string) {
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	//srchost := ""
	//dsthost := ""
	for h := range snet.Hosts {
		if snet.Hosts[h].IPAddr == dstIP { //network-independent
			//dsthost = snet.Hosts[h].Hostname
			dstid = snet.Hosts[h].ID
			dstMAC = snet.Hosts[h].MACAddr
		}

		if snet.Hosts[h].IPAddr == srcIP {
			//srchost = snet.Hosts[h].Hostname
			srcMAC = snet.Hosts[h].MACAddr
		}
	}

		//fmt.Printf("\nPonging %s from %s (dstid %s)\n", dsthost, srchost, dstid)

		s := constructSegment("pong!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f

	//check if host is found
	return
}
