package main

import (
	"log"

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
	"allegrolatam.cloud.aircorp.aero:11139",
	"travelvoucher.app.lan.com:21169",
	"www.latamairlines.com",
}

func main() {

	systemManager := services.NewWindowsSystemManager()

	systray.Run(onReady(systemManager), onExit)
}

func onReady(systemManager services.SystemManager) func() {
	return func() {
		iconData, err := services.GetIcon("resources/icono.ico")
		if err != nil {
			log.Printf("Error loading icon: %v\n", err)
			return
		}
		systray.SetIcon(iconData)
		systray.SetTitle("ScrapeBlocker")
		systray.SetTooltip("ScrapeBlocker")

		mStatus := systray.AddMenuItem("ScrapeBlocker V0.1 - LATAM", "Almacontact")

		go services.MonitorProcesses(systemManager, processesToMonitor, urlsToBlock)

		<-mStatus.ClickedCh
	}
}

func onExit() {
}
