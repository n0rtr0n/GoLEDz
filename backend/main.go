package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Patterns map[string]Pattern

var config *Config

func main() {

	config = loadConfig()

	// websocket connections. will no longer block if we're not connected to a websocket
	// var subscribers []chan *PixelMa
	ch := make(chan *PixelMap)
	defer close(ch)

	// left front leg
	pixels := buildLegSegment(1, 1, 350, 500, 180)
	*pixels = append(*pixels, *buildLegSegment(2, 1, 250, 500, 180)...)
	*pixels = append(*pixels, *buildLegSegment(3, 1, 150, 500, 180)...)

	// right front leg
	*pixels = append(*pixels, *buildLegSegment(4, 1, 450, 490, 0)...)
	*pixels = append(*pixels, *buildLegSegment(5, 1, 550, 490, 0)...)
	*pixels = append(*pixels, *buildLegSegment(6, 1, 650, 490, 0)...)

	// left rear leg
	*pixels = append(*pixels, *buildLegSegment(7, 1, 350, 190, 135)...)
	*pixels = append(*pixels, *buildLegSegment(8, 1, 270, 110, 135)...)

	// right rear leg
	*pixels = append(*pixels, *buildLegSegment(9, 1, 450, 190, 45)...)
	*pixels = append(*pixels, *buildLegSegment(10, 1, 530, 110, 45)...)

	// body
	*pixels = append(*pixels, *buildLegSegment(11, 1, 400, 500, 90)...)
	*pixels = append(*pixels, *buildLegSegment(12, 1, 400, 400, 90)...)
	*pixels = append(*pixels, *buildLegSegment(13, 1, 400, 300, 90)...)

	// head
	*pixels = append(*pixels, *buildLegSegment(14, 1, 410, 550, 270)...)

	// tusks
	*pixels = append(*pixels, *buildTuskSegment(15, 1, 350, 550, 225)...)
	*pixels = append(*pixels, *buildTuskSegment(16, 1, 460, 545, 315)...)

	pixelMap := PixelMap{
		// pixels: buildPixelGrid(),
		pixels: pixels,
		// pixels: build2ChannelsOfPixels(),
	}

	handler, err := NewSACNHandler()
	if err != nil {
		log.Fatal(err)
	}

	universeNumbers := []uint16{1}
	handler.Setup(universeNumbers, config.ControllerAddress)

	universes, errorTracker := handler.GetUniverses(), handler.GetErrorTracker()
	for _, universe := range universes {
		defer close(universe)
	}

	patterns := registerPatterns(&pixelMap)
	if len(patterns) == 0 {
		log.Fatal("no patterns registered")
	}

	initialPattern := patterns["rainbowDiagonal"]

	controller := NewPixelController(
		universes,
		errorTracker,
		config.TargetFramesPerSecond,
		initialPattern,
		&pixelMap,
	)
	server := NewLEDServer(controller, &pixelMap, patterns)

	// start the web server first
	address := fmt.Sprintf("%v:%v", config.HostAddress, config.HostPort)
	if err := server.Start(address); err != nil {
		log.Fatal(err)
	}

	// then start the controller
	if err := controller.Start(&pixelMap); err != nil {
		log.Fatal(err)
	}

	// wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// cleanup
	controller.Stop()
}
