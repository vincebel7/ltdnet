package model

import "time"

type MACEntry struct {
	Interface  int       `json:"interface"`
	State      string    `json:"state"`
	ExpireTime time.Time `json:"expireTime"`
}
