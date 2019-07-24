package main

import(
	"fmt"
	//"encoding/json"
	//"os"
	//"log"
	//"math/rand"
	//"time"
	//"bufio"
	//"strings"
	//"strconv"
	//"path/filepath"
	//"io/ioutil"
)

var channelMap = map[int]string{}
var channels = make([]chan int, len(snet.Hosts))

func Listener() {
	k := 0
	for j := range snet.Hosts{
		if j == 0 {}
		k = k+1
	}
	var channels = make([]chan int, k)
	for i := range snet.Hosts {
		if (len(snet.Hosts) == 3) {
		}
		//create channel and map channel ID
		channelMap[i] = snet.Hosts[i].ID
		channels[i] = make(chan int)
		//go listen(i)
		fmt.Printf("\n%s listening", snet.Hosts[i].Hostname)
		fmt.Printf(channelMap[i])
	}
}

func listen(i int) {
	fmt.Println("spawned goroutine %d\n", i)
}
