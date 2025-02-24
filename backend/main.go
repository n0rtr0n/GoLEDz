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

	// left front leg
	pixels := buildMammothSegment(1, 1, 350, 500, 180, limb_sections)
	*pixels = append(*pixels, *buildMammothSegment(2, 1, 250, 500, 180, limb_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(3, 1, 150, 500, 180, limb_sections)...)

	// right front leg
	*pixels = append(*pixels, *buildMammothSegment(4, 1, 450, 490, 0, limb_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(5, 1, 550, 490, 0, limb_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(6, 1, 650, 490, 0, limb_sections)...)

	// left rear leg
	*pixels = append(*pixels, *buildMammothSegment(7, 1, 350, 190, 135, limb_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(8, 1, 270, 110, 135, limb_sections)...)

	// right rear leg
	*pixels = append(*pixels, *buildMammothSegment(9, 1, 450, 190, 45, limb_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(10, 1, 530, 110, 45, limb_sections)...)

	// torso
	torso_sections := []Section{
		sections["all"],
		sections["torso"],
	}
	*pixels = append(*pixels, *buildMammothSegment(11, 1, 400, 500, 90, torso_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(12, 1, 400, 400, 90, torso_sections)...)
	*pixels = append(*pixels, *buildMammothSegment(13, 1, 400, 300, 90, torso_sections)...)

	// head
	head_sections := []Section{
		sections["all"],
		sections["head"],
	}

	*pixels = append(*pixels, *buildMammothSegment(14, 1, 410, 550, 270, head_sections)...)

	// tusks
	tusk_sections := []Section{
		sections["all"],
		sections["tusks"],
	}
	*pixels = append(*pixels, *buildTuskSegment(15, 1, 350, 550, 225, tusk_sections)...)
	*pixels = append(*pixels, *buildTuskSegment(16, 1, 460, 545, 315, tusk_sections)...)

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

	// create controller with temp pattern
	controller := NewPixelController(
		universes,
		errorTracker,
		config.TargetFramesPerSecond,
		patterns["solidColor"], // temporary initial pattern
		&pixelMap,
		config.TransitionDuration,
	)

	// now register patterns with controller

	// set the real initial pattern
	controller.SetPattern(patterns["spiral"])

	modes := registerModes(&pixelMap, patterns)

	// finally create server
	server := NewLEDServer(controller, &pixelMap, patterns, modes, &ServerConfig{
		TransitionDuration: config.TransitionDuration,
		TransitionEnabled:  config.TransitionEnabled,
	})

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

func registerModes(pixelMap *PixelMap, patterns map[string]Pattern) map[string]PatternMode {
	modes := make(map[string]PatternMode)

	randomMode := &RandomMode{
		pixelMap: pixelMap,
		patterns: patterns,
		Label:    "Random",
		Parameters: RandomParameters{
			SwitchInterval: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   60.0,
				Value: 15.0,
				Type:  "float",
			},
		},
	}

	modes[randomMode.GetName()] = randomMode
	return modes
}
