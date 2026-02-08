package primitive

import (
	"io"
)

type FileUpload struct {
	FileName string
	FileSize int64
	File     io.ReadCloser
}

func (c *FileUpload) Close() error {
	if c.File != nil {
		return c.File.Close()
	}

	return nil
}
