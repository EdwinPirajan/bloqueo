package services

import (
	"log"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
)

func MonitorProcesses(systemManager SystemManager, processesToMonitor []string, urlsToBlock []string, user *domain.User) {
	chromeService := NewChromeService("https://apps.mypurecloud.com")
	appManager := NewWindowsApplicationManager(systemManager)

	selectors := []string{
		"Finalizar llamada",
		"sms-textarea message-input form-control",
		"interaction-icon roster-email ember-view",
	}

	var previousMatchingProcesses []ProcessInfo
	var previousShouldBlock bool

	for {
		// Verificar si el usuario está activo
		if !user.Active {
			log.Println("El usuario no está activo. Desbloqueando todos los aplicativos.")
			err := RemoveURLsFromHostsFile(urlsToBlock)
			if err != nil {
				log.Printf("Error eliminando las URLs del archivo hosts: %v\n", err)
			} else {
				log.Println("Todas las URLs desbloqueadas correctamente.")
			}

			// Saltar a la siguiente iteración sin realizar más acciones
			time.Sleep(2 * time.Second)
			continue
		}

		shouldBlock := true

		// Obtener el HTML completo de la página
		htmlContent, err := chromeService.GetFullPageHTML()
		if err != nil {
			log.Printf("Error obteniendo el HTML de la página: %v\n", err)
			shouldBlock = true
		} else {
			// Verificar si alguno de los selectores está presente en la página
			for _, selector := range selectors {
				if strings.Contains(htmlContent, selector) {
					shouldBlock = false
					break
				}
			}
		}

		if shouldBlock {
			err := chromeService.BlockURLsInHosts(urlsToBlock)
			if err != nil {
				log.Printf("Error bloqueando las URLs en el archivo hosts: %v\n", err)
			} else {
				log.Printf("URLs bloqueadas exitosamente en el archivo hosts.")
			}
		} else {
			// Desbloquear las URLs eliminándolas del archivo hosts
			err := RemoveURLsFromHostsFile(urlsToBlock)
			if err != nil {
				log.Printf("Error eliminando las URLs del archivo hosts: %v\n", err)
			} else {
				log.Printf("URLs eliminadas exitosamente del archivo hosts.")
			}

			// Navegar de regreso a las URLs anteriores
			err = chromeService.NavigateBackToPreviousURLs()
			if err != nil {
				log.Printf("Error navegando de regreso a las URLs anteriores: %v\n", err)
			} else {
				log.Printf("Navegación de regreso a las URLs anteriores completada exitosamente.")
			}
		}

		activeProcesses, err := appManager.ListApplicationsInCurrentSession()
		if err != nil {
			log.Printf("Error listando las aplicaciones: %v\n", err)
			continue
		}

		matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
		log.Printf("Procesos coincidentes: %v\n", matchingProcesses)

		if !appManager.EqualProcessSlices(matchingProcesses, previousMatchingProcesses) || shouldBlock != previousShouldBlock {
			for _, process := range matchingProcesses {
				handles, err := appManager.GetProcessHandlesInCurrentSession(process.Name)
				if err != nil {
					log.Printf("Error obteniendo los manejadores del proceso %s: %v\n", process.Name, err)
					continue
				}
				for _, handle := range handles {
					if shouldBlock {
						log.Printf("Intentando suspender el proceso %s con el manejador %v\n", process.Name, handle)
						err := appManager.SuspendProcess(handle)
						if err != nil {
							log.Printf("Error suspendiendo el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s suspendido.\n", process.Name)
						}
					} else {
						log.Printf("Intentando reanudar el proceso %s con el manejador %v\n", process.Name, handle)
						err := appManager.ResumeProcess(handle)
						if err != nil {
							log.Printf("Error reanudando el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s reanudado.\n", process.Name)
						}
					}
				}
			}
			previousMatchingProcesses = matchingProcesses
			previousShouldBlock = shouldBlock
		}

		time.Sleep(2 * time.Second)
	}
}

func convertProcessNamesToProcessInfo(processNames []string) []ProcessInfo {
	var processInfos []ProcessInfo
	for _, name := range processNames {
		processInfos = append(processInfos, ProcessInfo{Name: name})
	}
	return processInfos
}
