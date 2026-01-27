package entity

// File represents a file entity with a name, binary data, and MIME type.
type File struct {
	fileName      string
	fileData      []byte
	fileExtension string
}

func NewFile(name string, data []byte, fileExtension string) File {
	return File{fileName: name, fileData: data, fileExtension: fileExtension}
}

// GetFileName returns the name of the file.
func (f *File) GetFileName() string {
	return f.fileName
}

// GetFileData returns the binary data of the file.
func (f *File) GetFileData() []byte {
	return f.fileData
}

// GetFileExtension returns the extension of the file.
func (f *File) GetFileExtension() string {
	return f.fileExtension
}
