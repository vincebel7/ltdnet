/*
File:		router.go
Author: 	https://bitbucket.org/vincebel
Purpose:	Router-specific functions
*/

package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/vincebel7/ltdnet/iphelper"
)

type Router struct {
	ID           string   `json:"id"`
	Model        string   `json:"model"`
	MACAddr      string   `json:"macaddr"` // LAN-facing interface
	Hostname     string   `json:"hostname"`
	Gateway      string   `json:"gateway"`
	DHCPPoolSize int      `json:"dhcp_pool_size"` //total addresses in DHCP pool
	VSwitch      Switch   `json:"vswitchid"`      // Virtual built-in switch to router
	DHCPPool     DHCPPool `json:"dhcp_pool"`      // Instance of DHCPPool
}

type DHCPPool struct {
	DHCPPoolStart  string            `json:"dhcp_pool_start"`  // Starting IP address of DHCP pool
	DHCPPoolEnd    string            `json:"dhcp_pool_end"`    // Ending IP address of DHCP pool
	DHCPPoolLeases map[string]string `json:"dhcp_pool_leases"` // Maps IP address to MAC address
}

const BOBCAT_PORTS = 4
const OSIRIS_PORTS = 2

func NewDHCPPool(start_addr string, end_addr string) DHCPPool {
	pool := DHCPPool{}
	pool.DHCPPoolStart = start_addr
	pool.DHCPPoolEnd = end_addr
	pool.DHCPPoolLeases = make(map[string]string)

	return pool
}

func NewBobcat(hostname string) Router {
	b := Router{}
	b.ID = idgen(8)
	b.Model = "Bobcat 100"
	b.MACAddr = macgen()
	b.Hostname = hostname
	b.DHCPPoolSize = 253

	v := addVirtualSwitch(BOBCAT_PORTS)

	b.VSwitch = v
	return b
}

func NewOsiris(hostname string) Router {
	o := Router{}
	o.ID = idgen(8)
	o.Model = "Osiris 2-I"
	o.MACAddr = macgen()
	o.Hostname = hostname
	o.DHCPPoolSize = 2

	v := addVirtualSwitch(OSIRIS_PORTS)

	o.VSwitch = v
	return o
}

func addVirtualSwitch(maxports int) Switch {
	v := Switch{}
	v.ID = idgen(8)
	v.Model = "virtual"
	v.Hostname = "V-" + v.ID
	v.Maxports = maxports

	v.PortIDs = make([]string, v.Maxports)
	for i := range v.PortIDs {
		v.PortIDs[i] = idgen(8)
	}

	v.Ports = make([]string, v.Maxports)
	for i := range v.Ports {
		v.Ports[i] = ""
	}

	v.MACTable = make(map[string]int)

	return v
}

func addRouter(routerHostname string, routerModel string) {
	routerModel = strings.ToUpper(routerModel)

	// input validation
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

	if snet.Netsize == "8" {
		r.Gateway = "10.0.0.1"
	} else if snet.Netsize == "16" {
		r.Gateway = "172.16.0.1"
	} else if snet.Netsize == "24" {
		r.Gateway = "192.168.0.1"
	}

	network_portion := strings.TrimSuffix(r.Gateway, "1")

	// DHCP (new)
	start_ip := network_portion + "2"
	end_iph, _ := iphelper.NewIPHelper(start_ip)
	end_ip := end_iph.IncreaseIPByConstant(r.DHCPPoolSize)
	r.DHCPPool = NewDHCPPool(start_ip, end_ip.String())

	snet.Router = r

	generateRouterChannels()

	assignSwitchport(snet.Router.VSwitch, snet.Router.ID)
}

func delRouter(hostname string) {
	r := Router{}

	r.ID = ""
	r.Model = ""
	r.MACAddr = ""
	r.Hostname = ""
	r.DHCPPoolSize = 0
	r.DHCPPool = NewDHCPPool("0.0.0.0", "0.0.0.0")
	r.VSwitch = addVirtualSwitch(0)

	snet.Router = r
	fmt.Printf("\nRouter deleted\n")
}

func next_free_addr() string {
	pool := snet.Router.DHCPPool.GetPoolAddresses()
	for i := range pool {
		current_addr := pool[i]
		if snet.Router.DHCPPool.DHCPPoolLeases[current_addr] == "" {
			return current_addr
		}
	}

	return ""
}

func (p *DHCPPool) GetPoolAddresses() []string {
	startIP, _ := iphelper.NewIPHelper(p.DHCPPoolStart)
	endIP, _ := iphelper.NewIPHelper(p.DHCPPoolEnd)

	startIPInt := startIP.IPToBigInt()
	endIPInt := endIP.IPToBigInt()

	var pool []string
	for i := new(big.Int).Set(startIPInt); i.Cmp(endIPInt) <= 0; i.Add(i, big.NewInt(1)) {
		pool = append(pool, iphelper.BigIntToIP(i)) // Convert back to string IP
	}
	return pool
}
