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
	r := Net().Router
	if r != nil && isSwitchportID(r.VSwitch, link) {
		for i := range r.VSwitch.PortLinksLocal {
			if r.VSwitch.PortLinksLocal[i] == link {
				return i
			}
		}
		return -1
	}
	// fall back to other switches
	switchID := getSwitchIDFromLink(link)
	if switchID == "" {
		return -1
	}
	swIdx := getSwitchIndexFromID(switchID)
	if swIdx < 0 {
		return -1
	}
	for i := range Net().Switches[swIdx].PortLinksLocal {
		if Net().Switches[swIdx].PortLinksLocal[i] == link {
			return i
		}
	}
	return -1
}

func getSwitchIDFromLink(link string) string {
	r := Net().Router
	if r != nil && isSwitchportID(r.VSwitch, link) {
		return r.VSwitch.ID
	}
	for i := range Net().Switches {
		if isSwitchportID(Net().Switches[i], link) {
			return Net().Switches[i].ID
		}
	}
	if r != nil {
		return r.VSwitch.ID
	}
	return ""
}

func getIDfromMAC(mac string) string {
	r := Net().Router
	if r != nil {
		if eth0, ok := r.Interfaces["eth0"]; ok && eth0.MACAddr == mac {
			return r.ID
		}
	}
	for h := range Net().Hosts {
		if iface, ok := Net().Hosts[h].Interfaces["eth0"]; ok && iface.MACAddr == mac {
			return Net().Hosts[h].ID
		}
	}
	return ""
}

func getHostnameFromID(id string) string {
	hostname := ""
	deviceType := getDeviceType(id)
	if deviceType == "host" {
		if getHostIndexFromID(id) != -1 {
			hostname = Net().Hosts[getHostIndexFromID(id)].Hostname
		} else {
			hostname = id
		}
	} else if deviceType == "switch" {
		if getSwitchIndexFromID(id) != -1 {
			hostname = Net().Switches[getSwitchIndexFromID(id)].Hostname
		} else {
			hostname = id
		}
	} else if deviceType == "vswitch" {
		hostname = Net().Router.VSwitch.Hostname
	} else if deviceType == "router" {
		hostname = Net().Router.Hostname
	} else {
		hostname = id
	}

	return hostname
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
