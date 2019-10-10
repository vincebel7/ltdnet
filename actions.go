/*
File:		actions.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Defines network functions such as ARP, DHCP, etc.
*/

package main

import(
	"fmt"
	"time"
	"strings"
	"net"
)

func ping(srcid string, dstIP string, secs int) {
	debug(4, "ping", srcid, "About to ping")

	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""

	if snet.Router.ID == srcid { //leave this in here until i implement controlling router and can ping from rtr
		srchost = snet.Router.Hostname
		srcIP = snet.Router.Gateway
		srcMAC = snet.Router.MACAddr

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcid, "router", dstIP)
		fmt.Println("Got dstmac", dstMAC) //leave this in here until implemented

		//get link to send ping to
		linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
	}

	if dstMAC == "" || srcMAC == "" {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcid {
				linkID = snet.Hosts[h].UplinkID
				srchost = snet.Hosts[h].Hostname
				srcIP = snet.Hosts[h].IPAddr
				srcMAC = snet.Hosts[h].MACAddr

				//TODO: Implement MAC learning to avoid ARPing every time
				dstMAC = arp_request(srcid, "host", dstIP)
			}
		}
	}
	fmt.Printf("\nPinging %s from %s\n", dstIP, srchost)
	timeoutCounter := 0
	sendCount := 0
	recvCount := 0
	lossCount := 0
	for i := 0; i < secs; i++ {
		s := constructSegment("ping!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)

		channels[linkID]<-f
		sendCount++
		pong := make(chan bool, 1)
		select {
			case pongdata := <-internal[srcid]:
				if(pongdata.Data.Data.Data == "pong!") {
					recvCount++
					pong<-true
				}
			case <-time.After(time.Second * 4):
				lossCount++
				pong<-false
		}

		if(<-pong) {
			fmt.Printf("Reply from %s\n", dstIP)
			timeoutCounter = 0
		} else {
			fmt.Printf("Request timed out.\n")
			timeoutCounter++
			i--
		}

		if(timeoutCounter == 4) { //Skip rest of pings if timeout
			i = secs
		}
		
		if(i < secs - 1) { //Only wait a second if 
			time.Sleep(time.Second)
		}
	}
	actionsync[srcid]<-1

	// Ping stats
	fmt.Printf("\nPing statistics for %s:\n", dstIP)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))


	return
}

func pong(srcid string, dstIP string, frame Frame) {
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := frame.SrcMAC
	if snet.Router.ID == srcid {
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcid, "router", dstIP)

		//get link to send ping to
		linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
	} else {
		index := getHostIndexFromID(srcid)
		srcMAC = snet.Hosts[index].MACAddr
		srcIP = snet.Hosts[index].IPAddr

		linkID = snet.Hosts[index].UplinkID
	}

	s := constructSegment("pong!")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
	return
}

func arp_request(srcid string, device_type string, dstIP string) string {
	debug(4, "arp_request", srcid, "About to ARP request")
	//Construct frame
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := "FF:FF:FF:FF:FF:FF"

	if(device_type == "router") {
		srcIP = snet.Router.Gateway
		srcMAC = snet.Router.MACAddr
		linkID = "FFFFFFFF"
	} else {
		index := getHostIndexFromID(srcid)
		srcIP = snet.Hosts[index].IPAddr
		srcMAC = snet.Hosts[index].MACAddr
		linkID = "FFFFFFFF"
	}

	s := constructSegment("ARPREQUEST")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	channels[linkID]<-f
	debug(4, "arp_request", srcid, "Sent ARP request")
	//computer with address will respond with its MAC
	replyframe := <-internal[srcid]
	return replyframe.Data.Data.Data[9:]
}

func arp_reply(i int, device_type string, frame Frame) {
	request_addr := frame.Data.DstIP
	linkID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := frame.SrcMAC
	dstIP := frame.Data.SrcIP

	if (device_type == "router") {
		if (request_addr != snet.Router.Gateway) {
			return
		} else {
			//fmt.Printf("[Router] THIS ME! I am %s\n", snet.Router.Hostname, request_addr)
			srcIP = snet.Router.Gateway
			srcMAC = snet.Router.MACAddr

			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		}
	} else {
		if (request_addr != snet.Hosts[i].IPAddr) {
			return
		} else {
			//fmt.Printf("[Host %s] THIS ME! I am %s\n", snet.Hosts[i].Hostname, request_addr)
			srcIP = snet.Hosts[i].IPAddr
			srcMAC = snet.Hosts[i].MACAddr

			linkID = snet.Hosts[i].UplinkID
		}
	}

	message := "ARPREPLY " + srcMAC
	s := constructSegment(message)
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
}

