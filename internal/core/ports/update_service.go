package ports

// UpdateService es la interfaz del servicio de actualizaciones.
type UpdateService interface {
	CheckForUpdates()
}
