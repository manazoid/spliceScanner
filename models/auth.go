package models

type (
	InputAccount struct {
		Login   string `json:"login"`
		Credits int    `json:"credits"`
	}
)