func dhcp_discover(host Host) {
	//get info
	srcIP := host.IPAddr
	srcMAC := host.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := "FF:FF:FF:FF:FF:FF"
	linkID := host.UplinkID

	s := constructSegment("DHCPDISCOVER")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	//need to give it to uplink
	channels[linkID]<-f
	debug(2, "dhcp_discover", host.ID, "DHCPDISCOVER sent")
	offer := <-internal[srcID]
	if(offer.Data.Data.Data != "") {
		if offer.Data.Data.Data == "DHCPOFFER NOAVAILABLE" {
			debug(1, "dhcp_discover", srcID, "Failed to obtain IP address: No free addresses available")
		} else {
			word := strings.Fields(offer.Data.Data.Data)
			if(len(word) > 0){
				word2 := word[1]
				gateway := word[2]
				snetmask := word[3]

				message := "DHCPREQUEST " + word2
				dstIP = offer.Data.SrcIP
				s = constructSegment(message)
				p = constructPacket(srcIP, dstIP, s)
				f = constructFrame(p, srcMAC, dstMAC)
				channels[linkID]<-f
				debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent - " + word2)
				//wait for acknowledgement

				ack := <-internal[srcID]
				if(ack.Data.Data.Data != "") {
					debug(2, "dhcp_discover", srcID, "DHCPACKNOWLEDGEMENT received - " + ack.Data.Data.Data)

					word = strings.Fields(ack.Data.Data.Data)
					if(len(word) > 1) {
						fmt.Println("")
						confirmed_addr := word[1]
						dynamic_assign(srcID, confirmed_addr, gateway, snetmask)
					} else {
						debug(1, "dhcp_discover", srcID, "Error 5: Empty DHCP acknowledgement\n")
					}
				}


			} else {
				debug(1, "dhcp_discover", srcID, "Error 2: Empty DHCP offer")
			}
		}
	}
	actionsync[srcID]<-1
}

func dhcp_offer(inc_f Frame){
	srcIP := snet.Router.Gateway
	dstIP := "255.255.255.255"
	srcMAC := snet.Router.MACAddr
	dstMAC := inc_f.SrcMAC
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid
	//linkID := snet.Hosts[getHostIndexFromID(dstid)].UplinkID

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


	message := ""
	if addr_to_give == "" {
		message = "DHCPOFFER NOAVAILABLE"
	} else {
		message = "DHCPOFFER " + addr_to_give + " " + gateway + " " + subnetmask
	}
	s := constructSegment(message)
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - " + addr_to_give)

	//acknowledge
	request := <-internal[snet.Router.ID]
	message = ""
	if(request.Data.Data.Data != "") {
		word := strings.Fields(request.Data.Data.Data)
		if(len(word) > 1) {
			if(word[1] == addr_to_give) {
				message = "DHCPACKNOWLEDGEMENT " + addr_to_give
			} else {
				debug(1, "dhcp_offer", snet.Router.ID, "Error 4: DHCP address requested is not same as offer")
			}
		} else {
			debug(1, "dhcp_offer", snet.Router.ID, "Error 3: Empty DHCP request")
		}
	}

	dstIP = request.Data.SrcIP
	s = constructSegment(message)
	p = constructPacket(srcIP, dstIP, s)
	f = constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f

	// Setting leasee's MAC in pool
	network_portion := strings.TrimSuffix(snet.Router.Gateway, "1")
	for i := 0; i < len(snet.Router.DHCPIndex); i++ {
		//fmt.Printf("\n[Router] Debugging sum: %s\n", network_portion + snet.Router.DHCPIndex[i])
		if (network_portion + snet.Router.DHCPIndex[i]) == addr_to_give {
			debug(2, "dhcp_offer", snet.Router.ID, "Removing address " + addr_to_give + " from pool")
			snet.Router.DHCPTable[snet.Router.DHCPIndex[i]] = getMACfromID(dstid) //NI TODO have client pass their MAC in DHCPREQUEST instead of relying on this NI
		}
	}
}

func ipset(hostname string) {
	fmt.Printf(" IP configuration for %s\n", hostname)

	correct := false
	var ipaddr, subnetmask, defaultgateway string
	for !correct {
		fmt.Print("IP Address: ")
		scanner.Scan()
		ipaddr = scanner.Text()

		fmt.Print("\nSubnet mask: ")
		scanner.Scan()
		subnetmask = scanner.Text()

		fmt.Print("\nDefault gateway: ")
		scanner.Scan()
		defaultgateway = scanner.Text()

		fmt.Printf("\nIP Address: %s\nSubnet mask: %s\nDefault gateway: %s\n", ipaddr, subnetmask, defaultgateway)
		fmt.Print("\nIs this correct? [Y/n/exit]")
		scanner.Scan()
		affirmation := scanner.Text()

		if(strings.ToUpper(affirmation) == "Y") {
			// error checking
			error := false
			if net.ParseIP(ipaddr).To4() == nil {
				error = true
				fmt.Printf("Error: '%s' is not a valid IP address\n", ipaddr)
			}
			if net.ParseIP(subnetmask).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid subnet mask\n", subnetmask)
			}
			if net.ParseIP(defaultgateway).To4() == nil {
				fmt.Printf("Error: '%s' is not a valid default gateway\n", defaultgateway)
			}

			if(!error) {
				correct = true
			}
		 } else if(strings.ToUpper(affirmation) == "EXIT") {
			fmt.Println("Network changes reverted")
			return
		 }
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			snet.Hosts[h].IPAddr = ipaddr
			snet.Hosts[h].SubnetMask = subnetmask
			snet.Hosts[h].DefaultGateway = defaultgateway
			fmt.Println("Network configuration updated")
		}
	}

}

func ipclear(id string) {
	index := getHostIndexFromID(id)
	snet.Hosts[index].IPAddr = ""
	snet.Hosts[index].SubnetMask = ""
	snet.Hosts[index].DefaultGateway = ""
	fmt.Println("Network configuration cleared")
}
