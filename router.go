/*
File:		router.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Router-specific functions
*/

package main

import(
	"fmt"
	"strings"
	"strconv"
)

const BOBCAT_PORTS = 4
const OSIRIS_PORTS = 2

func NewBobcat(hostname string) Router {
	b := Router{}
	b.ID = idgen(8)
	b.Model = "Bobcat 100"
	b.MACAddr = macgen()
	b.Hostname = hostname
	b.DHCPPool = 253
	b.Downports = BOBCAT_PORTS

	return b
}

func NewOsiris(hostname string) Router {
	o := Router{}
	o.ID = idgen(8)
	o.Model = "Osiris 2-I"
	o.MACAddr = macgen()
	o.Hostname = hostname
	o.DHCPPool = 2
	o.Downports = OSIRIS_PORTS

	return o
}

func addRouter() {
	fmt.Println("What model?")
	fmt.Println("Available: Bobcat, Osiris")
	fmt.Print("Model: ")
	scanner.Scan()
	routerModel := scanner.Text()
	routerModel = strings.ToUpper(routerModel)

	fmt.Print("Hostname: ")
	scanner.Scan()
	routerHostname := scanner.Text()

	// input validation

	if routerHostname == "" {
		fmt.Println("Hostname cannot be blank. Please try again")
		return
	}

	if hostname_exists(routerHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	r := Router{}

	if routerModel == "BOBCAT" {
		r = NewBobcat(routerHostname)
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(routerHostname)
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	if snet.Class == "A" {
		r.Gateway = "10.0.0.1"
	} else if snet.Class == "B" {
		r.Gateway = "172.16.0.1"
	} else if snet.Class == "C" {
		r.Gateway = "192.168.0.1"
	}
	addrconstruct := ""

	network_portion := strings.TrimSuffix(r.Gateway, "1")

	r.DHCPTable = make(map[string]string)

	for k := 2; k < (r.DHCPPool + 2); k++ {
		r.DHCPIndex = append(r.DHCPIndex, strconv.Itoa(k))
	}

	for i := 0; i < len(r.DHCPIndex); i++ {
		addrconstruct = network_portion + r.DHCPIndex[i]
		r.DHCPTable[addrconstruct] = ""
	}

	snet.Router = r
}

func delRouter() {
	fmt.Printf("\nAre you sure you want do delete router %s? [Y/n]: ", snet.Router.Hostname)
	scanner.Scan()
	confirmation := scanner.Text()
	confirmation = strings.ToUpper(confirmation)
	if confirmation == "Y" {
		r := Router{}

		r.ID = ""
		r.Model = ""
		r.MACAddr = ""
		r.Hostname = ""
		r.DHCPPool = 0
		r.Downports = 0
		r.Ports = nil

		snet.Router = r
		fmt.Printf("\nRouter deleted\n")
	} else {
		fmt.Printf("\nRouter %s was not deleted.\n", snet.Router.Hostname)
	}
}


func routerforward(frame Frame) {
	srcIP := frame.Data.SrcIP
	dstIP := frame.Data.DstIP
	srcMAC := snet.Router.MACAddr
	dstMAC := arp_request(snet.Router.ID, "router", dstIP)
	linkID := snet.Hosts[getHostIndexFromID(getIDfromMAC(dstMAC))].ID

	s := frame.Data.Data
	p := constructPacket(srcIP, dstIP, s)
	f := constructFrame(p, srcMAC, dstMAC)
	channels[linkID]<-f
}
