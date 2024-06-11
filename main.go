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

const SE_DEBUG_NAME = "SeDebugPrivilege"
const MutexName = "Global\\ScrapeBlockerMutex"

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

func enableDebugPrivilege() error {
	var hToken windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &hToken)
	if err != nil {
		return err
	}
	defer hToken.Close()

	var tkp windows.Tokenprivileges
	tkp.PrivilegeCount = 1
	tkp.Privileges[0].Attributes = windows.SE_PRIVILEGE_ENABLED

	name, err := windows.UTF16PtrFromString(SE_DEBUG_NAME)
	if err != nil {
		return err
	}
	err = windows.LookupPrivilegeValue(nil, name, &tkp.Privileges[0].Luid)
	if err != nil {
		return err
	}

	err = windows.AdjustTokenPrivileges(hToken, false, &tkp, 0, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func checkForDuplicateInstance() (windows.Handle, error) {
	mutexHandle, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(MutexName))
	if err != nil {
		return 0, err
	}
	if err = windows.GetLastError(); err == windows.ERROR_ALREADY_EXISTS {
		return 0, fmt.Errorf("another instance of the application is already running")
	}
	return mutexHandle, nil
}

func main() {
	mutexHandle, err := checkForDuplicateInstance()
	if err != nil {
		fmt.Printf("Application already running: %v\n", err)
		return
	}
	defer windows.CloseHandle(mutexHandle)

	// logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	fmt.Printf("Error opening log file: %v\n", err)
	// 	return
	// }
	// defer logFile.Close()
	// log.SetOutput(logFile)

	err = enableDebugPrivilege()
	if err != nil {
		fmt.Printf("Error enabling debug privilege: %v\n", err)
		return
	}

	systray.Run(onReady, onExit)
}

func onReady() {
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

	go func() {
		chromeService := services.NewChromeService("https://apps.mypurecloud.com")
		appManager := services.NewWindowsApplicationManager()

		selectors := []string{"Finalizar llamada"}
		processesToMonitor := config.Processes

		var previousMatchingProcesses []string
		var previousShouldBlock bool
		shouldBlock := true

		for {
			htmlContent, err := chromeService.GetFullPageHTML()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				shouldBlock = true
			} else {
				shouldBlock = true
				for _, selector := range selectors {
					if strings.Contains(htmlContent, selector) { // Corrige el error importando el paquete strings
						shouldBlock = false
						break
					}
				}
			}

			activeProcesses, err := appManager.ListApplications()
			if err != nil {
				fmt.Printf("Error listing applications: %v\n", err)
				continue
			}

			matchingProcesses := appManager.Intersect(activeProcesses, processesToMonitor)
			fmt.Printf("Procesos coincidentes: %v\n", matchingProcesses)

			// Check for changes in matching processes or shouldBlock state
			if !appManager.EqualStringSlices(matchingProcesses, previousMatchingProcesses) || shouldBlock != previousShouldBlock {
				for _, process := range matchingProcesses {
					handles, err := appManager.GetProcessHandles(process)
					if err != nil {
						fmt.Printf("Error getting handles for %s: %v\n", process, err)
						continue
					}
					for _, handle := range handles {
						if shouldBlock {
							err := appManager.SuspendProcess(handle)
							if err != nil {
								fmt.Printf("Error suspending process %s: %v\n", process, err)
							} else {
								fmt.Printf("Process %s suspended.\n", process)
							}
						} else {
							err := appManager.ResumeProcess(handle)
							if err != nil {
								fmt.Printf("Error resuming process %s: %v\n", process, err)
							} else {
								fmt.Printf("Process %s resumed.\n", process)
							}
						}
					}
				}
				previousMatchingProcesses = matchingProcesses
				previousShouldBlock = shouldBlock
			}

			time.Sleep(2 * time.Second)
		}
	}()

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
