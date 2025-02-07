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

// GetUser obtiene el usuario actual del sistema y retorna un objeto domain.User.
func GetUser() (domain.User, error) {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error al obtener el usuario actual: %v\n", err)
		return domain.User{}, fmt.Errorf("error al obtener el usuario actual: %w", err)
	}

	fullUsername := currentUser.Username
	splitUsername := strings.Split(fullUsername, "\\")
	username := splitUsername[len(splitUsername)-1] // Última parte después de '\'

	user := domain.User{
		Name:           username,
		Active:         true,
		LastConnection: time.Now(),
		// Se asignará Client en main
	}

	log.Printf("Usuario actual obtenido: %+v\n", user)
	return user, nil
}

// ConnectAndKeepOpen establece la conexión WebSocket y la mantiene abierta.
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

		// Procesar el mensaje recibido
		err = handleWebSocketMessage(message, user)
		if err != nil {
			log.Printf("Error procesando el mensaje del WebSocket: %v\n", err)
		}
	}

	log.Println("Conexión cerrada")
	return nil
}

// handleWebSocketMessage procesa los mensajes del WebSocket y actualiza el estado del usuario.
func handleWebSocketMessage(message []byte, user *domain.User) error {
	// Estructura para deserializar el mensaje
	type WebSocketMessage struct {
		Type        string        `json:"type"`
		ActiveUsers []domain.User `json:"active_users"`
	}

	var wsMessage WebSocketMessage
	err := json.Unmarshal(message, &wsMessage)
	if err != nil {
		return fmt.Errorf("error deserializando el mensaje WebSocket: %v", err)
	}

	// Buscar al usuario correspondiente en la lista de usuarios activos.
	for _, activeUser := range wsMessage.ActiveUsers {
		if activeUser.Name == user.Name {
			user.Active = activeUser.Active // Actualizar el estado del usuario
			log.Printf("Estado del usuario '%s' actualizado a: %v\n", user.Name, user.Active)
			break
		}
	}

	return nil
}
