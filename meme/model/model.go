package model

//go:generate easyjson model.go

//easyjson:json
type KeysRes []string

//easyjson:json
type KeyInfoRes struct {
	Key      string   `json:"key"`
	Keywords []string `json:"keywords"`
	Params   struct {
		MinImages    int      `json:"min_images"`
		MaxImages    int      `json:"max_images"`
		MinTexts     int      `json:"min_texts"`
		MaxTexts     int      `json:"max_texts"`
		DefaultTexts []string `json:"default_texts"`
	} `json:"params"`
}
