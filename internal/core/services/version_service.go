package services

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/EdwinPirajan/bloqueo.git/internal/adapters/repository"
	"github.com/EdwinPirajan/bloqueo.git/internal/config"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/ports"
)

// UpdateService es la implementación del servicio de actualizaciones.
type UpdateService struct {
	repo          ports.VersionChecker
	systemManager SystemManager
}

// NewUpdateService crea una nueva instancia de UpdateService.
func NewUpdateService(cfg *config.Config, systemManager SystemManager) ports.UpdateService {
	repo := repository.NewVersionRepository(cfg.UpdateCheckURL)
	return &UpdateService{repo: repo, systemManager: systemManager}
}

// CheckForUpdates verifica y aplica actualizaciones.
func (s *UpdateService) CheckForUpdates() {
	versionInfo, err := s.repo.CheckForUpdates()
	if err != nil {
		log.Printf("Error checking for updates: %v\n", err)
		return
	}

	if versionInfo.Version != config.CurrentVersion {
		log.Println("Nueva versión disponible. Actualizando...")
		err := s.downloadAndUpdate(versionInfo.URL)
		if err != nil {
			log.Printf("Error al actualizar: %v", err)
			return
		}
		// log.Println("Actualización completada. Ejecutando el script de actualización...")

		// cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", "C:\\ScrapeBlocker\\update.ps1")
		// cmd.Dir = "C:\\ScrapeBlocker"
		// err = cmd.Run()
		// if err != nil {
		// 	log.Fatalf("Error running update: %v", err)
		// }

	} else {
		log.Println("No hay actualizaciones disponibles.")
	}
}

func (s *UpdateService) downloadAndUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tempFile, err := os.CreateTemp("", "update-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}

	err = s.unzip(tempFile.Name(), ".")
	if err != nil {
		return err
	}

	return nil
}

func (s *UpdateService) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

// func (s *UpdateService) restartApplication() {
// 	executable, err := os.Executable()
// 	if err != nil {
// 		log.Fatalf("Error al obtener el ejecutable: %v", err)
// 	}
// 	cmd := exec.Command(executable)
// 	cmd.Start()
// 	os.Exit(0)
// }
