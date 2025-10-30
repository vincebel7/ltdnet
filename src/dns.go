/*
File:		dns.go
Author: 	https://github.com/vincebel7
Purpose:	DNS functions
*/

package main

import (
	"fmt"

	"github.com/vincebel7/ltdnet/src/model"
)

func (dnsServer *DNSServer) dnsServerMenu() {
	// Calculate the record count
	aRecordCount := len(dnsServer.Records)
	fmt.Println("DNS server:")
	fmt.Printf("\tA Record count: %d\n", aRecordCount)

	// Step 1: Calculate maximum widths
	maxNameLen := len("Hostname") // Starting with header width
	maxAddrLen := len("Address")
	maxTypeLen := len("Type")

	for _, record := range dnsServer.Records {
		if len(record.Name) > maxNameLen {
			maxNameLen = len(record.Name)
		}
		if len(record.RData) > maxAddrLen {
			maxAddrLen = len(record.RData)
		}
	}

	// Step 2: Print header with dynamic padding
	fmt.Println("\nDNS server records:")
	fmt.Printf("%-*s  %-*s  %-*s  %s\n", maxNameLen, "Hostname", maxAddrLen, "Address", maxTypeLen, "Type", "TTL")

	// Step 3: Print each record using calculated widths
	for _, record := range dnsServer.Records {
		fmt.Printf("%-*s  %-*s  %-*c  %d\n", maxNameLen, record.Name, maxAddrLen, record.RData, maxTypeLen, record.Type, record.TTL)
	}
	fmt.Print("\n")
}

// Local type alias for model.DNSServer to allow method definition
type DNSServer model.DNSServer

func displayDNSTable(dnsTable map[string]model.DNSRecord) {
	// Step 1: Calculate maximum widths
	maxNameLen := len("Hostname") // Starting with header width
	maxAddrLen := len("Address")

	for _, record := range dnsTable {
		if len(record.Name) > maxNameLen {
			maxNameLen = len(record.Name)
		}
		if len(record.RData) > maxAddrLen {
			maxAddrLen = len(record.RData)
		}
	}

	// Step 2: Print header with dynamic padding
	fmt.Printf("Hosts:\n")
	fmt.Printf("\tA Record count: %d\n\n", len(dnsTable))
	fmt.Printf("Host records:\n")
	fmt.Printf("%-*s  %-*s  %-4s  %s\n", maxNameLen, "Hostname", maxAddrLen, "Address", "Type", "TTL")

	// Step 3: Print each record using calculated widths
	for _, record := range dnsTable {
		fmt.Printf("%-*s  %-*s  %-4c  %d\n", maxNameLen, record.Name, maxAddrLen, record.RData, record.Type, record.TTL)
	}

	fmt.Print("\n")
}
