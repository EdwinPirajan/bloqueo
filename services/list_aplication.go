package services

import (
	"fmt"
	"os/exec"
	"strings"
)

type ListProcess interface {
	ListProcess() ([]string, error)
	CompareProcess() ([]string, error)
}

type ListProcessImpl struct{}

func NewListProcess() ListProcess {
	return &ListProcessImpl{}
}

func (s *ListProcessImpl) CompareProcess() ([]string, error) {
	ListProcess, err := s.ListProcess()
	if err != nil {
		return nil, err
	}

	var appProcess []string

	for _, process := range ListProcess {
		if strings.Contains(process, "Clientes.exe") {
			appProcess = append(appProcess, process)
		}
	}

	if len(appProcess) == 0 {
		return nil, fmt.Errorf("no applications found")
	}

	fmt.Println("Chrome Applications:", appProcess)

	return appProcess, nil
}

func (s *ListProcessImpl) ListProcess() ([]string, error) {
	cmd := exec.Command("tasklist")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var apps []string
	for _, line := range lines {
		if strings.Contains(line, ".exe") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				apps = append(apps, fields[0])
			}
		}
	}

	if len(apps) == 0 {
		return nil, fmt.Errorf("no applications found")
	}

	// fmt.Println("Applications:", apps)

	return apps, nil
}
