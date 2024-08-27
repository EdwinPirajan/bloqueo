package services

import (
	"log"
	"strings"
	"time"
)

var urlsToBlock = []string{
	"https://www.ejemplo.com",
	"https://www.otroejemplo.com",
}

func MonitorProcesses(systemManager SystemManager, processesToMonitor []string) {
	chromeService := NewChromeService("https://apps.mypurecloud.com")
	appManager := NewWindowsApplicationManager(systemManager)

	// Selectores HTML a verificar en la página de PureCloud
	selectors := []string{
		"Finalizar llamada",
		"sms-textarea message-input form-control",
		"interaction-icon roster-email ember-view",
	}

	var previousMatchingProcesses []ProcessInfo
	var previousShouldBlock bool

	for {
		shouldBlock := false

		// Obtener la URL actual
		currentURL, err := chromeService.GetCurrentURL()
		if err != nil {
			log.Printf("Error obteniendo URL actual: %v\n", err)
			shouldBlock = true
		} else {
			// Verificar si la URL actual es una de las URLs a bloquear
			for _, url := range urlsToBlock {
				if strings.Contains(currentURL, url) {
					log.Printf("La URL %s está abierta y será bloqueada.\n", url)
					shouldBlock = true
					break
				}
			}

			// Si la URL actual es la de PureCloud, verificar los selectores
			if strings.Contains(currentURL, "https://apps.mypurecloud.com") {
				htmlContent, err := chromeService.GetFullPageHTML()
				if err != nil {
					log.Printf("Error obteniendo HTML de la página: %v\n", err)
					shouldBlock = true
				} else {
					for _, selector := range selectors {
						if strings.Contains(htmlContent, selector) {
							shouldBlock = true
							log.Printf("El selector %s fue encontrado en la URL %s.\n", selector, currentURL)
							break
						}
					}
				}
			}
		}

		// Suspender o reanudar procesos según sea necesario
		activeProcesses, err := appManager.ListApplicationsInCurrentSession()
		if err != nil {
			log.Printf("Error al listar aplicaciones: %v\n", err)
			continue
		}

		matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
		log.Printf("Procesos coincidentes: %v\n", matchingProcesses)

		if !appManager.EqualProcessSlices(matchingProcesses, previousMatchingProcesses) || shouldBlock != previousShouldBlock {
			for _, process := range matchingProcesses {
				handles, err := appManager.GetProcessHandlesInCurrentSession(process.Name)
				if err != nil {
					log.Printf("Error obteniendo manejos para %s: %v\n", process.Name, err)
					continue
				}
				for _, handle := range handles {
					if shouldBlock {
						log.Printf("Intentando suspender el proceso %s con el manejador %v\n", process.Name, handle)
						err := appManager.SuspendProcess(handle)
						if err != nil {
							log.Printf("Error al suspender el proceso %s: %v\n", process.Name, err)
						} else {
							log.Printf("Proceso %s suspendido.\n", process.Name)
						}
					} else {
						log.Printf("Intentando reanudar el proceso %s con el manejador %v\n", process.Name, handle)
						err := appManager.ResumeProcess(handle)
						if err != nil {
							log.Printf("Error al reanudar el proceso %s: %v\n", process.Name, err)
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
