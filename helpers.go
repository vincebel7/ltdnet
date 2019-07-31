package main

import(
	//"fmt"
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
		if snet.Router.DHCPTable[v] == "" {
			net_prefix := ""
			//get network portion
			if(snet.Class == "A") {
				net_prefix = "10.0.0."
			} else if(snet.Class == "B") {
				net_prefix = "172.16.0."
			} else if(snet.Class == "C") {
				net_prefix = "192.168.0."
			}
			ipaddr := net_prefix + v
			return ipaddr
		}
	}
	return ""
}
