package entity

type RequestBody struct {
	UserPrompt   string
	SystemPrompt string
}

type Request struct {
	Model string
	Body  RequestBody
}

type Response struct {
	Output string
	Id     string
}
