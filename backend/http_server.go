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
	colorMasks        map[string]ColorMaskPattern
	options           Options
}

type ServerConfig struct {
	Options Options
}

type PatternsResponse struct {
	Patterns   map[string]PatternInfo   `json:"patterns"`
	ColorMasks map[string]ColorMaskInfo `json:"colorMasks"`
	Options    Options                  `json:"options"`
}

type ColorMaskInfo struct {
	Label      string               `json:"label"`
	Parameters AdjustableParameters `json:"parameters"`
}

type PatternInfo struct {
	Label      string               `json:"label"`
	Parameters AdjustableParameters `json:"parameters"`
}

func NewLEDServer(controller *PixelController, pixelMap *PixelMap, patterns map[string]Pattern, config *ServerConfig) *LEDServer {
	if config == nil {
		config = &ServerConfig{
			Options: *DefaultOptions(),
		}
	}

	server := &LEDServer{
		controller:  controller,
		pixelMap:    pixelMap,
		patterns:    patterns,
		colorMasks:  registerColorMasks(),
		subscribers: make([]chan *PixelMap, 0),
		options:     config.Options,
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

	// color mask management
	mux.HandleFunc("PUT /colorMasks/{mask}", s.handleSetColorMask)
	mux.HandleFunc("DELETE /colorMasks", s.handleDisableColorMask)

	// options endpoints
	mux.HandleFunc("GET /options", s.handleGetOptions)
	mux.HandleFunc("PUT /options/{option}", s.handleUpdateOption)

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
	response := struct {
		Patterns   map[string]interface{} `json:"patterns"`
		ColorMasks map[string]interface{} `json:"colorMasks"`
		Options    Options                `json:"options"`
	}{
		Patterns:   make(map[string]interface{}),
		ColorMasks: make(map[string]interface{}),
		Options:    s.options,
	}

	// add patterns to response
	for name, pattern := range s.patterns {
		patternResponse := struct {
			Label      string               `json:"label"`
			Parameters AdjustableParameters `json:"parameters"`
		}{
			Label:      pattern.GetLabel(),
			Parameters: pattern.GetPatternUpdateRequest().GetParameters(),
		}
		response.Patterns[name] = patternResponse
	}

	// add color masks to response
	for name, mask := range s.colorMasks {
		// create a response that includes both the mask update request and the label
		maskResponse := struct {
			Label      string               `json:"label"`
			Parameters AdjustableParameters `json:"parameters"`
		}{
			Label:      mask.GetLabel(), // Add the label
			Parameters: mask.GetPatternUpdateRequest().GetParameters(),
		}
		response.ColorMasks[name] = maskResponse
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

func (s *LEDServer) handleGetOptions(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.options)
}

func (s *LEDServer) handleUpdateOption(w http.ResponseWriter, r *http.Request) {
	optionID := r.PathValue("option")

	var valueMap map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&valueMap); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	value, exists := valueMap["value"]
	if !exists {
		http.Error(w, "Request must include a 'value' field", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// update the option
	if err := s.options.SetOption(optionID, value); err != nil {
		if err == ErrOptionNotFound {
			http.Error(w, "Unknown option: "+optionID, http.StatusBadRequest)
		} else if err == ErrInvalidOptionValue {
			http.Error(w, "Invalid value type for option: "+optionID, http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// update the controller with the new options
	s.controller.UpdateOptions(s.options)

	// return the updated options
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.options)
}
