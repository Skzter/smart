package entity

type Request struct {
	Header      []string `json:"header"`
	Prompt      string   `json:"prompt"`
	Destination string   `json:"destination"`
	Request     string   `json:"request"`
}
