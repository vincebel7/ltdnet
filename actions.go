/*
File:		actions.go
Author: 	https://github.com/vincebel7
Purpose:	Defines network functions such as ARP, DHCP, etc.
*/

package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func ping(srcID string, dstIP string, secs int) {
	debug(4, "ping", srcID, "About to ping")

	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := ""
	srchost := ""

	// Get srchost
	if snet.Router.ID == srcID {
		srchost = snet.Router.Hostname
	} else {
		for h := range snet.Hosts {
			if snet.Hosts[h].ID == srcID {
				srchost = snet.Hosts[h].Hostname
			}
		}
	}

	fmt.Printf("\nPinging %s from %s\n", dstIP, srchost)
	timeoutCounter := 0
	sendCount := 0
	recvCount := 0
	lossCount := 0

	for i := 0; i < secs; i++ {
		// Get MAC addresses
		if snet.Router.ID == srcID {
			srchost = snet.Router.Hostname
			srcIP = snet.Router.Gateway.String()
			srcMAC = snet.Router.MACAddr

			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, "router", dstIP)

			//get link to send ping to
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}

		} else { // Assumed to be host source.
			//TODO: Implement MAC learning to avoid ARPing every time
			dstMAC = arp_request(srcID, "host", dstIP)
		}

		if srcMAC == "" {
			for h := range snet.Hosts {
				if snet.Hosts[h].ID == srcID {
					linkID = snet.Hosts[h].UplinkID
					srchost = snet.Hosts[h].Hostname
					srcIP = snet.Hosts[h].IPAddr.String()
					srcMAC = snet.Hosts[h].MACAddr
				}
			}
		}

		s := constructSegment("ping!")
		p := constructPacket(srcIP, dstIP, s)
		f := constructFrame(p, srcMAC, dstMAC)

		channels[linkID] <- f
		sendCount++
		debug(2, "ping", srcID, "Ping sent")
		pong := make(chan bool, 1)
		select {
		case pongdata := <-internal[srcID]:
			if pongdata.Data.Data.Data == "pong!" {
				recvCount++
				pong <- true
			} else {
				debug(1, "ping", srcID, "Out-of-order channel error")
			}
		case <-time.After(time.Second * 4):
			lossCount++
			pong <- false
		}

		if <-pong {
			fmt.Printf("Reply from %s: seq=%d\n", dstIP, i)
			timeoutCounter = 0
		} else {
			fmt.Printf("Request timed out.\n")
			timeoutCounter++
			i--
		}

		if timeoutCounter == 4 { //Skip rest of pings if timeout
			i = secs
		}

		if i < secs-1 { //Only wait a second if
			time.Sleep(time.Second)
		}
	}
	actionsync[srcID] <- 1

	// Ping stats
	fmt.Printf("\nPing statistics for %s:\n", dstIP)
	fmt.Printf("\tPackets: Sent = %d, Received = %d, Lost = %d (%d%% loss)\n\n", sendCount, recvCount, lossCount, (lossCount / sendCount * 100))
}

func pong(srcID string, dstIP string, frame Frame) {
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := frame.SrcMAC
	if snet.Router.ID == srcID {
		srcMAC = snet.Router.MACAddr
		srcIP = snet.Router.Gateway.String()

		//TODO: Implement MAC learning to avoid ARPing every time
		dstMAC = arp_request(srcID, "router", dstIP)
		debug(4, "pong", srcID, "ARP completed. Dstmac acquired. dstMAC: "+dstMAC)
		//get link to send ping to

		//TODO get link to send ping to
		dstID := getIDfromMAC(dstMAC)
		if getHostIndexFromID(dstID) != -1 {
			//TODO check if host or router!
			linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
		} else if snet.Router.ID == dstID {
			linkID = snet.Router.ID
		}
		//END TODO
	} else {
		index := getHostIndexFromID(srcID)
		srcMAC = snet.Hosts[index].MACAddr
		srcIP = snet.Hosts[index].IPAddr.String()

		linkID = snet.Hosts[index].UplinkID
	}

	s := constructSegment("pong!")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	debug(4, "pong", srcID, "Awaiting pong send")
	channels[linkID] <- f
	debug(2, "pong", srcID, "Pong sent")
}

func arp_request(srcID string, device_type string, dstIP string) string {
	debug(4, "arp_request", srcID, "About to ARP request")
	//Construct frame
	linkID := ""
	srcIP := ""
	srcMAC := ""
	dstMAC := "FF:FF:FF:FF:FF:FF"

	if device_type == "router" {
		srcIP = snet.Router.Gateway.String()
		srcMAC = snet.Router.MACAddr
		linkID = "FFFFFFFF"
	} else {
		index := getHostIndexFromID(srcID)
		srcIP = snet.Hosts[index].IPAddr.String()
		srcMAC = snet.Hosts[index].MACAddr
		linkID = "FFFFFFFF"
	}

	s := constructSegment("ARPREQUEST")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	channels[linkID] <- f
	debug(2, "arp_request", srcID, "ARPREQUEST sent")
	//computer with address will respond with its MAC
	replyframe := <-internal[srcID]

	return replyframe.Data.Data.Data[9:]
}

