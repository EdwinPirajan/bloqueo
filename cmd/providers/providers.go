package providers

import (
	"fmt"
	"log"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func Run() {
	cliente := "bac"
	baseWebSocketURL := "ws://10.96.16.68:8080/api/v1/ws"

	user, err := services.GetUser()
	if err != nil {
		log.Fatalf("No se pudo obtener el usuario: %v", err)
	}
	user.Client = cliente

	config, err := services.FetchConfiguration(cliente)
	if err != nil {
		log.Printf("Error obteniendo la configuración: %v", err)
	}

	wsURL := fmt.Sprintf("%s/%s/%s", baseWebSocketURL, user.Name, cliente)

	systemManager := services.NewWindowsSystemManager()

	systray.Run(func() { onReady(systemManager, config, wsURL, &user) }, onExit)
}

func onReady(systemManager services.SystemManager, config services.ConfigResponse, wsURL string, user *domain.User) {
	iconData, err := services.GetIcon("resources/icono.ico")
	if err != nil {
		log.Printf("Error loading icon: %v", err)
		return
	}
	systray.SetIcon(iconData)
	systray.SetTitle("ScrapeBlocker")
	systray.SetTooltip("ScrapeBlocker")

	mStatus := systray.AddMenuItem("ScrapeBlocker - LATAM Airlines v0.2 - Almacontact", "Almacontact")

	go func() {
		if err := services.ConnectAndKeepOpen(wsURL, user); err != nil {
			log.Printf("Error en WebSocket: %v", err)
		}
	}()

	log.Printf("Iniciando MonitorProcesses con config: %+v y usuario: %+v", config, user)
	go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock, user)

	<-mStatus.ClickedCh
}

func onExit() {
	log.Println("Saliendo de la aplicación systray...")
}
