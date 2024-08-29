package http

import (
	"io"
	"net/http"
	"os"
)

type FileDownloader struct{}

func NewFileDownloader() *FileDownloader {
	return &FileDownloader{}
}

func (fd *FileDownloader) Download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
