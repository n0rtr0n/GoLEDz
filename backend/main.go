package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Patterns map[string]Pattern

var config *Config

func main() {

	config = loadConfig()

	// websocket connections. will no longer block if we're not connected to a websocket
	var subscribers []chan *PixelMap
	ch := make(chan *PixelMap)
	defer close(ch)

	// left front leg
	pixels := buildLegSegment(1, 1, 350, 500, 180)
	*pixels = append(*pixels, *buildLegSegment(1, 2, 250, 500, 180)...)
	*pixels = append(*pixels, *buildLegSegment(1, 3, 150, 500, 180)...)

	// right front leg
	*pixels = append(*pixels, *buildLegSegment(1, 4, 450, 490, 0)...)
	*pixels = append(*pixels, *buildLegSegment(1, 5, 550, 490, 0)...)
	*pixels = append(*pixels, *buildLegSegment(1, 6, 650, 490, 0)...)

	// left rear leg
	*pixels = append(*pixels, *buildLegSegment(1, 7, 350, 190, 135)...)
	*pixels = append(*pixels, *buildLegSegment(1, 8, 270, 110, 135)...)

	// right rear leg
	*pixels = append(*pixels, *buildLegSegment(1, 9, 450, 190, 45)...)
	*pixels = append(*pixels, *buildLegSegment(1, 10, 530, 110, 45)...)

	// body
	*pixels = append(*pixels, *buildLegSegment(1, 11, 400, 500, 90)...)
	*pixels = append(*pixels, *buildLegSegment(1, 12, 400, 400, 90)...)
	*pixels = append(*pixels, *buildLegSegment(1, 13, 400, 300, 90)...)

	// head
	*pixels = append(*pixels, *buildLegSegment(1, 14, 410, 550, 270)...)

	// tusks
	*pixels = append(*pixels, *buildTuskSegment(1, 15, 350, 550, 225)...)
	*pixels = append(*pixels, *buildTuskSegment(1, 16, 460, 545, 315)...)

	pixelMap := PixelMap{
		// pixels: buildPixelGrid(),
		pixels: pixels,
		//pixels: build2ChannelsOfPixels(),
	}

	patterns := registerPatterns(&pixelMap)

	currentPattern := patterns["rainbowDiagonal"]

	universes := setupSACN()
	for _, universe := range universes {
		defer close(universe)
	}

	/*
		we're currently using the paradigm of DMX over ethernet, so we think about the world
		in terms of universes == channels. for that reason, we're going to denormalize our pixel
		map a little bit by creating a map of pointers to pixels. this will save us a ton of compute
		cycles when we go to send the data over the wire, because this map will be the ordered
		representation by universe/channel of each pixel position, eliminating the need for
		expensive lookups
	*/
	pixelsByUniverse := make(map[uint16][]*Pixel)

	for i, pixel := range *pixelMap.pixels {
		pixelsByUniverse[pixel.universe] = append(pixelsByUniverse[pixel.universe], &(*pixelMap.pixels)[i])
	}

	/*
		our primarily loop. effectively, we want to parse input, update the pixel map, and display
		the new map. for now, this will be a crude and brute force implementation where we send
		every pixel for every frame
	*/

	updateInterval := time.Second / time.Duration(config.TargetFramesPerSecond)
	updateTicker := time.Tick(updateInterval)

	updateWithTimeout := func(timeout time.Duration) {

		done := make(chan bool)

		go func() {
			// update pixel map
			currentPattern.Update()

			for i, universe := range universes {
				bytes := make([]byte, 512)
				for _, pixel := range pixelsByUniverse[i] {
					// fmt.Println(pixel)
					pos := pixel.channelPosition - 1
					// fmt.Println(pos)
					startIndex := pos * 3
					endIndex := startIndex + 3
					copy(bytes[startIndex:endIndex], pixel.color.toString())
				}

				universe <- bytes
			}

			for _, subscriber := range subscribers {
				subscriber <- &pixelMap
			}
			done <- true
		}()

		select {
		case <-done:
		case <-time.After(timeout):
			fmt.Println("error: time limit exceeded for frame update")
		}
	}

	go func() {
		for range updateTicker {
			// Start a goroutine to execute the update function with a timeout
			go func() {
				updateWithTimeout(updateInterval)
			}()
		}
	}()

	mux := http.NewServeMux()

	// the websocket is primarily to feed our pixel map into the visualizer
	mux.HandleFunc("GET /socket", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("establishing websocket connection handler")
		socketHandler(w, r, ch)
		subscribers = append(subscribers, ch)
	})

	mux.HandleFunc("GET /patterns", func(w http.ResponseWriter, r *http.Request) {
		getPatternsHandler(w, r, &patterns)
	})

	// this is pretty nice feature of go 1.22;
	// i don't think i need gorilla/mux or gin to build a REST API
	// this endpoint will allow us to update the current pattern and/or pattern params
	// TODO: add http.Error handling vs printlines
	mux.HandleFunc("PUT /patterns/{pattern}", func(w http.ResponseWriter, r *http.Request) {
		updatePatternHandler(w, r, patterns, &currentPattern)
	})

	mux.HandleFunc("GET /", rootHandler)

	fmt.Println("starting webserver")
	// TODO: this seems to error out when not connected a network. need to find some way to handle that
	address := fmt.Sprintf("%v:%v", config.HostAddress, config.HostPort)
	log.Fatal(http.ListenAndServe(address, mux))
}
