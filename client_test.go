package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func loadTestNetwork(networkName string) Network {

	// TODO load these from saves/test_saves/
	testNetMap := map[string]string{
	}

	var testnet Network
	err := json.Unmarshal([]byte(testNetMap[networkName]), &testnet)
	if err != nil {
		fmt.Printf("Unmarshalling error: %v", err)
	}
	fmt.Printf("Loaded %s\n", testnet.Name)
	return testnet
}

func TestLinkHost(t *testing.T) {
	snet = loadTestNetwork("testNet1")

	if snet.Router.Gateway != "192.168.0.1" {
		t.Errorf("gateway First test failed")
	}

}
