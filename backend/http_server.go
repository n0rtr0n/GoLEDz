package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type LEDServer struct {
	controller        *PixelController
	patterns          map[string]Pattern
	currentPattern    Pattern
	subscribers       []chan *PixelMap
	mu                sync.RWMutex
	pixelMap          *PixelMap
	defaultTransition TransitionConfig
	modes             map[string]PatternMode
}

type ServerConfig struct {
	TransitionDuration time.Duration
	TransitionEnabled  bool
}

func NewLEDServer(controller *PixelController, pixelMap *PixelMap, patterns map[string]Pattern, modes map[string]PatternMode, config *ServerConfig) *LEDServer {
	if config == nil {
		config = &ServerConfig{
			TransitionDuration: 2 * time.Second,
			TransitionEnabled:  true,
		}
	}

	server := &LEDServer{
		controller:  controller,
		pixelMap:    pixelMap,
		patterns:    patterns,
		modes:       modes,
		subscribers: make([]chan *PixelMap, 0),
		defaultTransition: TransitionConfig{
			Duration: config.TransitionDuration,
			Enabled:  config.TransitionEnabled,
		},
	}

	if pattern, ok := patterns["spiral"]; ok {
		server.currentPattern = pattern
	} else {
		// get first available pattern
		for _, pattern := range patterns {
			server.currentPattern = pattern
			break
		}
	}

	return server
}

func (s *LEDServer) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// websocket route for visualizer
	mux.HandleFunc("GET /socket", s.handleWebSocket)

	// pattern management
	mux.HandleFunc("GET /patterns", s.handleGetPatterns)
	mux.HandleFunc("PUT /patterns/{pattern}", s.handleUpdatePattern)

	// health check
	mux.HandleFunc("GET /health", s.handleHealthCheck)

	// transition config
	mux.HandleFunc("PUT /transition", s.handleUpdateTransition)

	// mode management
	mux.HandleFunc("PUT /modes/{mode}", s.handleSetMode)
	mux.HandleFunc("DELETE /modes/current", s.handleDisableMode)

	return mux
}

func (s *LEDServer) Start(address string) error {
	s.controller.SetUpdateCallback(func(pixelMap *PixelMap) {
		s.NotifySubscribers()
	})

	server := &http.Server{
		Addr:    address,
		Handler: s.SetupRoutes(),
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (s *LEDServer) NotifySubscribers() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, ch := range s.subscribers {
		select {
		case ch <- s.pixelMap:
		default:
			// channel is full or blocked, skip this update for this subscriber
			log.Println("Skipped update for blocked subscriber")
		}
	}
}

func (s *LEDServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("establishing websocket connection handler")

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO: implement origin check
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	ch := make(chan *PixelMap, 10) // buffer of 10 to prevent blocking

	s.mu.Lock()
	s.subscribers = append(s.subscribers, ch)
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		for i, subscriber := range s.subscribers {
			if subscriber == ch {
				s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
		close(ch)
	}()

	for pixelMap := range ch {
		data, err := pixelMap.toJSON()
		if err != nil {
			log.Printf("error marshaling pixel map: %v", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("error writing to websocket: %v", err)
			return
		}
	}
}

func (s *LEDServer) handleGetPatterns(w http.ResponseWriter, r *http.Request) {
	type AllPatternsRequest struct {
		Patterns Patterns `json:"patterns"`
	}

	patternsReq := AllPatternsRequest{
		Patterns: s.patterns,
	}

	jsonData, err := json.Marshal(patternsReq)
	if err != nil {
		fmt.Printf("could not marshal json: %s\n", err)
		return
	}

	fmt.Fprint(w, string(jsonData))
}

func (s *LEDServer) handleUpdatePattern(w http.ResponseWriter, r *http.Request) {
	patternName := r.PathValue("pattern")

	pattern, exists := s.patterns[patternName]
	if !exists {
		http.Error(w, "Pattern not found", http.StatusNotFound)
		return
	}

	parameters := pattern.GetPatternUpdateRequest()
	if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := pattern.UpdateParameters(parameters.GetParameters()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// disable any active mode before setting pattern
	if s.controller.currentMode != nil {
		s.controller.currentMode.Stop()
		s.controller.currentMode = nil
	}

	s.controller.SetPattern(pattern)
	w.WriteHeader(http.StatusOK)
}

func (s *LEDServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Healthy")
}

func (s *LEDServer) handleUpdateTransition(w http.ResponseWriter, r *http.Request) {
	var configReq TransitionConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&configReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration := time.Duration(configReq.Duration) * time.Millisecond

	s.defaultTransition = TransitionConfig{
		Duration: duration,
		Enabled:  configReq.Enabled,
	}

	s.controller.SetTransitionDuration(duration)

	w.WriteHeader(http.StatusOK)
}

func (s *LEDServer) handleSetMode(w http.ResponseWriter, r *http.Request) {
	modeName := r.PathValue("mode")

	mode, exists := s.modes[modeName]
	if !exists {
		http.Error(w, "Mode not found", http.StatusNotFound)
		return
	}

	// only try to decode parameters if there's a request body
	if r.ContentLength > 0 {
		parameters := mode.GetPatternUpdateRequest()
		if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := mode.UpdateParameters(parameters.GetParameters()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s.controller.SetPattern(mode)
	w.WriteHeader(http.StatusOK)
}

func (s *LEDServer) handleDisableMode(w http.ResponseWriter, r *http.Request) {
	if s.controller.currentMode != nil {
		s.controller.currentMode.Stop()
		s.controller.currentMode = nil
	}
	w.WriteHeader(http.StatusOK)
}
