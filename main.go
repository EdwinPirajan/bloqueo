package main

import (
	"fmt"
	"log"

	"github.com/EdwinPirajan/bloqueo.git/internal/config"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/ports"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

var processesToMonitor = []string{
	"Seguridad.exe", "Gestion.exe", "Mensajería.exe", "Red.exe", "Referencia.exe",
	"Producto.exe", "version.exe", "javaw.exe", "Batch.exe", "Control.exe",
	"Clientes.exe", "ATMAdmin.exe", "ATMCompensacion.exe", "ATMMonitor.exe", "ATMPersonaliza.exe",
	"PINPAD.exe", "Bonos.exe", "cartera.exe", "COBISCorp.eCOBIS.COBISExplorer.CommunicationManager.exe",
	"COBISExplorerApplicationsRemover", "Cce.exe", "Cci.exe", "Cde.exe", "Cdi.exe", "Ceadmin.exe",
	"Corresp.exe", "Grb-gra.exe", "Stb.exe", "Tre-trr.exe", "Cobconta.exe", "cobconci.exe", "cobpresu.exe",
	"COBRANZA.exe", "Admcred.exe", "Tramites.exe", "Buzon.exe", "Camara.exe", "person.exe", "prudepo.exe",
	"tadmin.exe", "tarifario.exe", "sit.exe", "af.exe", "brp.exe", "cxc.exe", "cxp.exe", "SB.exe",
	"Rechazos.exe", "Reportvb5.exe", "Depadmin.exe", "Depopera.exe", "PEB.exe", "garantia.exe", "Firmas.exe",
	"HerramientaCuadre.exe", "vrcAgrario.exe",
}

var urlsToBlock = []string{
	"www.almacontact.com.co",
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Println("ScrapeBlocker v1.0.5")
	systemManager := services.NewWindowsSystemManager()
	err = systemManager.EnableDebugPrivilege()
	if err != nil {
		log.Fatalf("Error enabling debug privilege: %v", err)
	}

	updateService := services.NewUpdateService(cfg, systemManager)

	systray.Run(onReady(systemManager, updateService), onExit)
}

func onReady(systemManager services.SystemManager, updateService ports.UpdateService) func() {
	return func() {
		iconData, err := services.GetIcon("resources/icono.ico")
		if err != nil {
			log.Printf("Error loading icon: %v\n", err)
			return
		}
		systray.SetIcon(iconData)
		systray.SetTitle("ScrapeBlocker")
		systray.SetTooltip("ScrapeBlocker")

		mStatus := systray.AddMenuItem("test", "test")

		go services.MonitorProcesses(systemManager, processesToMonitor, urlsToBlock)
		go updateService.CheckForUpdates()

		<-mStatus.ClickedCh
	}
}

func onExit() {
}
