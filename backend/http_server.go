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
	colorMasks        map[string]ColorMaskPattern
}

type ServerConfig struct {
	TransitionDuration time.Duration
	TransitionEnabled  bool
}

type PatternsResponse struct {
	Patterns   map[string]PatternInfo   `json:"patterns"`
	ColorMasks map[string]ColorMaskInfo `json:"colorMasks"`
}

type ColorMaskInfo struct {
	Label      string               `json:"label"`
	Parameters AdjustableParameters `json:"parameters"`
}

type PatternInfo struct {
	Label      string               `json:"label"`
	Parameters AdjustableParameters `json:"parameters"`
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
		colorMasks:  registerColorMasks(),
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

	// color mask management
	mux.HandleFunc("PUT /colorMasks/{mask}", s.handleSetColorMask)
	mux.HandleFunc("DELETE /colorMasks", s.handleDisableColorMask)

	return mux
}

// Permissive CORS for now
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *LEDServer) Start(address string) error {
	s.controller.SetUpdateCallback(func(pixelMap *PixelMap) {
		s.NotifySubscribers()
	})

	server := &http.Server{
		Addr:    address,
		Handler: CORS(s.SetupRoutes()),
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
	patterns := make(map[string]PatternInfo)
	for name, pattern := range s.patterns {
		patterns[name] = PatternInfo{
			Label:      pattern.GetLabel(),
			Parameters: pattern.GetPatternUpdateRequest().GetParameters(),
		}
	}

	colorMasks := make(map[string]ColorMaskInfo)
	for name, mask := range s.colorMasks {
		colorMasks[name] = ColorMaskInfo{
			Label:      mask.GetLabel(),
			Parameters: mask.GetPatternUpdateRequest().GetParameters(),
		}
	}

	response := PatternsResponse{
		Patterns:   patterns,
		ColorMasks: colorMasks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *LEDServer) handleUpdatePattern(w http.ResponseWriter, r *http.Request) {
	patternName := r.PathValue("pattern")

	// get the current pattern's update request structure
	parameters := s.controller.patterns[patternName].GetPatternUpdateRequest()
	if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// disable any active mode before setting pattern
	if s.controller.currentMode != nil {
		s.controller.currentMode.Stop()
		s.controller.currentMode = nil
	}

	// handles parameter updates without transition
	if err := s.controller.UpdatePattern(patternName, parameters); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

func (s *LEDServer) handleSetColorMask(w http.ResponseWriter, r *http.Request) {
	maskName := r.PathValue("mask")

	mask, exists := s.colorMasks[maskName]
	if !exists {
		http.Error(w, "Color mask not found", http.StatusNotFound)
		return
	}

	// only try to decode parameters if there's a request body
	if r.ContentLength > 0 {
		parameters := mask.GetPatternUpdateRequest()
		if err := json.NewDecoder(r.Body).Decode(&parameters); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := mask.UpdateParameters(parameters.GetParameters()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s.controller.SetColorMask(mask)
	w.WriteHeader(http.StatusOK)
}

func (s *LEDServer) handleDisableColorMask(w http.ResponseWriter, r *http.Request) {
	s.controller.SetColorMask(nil)
	w.WriteHeader(http.StatusOK)
}
