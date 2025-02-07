package main

import (
	"fmt"
	"log"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func main() {
	cliente := "bac"
	baseWebSocketURL := "ws://localhost:8080/ws"

	// Obtener el usuario actual
	user, err := services.GetUser()
	if err != nil {
		log.Fatalf("No se pudo obtener el usuario: %v\n", err)
	}

	// Asignar la variable cliente al usuario
	user.Client = cliente

	// Construir la URL completa con el nombre de usuario y el cliente
	config, err := services.FetchConfiguration(cliente)
	if err != nil {
		fmt.Printf("Error obteniendo la configuración: %v\n", err)
	}
	wsURL := baseWebSocketURL + "/" + user.Name + "/" + cliente

	systemManager := services.NewWindowsSystemManager()

	systray.Run(onReady(systemManager, config, wsURL, &user), onExit)
}

func onReady(systemManager services.SystemManager, config services.ConfigResponse, wsURL string, user *domain.User) func() {
	return func() {
		iconData, err := services.GetIcon("resources/icono.ico")
		if err != nil {
			log.Printf("Error loading icon: %v\n", err)
			return
		}
		systray.SetIcon(iconData)
		systray.SetTitle("ScrapeBlocker")
		systray.SetTooltip("ScrapeBlocker")

		mStatus := systray.AddMenuItem("ScrapeBlocker - LATAM Airlines v0.2 - Almacontact", "Almacontact")

		// Lanzar conexión WebSocket en una goroutine
		go func() {
			err := services.ConnectAndKeepOpen(wsURL, user)
			if err != nil {
				log.Printf("Error en WebSocket: %v\n", err)
			}
		}()

		// Lanzar MonitorProcesses en otra goroutine
		log.Printf("Iniciando MonitorProcesses con config: %+v y usuario: %+v", config, user)
		go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock, user)

		<-mStatus.ClickedCh
	}
}

func onExit() {
	log.Println("Saliendo de la aplicación systray...")
}
