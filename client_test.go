package main

import (
	"testing"
	"net"
)

func TestNetworkSetup(t *testing.T) {
	loadNetwork("blank-v0_2_10", "test")
	setDebug("4")

	addRouter("r1", "Bobcat")

	if snet.Router.Gateway.String() != "192.168.0.1" {
		t.Errorf("Router not (properly) created")
	}

	addHost("h1")
	
	if snet.Hosts[0].Model != "ProBox 1" {
		t.Errorf("Host not (properly) created")
	}

	linkHost("h1", "r1")

	if (snet.Hosts[0].UplinkID == "") || (snet.Router.VSwitch.Ports[1] == "") {
		t.Errorf("Host not (properly) linked")
	}

	
	id := snet.Hosts[0].ID
	go dhcp_discover(snet.Hosts[0])
	<-actionsync[id]

	snet.Hosts[0].IPAddr = net.ParseIP("192.168.5.5")
	print(snet.Hosts[0].IPAddr.String())
	if (snet.Hosts[0].IPAddr.String() != "192.168.0.2") {
		t.Errorf("DHCP failed for host")
	}

}