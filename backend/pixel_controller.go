package main

import (
	"fmt"
	"log"
	"math"
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

	// set all bytes to zero initially
	for i := range bytes {
		bytes[i] = 0
	}

	// now write the pixel data
	for _, pixel := range pc.pixelsByUniverse[universe] {
		// calculate the actual DMX position based on channel position and pixel type
		channelsPerPixel := int(pixel.pixelType) // 3 for RGB, 4 for RGBW
		pos := (pixel.channelPosition - 1) * uint16(channelsPerPixel)

		// write color values to consecutive channels based on color ordering
		if pos+uint16(channelsPerPixel)-1 < 512 {
			// map the color values according to the pixel's color order
			var colorValues [4]byte

			// Use the already-adjusted color values, but force W to 0
			colorValues[0] = byte(pixel.color.R)
			colorValues[1] = byte(pixel.color.G)
			colorValues[2] = byte(pixel.color.B)
			if pixel.pixelType == PixelRGBW {
				colorValues[3] = 0 // Always force W to 0
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
	// don't create transition if we're just updating parameters
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
	// check for color mask changes
	var newMask ColorMaskPattern
	select {
	case newMask = <-pc.colorMaskChange:
		colorMaskTransitionEnabledOpt, _ := pc.options.GetOption("colorMaskTransitionEnabled")
		colorMaskTransitionEnabled := colorMaskTransitionEnabledOpt.GetValue().(bool)

		// only create transition if it's a different mask, not updating parameters, and transitions are enabled
		if !pc.isParameterUpdate && colorMaskTransitionEnabled &&
			(pc.currentColorMask == nil || pc.currentColorMask.GetName() != newMask.GetName()) {
			// create transition pixels
			sourcePixels := make([]Pixel, len(*pc.pixelMap.pixels))
			copy(sourcePixels, *pc.pixelMap.pixels)

			colorMaskTransitionDurationOpt, _ := pc.options.GetOption("colorMaskTransitionDuration")
			colorMaskTransitionDuration := time.Duration(colorMaskTransitionDurationOpt.GetValue().(int)) * time.Millisecond

			// store current state
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
				targetPattern: pc.currentPattern, // same pattern, different mask
				startTime:     time.Now(),
				duration:      colorMaskTransitionDuration,
				sourcePixels:  sourcePixels,
				targetPixels:  nil,
				sourceMask:    pc.currentColorMask,
				targetMask:    newMask,
			}
		} else {
			// just update the mask without transition
			pc.currentColorMask = newMask
			if pc.currentPattern != nil {
				pc.currentPattern.SetColorMask(pc.currentColorMask)
			}
		}
	default:
		// no color mask change pending
	}

	// handle pattern changes
	select {
	case newPattern := <-pc.patternChange:
		patternTransitionEnabledOpt, _ := pc.options.GetOption("patternTransitionEnabled")
		patternTransitionEnabled := patternTransitionEnabledOpt.GetValue().(bool)

		// don't create transition if we're just updating parameters or transitions are disabled
		if pc.isParameterUpdate || !patternTransitionEnabled {
			break
		}
		// create transition pixels
		sourcePixels := make([]Pixel, len(*pc.pixelMap.pixels))
		copy(sourcePixels, *pc.pixelMap.pixels)

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

	// handle active transition
	if pc.transition != nil && !pc.isParameterUpdate {
		elapsed := time.Since(pc.transition.startTime)

		// Calculate raw progress
		rawProgress := float64(elapsed) / float64(pc.transition.duration)

		// Use a smoothed progress that extends slightly beyond 1.0 before completing
		// This allows the source pattern to fade out more gradually
		smoothedProgress := rawProgress

		// Only mark the transition as complete when we're well past 100%
		transitionComplete := false

		if rawProgress >= 1.0 {
			if rawProgress >= 1.1 { // Give 10% extra time for smooth completion
				transitionComplete = true
			} else {
				// Ease out the final part of the transition (1.0-1.1)
				// This makes the last 10% of the transition smoother
				smoothedProgress = 1.0 - 0.5*math.Pow(1.1-rawProgress, 2)
				if smoothedProgress > 0.999 {
					smoothedProgress = 0.999 // Keep it just under 1.0 to maintain blending
				}
			}
		}

		// check if this is a color mask transition
		if pc.transition.sourcePattern == pc.transition.targetPattern &&
			pc.transition.sourceMask != nil && pc.transition.targetMask != nil {

			// Make sure both masks are updated first
			pc.transition.sourceMask.Update()
			pc.transition.targetMask.Update()

			// create a custom blended color mask for this frame
			blendedMask := &blendedColorMask{
				sourceMask: pc.transition.sourceMask,
				targetMask: pc.transition.targetMask,
				progress:   smoothedProgress,
			}

			// apply the blended mask to the pattern
			pc.currentPattern.SetColorMask(blendedMask)

			// update the pattern with the blended mask
			pc.currentPattern.Update()
		} else {
			// regular pattern transition
			if pc.transition.targetPattern != nil {
				// set the appropriate color mask on the target pattern
				if pc.transition.targetMask != nil {
					pc.transition.targetPattern.SetColorMask(pc.transition.targetMask)
				} else {
					pc.transition.targetPattern.SetColorMask(pc.currentColorMask)
				}

				DefaultTransitionFromPattern(
					pc.transition.targetPattern,
					pc.transition.sourcePattern,
					smoothedProgress,
					pc.pixelMap,
				)
			}
		}

		if transitionComplete {
			pc.currentPattern = pc.transition.targetPattern
			if pc.transition.targetMask != nil {
				pc.currentColorMask = pc.transition.targetMask
				pc.currentPattern.SetColorMask(pc.currentColorMask)
			}
			pc.transition = nil
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
	if pc.currentPattern != nil {
		pc.currentPattern.Update()
	}

	// After all pattern updates are done, apply brightness scaling to the pixel map
	pc.applyBrightnessToPixelMap()
}

func (pc *PixelController) applyBrightnessToPixelMap() {
	// Get brightness option
	brightnessOpt, err := pc.options.GetOption("brightness")
	var brightnessScale float64 = 1.0
	if err == nil {
		brightnessScale = brightnessOpt.GetValue().(float64) / 100.0
	}

	// Get gamma correction option
	gammaOpt, err := pc.options.GetOption("gamma")
	var gamma float64 = 1.0
	if err == nil {
		gamma = gammaOpt.GetValue().(float64)
	}

	// Store original colors in a local variable for this function
	originalColors := make([]Color, len(*pc.pixelMap.pixels))
	for i, pixel := range *pc.pixelMap.pixels {
		originalColors[i] = pixel.color
	}

	// Apply brightness and gamma to each pixel
	for i := range *pc.pixelMap.pixels {
		originalColor := originalColors[i]

		// First apply brightness
		r := float64(originalColor.R) * brightnessScale
		g := float64(originalColor.G) * brightnessScale
		b := float64(originalColor.B) * brightnessScale

		// Then apply gamma correction
		if gamma != 1.0 {
			// Normalize to 0-1 range
			r = math.Pow(r/255.0, 1.0/gamma) * 255.0
			g = math.Pow(g/255.0, 1.0/gamma) * 255.0
			b = math.Pow(b/255.0, 1.0/gamma) * 255.0
		}

		(*pc.pixelMap.pixels)[i].color = Color{
			R: colorPigment(r),
			G: colorPigment(g),
			B: colorPigment(b),
			W: 0, // Always set W to 0
		}
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
	// Get colors from both masks
	sourceColor := b.sourceMask.GetColorAt(point)
	targetColor := b.targetMask.GetColorAt(point)

	// Apply easing function to progress for smoother transitions
	// This creates a more natural acceleration/deceleration curve
	easedProgress := b.progress
	if b.progress < 0.5 {
		// Ease in (slow start)
		easedProgress = 2.0 * b.progress * b.progress
	} else {
		// Ease out (slow end)
		easedProgress = 1.0 - math.Pow(-2.0*b.progress+2.0, 2)/2.0
	}

	// For very low or very high progress values, just return the appropriate color
	// This prevents artifacts at the beginning and end of transitions
	if easedProgress < 0.01 {
		return Color{
			R: sourceColor.R,
			G: sourceColor.G,
			B: sourceColor.B,
			W: 0, // Always set W to 0
		}
	}
	if easedProgress > 0.99 {
		return Color{
			R: targetColor.R,
			G: targetColor.G,
			B: targetColor.B,
			W: 0, // Always set W to 0
		}
	}

	// Convert RGB to HSV for better blending
	sourceR, sourceG, sourceB := float64(sourceColor.R)/255.0, float64(sourceColor.G)/255.0, float64(sourceColor.B)/255.0
	targetR, targetG, targetB := float64(targetColor.R)/255.0, float64(targetColor.G)/255.0, float64(targetColor.B)/255.0

	// Skip HSV conversion for black or white colors to prevent artifacts
	if (sourceR < 0.01 && sourceG < 0.01 && sourceB < 0.01) ||
		(targetR < 0.01 && targetG < 0.01 && targetB < 0.01) {
		// For black, do direct RGB blending
		return Color{
			R: colorPigment(float64(sourceColor.R)*(1-easedProgress) + float64(targetColor.R)*easedProgress),
			G: colorPigment(float64(sourceColor.G)*(1-easedProgress) + float64(targetColor.G)*easedProgress),
			B: colorPigment(float64(sourceColor.B)*(1-easedProgress) + float64(targetColor.B)*easedProgress),
			W: 0, // Always set W to 0
		}
	}

	// Convert to HSV
	sourceH, sourceS, sourceV := RGBtoHSV(sourceR, sourceG, sourceB)
	targetH, targetS, targetV := RGBtoHSV(targetR, targetG, targetB)

	// Special case for colors with very low saturation (near white/gray/black)
	// Use the hue of the more saturated color
	if sourceS < 0.1 && targetS > 0.1 {
		sourceH = targetH
	} else if targetS < 0.1 && sourceS > 0.1 {
		targetH = sourceH
	}

	// Handle hue wrapping for shortest path
	if targetH-sourceH > 180 {
		sourceH += 360
	} else if sourceH-targetH > 180 {
		targetH += 360
	}

	// Interpolate in HSV space
	blendedH := sourceH*(1-easedProgress) + targetH*easedProgress
	if blendedH >= 360 {
		blendedH -= 360
	}

	// Use a non-linear blend for saturation to prevent washed-out colors
	// This gives more weight to the higher saturation
	blendedS := math.Sqrt(sourceS*sourceS*(1-easedProgress) + targetS*targetS*easedProgress)

	// Use a non-linear blend for value as well
	blendedV := math.Sqrt(sourceV*sourceV*(1-easedProgress) + targetV*targetV*easedProgress)

	// Convert back to RGB
	blendedR, blendedG, blendedB := HSVtoRGB(blendedH, blendedS, blendedV)

	// Create final color
	return Color{
		R: colorPigment(blendedR * 255),
		G: colorPigment(blendedG * 255),
		B: colorPigment(blendedB * 255),
		W: 0, // Always set W to 0
	}
}

func (b *blendedColorMask) Update() {
	// Both source and target masks should be updated separately
	b.sourceMask.Update()
	b.targetMask.Update()
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

func (pc *PixelController) UpdateOptions(options Options) {
	pc.patternMu.Lock()
	pc.options = options
	pc.patternMu.Unlock()

	pc.SetTransitionDuration(options.TransitionDuration)
}
