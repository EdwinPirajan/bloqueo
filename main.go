package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
		log.Fatalf("Application already running: %v", err)
	}
	defer windows.CloseHandle(mutexHandle)

	// logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("Error opening log file: %v", err)
	// }
	// defer logFile.Close()
	// log.SetOutput(logFile)

	err = enableDebugPrivilege()
	if err != nil {
		log.Fatalf("Error enabling debug privilege: %v", err)
	}

	systray.Run(onReady, onExit)
}

func onReady() {
	iconData, err := getIcon("icono.ico")
	if err != nil {
		log.Fatalf("Error loading icon: %v", err)
	}
	systray.SetIcon(iconData)
	systray.SetTitle("ScrapeBlocker")
	systray.SetTooltip("ScrapeBlocker")

	mStatus := systray.AddMenuItem("ScrapeBlocker - AlmaContact Desarrollo", "Estado de la aplicación")

	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	go func() {
		chromeService := services.NewChromeService("https://apps.mypurecloud.com")

		appManager := services.NewWindowsApplicationManager()

		selectors := []string{"Finalizar llamada"}
		processesToMonitor := config.Processes

		for {
			htmlContent, err := chromeService.GetFullPageHTML()
			shouldBlock := false

			if err != nil {
				log.Printf("Error: %v", err)
				shouldBlock = true
			} else {
				for _, selector := range selectors {
					if !strings.Contains(htmlContent, selector) {
						shouldBlock = true
						break
					}
				}
			}

			if shouldBlock {
				activeProcesses, err := appManager.ListApplications()
				if err != nil {
					log.Printf("Error listing applications: %v", err)
					continue
				}

				matchingProcesses := appManager.Intersect(activeProcesses, processesToMonitor)
				fmt.Printf("Procesos coincidentes: %v\n", matchingProcesses)

				for _, process := range matchingProcesses {
					handles, err := appManager.GetProcessHandles(process)
					if err != nil {
						log.Printf("Error getting handles for %s: %v\n", process, err)
						continue
					}
					for _, handle := range handles {
						err := appManager.SuspendProcess(handle)
						if err != nil {
							log.Printf("Error suspending process %s: %v\n", process, err)
						} else {
							log.Printf("Process %s suspended.\n", process)
						}
					}
				}
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
