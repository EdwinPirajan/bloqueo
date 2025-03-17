package services

import (
	"sync"
)

var (
	currentConfig ConfigResponse
	mutex         sync.Mutex
)

// SetCurrentConfig actualiza la configuración actual de forma segura.
func SetCurrentConfig(cfg ConfigResponse) {
	mutex.Lock()
	defer mutex.Unlock()
	currentConfig = cfg
}

// GetCurrentConfig retorna la configuración actual de forma segura.
func GetCurrentConfig() ConfigResponse {
	mutex.Lock()
	defer mutex.Unlock()
	return currentConfig
}