func arp_reply(i int, device_type string, frame Frame) {
	request_addr := frame.Data.DstIP
	linkID := ""
	srcMAC := ""
	srcIP := ""
	dstMAC := frame.SrcMAC
	dstIP := frame.Data.SrcIP
	srcID := ""

	if device_type == "router" {
		if request_addr != snet.Router.Gateway.String() {
			return
		} else {
			//fmt.Printf("[Router] THIS ME! I am %s\n", snet.Router.Hostname, request_addr)
			srcIP = snet.Router.Gateway.String()
			srcMAC = snet.Router.MACAddr
			srcID = snet.Router.ID

			// Determine linkID
			dstID := getIDfromMAC(dstMAC)
			if getHostIndexFromID(dstID) != -1 {
				linkID = snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID
			} else if snet.Router.ID == dstID {
				linkID = snet.Router.ID
			}
		}
	} else {
		if request_addr != snet.Hosts[i].IPAddr.String() {
			return
		} else {
			//fmt.Printf("[Host %s] THIS ME! I am %s\n", snet.Hosts[i].Hostname, request_addr)
			srcIP = snet.Hosts[i].IPAddr.String()
			srcMAC = snet.Hosts[i].MACAddr
			srcID = snet.Hosts[i].ID
			linkID = snet.Hosts[i].UplinkID
		}
	}

	message := "ARPREPLY " + srcMAC
	s := constructSegment(message)
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	//inspectFrame(f)
	debug(2, "arp_reply", srcID, "ARPREPLY sent")
	channels[linkID] <- f
}

func dhcp_discover(host Host) {
	debug(4, "dhcp_discover", host.ID, "Starting DHCPDISCOVER")
	//get info
	srcIP := host.IPAddr.String()
	srcMAC := host.MACAddr
	srcID := host.ID
	dstIP := "255.255.255.255"
	dstMAC := "FF:FF:FF:FF:FF:FF"
	linkID := host.UplinkID

	s := constructSegment("DHCPDISCOVER")
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)

	//need to give it to uplink
	channels[linkID] <- f
	debug(2, "dhcp_discover", host.ID, "DHCPDISCOVER sent")
	offer := <-internal[srcID]
	if offer.Data.Data.Data != "" {
		if offer.Data.Data.Data == "DHCPOFFER NOAVAILABLE" {
			debug(1, "dhcp_discover", srcID, "Failed to obtain IP address: No free addresses available")
		} else {
			word := strings.Fields(offer.Data.Data.Data)
			if len(word) > 0 {
				word2 := word[1]
				gateway := word[2]
				snetmask := word[3]

				message := "DHCPREQUEST " + word2
				dstIP = offer.Data.SrcIP
				s = constructSegment(message)
				p = constructPacket(srcIP, dstIP, s)
				f = constructFrame(p, srcMAC, dstMAC)
				channels[linkID] <- f
				debug(2, "dhcp_discover", srcID, "DHCPREQUEST sent - "+word2)
				//wait for acknowledgement

				ack := <-internal[srcID]
				if ack.Data.Data.Data != "" {
					debug(2, "dhcp_discover", srcID, "DHCPACKNOWLEDGEMENT received - "+ack.Data.Data.Data)

					word = strings.Fields(ack.Data.Data.Data)
					if len(word) > 1 {
						fmt.Println("")
						confirmed_addr := word[1]
						dynamic_assign(srcID, net.ParseIP(confirmed_addr), net.ParseIP(gateway), snetmask)
					} else {
						debug(1, "dhcp_discover", srcID, "Error 5: Empty DHCP acknowledgement\n")
					}
				}

			} else {
				debug(1, "dhcp_discover", srcID, "Error 2: Empty DHCP offer")
			}
		}
	}
	actionsync[srcID] <- 1
}

func dhcp_offer(inc_f Frame) {
	srcIP := snet.Router.Gateway.String()
	dstIP := "255.255.255.255"
	srcMAC := snet.Router.MACAddr
	dstMAC := inc_f.SrcMAC
	dstid := getIDfromMAC(dstMAC)
	linkID := dstid
	//linkID := snet.Hosts[getHostIndexFromID(dstid)].UplinkID

	//find open address
	addr_to_give := snet.Router.NextFreePoolAddress().String()
	gateway := snet.Router.Gateway.String()
	subnetmask := ""
	if snet.Netsize == "8" {
		subnetmask = "255.0.0.0"
	} else if snet.Netsize == "16" {
		subnetmask = "255.255.0.0"
	} else if snet.Netsize == "24" {
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
	channels[linkID] <- f
	debug(2, "dhcp_offer", snet.Router.ID, "DHCPOFFER sent - "+addr_to_give)

	//acknowledge
	request := <-internal[snet.Router.ID]
	message = ""
	if request.Data.Data.Data != "" {
		word := strings.Fields(request.Data.Data.Data)
		if len(word) > 1 {
			if word[1] == addr_to_give {
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
	channels[linkID] <- f

	// Setting leasee's MAC in pool (new)
	pool := snet.Router.GetDHCPPoolAddresses()
	for k := range pool {
		if pool[k].String() == addr_to_give {
			debug(2, "dhcp_offer", snet.Router.ID, "Assigning and removing address "+addr_to_give+" from pool")
			snet.Router.DHCPPool.DHCPPoolLeases[addr_to_give] = getMACfromID(dstid) //NI TODO have client pass their MAC in DHCPREQUEST instead of relying on this NI
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

		if strings.ToUpper(affirmation) == "Y" {
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

			if !error {
				correct = true
			}
		} else if strings.ToUpper(affirmation) == "EXIT" {
			fmt.Println("Network changes reverted")

			return
		}
	}

	//update info
	for h := range snet.Hosts {
		if snet.Hosts[h].Hostname == hostname {
			snet.Hosts[h].IPAddr = net.ParseIP(ipaddr)
			snet.Hosts[h].SubnetMask = subnetmask
			snet.Hosts[h].DefaultGateway = net.ParseIP(defaultgateway)
			fmt.Println("Network configuration updated")
		}
	}

}

func ipclear(id string) {
	index := getHostIndexFromID(id)
	snet.Hosts[index].IPAddr = nil
	snet.Hosts[index].SubnetMask = ""
	snet.Hosts[index].DefaultGateway = nil
	fmt.Println("Network configuration cleared")
}
