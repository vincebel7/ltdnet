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
	ID        string              `json:"id"`
	Model     string              `json:"model"`
	Hostname  string              `json:"hostname"`
	VSwitch   Switch              `json:"vswitchid"` // Virtual built-in switch to router
	DHCPPool  DHCPPool            `json:"dhcp_pool"` // Instance of DHCPPool
	ARPTable  map[string]ARPEntry `json:"arptable"`
	Interface Interface           `json:"interface"`
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

func NewBobcat() Router {
	bobcat := Router{}
	bobcat.Model = "Bobcat 100"

	vSwitch := addVirtualSwitch(BOBCAT_PORTS)
	bobcat.VSwitch = vSwitch

	return bobcat
}

func NewOsiris() Router {
	osiris := Router{}
	osiris.Model = "Osiris 2-I"

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
		r = NewBobcat()
		dhcpPoolSize = 253
	} else if routerModel == "OSIRIS" {
		r = NewOsiris()
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
	ipConfig := IPConfig{
		IPAddress:  gateway,
		SubnetMask: prefixLengthToSubnetMask(netsizeInt),
		DNSServer:  nil,
		ConfigType: "",
	}

	// Blank interface, no remote L1ID or IP configuration yet
	r.Interface = Interface{
		L1ID:     idgen(8),
		MACAddr:  macgen(),
		IPConfig: ipConfig,
	}

	network_portion := strings.TrimSuffix(r.GetIP(), "1")

	// Create DHCP Pool
	start_ip := net.ParseIP(network_portion + "2")
	end_iph, _ := iphelper.NewIPHelper(start_ip)
	end_ip := end_iph.IncreaseIPByConstant(dhcpPoolSize)
	r.DHCPPool = NewDHCPPool(start_ip, end_ip)

	snet.Router = r

	assignSwitchport(snet.Router.VSwitch, snet.Router.Interface.L1ID)

	snet.Router.Interface.RemoteL1ID = snet.Router.VSwitch.PortLinksLocal[0]

	generateRouterChannels()
	go listenRouterChannel()
	for i := 0; i < getActivePorts(snet.Router.VSwitch); i++ {
		go listenSwitchportChannel(snet.Router.VSwitch.ID, snet.Router.VSwitch.PortLinksLocal[i])
	}
	achievementTester(ROUTINE_BUSINESS)
}

func delRouter() {
	r := Router{}

	r.ID = ""
	r.Model = ""
	r.Interface = Interface{}
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
