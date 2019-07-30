package main

import(
	"fmt"
	"math/rand"
	"time"
	//"strings"
	//"strconv"
)

func idgen(n int) string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, n)

	rand.Seed(time.Now().UnixNano())
	for i := range id {
		id[i] = idchars[rand.Intn(len(idchars))]
	}

	return string(id)
}

func macgen() string {
	mac := idgen(2)
	for n := 0; n < 5; n++ {
		mac = mac + ":" + idgen(2)
	}
	return mac
}

func getMACfromID(id string) string {
	//Router
	if id == snet.Router.ID {
		return snet.Router.MACAddr
	}

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id {
			return snet.Hosts[h].MACAddr
		}
	}
	return ""
}

func getIDfromMAC(mac string) string {
	//Router
	if mac == snet.Router.MACAddr {
		return snet.Router.ID
	}

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].MACAddr == mac {
			return snet.Hosts[h].ID
		}
	}
	return ""
}

func next_free_addr() string {
	//find open address
	//fmt.Println(snet.Router.DHCPTable)
	for _, v := range snet.Router.DHCPIndex {
		fmt.Printf("key %s\n", v)
		if snet.Router.DHCPTable[v] == "" {
			fmt.Printf("\n%s is free\n", v)
			return v
		}
	}
	return ""
}
