package entity

type File struct {
	fileName string
	fileData []byte // fileData contains the binary data of the file
	mimeType string
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
