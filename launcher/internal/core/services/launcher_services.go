package services

import (
	"fmt"
	"launcher/internal/core/ports"
	"os/exec"
)

type LauncherService struct {
	checkService   ports.CheckService
	fileStorage    ports.FileStorage
	scriptRunner   ports.ScriptRunner
	zipExtractor   ports.ZipExtractor
	fileDownloader ports.FileDownloader
}

func NewLauncherService(cs ports.CheckService, fs ports.FileStorage, sr ports.ScriptRunner, ze ports.ZipExtractor, fd ports.FileDownloader) *LauncherService {
	return &LauncherService{
		checkService:   cs,
		fileStorage:    fs,
		scriptRunner:   sr,
		zipExtractor:   ze,
		fileDownloader: fd,
	}
}

func (ls *LauncherService) Launch() error {
	// Leer config.json
	localConfig, err := ls.fileStorage.LoadConfig()
	if err != nil {
		return fmt.Errorf("error cargando config.json: %v", err)
	}

	// Hacer la petición HTTP
	checkResponse, err := ls.checkService.CheckForUpdates()
	if err != nil {
		return fmt.Errorf("error realizando la petición: %v", err)
	}

	// Comparar versiones
	if localConfig.Version != checkResponse.Version {
		fmt.Println("Nueva versión detectada, iniciando descarga...")

		// Descargar el archivo
		if err := ls.fileDownloader.Download(checkResponse.URL, `C:\ScrapeBlocker\scrapeblocker.zip`); err != nil {
			return fmt.Errorf("error descargando el archivo: %v", err)
		}

		// Ejecutar update.bat
		if err := ls.scriptRunner.Run(); err != nil {
			return fmt.Errorf("error ejecutando update.bat: %v", err)
		}

		// Extraer archivo descargado
		if err := ls.zipExtractor.Extract(`C:\ScrapeBlocker\scrapeblocker.zip`, `C:\ScrapeBlocker`); err != nil {
			fmt.Printf("Error extrayendo el archivo: %v. Continuando con la ejecución...\n", err)
		} else {
			fmt.Println("Actualización completada exitosamente.")
		}
	} else {
		fmt.Println("La versión está actualizada.")
	}

	// Ejecutar ScrapeBlocker.exe
	if err := ls.runScrapeBlocker(); err != nil {
		return fmt.Errorf("error al ejecutar ScrapeBlocker.exe: %v", err)
	}

	return nil
}

// runScrapeBlocker ejecuta el archivo ScrapeBlocker.exe
func (ls *LauncherService) runScrapeBlocker() error {
	scriptPath := `C:\ScrapeBlocker\config\runsb.bat`

	// Comprobar si el archivo runsb.bat existe
	if _, err := exec.LookPath(scriptPath); err != nil {
		return fmt.Errorf("no se encontró runsb.bat en la ruta especificada: %v", scriptPath)
	}

	cmd := exec.Command("cmd", "/C", scriptPath)
	err := cmd.Start() // Utilizamos Start() para que no bloquee el proceso actual
	if err != nil {
		return fmt.Errorf("error al iniciar runsb.bat: %v", err)
	}

	fmt.Println("runsb.bat se ha iniciado correctamente.")
	return nil
}
