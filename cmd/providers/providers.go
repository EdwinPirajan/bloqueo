package providers

import (
	"fmt"
	"log"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func Run() {
	cliente := "bacb"
	baseWebSocketURL := "ws://10.96.16.67:8080/api/v1/ws"

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

	mStatus := systray.AddMenuItem("ScrapeBlocker - Demo Banco de Bogotá v0.1 - Almacontact", "Almacontact")

	// Lanza la conexión WebSocket en una goroutine
	go func() {
		if err := services.ConnectAndKeepOpen(wsURL, user); err != nil {
			log.Printf("Error en WebSocket: %v", err)
		}
	}()

	// Lanza el monitor de procesos en otra goroutine
	log.Printf("Iniciando MonitorProcesses con config: %+v y usuario: %+v", config, user)
	go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock, user)

	<-mStatus.ClickedCh
}

func onExit() {
	log.Println("Saliendo de la aplicación systray...")
}
