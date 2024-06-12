package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/services"
	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

type Config struct {
	Processes []string `json:"processes"`
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func main() {
	systemManager := services.NewWindowsSystemManager()

	mutexHandle, err := systemManager.CheckForDuplicateInstance()
	if err != nil {
		fmt.Printf("Application already running: %v\n", err)
		return
	}
	defer windows.CloseHandle(mutexHandle)

	err = systemManager.EnableDebugPrivilege()
	if err != nil {
		fmt.Printf("Error enabling debug privilege: %v\n", err)
		return
	}

	systray.Run(func() { onReady(systemManager) }, onExit)
}

func onReady(systemManager services.SystemManager) {
	iconData, err := getIcon("icono.ico")
	if err != nil {
		fmt.Printf("Error loading icon: %v\n", err)
		return
	}
	systray.SetIcon(iconData)
	systray.SetTitle("ScrapeBlocker")
	systray.SetTooltip("ScrapeBlocker")

	mStatus := systray.AddMenuItem("ScrapeBlocker - AlmaContact Desarrollo", "Estado de la aplicación")

	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	go func(systemManager services.SystemManager) {
		chromeService := services.NewChromeService("https://apps.mypurecloud.com")
		appManager := services.NewWindowsApplicationManager(systemManager)

		selectors := []string{"Finalizar llamada"}
		processesToMonitor := config.Processes

		var previousMatchingProcesses []services.ProcessInfo
		var previousShouldBlock bool

		for {
			shouldBlock := false

			htmlContent, err := chromeService.GetFullPageHTML()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				shouldBlock = true
			} else {
				for _, selector := range selectors {
					if !strings.Contains(htmlContent, selector) {
						shouldBlock = true
						break
					}
				}
			}

			activeProcesses, err := appManager.ListApplications()
			if err != nil {
				fmt.Printf("Error listing applications: %v\n", err)
				continue
			}

			matchingProcesses := appManager.Intersect(activeProcesses, convertProcessNamesToProcessInfo(processesToMonitor))
			fmt.Printf("Procesos coincidentes: %v\n", matchingProcesses)

			// Check for changes in matching processes or shouldBlock state
			if !appManager.EqualProcessSlices(matchingProcesses, previousMatchingProcesses) || shouldBlock != previousShouldBlock {
				for _, process := range matchingProcesses {
					handles, err := appManager.GetProcessHandles(process.Name)
					if err != nil {
						fmt.Printf("Error getting handles for %s: %v\n", process.Name, err)
						continue
					}
					for _, handle := range handles {
						if shouldBlock {
							fmt.Printf("Attempting to suspend process %s with handle %v\n", process.Name, handle)
							err := appManager.SuspendProcess(handle)
							if err != nil {
								fmt.Printf("Error suspending process %s: %v\n", process.Name, err)
							} else {
								fmt.Printf("Process %s suspended.\n", process.Name)
							}
						} else {
							fmt.Printf("Attempting to resume process %s with handle %v\n", process.Name, handle)
							err := appManager.ResumeProcess(handle)
							if err != nil {
								fmt.Printf("Error resuming process %s: %v\n", process.Name, err)
							} else {
								fmt.Printf("Process %s resumed.\n", process.Name)
							}
						}
					}
				}
				previousMatchingProcesses = matchingProcesses
				previousShouldBlock = shouldBlock
			}

			time.Sleep(2 * time.Second)
		}
	}(systemManager)

	<-mStatus.ClickedCh
}

func onExit() {
}

func getIcon(path string) ([]byte, error) {
	icon, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return icon, nil
}

func convertProcessNamesToProcessInfo(processNames []string) []services.ProcessInfo {
	var processInfos []services.ProcessInfo
	for _, name := range processNames {
		processInfos = append(processInfos, services.ProcessInfo{Name: name})
	}
	return processInfos
}
