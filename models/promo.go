package models

type (
	InputLast struct {
		Pending bool   `json:"pending"`
		Promo   string `json:"last"`
		Hash    string `json:"hash"`
	}

	OutputPromo struct {
		Pending bool   `json:"pending"`
		Promo   string `json:"promo"`
		Hash    string `json:"hash"`
	}

	OutputUpdate struct {
		Hash string `json:"hash"`
	}
)
