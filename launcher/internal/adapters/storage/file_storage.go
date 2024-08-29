package storage

import (
	"encoding/json"
	"io/ioutil"

	"launcher/internal/core/domain"
)

type FileStorage struct {
	configPath string
}

func NewFileStorage(configPath string) *FileStorage {
	return &FileStorage{configPath: configPath}
}

func (fs *FileStorage) LoadConfig() (*domain.Config, error) {
	file, err := ioutil.ReadFile(fs.configPath)
	if err != nil {
		return nil, err
	}

	var config domain.Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
