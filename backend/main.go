package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// websocket connections. will no longer block if we're not connected to a websocket
	var subscribers []chan *PixelMap
	ch := make(chan *PixelMap)
	defer close(ch)

	pixelMap := PixelMap{
		pixels: buildPixelGrid(),
	}

	// register patterns
	patterns := make(map[string]Pattern)

	rainbowPattern := RainbowPattern{
		pixelMap:   &pixelMap,
		speed:      1.0,
		currentHue: 1.0,
	}
	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap:   &pixelMap,
		currentHue: 0.0,
		speed:      6.0,
		reversed:   true,
		size:       0.5,
	}
	solidColorFadePattern := SolidColorFadePattern{
		pixelMap:   &pixelMap,
		currentHue: 0.0,
		speed:      5.0,
	}
	verticalStripesPattern := VerticalStripesPattern{
		pixelMap: &pixelMap,
		color:    Color{r: 255, g: 0, b: 255}, // purple
		speed:    10.0,
		size:     20.0,
	}

	patterns["rainbow"] = &rainbowPattern
	patterns["rainbowDiagonal"] = &rainbowDiagonalPattern
	patterns["solidColorFade"] = &solidColorFadePattern
	patterns["verticalStripes"] = &verticalStripesPattern

	// starting with just one single pattern and no ability to change patterns
	currentPattern := patterns["verticalStripes"]

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
	go func() {
		for {
			// we'll eventually handle frame rate. For now, we'll update up to 20 times/second
			time.Sleep(50 * time.Millisecond)

			// update pixel map
			currentPattern.Update()

			for i, universe := range universes {
				bytes := make([]byte, 512)
				for _, pixel := range pixelsByUniverse[i] {
					pos := pixel.channelPosition - 1
					startIndex := pos * 3
					endIndex := startIndex + 3
					copy(bytes[startIndex:endIndex], pixel.color.toString())
				}

				universe <- bytes
			}

			for _, subscriber := range subscribers {
				subscriber <- &pixelMap
			}
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
		patternsList := []map[string]string{}
		for k, _ := range patterns {
			patternsList = append(patternsList, map[string]string{"id": k})
		}

		patternsRoot := make(map[string][]map[string]string)
		patternsRoot["patterns"] = patternsList
		jsonData, err := json.Marshal(patternsRoot)
		if err != nil {
			fmt.Printf("could not marshal json: %s\n", err)
			return
		}

		fmt.Fprint(w, string(jsonData))
	})

	// this is pretty nice feature of go 1.22;
	// i don't think i need gorilla/mux or gin to build a REST API
	// this endpoint will allow us to update the current pattern and/or pattern params
	mux.HandleFunc("PUT /patterns/{pattern}", func(w http.ResponseWriter, r *http.Request) {
		patternName := r.PathValue("pattern")
		pattern, ok := patterns[patternName]
		if !ok {
			fmt.Println(patternName, " not found in registered patterns, skipping update")
			return
		}
		fmt.Println("new pattern", patternName)
		currentPattern = pattern
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GoLEDz web server")
	})

	fmt.Println("starting webserver")
	log.Fatal(http.ListenAndServe(":8008", mux))
}
