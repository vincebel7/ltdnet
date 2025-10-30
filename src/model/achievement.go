package model

type Achievement struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Hint        string `json:"hint"`
}
