package services

import (
	"fmt"
	"log"
	"os/user"
	"strings"
	"time"

	"github.com/EdwinPirajan/bloqueo.git/internal/core/domain"
	"github.com/gorilla/websocket"
)

// GetUser obtiene el usuario actual del sistema
func GetUser() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error al obtener el usuario actual: %v\n", err)
		return "", fmt.Errorf("error al obtener el usuario actual: %w", err)
	}

	// Extraer el nombre después de la barra inversa
	fullUsername := currentUser.Username
	splitUsername := strings.Split(fullUsername, "\\")
	username := splitUsername[len(splitUsername)-1] // Tomar la última parte después de '\'

	log.Printf("Usuario actual obtenido: %s\n", username)
	return username, nil
}

// CreateUser crea un objeto User basado en el nombre y asigna valores predeterminados
func CreateUser(name string) domain.User {
	return domain.User{
		ID:             0,
		Name:           name,
		Active:         true,
		LastConnection: time.Now(),
	}
}

// ConnectAndKeepOpen establece la conexión y la mantiene abierta
func ConnectAndKeepOpen(wsURL string) error {
	// Establecer conexión con el servidor WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	fmt.Println(wsURL)
	if err != nil {
		log.Printf("Error al conectar con el WebSocket: %v\n", err)
		return fmt.Errorf("error al conectar con el WebSocket: %w", err)
	}
	defer conn.Close()

	log.Println("Conexión establecida con el WebSocket")

	// Bucle para mantener la conexión abierta
	for {
		// Leer mensajes del servidor (si el servidor envía mensajes)
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error al leer mensaje del servidor: %v\n", err)
			break
		}

		log.Printf("Mensaje recibido del servidor: %s\n", message)
	}

	log.Println("Conexión cerrada")
	return nil
}
