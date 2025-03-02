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
	// var subscribers []chan *PixelMap
	ch := make(chan *PixelMap)
	defer close(ch)

	sections := map[string]Section{
		"all":   {name: "all", label: "All"},
		"limbs": {name: "limbs", label: "Limbs"},
		"torso": {name: "torso", label: "Torso"},
		"head":  {name: "head", label: "head"},
		"tusks": {name: "tusks", label: "tusks"},
	}

	limb_sections := []Section{
		sections["all"],
		sections["limbs"],
	}

	// define pixel types for different segments
	limbPixelType := PixelRGBW
	torsoPixelType := PixelRGBW
	headPixelType := PixelRGBW
	tuskPixelType := PixelRGB

	// left front leg
	pixels := buildMammothSegment(1, 1, 350, 500, 180, limb_sections, limbPixelType, RGB)
	*pixels = append(*pixels, *buildMammothSegment(1, 21, 250, 500, 180, limb_sections, limbPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(1, 41, 150, 500, 180, limb_sections, limbPixelType, RGB)...)

	// right front leg
	*pixels = append(*pixels, *buildMammothSegment(2, 1, 450, 490, 0, limb_sections, limbPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(2, 21, 550, 490, 0, limb_sections, limbPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(2, 41, 650, 490, 0, limb_sections, limbPixelType, RGB)...)

	// left rear leg
	*pixels = append(*pixels, *buildMammothSegment(3, 1, 350, 190, 135, limb_sections, limbPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(3, 21, 270, 110, 135, limb_sections, limbPixelType, RGB)...)

	// right rear leg
	*pixels = append(*pixels, *buildMammothSegment(4, 1, 450, 190, 45, limb_sections, limbPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(4, 21, 530, 110, 45, limb_sections, limbPixelType, RGB)...)

	// torso
	torso_sections := []Section{
		sections["all"],
		sections["torso"],
	}
	*pixels = append(*pixels, *buildMammothSegment(5, 1, 400, 500, 90, torso_sections, torsoPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(5, 21, 400, 400, 90, torso_sections, torsoPixelType, RGB)...)
	*pixels = append(*pixels, *buildMammothSegment(5, 41, 400, 300, 90, torso_sections, torsoPixelType, RGB)...)

	// head
	head_sections := []Section{
		sections["all"],
		sections["head"],
	}

	*pixels = append(*pixels, *buildMammothSegment(6, 1, 410, 550, 270, head_sections, headPixelType, RGB)...)

	// tusks
	tusk_sections := []Section{
		sections["all"],
		sections["tusks"],
	}
	*pixels = append(*pixels, *buildTuskSegment(31, 1, 350, 550, 225, tusk_sections, tuskPixelType, BRG)...)
	*pixels = append(*pixels, *buildTuskSegment(32, 1, 460, 545, 315, tusk_sections, tuskPixelType, BRG)...)

	pixelMap := PixelMap{
		// pixels: buildPixelGrid(),
		pixels: pixels,
		// pixels: build2ChannelsOfPixels(),
	}

	handler, err := NewSACNHandler()
	if err != nil {
		log.Fatal(err)
	}

	universeNumbers := []uint16{1, 2, 3, 4, 5, 6, 31, 32}
	handler.Setup(universeNumbers, config.ControllerAddress)

	// verify universes are working
	if err := handler.VerifyUniverses(); err != nil {
		log.Printf("Warning: Failed to verify universes: %v", err)
	}

	universes, errorTracker := handler.GetUniverses(), handler.GetErrorTracker()
	for _, universe := range universes {
		defer close(universe)
	}

	// now register patterns with controller
	patterns := registerPatterns(&pixelMap)
	if len(patterns) == 0 {
		log.Fatal("no patterns registered")
	}

	// Create default options
	options := DefaultOptions()

	// Create controller with initial pattern
	controller := NewPixelController(
		universes,
		errorTracker,
		config.TargetFramesPerSecond,
		patterns["maskOnly"],
		&pixelMap,
		*options,
	)

	// now register modes with server
	modes := registerModes(&pixelMap, patterns)

	// Create server config
	serverConfig := &ServerConfig{
		Options: *options,
	}

	// Create server
	server := NewLEDServer(controller, &pixelMap, patterns, modes, serverConfig)

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
