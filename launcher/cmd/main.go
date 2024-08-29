package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"launcher/internal/adapters/http"
	"launcher/internal/adapters/script"
	"launcher/internal/adapters/storage"
	"launcher/internal/adapters/zip"
	"launcher/internal/core/domain"
	"launcher/internal/core/services"
	"log"
)

func main() {
	// Leer la configuración desde el archivo
	configPath := `C:\ScrapeBlocker\config\config.json`
	configData, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error leyendo el archivo de configuración: %v", err)
	}

	var config domain.Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Error deserializando el archivo de configuración: %v", err)
	}

	// Inicializamos los adaptadores
	httpCheckService := http.NewHTTPCheckService("http://10.96.16.68:8080/check")
	fileStorage := storage.NewFileStorage(configPath)
	scriptRunner := script.NewScriptRunner(`C:\ScrapeBlocker\config\update.bat`)
	zipExtractor := zip.NewZipExtractor()
	fileDownloader := http.NewFileDownloader()

	launcherService := services.NewLauncherService(httpCheckService, fileStorage, scriptRunner, zipExtractor, fileDownloader)

	// Ejecutar el servicio de lanzamiento, pasando el cliente desde la configuración
	if err := launcherService.Launch(config.Client); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Ruta del archivo ZIP a eliminar después de la extracción
	zipPath := `C:\ScrapeBlocker\scrapeblocker.zip`

	// Llamar al método DeleteZip para eliminar el archivo ZIP
	if err := zipExtractor.DeleteZip(zipPath); err != nil {
		fmt.Printf("Error eliminando el archivo ZIP: %v\n", err)
	} else {
		fmt.Println("Archivo ZIP eliminado con éxito.")
	}
}
