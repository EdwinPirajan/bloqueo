package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type ChromeService interface {
	Connect() error
	GetFullPageHTML() (string, error)
	Close()
}

type chromeServiceImpl struct {
	conn          *websocket.Conn
	urlToMonitor  string
	systemManager SystemManager
}

func NewChromeService(urlToMonitor string, systemManager SystemManager) ChromeService {
	return &chromeServiceImpl{urlToMonitor: urlToMonitor, systemManager: systemManager}
}

func (s *chromeServiceImpl) Connect() error {
	url := "http://localhost:9222/json"
	var resp *http.Response
	var err error

	// Retry logic for connection
	for i := 0; i < 5; i++ {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		log.Printf("Error connecting to Chrome DevTools, retrying... (%d/5)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("error connecting to Chrome DevTools: %v", err)
	}
	defer resp.Body.Close()

	var targets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return fmt.Errorf("error decoding targets: %v", err)
	}

	for _, target := range targets {
		if target["type"] == "page" {
			wsURL := target["webSocketDebuggerUrl"].(string)
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err == nil {
				s.conn = conn
				log.Printf("Connected to WebSocket for target %s\n", target["url"])
				return nil
			}
			log.Printf("Error connecting to WebSocket for target %s: %v\n", target["url"], err)
		}
	}

	return fmt.Errorf("no suitable tab found")
}

func (s *chromeServiceImpl) Close() {
	if s.conn != nil {
		s.conn.Close()
		fmt.Println("Conexión a Chrome cerrada")
	}
}

func (s *chromeServiceImpl) GetFullPageHTML() (string, error) {
	err := s.Connect()
	if err != nil {
		return "", err
	}
	defer s.Close()

	request := map[string]interface{}{
		"id":     1,
		"method": "DOM.getDocument",
	}

	if err := s.conn.WriteJSON(request); err != nil {
		return "", fmt.Errorf("error sending command: %v", err)
	}

	var response map[string]interface{}
	if err := s.conn.ReadJSON(&response); err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	result, ok := response["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: %v", response)
	}

	root, ok := result["root"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected root format: %v", result)
	}

	rootNodeID, ok := root["nodeId"].(float64)
	if !ok {
		return "", fmt.Errorf("unexpected nodeId format: %v", root)
	}

	request = map[string]interface{}{
		"id":     2,
		"method": "DOM.getOuterHTML",
		"params": map[string]interface{}{
			"nodeId": rootNodeID,
		},
	}

	if err := s.conn.WriteJSON(request); err != nil {
		return "", fmt.Errorf("error sending command: %v", err)
	}

	if err := s.conn.ReadJSON(&response); err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	result, ok = response["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format: %v", response)
	}

	htmlContent, ok := result["outerHTML"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected outerHTML format: %v", result)
	}

	fmt.Println("Contenido HTML de la página obtenido exitosamente")
	return htmlContent, nil
}
