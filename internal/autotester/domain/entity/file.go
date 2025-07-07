package entity

// File represents a file entity with a name, binary data, and MIME type.
type File struct {
	fileName string
	fileData []byte
	mimeType string
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
