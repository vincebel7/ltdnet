package main

import(
	"fmt"
	"time"
	"strings"
	//"strconv"
)

func ping(srcid string, dstIP string, secs int) {
	srcIP := ""
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""

	if snet.Router.ID == srcid {
		srchost = snet.Router.Hostname
		srcIP = snet.Router.Gateway
		srcMAC = snet.Router.MACAddr

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcid, "router", dstIP)
	}

	if snet.Router.Gateway == dstIP { //NI
		dstid = snet.Router.ID
	}

	if dstMAC == "" || srcMAC == "" {
		for h := range snet.Hosts {
			if snet.Hosts[h].IPAddr == dstIP { // NI
				dstid = snet.Hosts[h].ID
			}

			if snet.Hosts[h].ID == srcid {

				srchost = snet.Hosts[h].Hostname
				srcIP = snet.Hosts[h].IPAddr
				srcMAC = snet.Hosts[h].MACAddr

				//dstMAC = getMACfromID(snet.Hosts[h].UplinkID) //eventually save uplink as MAC
				//TODO: Implement MAC learning to avoid ARPing every time
				dstMAC = arp_request(srcid, "host", dstIP)
				//about to leave for the day, but why did i just do this? should be next-hop MAC for hosts.
			}
		}
	}
	fmt.Printf("\nPinging %s from %s (dstid %s)\n", dstIP, srchost, dstid)
	for i := 0; i < secs; i++ {
		s := constructSegment("ping!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f //NI
		pong := <-internal[srcid]
		if(pong.Data.Data.Data == "pong!") {
			fmt.Printf("Reply from %s\n", dstIP)
		}
		time.Sleep(time.Second)
	}
	actionsync[srcid]<-1
	return
}


func pong(srcid string, dstIP string) {
	srcIP := ""
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	if snet.Router.ID == srcid {
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcid, "router", dstIP)
	}

	if snet.Router.Gateway == dstIP { //NI
		dstid = snet.Router.ID
	}

	for h := range snet.Hosts {
		if snet.Hosts[h].IPAddr == dstIP { //NI
			dstid = snet.Hosts[h].ID
		}

		if snet.Hosts[h].ID == srcid {
			srcMAC = snet.Hosts[h].MACAddr
			srcIP = snet.Hosts[h].IPAddr

			//dstMAC = getMACfromID(snet.Hosts[h].UplinkID)
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcid, "host", dstIP)
			//about to leave for the day, but why did i just do this? should be next-hop MAC for hosts.
		}
	}

		s := constructSegment("pong!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f //NI

	return
}
func arp_request(srcid string, device_type string, dstIP string) string {
	//send broadcast ARP request FF, requesting answer for IP address

	//Construct frame
	srcIP := ""
	srcMAC := ""
	dstMAC := "FF:FF:FF:FF:FF:FF"
	dstuplink := "FFFFFFFF" //NI do away with this once relying on MAC address
	if(device_type == "router") {
		srcIP = snet.Router.Gateway
		srcMAC = snet.Router.MACAddr
	} else {
		index := getHostIndexFromID(srcid)
		srcIP = snet.Hosts[index].IPAddr
		srcMAC = snet.Hosts[index].MACAddr
	}

	s := constructSegment("ARPREQUEST")
	p := constructPacket(srcIP, dstIP,s)
	f := constructFrame(p, srcMAC, dstMAC)

	channels[dstuplink]<-f
	//computer with address will respond with its MAC

	replyframe := <-internal[srcid]
	return replyframe.Data.Data.Data[9:]
}

func arp_reply(i int, frame Frame) {
	request_addr := frame.Data.DstIP

	if (request_addr != snet.Hosts[i].IPAddr) {
		return
	} else {
		//fmt.Printf("[Host %s] THIS ME!\n", snet.Hosts[i].Hostname)

		srcIP := snet.Hosts[i].IPAddr
		srcMAC := snet.Hosts[i].MACAddr
		dstMAC := frame.SrcMAC
		dstIP := frame.Data.SrcIP

		//NI get rid of this with MAC switchover
		dstid := ""
		if snet.Router.Gateway == dstIP { //NI
			dstid = snet.Router.ID
		}
		for h := range snet.Hosts {
			if snet.Hosts[h].IPAddr == dstIP { //NI
				dstid = snet.Hosts[h].ID
			}
		}

		message := "ARPREPLY " + srcMAC
		s := constructSegment(message)
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f //NI
	}
}

