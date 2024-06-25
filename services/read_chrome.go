package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type ChromeService interface {
	Connect() error
	GetFullPageHTML() (string, error)
	Close()
}

type chromeServiceImpl struct {
	conn         *websocket.Conn
	urlToMonitor string
}

func NewChromeService(urlToMonitor string) ChromeService {
	return &chromeServiceImpl{urlToMonitor: urlToMonitor}
}

func (s *chromeServiceImpl) Connect() error {
	url := "http://localhost:9222/json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error connecting to Chrome DevTools: %v", err)
	}
	defer resp.Body.Close()

	var targets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return fmt.Errorf("error decoding targets: %v", err)
	}

	var wsURL string
	for _, target := range targets {
		if target["type"] == "page" && strings.Contains(target["url"].(string), s.urlToMonitor) {
			wsURL = target["webSocketDebuggerUrl"].(string)
			break
		}
	}

	if wsURL == "" {
		return fmt.Errorf("no suitable tab found with URL containing: %s", s.urlToMonitor)
	}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("error connecting to WebSocket: %v", err)
	}
	s.conn = conn
	return nil
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