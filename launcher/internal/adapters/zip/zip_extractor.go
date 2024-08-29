package zip

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ZipExtractor implementa la interfaz ZipExtractor
type ZipExtractor struct{}

func NewZipExtractor() *ZipExtractor {
	return &ZipExtractor{}
}

func (ze *ZipExtractor) Extract(zipPath, destDir string) error {
	// Abrir el archivo ZIP
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo ZIP: %v", err)
	}
	defer func() {
		if cerr := r.Close(); cerr != nil {
			fmt.Printf("error cerrando el archivo ZIP: %v\n", cerr)
		}
	}()

	// Iterar sobre los archivos en el ZIP
	for _, file := range r.File {
		fpath := filepath.Join(destDir, file.Name)

		// Comprobar si la ruta es válida
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("archivo en el ZIP contiene una ruta inválida: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// Crear directorios si es necesario
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return fmt.Errorf("error creando directorio %s: %v", fpath, err)
			}
			continue
		}

		// Crear los directorios si no existen
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("error creando directorio %s: %v", filepath.Dir(fpath), err)
		}

		// Crear archivo para escribir el contenido
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("error creando archivo %s: %v", fpath, err)
		}

		// Abrir el archivo dentro del ZIP
		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("error abriendo archivo dentro del ZIP %s: %v", file.Name, err)
		}

		// Copiar el contenido del archivo ZIP al archivo de destino
		_, err = io.Copy(outFile, rc)

		// Cerrar los archivos
		rc.Close()      // Cerrar el archivo dentro del ZIP
		outFile.Close() // Cerrar el archivo de salida

		if err != nil {
			return fmt.Errorf("error copiando archivo %s: %v", fpath, err)
		}
	}

	// Log de confirmación de extracción completada
	fmt.Println("Todos los archivos se han extraído correctamente.")

	// Intentar eliminar el archivo ZIP varias veces
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		// Esperar un breve momento antes de intentar eliminar
		time.Sleep(1 * time.Second)

		err = os.Remove(zipPath)
		if err == nil {
			// Log de confirmación de eliminación del archivo ZIP
			fmt.Println("Archivo ZIP eliminado correctamente.")
			return nil
		}

		fmt.Printf("Intento %d: Error eliminando el archivo ZIP: %v\n", i+1, err)
	}

	return fmt.Errorf("error eliminando el archivo ZIP después de varios intentos: %v", err)
}
