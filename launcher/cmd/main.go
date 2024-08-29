package main

import (
	"fmt"
	"launcher/internal/adapters/http"
	"launcher/internal/adapters/script"
	"launcher/internal/adapters/storage"
	"launcher/internal/adapters/zip"
	"launcher/internal/core/services"
)

func main() {
	// Inicializamos los adaptadores
	httpCheckService := http.NewHTTPCheckService("http://10.96.16.68:8080/check")
	fileStorage := storage.NewFileStorage(`C:\ScrapeBlocker\config\config.json`)
	scriptRunner := script.NewScriptRunner(`C:\ScrapeBlocker\config\update.bat`)
	zipExtractor := zip.NewZipExtractor()
	fileDownloader := http.NewFileDownloader()

	launcherService := services.NewLauncherService(httpCheckService, fileStorage, scriptRunner, zipExtractor, fileDownloader)

	if err := launcherService.Launch(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
