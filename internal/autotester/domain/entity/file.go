package entity

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
