package ports

type ZipExtractor interface {
	Extract(zipPath, destDir string) error
}
