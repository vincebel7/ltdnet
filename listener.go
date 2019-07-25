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

var channelMap = map[string]int{}
const chansize = len(snet.Hosts)
var channels = [chansize]chan string

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
		channelMap[snet.Hosts[i].ID] = i
		channels[i] = make(chan string)
		go listen(snet.Hosts[i].ID)
		fmt.Printf("\n%s listening", snet.Hosts[i].Hostname)
	}
}

func listen(id string) {
	//<-channels[channelMap[id]]
	//<-channels[1]
	fmt.Println("spawned goroutine for %d\n", id)
}
