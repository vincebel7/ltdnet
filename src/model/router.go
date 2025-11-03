package model

import (
	"math/big"
	"net"

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
	DNSServer  *DNSServer           `json:"dnsserver"` // DNS server hosted on the router (optional)
	Interfaces map[string]Interface `json:"interfaces"`
}

type DHCPPool struct {
	DHCPPoolStart  net.IP            `json:"dhcp_pool_start"`  // Starting IP address of DHCP pool
	DHCPPoolEnd    net.IP            `json:"dhcp_pool_end"`    // Ending IP address of DHCP pool
	DHCPPoolLeases map[string]string `json:"dhcp_pool_leases"` // Maps IP address to MAC address
}

func (r *Router) GetID() string       { return r.ID }
func (r *Router) GetHostname() string { return r.Hostname }

func NewDHCPPool(start_addr net.IP, end_addr net.IP) DHCPPool {
	pool := DHCPPool{}
	pool.DHCPPoolStart = start_addr
	pool.DHCPPoolEnd = end_addr
	pool.DHCPPoolLeases = make(map[string]string)

	return pool
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
