package model

type Switch struct {
	ID              string              `json:"id"`
	Model           string              `json:"model"`
	Hostname        string              `json:"hostname"`
	MACTable        map[string]MACEntry `json:"mactable"`
	Maxports        int                 `json:"maxports"`
	PortLinksRemote []string            `json:"links_remote"` // maps port # to remote link ID
	PortLinksLocal  []string            `json:"links_local"`  // maps port # to local link ID
	ARPTable        map[string]ARPEntry `json:"arptable"`
}

func (s *Switch) GetID() string       { return s.ID }
func (s *Switch) GetHostname() string { return s.Hostname }
