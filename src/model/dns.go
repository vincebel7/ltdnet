package model

type DNSQuestion struct {
	QName  string `json:"qname"`  // Domain name being queried
	QType  uint16 `json:"qtype"`  // Type of query (A, AAAA, etc.)
	QClass uint16 `json:"qclass"` // Class of query (IN for Internet)
}

type DNSRecord struct {
	Name     string `json:"name"`     // Domain name for this record
	Type     uint16 `json:"type"`     // Type of record (A, AAAA, NS, etc.)
	Class    uint16 `json:"class"`    // Class of record (IN for Internet)
	TTL      uint32 `json:"ttl"`      // Time to live, in seconds
	RDLength uint16 `json:"rdlength"` // Length of RData field
	RData    string `json:"rdata"`    // The actual data for this record (IP, NS name, etc.)
}

type DNSServer struct {
	ServerID string      `json:"serverid"` // Device ID of the device hosting the server
	Records  []DNSRecord `json:"records"`
	Enabled  bool        `json:"enabled"`
}

func (s *DNSServer) ARecordLookup(hostname string) (DNSRecord, bool) {
	for _, r := range s.Records {
		if r.Name == hostname {
			return r, true
		}
	}
	return DNSRecord{}, false
}
