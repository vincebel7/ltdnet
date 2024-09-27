/*
File:		iphelper.go
Author: 	https://github.com/vincebel7
Purpose:	Functions for working with IPv4 addresses (type net.IP) and IP helpers.
*/

package iphelper

import (
	"math/big"
	"net"
)

type IPHelper struct {
	IP net.IP
}

// Constructor
func NewIPHelper(ip net.IP) (*IPHelper, error) {
	return &IPHelper{IP: ip}, nil
}

func (iph *IPHelper) IPToBigInt() *big.Int {
	ip := iph.IP.To4() // Ensure IPv4
	return big.NewInt(0).SetBytes(ip)
}

func BigIntToIP(ipInt *big.Int) net.IP {
	ipBytes := ipInt.Bytes()
	for len(ipBytes) < 4 {
		ipBytes = append([]byte{0}, ipBytes...) // Pad with leading zeroes if needed
	}
	return net.IP(ipBytes)
}

// Calculate the difference between two IP addresses
// TODO does this not account for wrapping
func (iph *IPHelper) IPDifference(other *IPHelper) *big.Int {
	ipInt1 := iph.IPToBigInt()
	ipInt2 := other.IPToBigInt()

	diff := big.NewInt(0)
	diff.Sub(ipInt2, ipInt1)
	return diff
}

func (iph *IPHelper) IncreaseIPByConstant(constant int) net.IP {
	ipInt := iph.IPToBigInt()

	sum := big.NewInt(0)
	sum.Add(ipInt, big.NewInt(int64(constant)))

	return WrapRawSumToIP(sum)
}

func WrapRawSumToIP(sum *big.Int) net.IP {
	for {
		ipBytes := sum.Bytes()
		byteAddr := make([]byte, 4)
		copy(byteAddr[4-len(ipBytes):], ipBytes)

		// Check for octet overflow, and set overflow octets to 0
		octetOverflow := false
		for i := 3; i >= 0; i-- {
			if int(byteAddr[i]) > 255 {
				byteAddr[i] = 0
				octetOverflow = true
			}
		}

		// If there's no overflow, we can return the valid IP
		if !octetOverflow {
			return net.IPv4(byteAddr[0], byteAddr[1], byteAddr[2], byteAddr[3])
		}

		// If overflow occurred, we increment the next octet
		sum.SetBytes(byteAddr)
		sum.Add(sum, big.NewInt(1))
	}
}
