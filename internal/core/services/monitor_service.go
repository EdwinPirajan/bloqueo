package services

import (
	"log"
	"strings"
	"time"
)

func MonitorProcesses(systemManager SystemManager, processesToMonitor []string, urlsToBlock []string) {
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

		// Si no se encuentran los selectores, cerrar la conexión con las páginas bloqueadas
		if shouldBlock {
			err := chromeService.CloseTabsWithURLs(urlsToBlock)
			if err != nil {
				log.Printf("Error cerrando la conexión con las URLs: %v\n", err)
			}
		} else {
			log.Printf("Selectores encontrados, se permite la interacción en las URLs bloqueadas.")
		}

		// Monitorear procesos activos (mantiene la lógica original)
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
