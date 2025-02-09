package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type LEDServer struct {
	controller     *PixelController
	patterns       map[string]Pattern
	currentPattern Pattern
	subscribers    []chan *PixelMap
	mu             sync.RWMutex
	pixelMap       *PixelMap
}

func NewLEDServer(controller *PixelController, pixelMap *PixelMap, patterns map[string]Pattern) *LEDServer {

	server := &LEDServer{
		controller:  controller,
		pixelMap:    pixelMap,
		patterns:    patterns,
		subscribers: make([]chan *PixelMap, 0),
	}

	if pattern, ok := patterns["rainbowCircle"]; ok {
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

	// pattern management routes
	mux.HandleFunc("GET /patterns", s.handleGetPatterns)
	mux.HandleFunc("PUT /patterns/{pattern}", s.handleUpdatePattern)

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

	ch := make(chan *PixelMap, 10) // Buffer of 10 to prevent blocking

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

	for {
		select {
		case pixelMap, ok := <-ch:
			if !ok {
				return
			}

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

	s.mu.Lock()

	fmt.Println("handling pattern update request")
	pattern, exists := s.patterns[patternName]
	if !exists {
		s.mu.Unlock()
		http.Error(w, "Pattern not found", http.StatusNotFound)
		return
	}

	s.currentPattern = pattern

	parameters := pattern.GetPatternUpdateRequest()

	err := json.NewDecoder(r.Body).Decode(&parameters)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	err = pattern.UpdateParameters(parameters.GetParameters())
	if err != nil {
		fmt.Println(err)
	}

	// update the pixel map with the new pattern
	s.controller.UpdatePattern(pattern)
	s.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}
