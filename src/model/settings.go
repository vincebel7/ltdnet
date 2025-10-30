package model

type Settings struct {
	ID             string              `json:"id"`
	Author         string              `json:"author"`
	Achievements   map[int]Achievement `json:"achievements"`
	AchievementsOn bool                `json:"achievements_on"`
	ProgramVer     string              `json:"program_ver"`
}
