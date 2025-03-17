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
	BlockURLsInHosts(urlsToBlock []string) error
	NavigateBackToPreviousURLs() error
}

type TabInfo struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type chromeServiceImpl struct {
	conn         *websocket.Conn
	urlToMonitor string
	tabPrevURLs  map[string]string
}

func NewChromeService(urlToMonitor string) ChromeService {
	return &chromeServiceImpl{
		urlToMonitor: urlToMonitor,
		tabPrevURLs:  make(map[string]string),
	}
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
		err := s.conn.Close()
		if err != nil {
			fmt.Printf("Error al cerrar la conexión: %v\n", err)
		} else {
			fmt.Println("Conexión a Chrome cerrada")
		}
		s.conn = nil
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

func (s *chromeServiceImpl) BlockURLsInHosts(urlsToBlock []string) error {
	// Agregar las URLs al archivo hosts
	err := AddURLsToHostsFile(urlsToBlock)
	if err != nil {
		return err
	}

	// Obtener todas las pestañas abiertas
	url := "http://localhost:9222/json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error al conectar con Chrome DevTools: %v", err)
	}
	defer resp.Body.Close()

	var targets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return fmt.Errorf("error al decodificar los targets: %v", err)
	}

	// Recorrer todas las pestañas y aplicar la acción a las que coincidan
	for _, target := range targets {
		pageURL, ok := target["url"].(string)
		if !ok {
			continue
		}
		tabID, ok := target["id"].(string)
		if !ok {
			continue
		}

		// Verificar si la URL de la pestaña coincide con alguna de las URLs a bloquear
		for _, blockedURL := range urlsToBlock {
			if strings.Contains(pageURL, blockedURL) {
				wsURL, ok := target["webSocketDebuggerUrl"].(string)
				if !ok {
					fmt.Printf("No se pudo obtener wsURL para la pestaña con URL %s\n", pageURL)
					continue
				}

				// Almacenar la URL anterior
				s.tabPrevURLs[tabID] = pageURL

				// Conectar a la pestaña y aplicar la acción
				if err := s.applyActionToTab(wsURL); err != nil {
					fmt.Printf("Error al aplicar la acción a la pestaña con URL %s: %v\n", pageURL, err)
				} else {
					fmt.Printf("Acción aplicada a la pestaña con URL %s\n", pageURL)
				}

				// No necesitamos seguir verificando otras URLs para esta pestaña
				break
			}
		}
	}

	return nil
}

func (s *chromeServiceImpl) NavigateBackToPreviousURLs() error {
	// Obtener todas las pestañas abiertas
	url := "http://localhost:9222/json"
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error al conectar con Chrome DevTools: %v", err)
	}
	defer resp.Body.Close()

	var targets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		return fmt.Errorf("error al decodificar los targets: %v", err)
	}

	// Recorrer todas las pestañas y navegar de regreso a las URLs anteriores
	for _, target := range targets {
		tabID, ok := target["id"].(string)
		if !ok {
			continue
		}

		prevURL, exists := s.tabPrevURLs[tabID]
		if !exists {
			continue
		}

		wsURL, ok := target["webSocketDebuggerUrl"].(string)
		if !ok {
			fmt.Printf("No se pudo obtener wsURL para la pestaña con ID %s\n", tabID)
			continue
		}

		// Conectar a la pestaña y navegar de regreso
		if err := s.navigateToURLInTab(wsURL, prevURL); err != nil {
			fmt.Printf("Error al navegar de regreso en la pestaña con ID %s: %v\n", tabID, err)
		} else {
			fmt.Printf("Navegado de regreso a %s en la pestaña con ID %s\n", prevURL, tabID)
		}

		// Remover el tabID del mapa, ya que hemos navegado de regreso
		delete(s.tabPrevURLs, tabID)
	}

	return nil
}

func (s *chromeServiceImpl) navigateToURLInTab(wsURL string, urlToNavigate string) error {
	// Conectar al WebSocket de la pestaña
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("error al conectar al WebSocket: %v", err)
	}
	defer conn.Close()

	// Contador de IDs para las solicitudes
	var requestID int = 1

	// Habilitar el dominio Page
	request := map[string]interface{}{
		"id":     requestID,
		"method": "Page.enable",
	}
	requestID++

	if err := conn.WriteJSON(request); err != nil {
		return fmt.Errorf("error al enviar Page.enable: %v", err)
	}

	var response map[string]interface{}
	if err := conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("error al leer la respuesta de Page.enable: %v", err)
	}

	// Verificar si hay errores en la respuesta
	if errObj, exists := response["error"]; exists {
		return fmt.Errorf("error desde CDP en Page.enable: %v", errObj)
	}

	// Navegar a la URL anterior
	request = map[string]interface{}{
		"id":     requestID,
		"method": "Page.navigate",
		"params": map[string]interface{}{
			"url": urlToNavigate,
		},
	}
	requestID++

	if err := conn.WriteJSON(request); err != nil {
		return fmt.Errorf("error al enviar Page.navigate: %v", err)
	}

	if err := conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("error al leer la respuesta de Page.navigate: %v", err)
	}

	// Verificar si hay errores en la respuesta
	if errObj, exists := response["error"]; exists {
		return fmt.Errorf("error desde CDP en Page.navigate: %v", errObj)
	}

	fmt.Println("Página navegada exitosamente en la pestaña")
	return nil
}

func (s *chromeServiceImpl) applyActionToTab(wsURL string) error {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("error al conectar al WebSocket: %v", err)
	}
	defer conn.Close()

	var requestID int = 1

	request := map[string]interface{}{
		"id":     requestID,
		"method": "Page.enable",
	}
	requestID++

	if err := conn.WriteJSON(request); err != nil {
		return fmt.Errorf("error al enviar Page.enable: %v", err)
	}

	var response map[string]interface{}
	if err := conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("error al leer la respuesta de Page.enable: %v", err)
	}

	if errObj, exists := response["error"]; exists {
		return fmt.Errorf("error desde CDP en Page.enable: %v", errObj)
	}

	request = map[string]interface{}{
		"id":     requestID,
		"method": "Page.navigate",
		"params": map[string]interface{}{
			"url": "about:blank",
		},
	}
	requestID++

	if err := conn.WriteJSON(request); err != nil {
		return fmt.Errorf("error al enviar Page.navigate: %v", err)
	}

	if err := conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("error al leer la respuesta de Page.navigate: %v", err)
	}

	if errObj, exists := response["error"]; exists {
		return fmt.Errorf("error desde CDP en Page.navigate: %v", errObj)
	}

	fmt.Println("Página redirigida exitosamente en la pestaña")
	return nil
}
