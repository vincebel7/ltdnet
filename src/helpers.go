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

	"github.com/vincebel7/ltdnet/src/model"
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

func ephemeralPortGen() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(65535-1024) + 1024
}

func getDeviceType(id string) string {
	r := Net().Router
	if r != nil {
		if r.ID == id {
			return "router"
		}
		if r.VSwitch.ID == id {
			return "vswitch"
		}
	}
	for s := range Net().Switches {
		if Net().Switches[s].ID == id {
			return "switch"
		}
	}

	return "host"
}

func getHostIndexFromID(id string) int {
	for h := range Net().Hosts {
		if Net().Hosts[h].ID == id {
			return h
		}
	}
	return -1
}

func getHostIndexFromLinkID(id string) int {
	for h := range Net().Hosts {
		if Net().Hosts[h].Interfaces["eth0"].RemoteL1ID == id {
			return h
		}
	}
	return -1
}

func getSwitchIndexFromID(id string) int {
	for s := range Net().Switches {
		if Net().Switches[s].ID == id {
			return s
		}
	}

	return -1
}

func getSwitchportIDFromLink(link string) int {
	switchID := getSwitchIDFromLink(link)

	s := Net().Router.VSwitch
	if switchID != Net().Router.VSwitch.ID {
		s = Net().Switches[getSwitchIndexFromID(switchID)]
	}

	for i := range s.PortLinksLocal {
		if s.PortLinksLocal[i] == link {
			return i
		}
	}

	return -1
}

func getSwitchIDFromLink(link string) string {
	s := Net().Router.VSwitch

	if isSwitchportID(Net().Router.VSwitch, link) {
		s = Net().Router.VSwitch
	} else {
		for i := range Net().Switches {
			if isSwitchportID(Net().Switches[i], link) {
				return Net().Switches[i].ID
			}
		}
	}

	return s.ID
}

func getIDfromMAC(mac string) string {
	//Router
	if mac == Net().Router.Interfaces["eth0"].MACAddr {
		return Net().Router.ID
	}

	//Hosts
	for h := range Net().Hosts {
		if Net().Hosts[h].Interfaces["eth0"].MACAddr == mac {
			return Net().Hosts[h].ID
		}
	}

	return ""
}

func dynamic_assign(id string, ipaddr net.IP, defaultgateway net.IP, subnetMask string) {
	for h := range Net().Hosts {
		if Net().Hosts[h].ID == id {
			iface := Net().Hosts[h].Interfaces["eth0"]

			iface.IPConfig.IPAddress = ipaddr
			iface.IPConfig.SubnetMask = subnetMask
			iface.IPConfig.DefaultGateway = defaultgateway

			Net().Hosts[h].Interfaces["eth0"] = iface

			fmt.Println("Network configuration updated")
		}
	}

}

func hostname_exists(hostname string) bool {
	target := strings.ToUpper(hostname)
	r := Net().Router
	if r != nil {
		if strings.ToUpper(r.Hostname) == target {
			return true
		}
		if strings.ToUpper(r.VSwitch.Hostname) == target {
			return true
		}
	}
	for s := range Net().Switches {
		if strings.ToUpper(Net().Switches[s].Hostname) == target {
			return true
		}
	}
	for h := range Net().Hosts {
		if strings.ToUpper(Net().Hosts[h].Hostname) == target {
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

func removeHostFromSlice(s []model.Host, i int) []model.Host {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeSwitchFromSlice(s []model.Switch, i int) []model.Switch {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func PadRight(str string, length int) string {
	if len(str) >= length {
		return str // Return the original string if it's already the desired length or longer
	}
	return str + fmt.Sprintf("%*s", length-len(str), "")
}
