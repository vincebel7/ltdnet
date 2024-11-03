/*
File:		interface.go
Author: 	https://github.com/vincebel7
Purpose: 	Network interface
*/

package main

import "net"

type Interface struct {
	Name       string   `json:"name"`
	L1ID       string   `json:"id"`             // L1ID establishes one end of a Layer-1 connection
	RemoteL1ID string   `json:"remote_link_id"` // Remote L1ID this interface is connected to
	MACAddr    string   `json:"macaddr"`        // MAC Address
	IPConfig   IPConfig `json:"ipconfig"`
}

type IPConfig struct {
	IPAddress      net.IP `json:"ip"`
	SubnetMask     string `json:"subnetmask"`
	DefaultGateway net.IP `json:"gateway"`
	DNSServer      net.IP `json:"dns"`
	ConfigType     string `json:"configtype"` // Static or DHCP
}

func (h *Host) GetIP(interfaceName string) string {
	return h.Interfaces[interfaceName].IPConfig.IPAddress.String()
}

func (h *Host) GetMask(interfaceName string) string {
	return h.Interfaces[interfaceName].IPConfig.SubnetMask
}

func (h *Host) GetGateway(interfaceName string) string {
	return h.Interfaces[interfaceName].IPConfig.DefaultGateway.String()
}

func (r *Router) GetIP(interfaceName string) string {
	return r.Interfaces[interfaceName].IPConfig.IPAddress.String()
}

func (r *Router) GetMask(interfaceName string) string {
	return r.Interfaces[interfaceName].IPConfig.SubnetMask
}
