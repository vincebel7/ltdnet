/*
File:		dns.go
Author: 	https://github.com/vincebel7
Purpose:	DNS
*/

package main

import (
	"fmt"
	"strings"
)

type DNSServer struct {
	ServerID string      `json:"serverid"` // Device ID of the device hosting the server
	Records  []DNSRecord `json:"records"`
	Enabled  bool        `json:"enabled"`
}

func (dnsServer *DNSServer) dnsServerMenu() {
	aRecordCount := len(dnsServer.Records)
	fmt.Println("DNS server:")
	fmt.Printf("\tA Record count: %d\n\n", aRecordCount)
}

func (dnsServer *DNSServer) aRecordLookup(hostname string) DNSRecord {
	for record := range dnsServer.Records {
		if dnsServer.Records[record].Name == hostname {
			return dnsServer.Records[record]
		}
	}

	return DNSRecord{}
}

func (dnsServer *DNSServer) addDNSRecordToServer(recordType uint16, hostname string, address string) {
	switch recordType {
	case 'A':
		if dnsServer.aRecordLookup(hostname).Name != "" {
			fmt.Print("Record already exists. are you sure you want to overwrite? [y/n]: ")
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.ToUpper(confirmation)

			if confirmation != "Y" {
				return
			}
		}

		newRecord := DNSRecord{
			Name:  hostname,
			Type:  recordType,
			Class: 1,
			TTL:   65535,
			RData: address,
		}
		dnsServer.Records = append(dnsServer.Records, newRecord)
		debug(3, "addDNSRecordToServer", snet.Router.ID, "Record added to DNS server")

	default:
		fmt.Println("[Error] DNS record type not yet supported by DNS server")
	}
}
