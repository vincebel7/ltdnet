package main

import(
	"fmt"
	//"time"
	"strings"
	"strconv"
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
	hostindex := -1
	for i := range snet.Hosts {
		if(snet.Hosts[i].ID == id){
			hostindex = i
			host = snet.Hosts[i]
		}
	}
	if host.ID == "" {
		fmt.Println("Error: ID cannot be located. Please try again")
	}

	//interface
	fmt.Printf("\n")
	action_selection := ""
	for strings.ToUpper(action_selection) != "EXIT" {
		host = snet.Hosts[hostindex]

		fmt.Printf("%s> ", host.Hostname)
		scanner.Scan()
		action_selection := scanner.Text()
		actionword1 := ""
		if action_selection != "" {
		action := strings.Fields(action_selection)
		if(len(action) > 0){
			actionword1 = action[0]
		}

		switch actionword1 {
			case "":
			case "ping":
				if(host.UplinkID == "") {
					fmt.Println("Device is not connected. Please set an uplink")
				} else if(host.IPAddr == "0.0.0.0") {
					fmt.Println("Device does not have IP configuration. Please use DHCP or statically assign an IP configuration")
				}else {
					if len(action) > 1 {
						if len(action) > 2 { //if seconds is specified
							seconds, _ := strconv.Atoi(action[2])
							go ping(host.ID, action[1], seconds)
						} else {
							go ping(host.ID, action[1], 1)
						}
						<-actionsync[id]
					}
				}
			case "dhcp":
				if(host.UplinkID == "") {
					fmt.Println("Device is not connected. Please set an uplink")
				} else {

					go dhcp_discover(host)
					<-actionsync[id]
					save()
				}
			case "exit":
				return
			case "help":
				fmt.Println("",
				"ping <dest_ip> [seconds]\tPings IP\n",
				"dhcp\t\t\t\tGets address via DHCP")
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
