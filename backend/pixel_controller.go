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
	patterns         map[string]Pattern
	errorTracker     *ErrorTracker
	pixelsByUniverse map[uint16][]*Pixel
	updateInterval   time.Duration
	running          bool
	stopChan         chan struct{}
	wg               sync.WaitGroup
	currentMode      PatternMode
	currentPattern   Pattern
	patternMu        sync.RWMutex
	onUpdate         func(*PixelMap)
	pixelMap         *PixelMap
	transition       *struct {
		sourcePattern Pattern
		targetPattern Pattern
		startTime     time.Time
		duration      time.Duration
		sourcePixels  []Pixel
		targetPixels  []Pixel
	}
	transitionMutex    sync.RWMutex
	transitionDuration time.Duration
	patternChange      chan Pattern
	currentColorMask   ColorMaskPattern
	colorMaskChange    chan ColorMaskPattern
	isParameterUpdate  bool
}

func getDefaultColorMask() ColorMaskPattern {
	// return &GradientColorMask{
	// 	Parameters: GradientParameters{
	// 		Color1: ColorParameter{
	// 			Value: Color{R: 255, G: 0, B: 0},
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		Color2: ColorParameter{
	// 			Value: Color{R: 0, G: 0, B: 255},
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		Speed: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   20.0,
	// 			Value: 1.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		Reversed: BooleanParameter{
	// 			Value: false,
	// 			Type:  TYPE_BOOL,
	// 		},
	// 	},
	// 	Label: "Gradient",
	// }
	return &RainbowCircleMask{
		Parameters: RainbowCircleParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   25.0,
				Value: 6.0,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		Label: "Rainbow Circle",
	}
}

func NewPixelController(universes map[uint16]chan<- []byte, errorTracker *ErrorTracker, fps int, initialPattern Pattern, pixelMap *PixelMap, transitionDuration time.Duration) *PixelController {
	if initialPattern == nil {
		panic("initialPattern cannot be nil")
	}

	controller := &PixelController{
		universes:          universes,
		errorTracker:       errorTracker,
		updateInterval:     time.Second / time.Duration(fps),
		stopChan:           make(chan struct{}),
		currentPattern:     initialPattern,
		pixelMap:           pixelMap,
		transitionDuration: transitionDuration,
		patternChange:      make(chan Pattern, 1),
		colorMaskChange:    make(chan ColorMaskPattern, 1),
		currentColorMask:   getDefaultColorMask(),
	}

	controller.patterns = registerPatterns(pixelMap)
	return controller
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

// updates all universes with current pixel data
func (pc *PixelController) updateAllUniverses() error {
	pc.transitionMutex.RLock()
	defer pc.transitionMutex.RUnlock()

	pc.Update()

	if pc.onUpdate != nil {
		pc.onUpdate(pc.pixelMap)
	}

	// send updated pixels to universes
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

func (pc *PixelController) SetPattern(pattern interface{}) error {
	switch p := pattern.(type) {
	case PatternMode:
		if pc.currentMode != nil {
			pc.currentMode.Stop()
		}
		pc.currentMode = p
		p.SetController(pc)
		if pattern := p.GetCurrentPattern(); pattern != nil {
			pc.currentPattern = pattern
		}
		p.Start()
		return nil

	case Pattern:
		if pc.isParameterUpdate {
			// Set color mask before updating pattern
			if pc.currentColorMask != nil {
				p.SetColorMask(pc.currentColorMask)
			}
			pc.currentPattern = p
			return nil
		}
		// Set color mask before sending to pattern change channel
		if pc.currentColorMask != nil {
			p.SetColorMask(pc.currentColorMask)
		}
		select {
		case pc.patternChange <- p:
			return nil
		default:
			return fmt.Errorf("pattern change channel full, try again later")
		}

	default:
		return fmt.Errorf("unknown pattern type: %T", pattern)
	}
}

func (pc *PixelController) SetUpdateCallback(callback func(*PixelMap)) {
	pc.patternMu.Lock()
	pc.onUpdate = callback
	pc.patternMu.Unlock()
}

func (pc *PixelController) SetColorMask(mask ColorMaskPattern) error {
	select {
	case pc.colorMaskChange <- mask:
		return nil
	default:
		return fmt.Errorf("color mask change channel full, try again later")
	}
}

func (pc *PixelController) Update() {
	// Handle color mask changes
	select {
	case newMask := <-pc.colorMaskChange:
		pc.currentColorMask = newMask
	default:
		// no color mask change pending
	}

	// Update color mask if it exists
	if pc.currentColorMask != nil {
		pc.currentColorMask.Update()
		if pc.currentPattern != nil {
			pc.currentPattern.SetColorMask(pc.currentColorMask)
		}
	}

	select {
	case newPattern := <-pc.patternChange:
		// Don't create transition if we're just updating parameters
		if pc.isParameterUpdate {
			break
		}
		// Create transition pixels
		sourcePixels := make([]Pixel, len(*pc.pixelMap.pixels))
		targetPixels := make([]Pixel, len(*pc.pixelMap.pixels))
		copy(sourcePixels, *pc.pixelMap.pixels)
		copy(targetPixels, *pc.pixelMap.pixels)

		pc.transition = &struct {
			sourcePattern Pattern
			targetPattern Pattern
			startTime     time.Time
			duration      time.Duration
			sourcePixels  []Pixel
			targetPixels  []Pixel
		}{
			sourcePattern: pc.currentPattern,
			targetPattern: newPattern,
			startTime:     time.Now(),
			duration:      pc.transitionDuration,
			sourcePixels:  sourcePixels,
			targetPixels:  targetPixels,
		}
	default:
		// no pattern change pending, continue with normal update
	}

	// handle active transition
	if pc.transition != nil && !pc.isParameterUpdate {
		elapsed := time.Since(pc.transition.startTime)
		progress := float64(elapsed) / float64(pc.transition.duration)

		DefaultTransitionFromPattern(
			pc.transition.targetPattern,
			pc.transition.sourcePattern,
			progress,
			pc.pixelMap,
		)

		if progress >= 1.0 {
			pc.currentPattern = pc.transition.targetPattern
			pc.transition = nil
			if pc.currentMode != nil {
				if mode, ok := pc.currentMode.(*RandomMode); ok {
					mode.TransitionComplete()
				}
			}
		}
		return
	}

	// normal pattern update
	if pc.currentMode != nil {
		pc.currentMode.Update()
	} else if pc.currentPattern != nil {
		pc.currentPattern.Update()
	}
}

func (pc *PixelController) SetTransitionDuration(duration time.Duration) {
	pc.transitionMutex.Lock()
	defer pc.transitionMutex.Unlock()

	pc.transitionDuration = duration

	// if there's an active transition, update its duration
	if pc.transition != nil {
		elapsed := time.Since(pc.transition.startTime)
		progress := float64(elapsed) / float64(pc.transition.duration)
		pc.transition.startTime = time.Now().Add(-time.Duration(float64(duration) * progress))
		pc.transition.duration = duration
	}
}

func (c *PixelController) UpdatePattern(patternName string, request PatternUpdateRequest) error {
	pattern, exists := c.patterns[patternName]
	if !exists {
		return fmt.Errorf("pattern %s not found", patternName)
	}

	// If updating same pattern, just update parameters directly
	if c.currentPattern != nil && c.currentPattern.GetName() == patternName {
		c.isParameterUpdate = true
		defer func() { c.isParameterUpdate = false }()
		return c.currentPattern.UpdateParameters(request.GetParameters())
	}

	// For different patterns, do the normal transition
	if err := pattern.UpdateParameters(request.GetParameters()); err != nil {
		return err
	}
	return c.SetPattern(pattern)
}
