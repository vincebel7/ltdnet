package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func loadTestNetwork(networkName string) Network {

	testNetMap := map[string]string{
		"testNet0": `{"id":"6715925f","name":"mynet","author":"vincebel7","netsize":"24","router":{"id":"c54f180b","model":"Bobcat 100","macaddr":"71:87:b6:50:fe:ac","hostname":"r1","gateway":"192.168.0.1","dhcppool":253,"vswitchid":{"id":"08d37525","model":"virtual","hostname":"V-08d37525","mgmtip":"","mactable":{"1b:77:76:29:be:53":1,"27:4b:57:6a:e8:9d":3,"9b:0a:d6:55:e1:f8":2},"maxports":4,"ports":["c54f180b","f8386e73","39625458","62b72e3f"],"portids":["c706367a","6ca08f45","5d59dc8d","4c1553ce"],"portmacs":null},"dhcptable":{"192.168.0.1":"","192.168.0.2":"1b:77:76:29:be:53","192.168.0.3":"9b:0a:d6:55:e1:f8","192.168.0.4":"27:4b:57:6a:e8:9d","192.168.0.5":""}}}`,
		"testNet1": `{"id":"38b760d8","name":"testNet1","author":"vince","netsize":"24","router":{"id":"e4846ce3","model":"Bobcat 100","macaddr":"12:7b:c6:a2:f4:b8","hostname":"r1","gateway":"192.168.0.1","dhcppool":253,"vswitchid":{"id":"409ab78c","model":"virtual","hostname":"V-409ab78c","mgmtip":"","mactable":{},"maxports":4,"ports":["e4846ce3","","",""],"portids":["0d4eb68a","916d6402","4b0966df","17263a26"],"portmacs":null},"dhcptable":{"192.168.0.1":"","192.168.0.2":"","192.168.0.3":"","192.168.0.4":"","192.168.0.5":""},"dhcptableorder":["192.168.0.1","192.168.0.2","192.168.0.3","192.168.0.4","192.168.0.5"]},"switches":null,"hosts":[{"id":"6541fa16","model":"ProBox 1","macaddr":"e7:8b:c1:22:9e:b0","hostname":"h1","ipaddr":"0.0.0.0","mask":"","gateway":"","uplinkid":""},{"id":"7ff8741c","model":"ProBox 1","macaddr":"7b:8e:cf:a3:e7:f3","hostname":"h2","ipaddr":"0.0.0.0","mask":"","gateway":"","uplinkid":""},{"id":"c63073c5","model":"ProBox 1","macaddr":"71:b7:aa:3b:27:9b","hostname":"h3","ipaddr":"0.0.0.0","mask":"","gateway":"","uplinkid":""},{"id":"19c35a9a","model":"ProBox 1","macaddr":"6d:30:1b:d8:3b:36","hostname":"h4","ipaddr":"0.0.0.0","mask":"","gateway":"","uplinkid":""}],"debug_level":1}`,
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
