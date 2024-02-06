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

	color := Color{0, 0, 255}

	// initialize a grid of pixels for fun coordinate-based pattern development
	pixels := []Pixel{}

	var xPos int16
	var yPos int16
	xStart := 100
	yStart := 100
	spacing := 10
	for i := 0; i < 40; i++ {
		xPos = int16(xStart + i*spacing)
		for j := 0; j < 40; j++ {
			yPos = int16(yStart + j*spacing)
			pixels = append(pixels, Pixel{x: xPos, y: yPos, color: color})
		}
	}

	pixelMap := PixelMap{
		pixels: &pixels,
	}

	/*
		starting with just one single pattern and no ability to change patterns
	*/
	currentPattern := SolidColorFadePattern{
		pixelMap:   &pixelMap,
		currentHue: 0.0, // effectively red
		speed:      1,
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
