package entity

// Content represents a Wrapper for possible prompt elements, including text and files in addition it has a tokenCount
type Content struct {
	Text       string // Text contains the written part of the prompt
	Files      []File // Files contains any files associated with the prompt
	TokenCount int    // TokenCount specifies the number of tokens within the prompt
}

// File represents a file with name, binary data and the file MIME type
type File struct {
	fileName string // fileName is the name of the file
	fileData []byte // fileData contains the binary data of the file
	mimeType string // mimeType is the MIME type of the file
}

func (f *File) GetFileName() string {
	return f.fileName
}

func (f *File) GetFileData() []byte {
	return f.fileData
}

func (f *File) GetMimeType() string {
	return f.mimeType
}

// Prompt is an abstract prompt type that contains the common elements shared by User and SystemPrompt.
type Prompt struct {
	content      Content  // content holds the actual prompt
	languageCode string   // languageCode specifies the language of the prompt
	logStamp     LogStamp // logStamp tracks the logging data
}

// UserPrompt specifies a Prompt
type UserPrompt struct {
	Prompt
	sessionID SessionID
}

// SystemPrompt specifies a Prompt
type SystemPrompt struct {
	Prompt
}
