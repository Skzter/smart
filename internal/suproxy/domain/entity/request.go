package entity

// Request represents a request entity with header, prompt, destination, and request content.
type Request struct {
	Header      map[string]string `json:"header"`
	Prompt      string            `json:"prompt"`
	Destination string            `json:"destination"`
	Request     string            `json:"request"`
}

// RequestNoHeader represents a request entity without the header field.
type RequestNoHeader struct {
	Prompt      string `json:"prompt"`
	Destination string `json:"destination"`
	Request     string `json:"request"`
}

// ConvertToNoHeader transforms a Request into a RequestNoHeader by omitting the Header field.
func ConvertToNoHeader(r *Request) RequestNoHeader {
	if r == nil {
		return RequestNoHeader{}
	}
	return RequestNoHeader{
		Prompt:      r.Prompt,
		Destination: r.Destination,
		Request:     r.Request,
	}
}
