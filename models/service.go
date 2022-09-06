package models

type (
	Cookie struct {
		Cookie string `json:"cookie"`
	}

	OutputService struct {
		Promo  string `json:"promo"`
		Info   string `json:"info"`
		Status string `json:"status"`
	}
)
