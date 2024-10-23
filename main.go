package main

import (
	"fmt"
	"log"
	"os"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/services"
	"github.com/getlantern/systray"
)

func main() {
	cliente := "bac"

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

		mStatus := systray.AddMenuItem("ScrapeBlocker - LATAM Arilines v0.2 - Almacontact", "Almacontact")

		userName := os.Getenv("USERNAME")

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
