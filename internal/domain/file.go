package domain

import "net/url"

const (
	ImageType = "Image"
	DocumentType = "Document"
)

type FileMetadata struct {
	Name string
    Filename string
    Type string
    MimeType string
    SizeBytes int64
	UploaderId int64
	Local bool
}

type File struct {
	FileMetadata
	Digest  string
    Path string
    ApId *url.URL
    Url *url.URL
}