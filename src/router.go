/*
File:		router.go
Author: 	https://github.com/vincebel7
Purpose:	Router-specific functions
*/

package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/vincebel7/ltdnet/iphelper"
	"github.com/vincebel7/ltdnet/src/model"
)

const BOBCAT_PORTS = 4
const OSIRIS_PORTS = 2

func NewDHCPPool(start_addr net.IP, end_addr net.IP) model.DHCPPool {
	pool := model.DHCPPool{}
	pool.DHCPPoolStart = start_addr
	pool.DHCPPoolEnd = end_addr
	pool.DHCPPoolLeases = make(map[string]string)

	return pool
}

func NewBobcat(r model.Router) model.Router {
	r.Model = "Bobcat 100"

	vSwitch := addVirtualSwitch(BOBCAT_PORTS)
	r.VSwitch = vSwitch

	return r
}

func NewOsiris(r model.Router) model.Router {
	r.Model = "Osiris 2-I"

	vSwitch := addVirtualSwitch(OSIRIS_PORTS)
	r.VSwitch = vSwitch

	return r
}

func addRouter(routerHostname string, routerModel string) {
	routerModel = strings.ToUpper(routerModel)

	// input validation
	if hostname_exists(routerHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	r_check := Net().Router
	if r_check != nil {
		fmt.Printf("Network already has a router, %s.\n", Net().Router.Hostname)
		return
	}

	r := model.Router{}

	dhcpPoolSize := 0

	if routerModel == "BOBCAT" {
		r = NewBobcat(r)
		dhcpPoolSize = 253
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(r)
		dhcpPoolSize = 2
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	var gateway net.IP
	if Net().Netsize == "8" {
		gateway = net.ParseIP("10.0.0.1")
	} else if Net().Netsize == "16" {
		gateway = net.ParseIP("172.16.0.1")
	} else if Net().Netsize == "24" {
		gateway = net.ParseIP("192.168.0.1")
	}

	r.ID = idgen(8)
	r.Hostname = routerHostname
	r.ARPTable = make(map[string]model.ARPEntry)

	netsizeInt, _ := strconv.Atoi(Net().Netsize)

	// Interfaces
	r.Interfaces = make(map[string]model.Interface)

	loopbackIPConfig := model.IPConfig{
		IPAddress:      net.ParseIP("127.0.0.1"),
		SubnetMask:     "255.0.0.0",
		DefaultGateway: nil,
		DNSServer:      net.ParseIP("127.0.0.1"),
		ConfigType:     "static",
	}
	eth0IPConfig := model.IPConfig{
		IPAddress:  gateway,
		SubnetMask: prefixLengthToSubnetMask(netsizeInt),
		DNSServer:  gateway,
		ConfigType: "",
	}

	r.Interfaces["lo"] = model.Interface{
		Name:     "lo",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: loopbackIPConfig,
	}
	r.Interfaces["eth0"] = model.Interface{
		Name:     "eth0",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: eth0IPConfig,
	}

	// DNS table
	r.DNSTable = make(map[string]model.DNSRecord)

	r.DNSTable[r.Hostname] = model.DNSRecord{
		Name:  r.Hostname,
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}
	r.DNSTable["localhost"] = model.DNSRecord{
		Name:  "localhost",
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}

	network_portion := strings.TrimSuffix(r.GetIP("eth0"), "1")

	// Create DHCP Pool
	start_ip := net.ParseIP(network_portion + "2")
	end_iph, _ := iphelper.NewIPHelper(start_ip)
	end_ip := end_iph.IncreaseIPByConstant(dhcpPoolSize)
	r.DHCPPool = NewDHCPPool(start_ip, end_ip)

	Net().Router = &r

	assignSwitchport(Net().Router.VSwitch, Net().Router.Interfaces["eth0"].L1ID)

	iface := Net().Router.Interfaces["eth0"]
	iface.RemoteL1ID = Net().Router.VSwitch.PortLinksLocal[0]
	Net().Router.Interfaces["eth0"] = iface

	generateRouterChannels()
	go listenRouterChannel("lo")
	go listenRouterChannel("eth0")

	for i := 0; i < getActivePorts(Net().Router.VSwitch); i++ {
		go listenSwitchportChannel(Net().Router.VSwitch.ID, Net().Router.VSwitch.PortLinksLocal[i])
	}
	achievementTester(ROUTINE_BUSINESS)
}

func delRouter() {
	r := model.Router{}

	r.ID = ""
	r.Model = ""
	r.Interfaces["eth0"] = model.Interface{}
	r.Hostname = ""
	r.DHCPPool = NewDHCPPool(net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.0"))
	r.VSwitch = addVirtualSwitch(0)

	Net().Router = &r
	systemLog(2, "delRouter", "Router deleted")
}
