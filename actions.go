package main

import(
	"fmt"
	"time"
	"strings"
	//"strconv"
)

func ping(srcIP string, dstIP string, secs int) {
	srcid := ""
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""
	dsthost := ""

	if snet.Router.Gateway == srcIP {
		srchost = snet.Router.Hostname
		srcid = snet.Router.ID
		srcMAC = snet.Router.MACAddr
	}

	if snet.Router.Gateway == dstIP {
		dsthost = snet.Router.Hostname
		dstid = snet.Router.ID
		dstMAC = snet.Router.MACAddr
	}

	if dsthost == "" || srchost == "" {
		for h := range snet.Hosts {
			if snet.Hosts[h].IPAddr == dstIP { // NI
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
	}

	for i := 0; i < secs; i++ {
		fmt.Printf("\nPinging %s from %s (dstid %s)\n", dsthost, srchost, dstid)

		s := constructSegment("ping!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f //NI
		pong := <-internal[srcid]
		if(pong.Data.Data.Data == "pong!") {
			fmt.Println("Received")
		}
		time.Sleep(time.Second)
	}
	actionsync[srcid]<-1
	return
}


func pong(srcIP string, dstIP string) {
	dstid := ""
	srcMAC := ""
	dstMAC := ""
	//srchost := ""
	//dsthost := ""
	for h := range snet.Hosts {
		if snet.Hosts[h].IPAddr == dstIP { //NI
			//dsthost = snet.Hosts[h].Hostname
			dstid = snet.Hosts[h].ID
			dstMAC = snet.Hosts[h].MACAddr
		}

		if snet.Hosts[h].IPAddr == srcIP {
			//srchost = snet.Hosts[h].Hostname
			srcMAC = snet.Hosts[h].MACAddr
		}
	}

		//fmt.Printf("\nPonging from %s\n", dstid)

		s := constructSegment("pong!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)
		channels[dstid]<-f //NI

	return
}

func dhcp_discover(host Host) {
	//get info
	srchost := host.Hostname
	srcIP := host.IPAddr
	srcMAC := host.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := getMACfromID(host.UplinkID)

	fmt.Printf("\n[Host] Host %s is initiating a DHCP Discover broadcast\n", srchost)
	s := constructSegment("DHCPDISCOVER")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	//need to give it to uplink
	channels[host.UplinkID]<-f //NI
	offer := <-internal[srcID]
	if(offer.Data.Data.Data != "") {
		dstid := getIDfromMAC(offer.SrcMAC)

		if offer.Data.Data.Data == "DHCPOFFER NOAVAILABLE" {
			fmt.Println("\n[Host] Failed to obtain IP address: No free addresses available")
		} else {
			word := strings.Fields(offer.Data.Data.Data)
			if(len(word) > 0){
				word2 := word[1]
				gateway := word[2]
				snetmask := word[3]
				fmt.Printf("\n[Host] Requesting IP Address %s\n", word2)

				message := "DHCPREQUEST " + word2
				s = constructSegment(message)
				p = constructPacket(srcIP, dstIP, s)
				f = constructFrame(p, srcMAC, dstMAC)
				channels[dstid]<-f //NI

				//wait for acknowledgement

				ack := <-internal[srcID]
				fmt.Println("\n[Host] I have received an acknowledgement\n")
				if(ack.Data.Data.Data != "") {
					fmt.Printf("\n[Host] My acknowledgement is: %s\n", ack.Data.Data.Data)

					word = strings.Fields(ack.Data.Data.Data)
					if(len(word) > 1) {
						confirmed_addr := word[1]
						dynamic_assign(srcID, confirmed_addr, gateway, snetmask)
					} else {
						fmt.Printf("\n[Host] Error 5: Empty DHCP acknowledgement\n")
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

	fmt.Printf("\n[Router] Address to give: %s\n", addr_to_give)

	message := ""
	if addr_to_give == "" {
		message = "DHCPOFFER NOAVAILABLE"
	} else {
		message = "DHCPOFFER " + addr_to_give + " " + gateway + " " + subnetmask
	}
	s := constructSegment(message)
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[dstid]<-f //NI

	//acknowledge
	request := <-internal[snet.Router.ID]
	message = ""
	if(request.Data.Data.Data != "") {
		fmt.Printf("\n[Router] My request is: %s\n", request.Data.Data.Data)

		word := strings.Fields(request.Data.Data.Data)
		if(len(word) > 1) {
			if(word[1] == addr_to_give) {
				message = "DHCPACKNOWLEDGEMENT " + addr_to_give
			} else {
				fmt.Println("\n[Router] Error 4: DHCP address requested is not same as offer")
			}
		} else {
			fmt.Printf("\n[Router] Error 3: Empty DHCP request\n")
		}
	}

	s = constructSegment(message)
	p = constructPacket(srcIP, dstIP, s)
	f = constructFrame(p, srcMAC, dstMAC)
	channels[dstid]<-f //NI

	//REMOVE FROM POOL
	network_portion := strings.TrimSuffix(snet.Router.Gateway, "1")
	for i := 0; i < len(snet.Router.DHCPIndex); i++ {
		//fmt.Printf("\n[Router] Debugging sum: %s\n", network_portion + snet.Router.DHCPIndex[i])
		if (network_portion + snet.Router.DHCPIndex[i]) == addr_to_give {
			fmt.Printf("\n[Router] Removing address %s from pool\n", addr_to_give)
			snet.Router.DHCPTable[snet.Router.DHCPIndex[i]] = getMACfromID(dstid)
		}
	}
}
