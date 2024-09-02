package main

import (
	"fmt"
	"log"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func main() {
	cliente := "latam"

	config, err := services.FetchConfiguration(cliente)
	if err != nil {
		fmt.Printf("Error obteniendo la configuración: %v\n", err)
	}

	systemManager := services.NewWindowsSystemManager()

	systray.Run(onReady(systemManager, config), onExit)
}

func onReady(systemManager services.SystemManager, config services.ConfigResponse) func() {
	return func() {
		iconData, err := services.GetIcon("resources/icono.ico")
		if err != nil {
			log.Printf("Error loading icon: %v\n", err)
			return
		}
		systray.SetIcon(iconData)
		systray.SetTitle("ScrapeBlocker")
		systray.SetTooltip("ScrapeBlocker")

		mStatus := systray.AddMenuItem("Versión desactualizada", "Almacontact")

		go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock)

		<-mStatus.ClickedCh
	}
}

func onExit() {
}
