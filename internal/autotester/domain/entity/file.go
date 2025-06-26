package entity

// File represents a file entity with a name, binary data, and MIME type.
type File struct {
	fileName string // fileName is the name of the file
	fileData []byte // fileData contains the binary data of the file
	mimeType string // mimeType is the MIME type of the file
}

// GetFileName returns the name of the file.
func (f *File) GetFileName() string {
	return f.fileName
}

// GetFileData returns the binary data of the file.
func (f *File) GetFileData() []byte {
	return f.fileData
}

// GetMimeType returns the MIME type of the file.
func (f *File) GetMimeType() string {
	return f.mimeType
}