func dhcp_discover(host Host) {
	//get info
	srchost := host.Hostname
	srcIP := host.IPAddr
	srcMAC := host.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := getMACfromID(host.UplinkID)

	s := constructSegment("DHCPDISCOVER")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	//need to give it to uplink
	channels[host.UplinkID]<-f
	fmt.Printf("[Host %s] DHCPDISCOVER sent\n", srchost)
	offer := <-internal[srcID]
	if(offer.Data.Data.Data != "") {
		if offer.Data.Data.Data == "DHCPOFFER NOAVAILABLE" {
			fmt.Println("[Host %s] Failed to obtain IP address: No free addresses available", srchost)
		} else {
			word := strings.Fields(offer.Data.Data.Data)
			if(len(word) > 0){
				word2 := word[1]
				gateway := word[2]
				snetmask := word[3]
				fmt.Printf("[Host %s] DHCPREQUEST sent - %s\n", srchost, word2)

				message := "DHCPREQUEST " + word2
				dstIP = offer.Data.SrcIP
				s = constructSegment(message)
				p = constructPacket(srcIP, dstIP, s)
				f = constructFrame(p, srcMAC, dstMAC)
				channels[host.UplinkID]<-f

				//wait for acknowledgement

				ack := <-internal[srcID]
				if(ack.Data.Data.Data != "") {
					fmt.Printf("[Host %s] DCHPACKNOWLEDGEMENT received - %s\n", srchost, ack.Data.Data.Data)

					word = strings.Fields(ack.Data.Data.Data)
					if(len(word) > 1) {
						fmt.Println("")
						confirmed_addr := word[1]
						dynamic_assign(srcID, confirmed_addr, gateway, snetmask)
					} else {
						fmt.Printf("[Host %s] Error 5: Empty DHCP acknowledgement\n", srchost)
					}
				}


			} else {
				fmt.Println("Error 2: Empty DHCP offer")
			}
		}
	}

	actionsync[srcID]<-1
}

func dhcp_offer(inc_f Frame){
	srcIP := snet.Router.Gateway
	dstIP := inc_f.Data.SrcIP
	srcMAC := snet.Router.MACAddr
	dstMAC := inc_f.SrcMAC
	dstid := getIDfromMAC(dstMAC)

	//find open address
	addr_to_give := next_free_addr()
	gateway := snet.Router.Gateway
	subnetmask := ""
	if snet.Class == "A" {
		subnetmask = "255.0.0.0"
	} else if snet.Class == "B" {
		subnetmask = "255.255.0.0"
	} else if snet.Class == "C" {
		subnetmask = "255.255.255.0"
	}

	fmt.Printf("[Router] DHCPOFFER sent - %s\n", addr_to_give)

	message := ""
	if addr_to_give == "" {
		message = "DHCPOFFER NOAVAILABLE"
	} else {
		message = "DHCPOFFER " + addr_to_give + " " + gateway + " " + subnetmask
	}
	s := constructSegment(message)
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[dstid]<-f //NI because not using MAC learning (then MAC->ID)

	//acknowledge
	request := <-internal[snet.Router.ID]
	message = ""
	if(request.Data.Data.Data != "") {
		word := strings.Fields(request.Data.Data.Data)
		if(len(word) > 1) {
			if(word[1] == addr_to_give) {
				message = "DHCPACKNOWLEDGEMENT " + addr_to_give
			} else {
				fmt.Println("[Router] Error 4: DHCP address requested is not same as offer")
			}
		} else {
			fmt.Printf("[Router] Error 3: Empty DHCP request\n")
		}
	}

	dstIP = request.Data.SrcIP
	s = constructSegment(message)
	p = constructPacket(srcIP, dstIP, s)
	f = constructFrame(p, srcMAC, dstMAC)
	channels[dstid]<-f ///NI because not using MAC learning (then MAC->ID)

	// Setting leasee's MAC in pool
	network_portion := strings.TrimSuffix(snet.Router.Gateway, "1")
	for i := 0; i < len(snet.Router.DHCPIndex); i++ {
		//fmt.Printf("\n[Router] Debugging sum: %s\n", network_portion + snet.Router.DHCPIndex[i])
		if (network_portion + snet.Router.DHCPIndex[i]) == addr_to_give {
			fmt.Printf("[Router] Removing address %s from pool\n", addr_to_give)
			snet.Router.DHCPTable[snet.Router.DHCPIndex[i]] = getMACfromID(dstid) //NI TODO have client pass their MAC in DHCPREQUEST instead of relying on this NI
		}
	}
}
