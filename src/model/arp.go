package model

import "time"

type ARPEntry struct {
	MACAddr    string    `json:"macaddr"`
	ExpireTime time.Time `json:"expireTime"`
	Interface  string    `json:"interface"`
	State      string    `json:"state"`
}
