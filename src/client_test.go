package main

import (
	"os"
	"path/filepath"
	"testing"
)

func cleanTestSaves() {
	dir := "../ltdnet_saves/test_saves/"
	files, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, file := range files {
		os.Remove(file)
	}
}

func TestNetworkSetup(t *testing.T) {
	// Setup
	testnetName := "testnet-" + idgen(8)
	newNetwork(testnetName, "24", "test")
	go Listener()

	setDebug("4")

	// Test 1: Add router
	addRouter("r1", "Bobcat")

	if snet.Router.GetIP("eth0") != "192.168.0.1" {
		t.Errorf("Router not (properly) created")
	}

	// Test 2: Add hosts
	addHost("h1")
	addHost("h2")
	addHost("h3")
	addHost("h4")

	if snet.Hosts[0].Model != "ProBox 1" {
		t.Errorf("Host not (properly) created")
	}

	// Test 3: Link hosts
	linkHostTo("h1", "r1")
	linkHostTo("h2", "r1")
	linkHostTo("h3", "r1")

	if (snet.Hosts[0].Interfaces["eth0"].RemoteL1ID == "") || (snet.Router.VSwitch.PortLinksRemote[1] == "") {
		t.Errorf("Host not (properly) linked")
	}

	// Test 4: DHCP
	go dhcp_discover(snet.Hosts[0])
	<-actionsync[snet.Hosts[0].ID]
	if snet.Hosts[0].GetIP("eth0") != "192.168.0.2" {
		t.Errorf("DHCP failed for host")
	}

	go dhcp_discover(snet.Hosts[1])
	<-actionsync[snet.Hosts[1].ID]
	if snet.Hosts[1].GetIP("eth0") != "192.168.0.3" {
		t.Errorf("DHCP failed for host")
	}

	go dhcp_discover(snet.Hosts[2])
	<-actionsync[snet.Hosts[2].ID]
	if snet.Hosts[2].GetIP("eth0") != "192.168.0.4" {
		t.Errorf("DHCP failed for host")
	}

	// Test 5: Delete host
	delHost("h2")

	// Test 5: Pinging
	go ping(snet.Hosts[0].ID, "192.168.0.4", 1)
	lossCount := <-actionsync[snet.Hosts[0].ID]
	if lossCount != 0 {
		t.Errorf("Ping from h1 to h3 failed")
	}

	go ping(snet.Hosts[0].ID, "192.168.0.2", 1)
	lossCount = <-actionsync[snet.Hosts[0].ID]
	if lossCount != 0 {
		t.Errorf("Ping from h1 to h1 failed")
	}

	// Test 6: Switch linking
	addSwitch("s1")
	addHost("h5")
	addHost("h6")

	linkHostTo("h5", "s1")
	linkHostTo("h6", "s1")

	cleanTestSaves()
}
