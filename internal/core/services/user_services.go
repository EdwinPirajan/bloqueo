package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os/user"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
	"github.com/gorilla/websocket"
)

func GetUser() (domain.User, error) {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error al obtener el usuario actual: %v\n", err)
		return domain.User{}, fmt.Errorf("error al obtener el usuario actual: %w", err)
	}

	fullUsername := currentUser.Username
	splitUsername := strings.Split(fullUsername, "\\")
	username := splitUsername[len(splitUsername)-1]
	user := domain.User{
		Name:           username,
		Active:         true,
		LastConnection: time.Now(),
	}

	log.Printf("Usuario actual obtenido: %+v\n", user)
	return user, nil
}

func ConnectAndKeepOpen(wsURL string, user *domain.User) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("Error al conectar con el WebSocket: %v\n", err)
		return fmt.Errorf("error al conectar con el WebSocket: %w", err)
	}
	defer conn.Close()

	log.Println("Conexión establecida con el WebSocket")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error al leer mensaje del servidor: %v\n", err)
			break
		}

		log.Printf("Mensaje recibido del servidor: %s\n", message)

		err = handleWebSocketMessage(message, user)
		if err != nil {
			log.Printf("Error procesando el mensaje del WebSocket: %v\n", err)
		}
	}

	log.Println("Conexión cerrada")
	return nil
}

func handleWebSocketMessage(message []byte, user *domain.User) error {
	// Definimos la estructura del mensaje WS, incluyendo un campo opcional Data
	type WSMessage struct {
		Type        string        `json:"type"`
		ActiveUsers []domain.User `json:"active_users,omitempty"`
		// No incluimos Data si usamos la señal para hacer el fetch, pero podrías incluirlo si lo deseas:
		// Data json.RawMessage `json:"data,omitempty"`
	}

	var wsMessage WSMessage
	if err := json.Unmarshal(message, &wsMessage); err != nil {
		return fmt.Errorf("error deserializando el mensaje WebSocket: %v", err)
	}

	switch wsMessage.Type {
	case "update":
		// Procesamos el mensaje de actualización de usuarios
		for _, activeUser := range wsMessage.ActiveUsers {
			if activeUser.Name == user.Name {
				user.Active = activeUser.Active
				log.Printf("Estado del usuario '%s' actualizado a: %v\n", user.Name, user.Active)
				break
			}
		}
	case "refresh", "configuracion":
		// Mensaje que indica que se debe refrescar la configuración
		log.Printf("Mensaje de actualización de configuración recibido: %s\n", wsMessage.Type)
		// Se asume que existe la función FetchConfiguration que hace el fetch vía HTTP
		config, err := FetchConfiguration(user.Client)
		if err != nil {
			log.Printf("Error al obtener nueva configuración: %v", err)
		} else {
			// Se actualiza el store global usando sync.Mutex (definido en configstore)
			SetCurrentConfig(config)
			log.Printf("Nueva configuración actualizada: %+v", config)
		}
	default:
		log.Printf("Tipo de mensaje desconocido: %s", wsMessage.Type)
	}

	return nil
}
