package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func main() {
	// cliente := "bac"
	baseWebSocketURL := "ws://localhost:8080/ws"

	// Obtener el usuario actual
	username, err := services.GetUser()
	if err != nil {
		log.Fatalf("No se pudo obtener el usuario: %v\n", err)
	}

	// Construir la URL completa con el nombre de usuario
	wsURL := baseWebSocketURL + "/" + username

	// Establecer conexión WebSocket y mantenerla abierta
	err = services.ConnectAndKeepOpen(wsURL)
	if err != nil {
		log.Fatalf("Error al conectar con el WebSocket: %v\n", err)
	}

	log.Println("Conexión finalizada")

	// log.Println("Proceso completado con éxito")
	// config, err := services.FetchConfiguration(cliente)
	// if err != nil {
	// 	fmt.Printf("Error obteniendo la configuración: %v\n", err)
	// }

	// systemManager := services.NewWindowsSystemManager()

	// systray.Run(onReady(systemManager, config), onExit)
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

		mStatus := systray.AddMenuItem("ScrapeBlocker - LATAM Arilines v0.2 - Almacontact", "Almacontact")

		userName := os.Getenv("USERNAME")

		currentUser, err := user.Current()
		if err != nil {
			fmt.Printf("Error al obtener el usuario: %v\n", err)
			return
		}

		fmt.Printf("Usuario de red: %s\n", currentUser.Username) // Incluye dominio si está disponible (e.g., DOMAIN\\User)
		fmt.Printf("Directorio de inicio: %s\n", currentUser.HomeDir)

		go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock, userName)

		<-mStatus.ClickedCh
	}
}

func onExit() {
}

// package main

// import (
// 	"log"
// 	"os"

// 	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
// 	"github.com/getlantern/systray"
// )

// func main() {
// 	// Configura el logger para que imprima en stderr
// 	log.SetOutput(os.Stderr)

// 	cliente := "latam"

// 	config, err := services.FetchConfiguration(cliente)
// 	if err != nil {
// 		log.Printf("Error obteniendo la configuración: %v\n", err) // Usar log en vez de fmt
// 	}

// 	systemManager := services.NewWindowsSystemManager()

// 	// No uses paréntesis en 'onReady' para pasar la función
// 	systray.Run(onReady(systemManager, config), onExit)
// }

// func onReady(systemManager services.SystemManager, config services.ConfigResponse) func() {
// 	return func() {
// 		iconData, err := services.GetIcon("resources/icono.ico")
// 		if err != nil {
// 			log.Printf("Error loading icon: %v\n", err)
// 			return
// 		}
// 		systray.SetIcon(iconData)
// 		systray.SetTitle("ScrapeBlocker")
// 		systray.SetTooltip("ScrapeBlocker")

// 		// Añadir log para depurar clics en el menú
// 		mStatus := systray.AddMenuItem("ScrapeBlocker - LATAM Arilines v0.2 - Almacontact", "Almacontact")
// 		log.Println("Menu item creado: ScrapeBlocker - LATAM Airlines")

// 		// Obtener el nombre de usuario
// 		userName := os.Getenv("USERNAME")
// 		log.Printf("Usuario actual: %s\n", userName)

// 		// Iniciar la monitorización de procesos
// 		go services.MonitorProcesses(systemManager, config.ProcessesToMonitor, config.UrlsToBlock, userName)

// 		// Esperar a que el usuario haga clic en el menú
// 		<-mStatus.ClickedCh
// 		log.Println("Menu item clickeado. Saliendo...")
// 	}
// }

// func onExit() {
// 	log.Println("Saliendo de la aplicación systray...")
// }
