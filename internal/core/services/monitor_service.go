package services

import (
	"log"
	"strings"
	"time"
)

func MonitorProcesses(systemManager SystemManager, processesToMonitor []string) {
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

		htmlContent, err := chromeService.GetFullPageHTML()
		if err != nil {
			log.Printf("Error: %v\n", err)
			shouldBlock = true
		} else {
			for _, selector := range selectors {
				if strings.Contains(htmlContent, selector) {
					shouldBlock = false
					break
				}
			}
		}

		activeProcesses, err := appManager.ListApplicationsInCurrentSession()
		if err != nil {
			log.Printf("Error listing applications: %v\n", err)
			continue
		}

		matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
		log.Printf("Procesos coincidentes: %v\n", matchingProcesses)

		if !appManager.EqualProcessSlices(matchingProcesses, previousMatchingProcesses) || shouldBlock != previousShouldBlock {
			for _, process := range matchingProcesses {
				handles, err := appManager.GetProcessHandlesInCurrentSession(process.Name)
				if err != nil {
					log.Printf("Error getting handles for %s: %v\n", process.Name, err)
					continue
				}
				for _, handle := range handles {
					if shouldBlock {
						log.Printf("Attempting to suspend process %s with handle %v\n", process.Name, handle)
						err := appManager.SuspendProcess(handle)
						if err != nil {
							log.Printf("Error suspending process %s: %v\n", process.Name, err)
						} else {
							log.Printf("Process %s suspended.\n", process.Name)
						}
					} else {
						log.Printf("Attempting to resume process %s with handle %v\n", process.Name, handle)
						err := appManager.ResumeProcess(handle)
						if err != nil {
							log.Printf("Error resuming process %s: %v\n", process.Name, err)
						} else {
							log.Printf("Process %s resumed.\n", process.Name)
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
