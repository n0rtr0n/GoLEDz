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

	solidColorPattern := SolidColorPattern{
		pixelMap:   &pixelMap,
		parameters: AdjustableParameters{},
	}

	solidColorPattern.parameters["color"] = &ColorParameter{
		value: Color{R: 0, G: 0, B: 255},
	}

	rainbowPattern := RainbowPattern{
		pixelMap:   &pixelMap,
		parameters: AdjustableParameters{},
		currentHue: 1.0,
	}
	rainbowPattern.parameters["speed"] = &FloatParameter{
		value: 1.0,
		min:   0.1,
		max:   360.0,
	}

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap:   &pixelMap,
		parameters: AdjustableParameters{},
		currentHue: 0.0,
	}
	rainbowDiagonalPattern.parameters["speed"] = &FloatParameter{
		value: 6.0,
		min:   0.1,
		max:   360.0,
	}
	rainbowDiagonalPattern.parameters["size"] = &FloatParameter{
		value: 0.5,
		min:   0.1,
		max:   180.0,
	}
	rainbowDiagonalPattern.parameters["reversed"] = &BooleanParameter{
		value: true,
	}

	solidColorFadePattern := SolidColorFadePattern{
		pixelMap:   &pixelMap,
		parameters: AdjustableParameters{},
		currentHue: 0.0,
	}
	solidColorFadePattern.parameters["speed"] = &FloatParameter{
		value: 5.0,
		min:   0.1,
		max:   360.0,
	}
	verticalStripesPattern := VerticalStripesPattern{
		pixelMap:        &pixelMap,
		parameters:      AdjustableParameters{},
		currentPosition: 0.0,
	}
	verticalStripesPattern.parameters["color"] = &ColorParameter{
		value: Color{R: 0, G: 0, B: 255},
	}
	verticalStripesPattern.parameters["size"] = &FloatParameter{
		value: 25.0,
		min:   0.1,
		max:   360.0,
	}
	verticalStripesPattern.parameters["speed"] = &FloatParameter{
		value: 15.0,
		min:   0.1,
		max:   360.0,
	}

	patterns["solidColor"] = &solidColorPattern
	patterns["rainbow"] = &rainbowPattern
	patterns["rainbowDiagonal"] = &rainbowDiagonalPattern
	patterns["solidColorFade"] = &solidColorFadePattern
	patterns["verticalStripes"] = &verticalStripesPattern

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
	go func() {
		for {
			// TODO: we'll eventually handle frame rate. For now, we'll update up to 20 times/second
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
	// TODO: add http.Error handling vs printlines
	mux.HandleFunc("PUT /patterns/{pattern}", func(w http.ResponseWriter, r *http.Request) {
		patternName := r.PathValue("pattern")
		pattern, ok := patterns[patternName]
		if !ok {
			fmt.Println(patternName, " not found in registered patterns, skipping update")
			return
		}

		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		requestParameters, ok := body["parameters"].([]interface{})
		if !ok || len(requestParameters) == 0 {
			fmt.Println("Invalid or empty parameters provided")
		}

		for key, parameter := range *pattern.ListParameters() {
			for k, param := range requestParameters {
				fmt.Println("k ", k, " param: ", param)
				newParam := param.(map[string]interface{})
				parm, ok := newParam[key]
				if !ok {
					continue
				}
				err = nil
				switch key {
				case "color":
					colorParam, ok := parm.(map[string]interface{})
					if !ok {
						fmt.Println("Color parameter is not a map")
						continue
					}

					// Extract individual color components
					r, ok := colorParam["r"].(float64)
					if !ok {
						fmt.Println("Value of 'r' is not a float64")
						continue
					}
					g, ok := colorParam["g"].(float64)
					if !ok {
						fmt.Println("Value of 'g' is not a float64")
						continue
					}
					b, ok := colorParam["b"].(float64)
					if !ok {
						fmt.Println("Value of 'b' is not a float64")
						continue
					}

					color := Color{
						R: colorPigment(r),
						G: colorPigment(g),
						B: colorPigment(b),
					}

					err = parameter.Update(color)
				case "speed":
					speed, ok := parm.(float64)
					if !ok {
						fmt.Println("Speed parameter is not a float")
						continue
					}
					err = parameter.Update(speed)
				case "size":
					size, ok := parm.(float64)
					if !ok {
						fmt.Println("Size parameter is not a float")
						continue
					}
					err = parameter.Update(size)
				case "reversed":
					fmt.Println("parm: ", parm)
					reversed, ok := parm.(bool)
					if !ok {
						fmt.Println("Reversed parameter is not a boolean")
						continue
					}
					err = parameter.Update(reversed)
				}
				if err != nil {
					msg := fmt.Sprintf(
						"Error updating parameter %v on pattern %v: %v",
						key,
						patternName,
						err,
					)
					fmt.Println(msg)
				}
			}
		}

		fmt.Println("new pattern", patternName)
		currentPattern = pattern
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "GoLEDz web server")
	})

	fmt.Println("starting webserver")
	// TODO: this seems to error out when not connected a network. need to find some way to handle that
	log.Fatal(http.ListenAndServe(":8008", mux))
}
