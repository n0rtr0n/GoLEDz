package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	/*
		first goal is to light up ONE pixel. let's call it red
		1. send the pixel to the websocket location so we can see it digitally
		2. use a very simple implementation of an e1.31/sACN library so we can see it physically
	*/

	// channel for websocket connections
	ch := make(chan *PixelMap)
	defer close(ch)

	pixelMap := PixelMap{
		pixels: build2ChannelsOfPixels(),
	}

	/*
		starting with just one single pattern and no ability to change patterns
	*/
	currentPattern := SolidColorFadePattern{
		pixelMap:   &pixelMap,
		currentHue: 0.0, // effectively red
		speed:      1,
	}

	universes := setupSACN()
	for _, universe := range universes {
		defer close(universe)
	}

	// we're currently using the paradigm of DMX over ethernet, so we think about the world
	// in terms of universes == channels. for that reason, we're going to denormalize our pixel
	// map a little bit by creating a map of pointers to pixels. this will save us a ton of compute
	// cycles when we go to send the data over the wire, because this map will be the ordered
	// representation by universe/channel of each pixel position, eliminating the need for
	// expensive lookups
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

			// this will block until we have a websocket to receive this data
			ch <- &pixelMap
		}
	}()

	// the websocket is primarily to feed our pixel map into the visualizer
	http.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("establishing websocket connection handler")
		socketHandler(w, r, ch)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world")
	})

	fmt.Println("starting webserver")
	log.Fatal(http.ListenAndServe(":8008", nil))
}
