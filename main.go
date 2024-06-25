package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/services"
	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

var processesToMonitor = []string{
	"Seguridad.exe",
	"Gestion.exe",
	"Mensajería.exe",
	"Red.exe",
	"Referencia.exe",
	"Producto.exe",
	"version.exe",
	"javaw.exe",
	"Batch.exe",
	"Control.exe",
	"Clientes.exe",
	"ATMAdmin.exe",
	"ATMCompensacion.exe",
	"ATMMonitor.exe",
	"ATMPersonaliza.exe",
	"PINPAD.exe",
	"Bonos.exe",
	"Cartera.exe",
	"COBISCorp.eCOBIS.COBISExplorer.CommunicationManager.exe",
	"COBISExplorerApplicationsRemover",
	"Cce.exe",
	"Cci.exe",
	"Cde.exe",
	"Cdi.exe",
	"Ceadmin.exe",
	"Corresp.exe",
	"Grb-gra.exe",
	"Stb.exe",
	"Tre-trr.exe",
	"Cobconta.exe",
	"cobconci.exe",
	"cobpresu.exe",
	"COBRANZA.exe",
	"Admcred.exe",
	"Tramites.exe",
	"Buzon.exe",
	"Camara.exe",
	"person.exe",
	"prudepo.exe",
	"tadmin.exe",
	"tarifario.exe",
	"sit.exe",
	"af.exe",
	"brp.exe",
	"cxc.exe",
	"cxp.exe",
	"SB.exe",
	"Rechazos.exe",
	"Reportvb5.exe",
	"Depadmin.exe",
	"Depopera.exe",
	"PEB.exe",
	"garantia.exe",
	"Firmas.exe",
	"HerramientaCuadre.exe",
	"vrcAgrario.exe",
}

func main() {
	systemManager := services.NewWindowsSystemManager()

	sessionID, err := systemManager.GetCurrentSessionID()
	if err != nil {
		fmt.Printf("Error getting current session ID: %v\n", err)
		return
	}

	mutexName := fmt.Sprintf("ScrapeBlockerMutex_Session_%d", sessionID)
	mutexHandle, err := windows.CreateMutex(nil, false, windows.StringToUTF16Ptr(mutexName))
	if err != nil {
		fmt.Printf("Error creating mutex: %v\n", err)
		return
	}
	if err = windows.GetLastError(); err == windows.ERROR_ALREADY_EXISTS {
		fmt.Printf("Application already running in this session\n")
		return
	}
	defer windows.CloseHandle(mutexHandle)

	err = systemManager.EnableDebugPrivilege()
	if err != nil {
		fmt.Printf("Error enabling debug privilege: %v\n", err)
		return
	}

	go monitorSession(systemManager, sessionID)
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

	go func(systemManager services.SystemManager) {
		chromeService := services.NewChromeService("https://apps.mypurecloud.com")
		appManager := services.NewWindowsApplicationManager(systemManager)

		selectors := []string{
			"Finalizar llamada",
		}

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
	// No need to unblock processes globally on exit
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

func monitorSession(systemManager services.SystemManager, currentSessionID uint32) {
	for {
		activeSessionID, err := systemManager.GetCurrentActiveSessionID()
		if err != nil {
			fmt.Printf("Error getting active session ID: %v\n", err)
			return
		}

		if currentSessionID != activeSessionID {
			fmt.Println("User session has changed. Exiting application.")
			os.Exit(0)
		}

		time.Sleep(2 * time.Second) // Reduce the sleep interval for quicker response
	}
}
