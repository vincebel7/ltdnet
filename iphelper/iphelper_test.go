package iphelper

import (
	"net"
	"testing"
)

func TestIncreaseIPByConstant(t *testing.T) {
	iph, _ := NewIPHelper(net.ParseIP("192.168.3.240"))

	summedIP := iph.IncreaseIPByConstant(5)
	if summedIP.String() != "192.168.3.245" {
		t.Errorf("TestIncreaseIPByConstant: Non-overflow IP sum failed")
	}

	summedIP = iph.IncreaseIPByConstant(50)
	print(summedIP.String())
	if summedIP.String() != "192.168.4.34" {
		t.Errorf("TestIncreaseIPByConstant: Non-overflow IP sum failed")
	}

}

func TestIPInSameSubnet(t *testing.T) {
	ip1 := "192.168.0.5"
	ip2 := "192.168.1.2"
	ip1Mask := "255.255.255.0"

	if IPInSameSubnet(ip1, ip2, ip1Mask) {
		t.Errorf("IPInSameSubnet returned unexpected result, claims 192.168.1.2 is in 192.168.0.5/24")
	}

	ip1Mask = "255.255.0.0"
	if !IPInSameSubnet(ip1, ip2, ip1Mask) {
		t.Errorf("IPInSameSubnet returned unexpected result, claims 192.168.1.2 is not in 192.168.0.5/16")
	}
}
