package services

import (
	"os"
)

func GetIcon(path string) ([]byte, error) {
	icon, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return icon, nil
}
