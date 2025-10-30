package main

import (
	"fmt"

	"github.com/vincebel7/ltdnet/src/engine"
)

func showDNSServer() {
	eng := engine.Instance()
	recs := eng.DNSRecords()
	fmt.Println("DNS server:")
	fmt.Printf("\tA Record count: %d\n\n", len(recs))

	// dynamic column widths
	maxName := len("Hostname")
	maxAddr := len("Address")
	for _, r := range recs {
		if len(r.Name) > maxName {
			maxName = len(r.Name)
		}
		if len(r.RData) > maxAddr {
			maxAddr = len(r.RData)
		}
	}
	fmt.Printf("%-*s  %-*s  Type\n", maxName, "Hostname", maxAddr, "Address")
	for _, r := range recs {
		fmt.Printf("%-*s  %-*s  %d\n", maxName, r.Name, maxAddr, r.RData, r.Type)
	}
}
