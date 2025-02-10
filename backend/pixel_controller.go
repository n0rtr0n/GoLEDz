package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// PixelController manages the updating and display of pixels across universes
type PixelController struct {
	universes        map[uint16]chan<- []byte
	errorTracker     *ErrorTracker
	pixelsByUniverse map[uint16][]*Pixel
	updateInterval   time.Duration
	running          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	currentPattern   Pattern
	patternMu        sync.RWMutex
	onUpdate         func(*PixelMap)
	pixelMap         *PixelMap
}

func NewPixelController(universes map[uint16]chan<- []byte, errorTracker *ErrorTracker, fps int, initialPattern Pattern, pixelMap *PixelMap) *PixelController {
	if initialPattern == nil {
		panic("initialPattern cannot be nil")
	}

	return &PixelController{
		universes:      universes,
		errorTracker:   errorTracker,
		updateInterval: time.Second / time.Duration(fps),
		stopChan:       make(chan struct{}),
		currentPattern: initialPattern,
		pixelMap:       pixelMap,
	}
}

// we're currently using the paradigm of DMX over ethernet, so we think about the world
// in terms of universes == channels. for that reason, we're going to denormalize our pixel
// map a little bit by creating a map of pointers to pixels. this will save us a ton of compute
// cycles when we go to send the data over the wire, because this map will be the ordered
// representation by universe/channel of each pixel position, eliminating the need for
// expensive lookups

func (pc *PixelController) organizePixelsByUniverse(pixelMap *PixelMap) {
	pc.pixelsByUniverse = make(map[uint16][]*Pixel)
	for i, pixel := range *pixelMap.pixels {
		pc.pixelsByUniverse[pixel.universe] = append(
			pc.pixelsByUniverse[pixel.universe],
			&(*pixelMap.pixels)[i],
		)
	}
}

// prepares the byte data for a specific universe
func (pc *PixelController) prepareUniverseData(universe uint16) []byte {
	bytes := make([]byte, 512)
	for _, pixel := range pc.pixelsByUniverse[universe] {
		pos := pixel.channelPosition - 1
		startIndex := pos * 3
		endIndex := startIndex + 3
		rgbBytes := pixel.color.toString()
		copy(bytes[startIndex:endIndex], rgbBytes)
	}
	return bytes
}

// updateAllUniverses updates all universes with current pixel data
func (pc *PixelController) updateAllUniverses() error {
	pc.patternMu.RLock()
	if pc.currentPattern != nil {
		pc.currentPattern.Update()
		if pc.onUpdate != nil {
			pc.onUpdate(pc.pixelMap)
		}
	}
	pc.patternMu.RUnlock()

	for universe := range pc.universes {
		data := pc.prepareUniverseData(universe)
		pc.universes[universe] <- data
	}

	return nil
}

func (pc *PixelController) Start(pixelMap *PixelMap) error {
	if pc.running {
		return fmt.Errorf("controller is already running")
	}

	pc.running = true
	pc.organizePixelsByUniverse(pixelMap)

	pc.wg.Add(1)
	go func() {
		defer pc.wg.Done()
		updateTicker := time.NewTicker(pc.updateInterval)
		defer updateTicker.Stop()

		for {
			select {
			case <-pc.stopChan:
				return
			case <-updateTicker.C:
				if err := pc.updateAllUniverses(); err != nil {
					log.Printf("Update error: %v", err)
				}
			}
		}
	}()

	return nil
}

func (pc *PixelController) Stop() {
	if !pc.running {
		return
	}

	close(pc.stopChan)
	pc.wg.Wait()
	pc.running = false
}

func (pc *PixelController) UpdatePattern(pattern Pattern) {
	fmt.Println("Switching to pattern: ")
	fmt.Println(pattern.GetName())
	pc.patternMu.Lock()
	pc.currentPattern = pattern
	pc.patternMu.Unlock()
}

func (pc *PixelController) SetUpdateCallback(callback func(*PixelMap)) {
	pc.patternMu.Lock()
	pc.onUpdate = callback
	pc.patternMu.Unlock()
}
