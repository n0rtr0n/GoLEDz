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
	options          Options
	transition       *struct {
		sourcePattern Pattern
		targetPattern Pattern
		startTime     time.Time
		duration      time.Duration
		sourcePixels  []Pixel
		targetPixels  []Pixel
		sourceMask    ColorMaskPattern
		targetMask    ColorMaskPattern
	}
	transitionMutex    sync.RWMutex
	transitionDuration time.Duration
	patternChange      chan Pattern
	currentColorMask   ColorMaskPattern
	colorMaskChange    chan ColorMaskPattern
	isParameterUpdate  bool
}

func getDefaultColorMask() ColorMaskPattern {
	return &RainbowCircleMask{
		BasePattern: BasePattern{
			Label: "Rainbow Circle",
		},
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
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
		},
	}
}

func NewPixelController(universes map[uint16]chan<- []byte, errorTracker *ErrorTracker, fps int, initialPattern Pattern, pixelMap *PixelMap, options Options) *PixelController {
	if initialPattern == nil {
		panic("initialPattern cannot be nil")
	}

	controller := &PixelController{
		universes:        universes,
		errorTracker:     errorTracker,
		updateInterval:   time.Second / time.Duration(fps),
		stopChan:         make(chan struct{}),
		currentPattern:   initialPattern,
		pixelMap:         pixelMap,
		options:          options,
		patternChange:    make(chan Pattern, 1),
		colorMaskChange:  make(chan ColorMaskPattern, 1),
		currentColorMask: getDefaultColorMask(),
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

	// Set all bytes to zero initially
	for i := range bytes {
		bytes[i] = 0
	}

	// Now write the pixel data
	for _, pixel := range pc.pixelsByUniverse[universe] {
		// Calculate the actual DMX position based on channel position and pixel type
		channelsPerPixel := int(pixel.pixelType) // 3 for RGB, 4 for RGBW
		pos := (pixel.channelPosition - 1) * uint16(channelsPerPixel)

		// Write color values to consecutive channels based on color ordering
		if pos+uint16(channelsPerPixel)-1 < 512 {
			// map the color values according to the pixel's color order
			var colorValues [4]byte

			colorValues[0] = byte(pixel.color.R)
			colorValues[1] = byte(pixel.color.G)
			colorValues[2] = byte(pixel.color.B)
			if pixel.pixelType == PixelRGBW {
				colorValues[3] = byte(pixel.color.W)
			}

			switch pixel.colorOrder {
			case RGB:
				// already set correctly
			case RBG:
				colorValues[1], colorValues[2] = colorValues[2], colorValues[1]
			case BRG:
				r, g, b := colorValues[0], colorValues[1], colorValues[2]
				colorValues[0] = b
				colorValues[1] = r
				colorValues[2] = g
			case BGR:
				r, g, b := colorValues[0], colorValues[1], colorValues[2]
				colorValues[0] = b
				colorValues[1] = g
				colorValues[2] = r
			case GRB:
				r, g, b := colorValues[0], colorValues[1], colorValues[2]
				colorValues[0] = g
				colorValues[1] = r
				colorValues[2] = b
			case GBR:
				r, g, b := colorValues[0], colorValues[1], colorValues[2]
				colorValues[0] = g
				colorValues[1] = b
				colorValues[2] = r
			}

			// write the remapped values to the output buffer
			for i := range channelsPerPixel {
				bytes[pos+uint16(i)] = colorValues[i]
			}
		}
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
			// set color mask before updating pattern
			if pc.currentColorMask != nil {
				p.SetColorMask(pc.currentColorMask)
			}
			pc.currentPattern = p
			return nil
		}
		// set color mask before sending to pattern change channel
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
	// Don't create transition if we're just updating parameters
	if pc.isParameterUpdate {
		pc.currentColorMask = mask
		return nil
	}

	select {
	case pc.colorMaskChange <- mask:
		return nil
	default:
		return fmt.Errorf("color mask change channel full, try again later")
	}
}

func (pc *PixelController) Update() {
	// Check for color mask changes
	var newMask ColorMaskPattern
	select {
	case newMask = <-pc.colorMaskChange:
		// Get the colorMaskTransitionEnabled option
		colorMaskTransitionEnabledOpt, _ := pc.options.GetOption("colorMaskTransitionEnabled")
		colorMaskTransitionEnabled := colorMaskTransitionEnabledOpt.GetValue().(bool)

		// Only create transition if it's a different mask, not updating parameters, and transitions are enabled
		if !pc.isParameterUpdate && colorMaskTransitionEnabled &&
			(pc.currentColorMask == nil || pc.currentColorMask.GetName() != newMask.GetName()) {
			// Create transition pixels
			sourcePixels := make([]Pixel, len(*pc.pixelMap.pixels))
			copy(sourcePixels, *pc.pixelMap.pixels)

			// Get the colorMaskTransitionDuration option
			colorMaskTransitionDurationOpt, _ := pc.options.GetOption("colorMaskTransitionDuration")
			colorMaskTransitionDuration := time.Duration(colorMaskTransitionDurationOpt.GetValue().(int)) * time.Millisecond

			// Store current state
			pc.transition = &struct {
				sourcePattern Pattern
				targetPattern Pattern
				startTime     time.Time
				duration      time.Duration
				sourcePixels  []Pixel
				targetPixels  []Pixel
				sourceMask    ColorMaskPattern
				targetMask    ColorMaskPattern
			}{
				sourcePattern: pc.currentPattern,
				targetPattern: pc.currentPattern, // Same pattern, different mask
				startTime:     time.Now(),
				duration:      colorMaskTransitionDuration,
				sourcePixels:  sourcePixels,
				targetPixels:  nil,
				sourceMask:    pc.currentColorMask,
				targetMask:    newMask,
			}
		} else {
			// Just update the mask without transition
			pc.currentColorMask = newMask
			if pc.currentPattern != nil {
				pc.currentPattern.SetColorMask(pc.currentColorMask)
			}
		}
	default:
		// No color mask change pending
	}

	// Handle pattern changes
	select {
	case newPattern := <-pc.patternChange:
		// Get the patternTransitionEnabled option
		patternTransitionEnabledOpt, _ := pc.options.GetOption("patternTransitionEnabled")
		patternTransitionEnabled := patternTransitionEnabledOpt.GetValue().(bool)

		// don't create transition if we're just updating parameters or transitions are disabled
		if pc.isParameterUpdate || !patternTransitionEnabled {
			break
		}
		// create transition pixels
		sourcePixels := make([]Pixel, len(*pc.pixelMap.pixels))
		copy(sourcePixels, *pc.pixelMap.pixels)

		// Get the patternTransitionDuration option
		patternTransitionDurationOpt, _ := pc.options.GetOption("patternTransitionDuration")
		patternTransitionDuration := time.Duration(patternTransitionDurationOpt.GetValue().(int)) * time.Millisecond

		pc.transition = &struct {
			sourcePattern Pattern
			targetPattern Pattern
			startTime     time.Time
			duration      time.Duration
			sourcePixels  []Pixel
			targetPixels  []Pixel
			sourceMask    ColorMaskPattern
			targetMask    ColorMaskPattern
		}{
			sourcePattern: pc.currentPattern,
			targetPattern: newPattern,
			startTime:     time.Now(),
			duration:      patternTransitionDuration,
			sourcePixels:  sourcePixels,
			targetPixels:  nil,
			sourceMask:    pc.currentColorMask,
			targetMask:    pc.currentColorMask,
		}
	default:
		// no pattern change pending, continue with normal update
	}

	// Handle active transition
	if pc.transition != nil && !pc.isParameterUpdate {
		elapsed := time.Since(pc.transition.startTime)
		progress := float64(elapsed) / float64(pc.transition.duration)

		// Check if this is a color mask transition
		if pc.transition.sourcePattern == pc.transition.targetPattern &&
			pc.transition.sourceMask != nil && pc.transition.targetMask != nil {

			// Create a custom blended color mask for this frame
			blendedMask := &blendedColorMask{
				sourceMask: pc.transition.sourceMask,
				targetMask: pc.transition.targetMask,
				progress:   progress,
			}

			// Update both source and target masks
			pc.transition.sourceMask.Update()
			pc.transition.targetMask.Update()

			// Apply the blended mask to the pattern
			pc.currentPattern.SetColorMask(blendedMask)

			// Update the pattern with the blended mask
			pc.currentPattern.Update()
		} else {
			// Regular pattern transition
			if pc.transition.targetPattern != nil {
				// Set the appropriate color mask on the target pattern
				if pc.transition.targetMask != nil {
					pc.transition.targetPattern.SetColorMask(pc.transition.targetMask)
				} else {
					pc.transition.targetPattern.SetColorMask(pc.currentColorMask)
				}

				DefaultTransitionFromPattern(
					pc.transition.targetPattern,
					pc.transition.sourcePattern,
					progress,
					pc.pixelMap,
				)
			}
		}

		if progress >= 1.0 {
			pc.currentPattern = pc.transition.targetPattern
			if pc.transition.targetMask != nil {
				pc.currentColorMask = pc.transition.targetMask
				pc.currentPattern.SetColorMask(pc.currentColorMask)
			}
			pc.transition = nil
			if pc.currentMode != nil {
				if mode, ok := pc.currentMode.(*RandomMode); ok {
					mode.TransitionComplete()
				}
			}
		}
		return
	}

	// update color mask if it exists
	if pc.currentColorMask != nil {
		pc.currentColorMask.Update()
		if pc.currentPattern != nil {
			pc.currentPattern.SetColorMask(pc.currentColorMask)
		}
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

	// if updating same pattern, just update parameters directly
	if c.currentPattern != nil && c.currentPattern.GetName() == patternName {
		c.isParameterUpdate = true
		defer func() { c.isParameterUpdate = false }()
		return c.currentPattern.UpdateParameters(request.GetParameters())
	}

	// for different patterns, do the normal transition
	if err := pattern.UpdateParameters(request.GetParameters()); err != nil {
		return err
	}
	return c.SetPattern(pattern)
}

type blendedColorMask struct {
	BasePattern
	sourceMask ColorMaskPattern
	targetMask ColorMaskPattern
	progress   float64
}

func (b *blendedColorMask) GetColorAt(point Point) Color {
	sourceColor := b.sourceMask.GetColorAt(point)
	targetColor := b.targetMask.GetColorAt(point)

	// blend the colors based on transition progress
	return blendColors(sourceColor, targetColor, b.progress)
}

func (b *blendedColorMask) Update() {
	// both source and target masks are updated in the main Update method
}

func (b *blendedColorMask) GetName() string {
	return "blendedColorMask"
}

func (b *blendedColorMask) UpdateParameters(parameters AdjustableParameters) error {
	return nil // no parameters to update
}

func (b *blendedColorMask) GetPatternUpdateRequest() PatternUpdateRequest {
	return nil // no update request needed
}

func (b *blendedColorMask) TransitionFrom(source Pattern, progress float64) {
	// no transition needed for this temporary mask
}

func (pc *PixelController) SetMode(mode PatternMode) {
	pc.patternMu.Lock()
	defer pc.patternMu.Unlock()

	// Clear any existing transition
	pc.transition = nil

	// Set the new mode
	pc.currentMode = mode

	// If mode is nil, ensure we have a pattern active
	if mode == nil && pc.currentPattern == nil {
		// Set a default pattern if none is active
		for _, pattern := range pc.patterns {
			pc.currentPattern = pattern
			break
		}
	}
}

// Update the UpdateOptions method to handle mode changes
func (pc *PixelController) UpdateOptions(options Options) {
	pc.patternMu.Lock()
	pc.options = options
	pc.patternMu.Unlock()

	// Set transition duration
	pc.SetTransitionDuration(options.TransitionDuration)

	// Handle mode change if specified
	if options.ActiveMode != "" {
		// This will be handled by the server's handleUpdateOptions method
	}
}
