package entity

type Message struct {
	Actor       string `json:"agent"`
	MessageBody string `json:"data"`
}
