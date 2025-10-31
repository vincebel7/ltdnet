package model

import (
	"math/big"
	"net"
	"time"

	"github.com/vincebel7/ltdnet/iphelper"
)

type Network struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Netsize    string   `json:"netsize"`
	Router     *Router  `json:"router"`
	Switches   []Switch `json:"switches"`
	Hosts      []Host   `json:"hosts"`
	ProgramVer string   `json:"program_ver"`
}

type Host struct {
	ID         string               `json:"id"`
	Model      string               `json:"model"`
	Hostname   string               `json:"hostname"`
	ARPTable   map[string]ARPEntry  `json:"arptable"`
	DNSTable   map[string]DNSRecord `json:"dnstable"`
	Interfaces map[string]Interface `json:"interfaces"`
}

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

type Switch struct {
	ID              string              `json:"id"`
	Model           string              `json:"model"`
	Hostname        string              `json:"hostname"`
	MACTable        map[string]MACEntry `json:"mactable"`
	Maxports        int                 `json:"maxports"`
	PortLinksRemote []string            `json:"links_remote"` // maps port # to remote link ID
	PortLinksLocal  []string            `json:"links_local"`  // maps port # to local link ID
	ARPTable        map[string]ARPEntry `json:"arptable"`
}

type MACEntry struct {
	Interface  int       `json:"interface"`
	State      string    `json:"state"`
	ExpireTime time.Time `json:"expireTime"`
}

type ARPEntry struct {
	MACAddr    string    `json:"macaddr"`
	ExpireTime time.Time `json:"expireTime"`
	Interface  string    `json:"interface"`
	State      string    `json:"state"`
}

func (n *Network) ClearMACTables() {
	// Router ARP
	if n.Router != nil && n.Router.ARPTable != nil {
		for k := range n.Router.ARPTable {
			delete(n.Router.ARPTable, k)
		}
	}
	// Hosts ARP
	for i := range n.Hosts {
		if n.Hosts[i].ARPTable != nil {
			for k := range n.Hosts[i].ARPTable {
				delete(n.Hosts[i].ARPTable, k)
			}
		}
	}
	// Switch MAC address tables
	for i := range n.Switches {
		n.Switches[i].MACTable = make(map[string]MACEntry)
	}
	// VSwitch MAC table only if router exists
	if n.Router != nil {
		if n.Router.VSwitch.MACTable == nil {
			n.Router.VSwitch.MACTable = make(map[string]MACEntry)
		} else {
			for k := range n.Router.VSwitch.MACTable {
				delete(n.Router.VSwitch.MACTable, k)
			}
		}
	}
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
