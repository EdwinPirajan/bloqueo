package services

import (
	"archive/zip"
	"fmt"
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
	log.Println("Checking update status")

	if s.isUpdateInProgress() {
		log.Println("Update in progress marker file exists.")
		return
	}

	versionInfo, err := s.repo.CheckForUpdates()
	if err != nil {
		log.Printf("Error checking for updates: %v\n", err)
		return
	}

	if versionInfo.Version == "" {
		log.Println("Received invalid version information")
		return
	}

	if versionInfo.Version != config.CurrentVersion {
		log.Println("New version available:", versionInfo.Version)
		s.setUpdateInProgress(true)

		zipURL := versionInfo.URL
		log.Println("Download URL:", zipURL)

		err := s.downloadAndUpdate(zipURL)
		if err != nil {
			log.Printf("Error updating: %v", err)
			s.setUpdateInProgress(false)
			return
		}
		log.Println("Update completed. Restarting...")
		s.runUpdaterScript()
	} else {
		log.Println("No updates available.")
	}
}

func (s *UpdateService) downloadAndUpdate(url string) error {
	log.Println("Starting download from", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error downloading file: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to download file: Status code %d", resp.StatusCode)
		return fmt.Errorf("failed to download file: Status code %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "update-")
	if err != nil {
		log.Printf("Error creating temp directory: %v", err)
		return err
	}
	defer os.RemoveAll(tempDir)

	tempFilePath := filepath.Join(tempDir, "scrapeblockerV1.0.6.zip")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		return err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		log.Printf("Error saving downloaded file: %v", err)
		return err
	}

	log.Println("Download completed, starting unzip")
	err = s.unzip(tempFilePath, tempDir)
	if err != nil {
		log.Printf("Error unzipping file: %v", err)
		return err
	}

	updaterScript := "scripts/update.bat"
	err = copyFile(updaterScript, filepath.Join(tempDir, "updater.bat"))
	if err != nil {
		log.Printf("Error copying updater script: %v", err)
		return err
	}

	err = os.Rename(tempDir, "update")
	if err != nil {
		log.Printf("Error renaming temp directory: %v", err)
		return err
	}

	log.Println("Update downloaded and unzipped successfully")
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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (s *UpdateService) runUpdaterScript() {
	cmd := exec.Command("cmd", "/C", "update\\update.bat")
	cmd.Start()
	os.Exit(0)
}

func (s *UpdateService) isUpdateInProgress() bool {
	_, err := os.Stat(updateMarkerFile)
	if err == nil {
		log.Println("Update in progress marker file exists.")
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
