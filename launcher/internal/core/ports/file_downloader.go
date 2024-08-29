package ports

type FileDownloader interface {
	Download(url, dest string) error
}
