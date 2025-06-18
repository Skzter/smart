package entity

type DatabaseEntry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Tags     []string `json:"tags"`
}
