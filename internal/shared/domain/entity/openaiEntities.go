package entity

type RequestBody struct {
	UserPrompt   string
	SystemPrompt string
}

type Request struct {
	Model string
	Body  RequestBody
	Id    string
}

type Response struct {
	Output string
	Id     string
	Status string
}

func NewRequest(model string, body RequestBody) Request {
	return Request{Model: model, Body: body}
}

func NewRequestSession(model string, body RequestBody, session string) Request {
	return Request{Model: model, Body: body, Id: session}
}
