package services

import (
	"log"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
)

func MonitorProcesses(systemManager SystemManager, initialProcesses []string, initialUrls []string, user *domain.User) {
	chromeService := NewChromeService("https://apps.mypurecloud.com")
	appManager := NewWindowsApplicationManager(systemManager)

	selectors := []string{
		"participant call-participant text-center ember-view",
		"sms-textarea message-input form-control",
		"interaction-icon roster-email ember-view",
	}

	var previousMatchingProcesses []ProcessInfo
	var previousShouldBlock bool

	for {
		// 1) Obtener la configuración actual (actualizada vía WS) desde el store.
		cfg := GetCurrentConfig()

		// Si la configuración actual está vacía, se usan los valores iniciales.
		var processesToMonitor, urlsToBlock []string
		if len(cfg.ProcessesToMonitor) == 0 {
			processesToMonitor = initialProcesses
		} else {
			processesToMonitor = cfg.ProcessesToMonitor
		}
		if len(cfg.UrlsToBlock) == 0 {
			urlsToBlock = initialUrls
		} else {
			urlsToBlock = cfg.UrlsToBlock
		}

		// 2) Si el usuario no está activo, desbloquear todo y continuar.
		if !user.Active {
			log.Println("El usuario no está activo. Desbloqueando todos los aplicativos.")
			if err := RemoveURLsFromHostsFile(urlsToBlock); err != nil {
				log.Printf("Error eliminando las URLs del archivo hosts: %v\n", err)
			} else {
				log.Println("Todas las URLs desbloqueadas correctamente.")
			}

			activeProcesses, err := appManager.ListApplicationsInCurrentSession()
			if err != nil {
				log.Printf("Error listando las aplicaciones: %v\n", err)
			} else {
				matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
				log.Printf("Procesos a desbloquear: %v\n", matchingProcesses)
				for _, process := range matchingProcesses {
					handles, err := appManager.GetProcessHandlesInCurrentSession(process.Name)
					if err != nil {
						log.Printf("Error obteniendo los manejadores del proceso %s: %v\n", process.Name, err)
						continue
					}
					for _, handle := range handles {
						log.Printf("Intentando reanudar el proceso %s con el manejador %v\n", process.Name, handle)
						if err := appManager.ResumeProcess(handle); err != nil {
							log.Printf("Error reanudando el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s reanudado.\n", process.Name)
						}
					}
				}
			}
			previousMatchingProcesses = nil
			previousShouldBlock = false

			time.Sleep(2 * time.Second)
			continue
		}

		shouldBlock := true
		htmlContent, err := chromeService.GetFullPageHTML()
		if err != nil {
			log.Printf("Error obteniendo el HTML de la página: %v\n", err)
			shouldBlock = true
		} else {
			for _, selector := range selectors {
				if strings.Contains(htmlContent, selector) {
					shouldBlock = false
					break
				}
			}
		}

		if shouldBlock {
			if err := chromeService.BlockURLsInHosts(urlsToBlock); err != nil {
				log.Printf("Error bloqueando las URLs en el archivo hosts: %v\n", err)
			} else {
				log.Println("URLs bloqueadas exitosamente en el archivo hosts.")
			}
		} else {
			if err := RemoveURLsFromHostsFile(urlsToBlock); err != nil {
				log.Printf("Error eliminando las URLs del archivo hosts: %v\n", err)
			} else {
				log.Println("URLs eliminadas exitosamente del archivo hosts.")
			}

			if err := chromeService.NavigateBackToPreviousURLs(); err != nil {
				log.Printf("Error navegando de regreso a las URLs anteriores: %v\n", err)
			} else {
				log.Println("Navegación de regreso a las URLs anteriores completada exitosamente.")
			}
		}

		activeProcesses, err := appManager.ListApplicationsInCurrentSession()
		if err != nil {
			log.Printf("Error listando las aplicaciones: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
		log.Printf("Procesos coincidentes: %v\n", matchingProcesses)

		disappearedProcesses := difference(previousMatchingProcesses, matchingProcesses)
		if len(disappearedProcesses) > 0 {
			log.Printf("Procesos que ya no están en la config: %v\n", disappearedProcesses)
			for _, process := range disappearedProcesses {
				handles, err := appManager.GetProcessHandlesInCurrentSession(process.Name)
				if err != nil {
					log.Printf("Error obteniendo los manejadores del proceso %s: %v\n", process.Name, err)
					continue
				}
				for _, handle := range handles {
					log.Printf("Reanudando proceso %s (eliminado de la config) con manejador %v\n", process.Name, handle)
					if err := appManager.ResumeProcess(handle); err != nil {
						log.Printf("Error reanudando el proceso %s: %v\n", process.Name, err)
					} else {
						log.Printf("Proceso %s reanudado (ya no está en la config).\n", process.Name)
					}
				}
			}
		}

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
						if err := appManager.SuspendProcess(handle); err != nil {
							log.Printf("Error suspendiendo el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s suspendido.\n", process.Name)
						}
					} else {
						log.Printf("Intentando reanudar el proceso %s con el manejador %v\n", process.Name, handle)
						if err := appManager.ResumeProcess(handle); err != nil {
							log.Printf("Error reanudando el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s reanudado.\n", process.Name)
						}
					}
				}
			}
		}

		previousMatchingProcesses = matchingProcesses
		previousShouldBlock = shouldBlock

		time.Sleep(2 * time.Second)
	}
}

// difference retorna los procesos que estaban en oldList y ya no aparecen en newList.
func difference(oldList, newList []ProcessInfo) []ProcessInfo {
	newMap := make(map[string]bool)
	for _, proc := range newList {
		newMap[proc.Name] = true
	}

	var diff []ProcessInfo
	for _, proc := range oldList {
		if !newMap[proc.Name] {
			diff = append(diff, proc)
		}
	}
	return diff
}

func convertProcessNamesToProcessInfo(processNames []string) []ProcessInfo {
	var processInfos []ProcessInfo
	for _, name := range processNames {
		processInfos = append(processInfos, ProcessInfo{Name: name})
	}
	return processInfos
}
