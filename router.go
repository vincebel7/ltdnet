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
	"strings"

	"github.com/vincebel7/ltdnet/iphelper"
)

type Router struct {
	ID       string            `json:"id"`
	Model    string            `json:"model"`
	MACAddr  string            `json:"macaddr"` // LAN-facing interface
	Hostname string            `json:"hostname"`
	Gateway  net.IP            `json:"gateway"`
	VSwitch  Switch            `json:"vswitchid"` // Virtual built-in switch to router
	DHCPPool DHCPPool          `json:"dhcp_pool"` // Instance of DHCPPool
	ARPTable map[string]string `json:"arptable"`
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

func NewBobcat(hostname string) Router {
	bobcat := Router{}
	bobcat.ID = idgen(8)
	bobcat.Model = "Bobcat 100"
	bobcat.MACAddr = macgen()
	bobcat.Hostname = hostname
	bobcat.ARPTable = make(map[string]string)

	vSwitch := addVirtualSwitch(BOBCAT_PORTS)
	bobcat.VSwitch = vSwitch

	return bobcat
}

func NewOsiris(hostname string) Router {
	osiris := Router{}
	osiris.ID = idgen(8)
	osiris.Model = "Osiris 2-I"
	osiris.MACAddr = macgen()
	osiris.Hostname = hostname
	osiris.ARPTable = make(map[string]string)

	vSwitch := addVirtualSwitch(OSIRIS_PORTS)
	osiris.VSwitch = vSwitch

	return osiris
}

func addRouter(routerHostname string, routerModel string) {
	routerModel = strings.ToUpper(routerModel)

	// input validation
	if hostname_exists(routerHostname) {
		fmt.Println("Hostname already exists. Please try again")
		return
	}

	r := Router{}

	dhcpPoolSize := 0

	if routerModel == "BOBCAT" {
		r = NewBobcat(routerHostname)
		dhcpPoolSize = 253
	} else if routerModel == "OSIRIS" {
		r = NewOsiris(routerHostname)
		dhcpPoolSize = 2
	} else {
		fmt.Println("Invalid model. Please try again")
		return
	}

	if snet.Netsize == "8" {
		r.Gateway = net.ParseIP("10.0.0.1")
	} else if snet.Netsize == "16" {
		r.Gateway = net.ParseIP("172.16.0.1")
	} else if snet.Netsize == "24" {
		r.Gateway = net.ParseIP("192.168.0.1")
	}

	network_portion := strings.TrimSuffix(r.Gateway.String(), "1")

	// Create DHCP Pool
	start_ip := net.ParseIP(network_portion + "2")
	end_iph, _ := iphelper.NewIPHelper(start_ip)
	end_ip := end_iph.IncreaseIPByConstant(dhcpPoolSize)
	r.DHCPPool = NewDHCPPool(start_ip, end_ip)

	snet.Router = r

	assignSwitchport(snet.Router.VSwitch, snet.Router.ID)

	generateRouterChannels()
	go listenRouterChannel()
	for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
		go listenSwitchportChannel(snet.Router.VSwitch.PortIDs[i])
	}
	achievementTester(ROUTINE_BUSINESS)
}

func delRouter() {
	r := Router{}

	r.ID = ""
	r.Model = ""
	r.MACAddr = ""
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
