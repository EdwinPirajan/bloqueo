package services

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const (
	chromeCachePath   = "C:\\Users\\%s\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Cache"
	chromeCookiesPath = "C:\\Users\\%s\\AppData\\Local\\Google\\Chrome\\User Data\\Default\\Cookies"
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

// clearChromeCache limpia la caché de Chrome usando el protocolo de DevTools (Network.clearBrowserCache)

func modifyPermissions(userName string, restrict bool) error {
	cachePath := fmt.Sprintf(chromeCachePath, userName)
	cookiesPath := fmt.Sprintf(chromeCookiesPath, userName)

	// Cambiar permisos de caché
	err := setFilePermissions(cachePath, restrict)
	if err != nil {
		return fmt.Errorf("error al cambiar permisos de la caché de Chrome: %v", err)
	}

	// Cambiar permisos de cookies
	err = setFilePermissions(cookiesPath, restrict)
	if err != nil {
		return fmt.Errorf("error al cambiar permisos de las cookies de Chrome: %v", err)
	}

	if restrict {
		color.Red("Permisos de escritura de la caché y cookies bloqueados temporalmente.")
	} else {
		color.Green("Permisos de la caché y cookies restaurados.")
	}

	return nil
}

// setFilePermissions cambia los permisos de un archivo o directorio (bloqueo o restauración)
func setFilePermissions(path string, restrict bool) error {
	var permissions os.FileMode

	if restrict {
		permissions = 0444 // Solo lectura
	} else {
		permissions = 0644 // Lectura y escritura
	}

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return os.Chmod(p, permissions)
		}
		return nil
	})

	return err
}

func renameCacheAndCookies(userName string, restore bool) error {
	cachePath := fmt.Sprintf(chromeCachePath, userName)
	cookiesPath := fmt.Sprintf(chromeCookiesPath, userName)
	backupCachePath := cachePath + "_backup"
	backupCookiesPath := cookiesPath + "_backup"

	if restore {
		// Restaurar caché y cookies desde los nombres renombrados
		err := os.Rename(backupCachePath, cachePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error al restaurar la caché de Chrome: %v", err)
		}
		err = os.Rename(backupCookiesPath, cookiesPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error al restaurar las cookies de Chrome: %v", err)
		}
		color.Green("Caché y cookies de Chrome restauradas.")
	} else {
		// Renombrar caché y cookies para desactivar su uso por Chrome
		err := os.Rename(cachePath, backupCachePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error al renombrar la caché de Chrome: %v", err)
		}
		err = os.Rename(cookiesPath, backupCookiesPath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error al renombrar las cookies de Chrome: %v", err)
		}
		color.Red("Caché y cookies de Chrome renombradas temporalmente.")
	}

	return nil
}
