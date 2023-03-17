package utils

import (
	"net/http"
	"os"
)

func ExistData(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func DetectFileContentType(file *os.File) string {
	buffer := make([]byte, 512)
	file.Read(buffer)

	contentType := http.DetectContentType(buffer)

	file.Seek(0, 0)

	return contentType
}
