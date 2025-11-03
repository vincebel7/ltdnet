package model

type Achievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Hint        string `json:"hint"`
}

// Achievement IDs
const (
	AchRoutineBusiness = 1
	AchUnitedPingdom   = 2
	AchArpHot          = 3
	AchSniffFrames     = 4
	AchTenHosts        = 5
	AchMyName          = 6
)
