package services

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// AddURLsToHostsFile añade URLs al archivo hosts
func AddURLsToHostsFile(urls []string) error {
	const hostsFilePath = "C:\\Windows\\System32\\drivers\\etc\\hosts"

	// Abrimos el archivo en modo lectura y escritura
	file, err := os.OpenFile(hostsFilePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo hosts: %v", err)
	}
	defer file.Close()

	// Leemos las líneas existentes
	existingLines := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		existingLines[line] = true
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error al leer el archivo hosts: %v", err)
	}

	// Escribimos las nuevas líneas si no están presentes
	for _, url := range urls {
		entry := fmt.Sprintf("0.0.0.0 %s", url)
		if !existingLines[entry] {
			if _, err := file.WriteString(entry + "\n"); err != nil {
				return fmt.Errorf("error al escribir en el archivo hosts: %v", err)
			}
			color.Green("URL añadida al archivo hosts: %s", entry)
		} else {
			color.Yellow("URL ya existe en el archivo hosts: %s", entry)
		}
	}

	return nil
}

// RemoveURLsFromHostsFile elimina URLs del archivo hosts
func RemoveURLsFromHostsFile(urls []string) error {
	const hostsFilePath = "C:\\Windows\\System32\\drivers\\etc\\hosts"

	// Leer el archivo hosts actual
	file, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return fmt.Errorf("error al leer el archivo hosts: %v", err)
	}

	lines := strings.Split(string(file), "\n")
	var newLines []string

	// Filtrar las líneas para eliminar las URLs especificadas
	for _, line := range lines {
		shouldKeep := true
		for _, url := range urls {
			if strings.Contains(line, url) {
				shouldKeep = false
				color.Red("URL eliminada del archivo hosts: %s", line)
				break
			}
		}
		if shouldKeep {
			newLines = append(newLines, line)
		}
	}

	// Escribir las nuevas líneas al archivo hosts
	output := strings.Join(newLines, "\n")
	err = os.WriteFile(hostsFilePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo hosts: %v", err)
	}

	return nil
}
