/*
File:		dns.go
Author: 	https://github.com/vincebel7
Purpose:	DNS functions
*/

package main

import (
	"fmt"
	"strings"
)

type DNSQuestion struct {
	QName  string `json:"qname"`  // Domain name being queried
	QType  uint16 `json:"qtype"`  // Type of query (A, AAAA, etc.)
	QClass uint16 `json:"qclass"` // Class of query (IN for Internet)
}

type DNSRecord struct {
	Name     string `json:"name"`     // Domain name for this record
	Type     uint16 `json:"type"`     // Type of record (A, AAAA, NS, etc.)
	Class    uint16 `json:"class"`    // Class of record (IN for Internet)
	TTL      uint32 `json:"ttl"`      // Time to live, in seconds
	RDLength uint16 `json:"rdlength"` // Length of RData field
	RData    string `json:"rdata"`    // The actual data for this record (IP, NS name, etc.)
}

type DNSServer struct {
	ServerID string      `json:"serverid"` // Device ID of the device hosting the server
	Records  []DNSRecord `json:"records"`
	Enabled  bool        `json:"enabled"`
}

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
		snet.Router.DNSServer.Records = dnsServer.Records
		debug(3, "addDNSRecordToServer", snet.Router.ID, "Record added to DNS server")

	default:
		fmt.Println("[Error] DNS record type not yet supported by DNS server")
	}
}

func displayDNSTable(dnsTable map[string]DNSRecord) {
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
