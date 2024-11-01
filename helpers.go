/*
File:		helpers.go
Author: 	https://github.com/vincebel7
Purpose:	Various misc helper functions
*/

package main

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func idgen(n int) string {
	var idchars = []rune("abcdef1234567890")
	id := make([]rune, n)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range id {
		id[i] = idchars[r.Intn(len(idchars))]
	}

	return string(id)
}

func idgen_int(n int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	firstDigit := r.Intn(9) + 1
	numberStr := strconv.Itoa(firstDigit)
	for i := 1; i < n; i++ {
		digit := r.Intn(10)
		numberStr += strconv.Itoa(digit)
	}
	result, _ := strconv.Atoi(numberStr)

	return result
}

func macgen() string {
	mac := idgen(2)
	for n := 0; n < 5; n++ {
		mac = mac + ":" + idgen(2)
	}

	return mac
}

func getDeviceType(id string) string {
	if snet.Router.ID == id {
		return "router"
	}
	if snet.Router.VSwitch.ID == id {
		return "vswitch"
	}

	for s := range snet.Switches {
		if snet.Switches[s].ID == id {
			return "switch"
		}
	}

	return "host"
}

func getHostIndexFromID(id string) int {
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id {
			return h
		}
	}
	return -1
}

func getHostIndexFromLinkID(id string) int {
	for h := range snet.Hosts {
		if snet.Hosts[h].Interface.RemoteL1ID == id {
			return h
		}
	}
	return -1
}

func getSwitchIndexFromID(id string) int {
	for s := range snet.Switches {
		if snet.Switches[s].ID == id {
			return s
		}
	}

	return -1
}

func getSwitchportIDFromLink(link string) int {
	switchID := getSwitchIDFromLink(link)

	s := snet.Router.VSwitch
	if switchID != snet.Router.VSwitch.ID {
		s = snet.Switches[getSwitchIndexFromID(switchID)]
	}

	for i := range s.PortLinksLocal {
		if s.PortLinksLocal[i] == link {
			return i
		}
	}

	return -1
}

func getSwitchIDFromLink(link string) string {
	s := snet.Router.VSwitch

	if isSwitchportID(snet.Router.VSwitch, link) {
		s = snet.Router.VSwitch
	} else {
		for i := range snet.Switches {
			if isSwitchportID(snet.Switches[i], link) {
				return snet.Switches[i].ID
			}
		}
	}

	return s.ID
}

func getIDfromMAC(mac string) string {
	//Router
	if mac == snet.Router.Interface.MACAddr {
		return snet.Router.ID
	}

	//Hosts
	for h := range snet.Hosts {
		if snet.Hosts[h].Interface.MACAddr == mac {
			return snet.Hosts[h].ID
		}
	}

	return ""
}

func dynamic_assign(id string, ipaddr net.IP, defaultgateway net.IP, subnetMask string) {
	for h := range snet.Hosts {
		if snet.Hosts[h].ID == id {
			snet.Hosts[h].Interface.IPConfig.IPAddress = ipaddr
			snet.Hosts[h].Interface.IPConfig.SubnetMask = subnetMask
			snet.Hosts[h].Interface.IPConfig.DefaultGateway = defaultgateway
			fmt.Println("Network configuration updated")
		}
	}

}

func hostname_exists(hostname string) bool {
	hostname = strings.ToUpper(hostname)

	if strings.ToUpper(snet.Router.Hostname) == hostname {
		return true
	}
	if strings.ToUpper(snet.Router.VSwitch.Hostname) == hostname {
		return true
	}

	for s := range snet.Switches {
		if strings.ToUpper(snet.Switches[s].Hostname) == hostname {
			return true
		}
	}

	for h := range snet.Hosts {
		if strings.ToUpper(snet.Hosts[h].Hostname) == hostname {
			return true
		}
	}

	return false
}

func prefixLengthToSubnetMask(prefixLength int) string {
	subnetMask := "0.0.0.0"
	if prefixLength == 8 {
		subnetMask = "255.0.0.0"
	} else if prefixLength == 16 {
		subnetMask = "255.255.0.0"
	} else if prefixLength == 24 {
		subnetMask = "255.255.255.0"
	}

	return subnetMask
}

func removeHostFromSlice(s []Host, i int) []Host {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeSwitchFromSlice(s []Switch, i int) []Switch {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func PadRight(str string, length int) string {
	if len(str) >= length {
		return str // Return the original string if it's already the desired length or longer
	}
	return str + fmt.Sprintf("%*s", length-len(str), "")
}
