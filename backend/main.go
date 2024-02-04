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

	// the color red. yep.
	colorRed := Color{255, 0, 0}

	// initial map of one single but mighty pixel
	pixels := []Pixel{
		{
			x:     300,
			y:     300,
			color: colorRed,
		},
	}

	pixelMap := PixelMap{
		pixels: &pixels,
	}

	/*
		starting with just one single pattern and no ability to change patterns
		solid color is nice because no matter how many pixels we have, we can uniformly
		set them all to the same color, which makes this an excellent initial test case
	*/
	currentPattern := SolidColorPattern{
		pixelMap: &pixelMap,
		color:    colorRed,
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

			// this will block until we have a websocket to receive this data
			ch <- &pixelMap

			// TODO: render to physical device
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
