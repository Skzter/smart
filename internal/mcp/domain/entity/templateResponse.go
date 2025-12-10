package entity

// TemplateResponse repräsentiert die Antwort des Autotester-Backends auf eine Template-Abfrage.
type TemplateResponse struct {
	Content string `json:"template"`
}
