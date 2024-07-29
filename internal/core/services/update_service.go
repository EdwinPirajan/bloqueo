package services

import (
	"archive/zip"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/EdwinPirajan/bloqueo.git/internal/adapters/repository"
	"github.com/EdwinPirajan/bloqueo.git/internal/config"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/ports"
)

const updateMarkerFile = "update_in_progress"

type UpdateService struct {
	repo          ports.VersionChecker
	systemManager SystemManager
}

func NewUpdateService(cfg *config.Config, systemManager SystemManager) ports.UpdateService {
	repo := repository.NewVersionRepository(cfg.UpdateCheckURL)
	return &UpdateService{repo: repo, systemManager: systemManager}
}

func (s *UpdateService) CheckForUpdates() {
	if s.isUpdateInProgress() {
		log.Println("Update already in progress, skipping check.")
		return
	}

	versionInfo, err := s.repo.CheckForUpdates()
	if err != nil {
		log.Printf("Error checking for updates: %v\n", err)
		return
	}

	if versionInfo.Version != config.CurrentVersion {
		log.Println("New version available. Updating...")
		s.setUpdateInProgress(true)
		err := s.downloadAndUpdate(versionInfo.URL)
		if err != nil {
			log.Printf("Error updating: %v", err)
			s.setUpdateInProgress(false)
			return
		}
		log.Println("Update completed. Restarting...")
		s.restartApplication()
	} else {
		log.Println("No updates available.")
	}
}

func (s *UpdateService) downloadAndUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a temporary file for the new version
	tempFile, err := os.CreateTemp("", "update-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}

	// Unzip and replace files
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

func (s *UpdateService) restartApplication() {
	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable: %v", err)
	}
	cmd := exec.Command(executable)
	cmd.Start()
	os.Exit(0)
}

func (s *UpdateService) isUpdateInProgress() bool {
	if _, err := os.Stat(updateMarkerFile); err == nil {
		return true
	}
	return false
}

func (s *UpdateService) setUpdateInProgress(inProgress bool) {
	if inProgress {
		os.Create(updateMarkerFile)
	} else {
		os.Remove(updateMarkerFile)
	}
}
