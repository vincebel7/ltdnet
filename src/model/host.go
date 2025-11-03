package model

type Host struct {
	ID         string               `json:"id"`
	Model      string               `json:"model"`
	Hostname   string               `json:"hostname"`
	ARPTable   map[string]ARPEntry  `json:"arptable"`
	DNSTable   map[string]DNSRecord `json:"dnstable"`
	Interfaces map[string]Interface `json:"interfaces"`
}

func (h *Host) GetID() string       { return h.ID }
func (h *Host) GetHostname() string { return h.Hostname }
