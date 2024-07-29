package domain

// VersionInfo representa la estructura del archivo JSON de la versión en el servidor.
type VersionInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}
