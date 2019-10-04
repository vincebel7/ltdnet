/*
File:		debug.co
Author: 	https://bitbucket.org/vincebel
Purpose:	Functions related to debugging and testing
*/

package main

import(
	"fmt"
	"time"
)

func inspectFrame(f Frame) {
	p := f.Data
	s := p.Data

	fmt.Printf("\n ========== FRAME ========== \n")
	fmt.Printf("Source MAC:\t%s\n", f.SrcMAC)
	fmt.Printf("Dest MAC:\t%s\n", f.DstMAC)

	fmt.Printf("\n ========== PACKET ========== \n")
	fmt.Printf("Source IP:\t%s\n", p.SrcIP)
	fmt.Printf("Dest IP:\t%s\n", p.DstIP)

	fmt.Printf("\n ========== SEGMENT ========== \n")
	fmt.Printf("Data: %s", s.Data)

	fmt.Printf("\n\n")
}

func sleepDiv() {
	time.Sleep(time.Second)
	fmt.Println("---------------------")
	time.Sleep(time.Second)
}
