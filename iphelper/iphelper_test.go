package iphelper

import "testing"

func TestIncreaseIPByConstant(t *testing.T) {
	iph, _ := NewIPHelper("192.168.3.240")

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
