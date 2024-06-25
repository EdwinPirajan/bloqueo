package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/services"
	"github.com/getlantern/systray"
)

var processesToMonitor = []string{
	"Seguridad.exe", "Gestion.exe", "Mensajería.exe", "Red.exe", "Referencia.exe",
	"Producto.exe", "version.exe", "javaw.exe", "Batch.exe", "Control.exe",
	"Clientes.exe", "ATMAdmin.exe", "ATMCompensacion.exe", "ATMMonitor.exe", "ATMPersonaliza.exe",
	"PINPAD.exe", "Bonos.exe", "Cartera.exe", "COBISCorp.eCOBIS.COBISExplorer.CommunicationManager.exe",
	"COBISExplorerApplicationsRemover", "Cce.exe", "Cci.exe", "Cde.exe", "Cdi.exe", "Ceadmin.exe",
	"Corresp.exe", "Grb-gra.exe", "Stb.exe", "Tre-trr.exe", "Cobconta.exe", "cobconci.exe", "cobpresu.exe",
	"COBRANZA.exe", "Admcred.exe", "Tramites.exe", "Buzon.exe", "Camara.exe", "person.exe", "prudepo.exe",
	"tadmin.exe", "tarifario.exe", "sit.exe", "af.exe", "brp.exe", "cxc.exe", "cxp.exe", "SB.exe",
	"Rechazos.exe", "Reportvb5.exe", "Depadmin.exe", "Depopera.exe", "PEB.exe", "garantia.exe", "Firmas.exe",
	"HerramientaCuadre.exe", "vrcAgrario.exe",
}

func main() {
	systemManager := services.NewWindowsSystemManager()

	err := systemManager.EnableDebugPrivilege()
	if err != nil {
		log.Printf("Error enabling debug privilege: %v\n", err)
		return
	}

	systray.Run(onReady(systemManager), onExit)
}

func onReady(systemManager services.SystemManager) func() {
	return func() {
		iconData, err := getIcon("icono.ico")
		if err != nil {
			log.Printf("Error loading icon: %v\n", err)
			return
		}
		systray.SetIcon(iconData)
		systray.SetTitle("ScrapeBlocker")
		systray.SetTooltip("ScrapeBlocker")

		mStatus := systray.AddMenuItem("ScrapeBlocker - AlmaContact Desarrollo", "Estado de la aplicación")

		go monitorProcesses(systemManager)

		<-mStatus.ClickedCh
	}
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

func monitorProcesses(systemManager services.SystemManager) {
	chromeService := services.NewChromeService("https://apps.mypurecloud.com")
	appManager := services.NewWindowsApplicationManager(systemManager)

	selectors := []string{"Finalizar llamada"}

	var previousMatchingProcesses []services.ProcessInfo
	var previousShouldBlock bool

	for {
		shouldBlock := false

		htmlContent, err := chromeService.GetFullPageHTML()
		if err != nil {
			log.Printf("Error: %v\n", err)
			shouldBlock = true
		} else {
			for _, selector := range selectors {
				if !strings.Contains(htmlContent, selector) {
					shouldBlock = true
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

func convertProcessNamesToProcessInfo(processNames []string) []services.ProcessInfo {
	var processInfos []services.ProcessInfo
	for _, name := range processNames {
		processInfos = append(processInfos, services.ProcessInfo{Name: name})
	}
	return processInfos
}
