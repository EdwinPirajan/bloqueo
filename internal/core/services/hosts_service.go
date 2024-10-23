package services

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

//cambios

func AddURLsToHostsFile(urls []string) error {
	const hostsFilePath = "C:\\Windows\\System32\\drivers\\etc\\hosts"

	file, err := os.OpenFile(hostsFilePath, os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo hosts: %v", err)
	}
	defer file.Close()

	existingLines := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		existingLines[line] = true
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error al leer el archivo hosts: %v", err)
	}

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

func RemoveURLsFromHostsFile(urls []string) error {
	const hostsFilePath = "C:\\Windows\\System32\\drivers\\etc\\hosts"

	file, err := os.ReadFile(hostsFilePath)
	if err != nil {
		return fmt.Errorf("error al leer el archivo hosts: %v", err)
	}

	lines := strings.Split(string(file), "\n")
	var newLines []string

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

	output := strings.Join(newLines, "\n")
	err = os.WriteFile(hostsFilePath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("error al escribir el archivo hosts: %v", err)
	}

	return nil
}

func CloseChromeTabsWithURLs(urls []string) error {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq chrome.exe", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error al obtener la lista de procesos de Chrome: %v", err)
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		pid := strings.Trim(fields[1], "\"")

		cmdURLs := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "get", "CommandLine")
		outputURLs, err := cmdURLs.Output()
		if err != nil {
			log.Printf("Error al obtener las URLs para el proceso %s: %v", pid, err)
			continue
		}

		for _, url := range urls {
			if strings.Contains(string(outputURLs), url) {
				cmdKill := exec.Command("taskkill", "/PID", pid, "/F")
				err = cmdKill.Run()
				if err != nil {
					log.Printf("Error al cerrar el proceso %s: %v", pid, err)
				} else {
					log.Printf("Cerrada pestaña con URL: %s", url)
				}
				break
			}
		}
	}

	return nil
}

func HardReloadTabsWithURLs(urlsToBlock []string) error {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq chrome.exe", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error al obtener la lista de procesos de Chrome: %v", err)
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 2 {
			continue
		}

		pid := strings.Trim(fields[1], "\"")

		cmdURLs := exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%s", pid), "get", "CommandLine")
		outputURLs, err := cmdURLs.Output()
		if err != nil {
			log.Printf("Error al obtener las URLs para el proceso %s: %v", pid, err)
			continue
		}

		for _, url := range urlsToBlock {
			if strings.Contains(string(outputURLs), url) {
				err = hardReloadChromeTab(pid)
				if err != nil {
					log.Printf("Error al hacer hard reload de la pestaña con URL %s: %v", url, err)
				} else {
					log.Printf("Hard reload realizado con éxito para la pestaña con URL: %s", url)
				}
				break
			}
		}
	}

	fmt.Println("Hard reload de pestañas completado", urlsToBlock)

	return nil
}

func hardReloadChromeTab(pid string) error {
	cmd := exec.Command("taskkill", "/PID", pid, "/F")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al hacer hard reload de la pestaña con PID %s: %v", pid, err)
	}

	return nil
}
