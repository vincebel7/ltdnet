/*
File:		router.go
Author: 	https://github.com/vincebel7
Purpose:	Router-specific functions
*/

package main

import (
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	"github.com/vincebel7/ltdnet/iphelper"
)

type Router struct {
	ID         string               `json:"id"`
	Model      string               `json:"model"`
	Hostname   string               `json:"hostname"`
	VSwitch    Switch               `json:"vswitchid"` // Virtual built-in switch to router
	DHCPPool   DHCPPool             `json:"dhcp_pool"` // Instance of DHCPPool
	ARPTable   map[string]ARPEntry  `json:"arptable"`
	DNSTable   map[string]DNSRecord `json:"dnstable"`  // Local DNS table
	DNSServer  DNSServer            `json:"dnsserver"` // DNS server hosted on the router (optional)
	Interfaces map[string]Interface `json:"interfaces"`
}

type DHCPPool struct {
	DHCPPoolStart  net.IP            `json:"dhcp_pool_start"`  // Starting IP address of DHCP pool
	DHCPPoolEnd    net.IP            `json:"dhcp_pool_end"`    // Ending IP address of DHCP pool
	DHCPPoolLeases map[string]string `json:"dhcp_pool_leases"` // Maps IP address to MAC address
}

const BOBCAT_PORTS = 4
const OSIRIS_PORTS = 2

func NewDHCPPool(start_addr net.IP, end_addr net.IP) DHCPPool {
	pool := DHCPPool{}
	pool.DHCPPoolStart = start_addr
	pool.DHCPPoolEnd = end_addr
	pool.DHCPPoolLeases = make(map[string]string)

	return pool
}

func NewBobcat(r Router) Router {
	r.Model = "Bobcat 100"

	vSwitch := addVirtualSwitch(BOBCAT_PORTS)
	r.VSwitch = vSwitch

	return r
}

func NewOsiris(r Router) Router {
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

	if snet.Router.Hostname != "" {
		fmt.Printf("Network already has a router, %s.\n", snet.Router.Hostname)
		return
	}

	r := Router{}

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
	if snet.Netsize == "8" {
		gateway = net.ParseIP("10.0.0.1")
	} else if snet.Netsize == "16" {
		gateway = net.ParseIP("172.16.0.1")
	} else if snet.Netsize == "24" {
		gateway = net.ParseIP("192.168.0.1")
	}

	r.ID = idgen(8)
	r.Hostname = routerHostname
	r.ARPTable = make(map[string]ARPEntry)

	netsizeInt, _ := strconv.Atoi(snet.Netsize)

	// Interfaces
	r.Interfaces = make(map[string]Interface)

	loopbackIPConfig := IPConfig{
		IPAddress:      net.ParseIP("127.0.0.1"),
		SubnetMask:     "255.0.0.0",
		DefaultGateway: nil,
		DNSServer:      net.ParseIP("127.0.0.1"),
		ConfigType:     "static",
	}
	eth0IPConfig := IPConfig{
		IPAddress:  gateway,
		SubnetMask: prefixLengthToSubnetMask(netsizeInt),
		DNSServer:  gateway,
		ConfigType: "",
	}

	r.Interfaces["lo"] = Interface{
		Name:     "lo",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: loopbackIPConfig,
	}
	r.Interfaces["eth0"] = Interface{
		Name:     "eth0",
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: eth0IPConfig,
	}

	// DNS table
	r.DNSTable = make(map[string]DNSRecord)

	r.DNSTable[r.Hostname] = DNSRecord{
		Name:  r.Hostname,
		Type:  'A',
		Class: 0,
		TTL:   65535,
		RData: "127.0.0.1",
	}
	r.DNSTable["localhost"] = DNSRecord{
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

	snet.Router = r

	assignSwitchport(snet.Router.VSwitch, snet.Router.Interfaces["eth0"].L1ID)

	iface := snet.Router.Interfaces["eth0"]
	iface.RemoteL1ID = snet.Router.VSwitch.PortLinksLocal[0]
	snet.Router.Interfaces["eth0"] = iface

	generateRouterChannels()
	go listenRouterChannel("lo")
	go listenRouterChannel("eth0")

	for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
		go listenSwitchportChannel(snet.Router.VSwitch.ID, snet.Router.VSwitch.PortLinksLocal[i])
	}
	achievementTester(ROUTINE_BUSINESS)
}

func delRouter() {
	r := Router{}

	r.ID = ""
	r.Model = ""
	r.Interfaces["eth0"] = Interface{}
	r.Hostname = ""
	r.DHCPPool = NewDHCPPool(net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.0"))
	r.VSwitch = addVirtualSwitch(0)

	snet.Router = r
	fmt.Printf("\nRouter deleted\n")
}

func (router Router) NextFreePoolAddress() net.IP {
	poolAddrs := router.GetDHCPPoolAddresses()
	for i := range poolAddrs {
		current_addr := poolAddrs[i]
		if router.DHCPPool.DHCPPoolLeases[current_addr.String()] == "" {
			return current_addr
		}
	}

	return nil
}

func (router Router) GetDHCPPoolAddresses() []net.IP {
	pool := router.DHCPPool

	// Create IP Helper
	startIP, _ := iphelper.NewIPHelper(pool.DHCPPoolStart)
	endIP, _ := iphelper.NewIPHelper(pool.DHCPPoolEnd)

	// Convert to BigInt for arithmetic
	startIPInt := startIP.IPToBigInt()
	endIPInt := endIP.IPToBigInt()

	var poolAddrs []net.IP
	for i := new(big.Int).Set(startIPInt); i.Cmp(endIPInt) <= 0; i.Add(i, big.NewInt(1)) {
		poolAddrs = append(poolAddrs, iphelper.BigIntToIP(i)) // Convert back to string IP
	}

	return poolAddrs
}

func (router Router) IsAvailableAddress(testAddr net.IP) bool {
	poolAddrs := router.GetDHCPPoolAddresses()
	for i := range poolAddrs {
		currentAddr := poolAddrs[i]
		if currentAddr.Equal(testAddr) {
			if router.DHCPPool.DHCPPoolLeases[currentAddr.String()] == "" {
				return true
			} else {
				return false
			}
		}
	}

	return false
}

func (router Router) routeToInterface(dstIP string) Interface {
	for iface := range router.Interfaces {
		devIP := router.GetIP(iface)
		devMask := router.GetMask(iface)

		fmt.Printf("testing to see if %s is in same subnet as %s", devIP, dstIP)
		if iphelper.IPInSameSubnet(devIP, dstIP, devMask) {
			fmt.Printf("Routing %s to interface %s\n", dstIP, router.Interfaces[iface].Name)
			return router.Interfaces[iface]
		}
	}

	// Default gateway
	fmt.Println("Routing to interface %s", "eth0")

	return router.Interfaces["eth0"]
}
