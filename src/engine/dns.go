package engine

import (
	"errors"
	"net"

	"github.com/vincebel7/ltdnet/src/model"
)

func (e *Engine) AddDNSRecord(rtype, name, ip string) error {
	if e.Net.Router == nil || e.Net.Router.DNSServer == nil {
		return errors.New("dns server not initialized")
	}
	if rtype != "A" {
		return errors.New("unsupported record type")
	}
	if net.ParseIP(ip) == nil {
		return errors.New("invalid IP")
	}
	// ensure map
	if e.Net.Router.DNSServer.Records == nil {
		e.Net.Router.DNSServer.Records = make([]model.DNSRecord, 0)
	}
	var recordType uint16
	switch rtype {
	case "A":
		recordType = 1
	default:
		return errors.New("unsupported record type")
	}
	e.Net.Router.DNSServer.Records = append(e.Net.Router.DNSServer.Records, model.DNSRecord{
		Type:  recordType,
		Name:  name,
		RData: ip,
	})
	return nil
}

func (e *Engine) DNSRecords() []model.DNSRecord {
	if e.Net == nil || e.Net.Router == nil || e.Net.Router.DNSServer == nil {
		return nil
	}
	return e.Net.Router.DNSServer.Records
}
