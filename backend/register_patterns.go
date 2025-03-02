package main

import (
	"time"
)

const TYPE_FLOAT = "float"
const TYPE_INT = "int"
const TYPE_COLOR = "color"
const TYPE_BOOL = "bool"

func registerPatterns(pixelMap *PixelMap) map[string]Pattern {
	patterns := make(map[string]Pattern)

	maskOnlyPattern := MaskOnlyPattern{
		BasePattern: BasePattern{
			Label: "Color Mask",
		},
		pixelMap:   pixelMap,
		Parameters: MaskOnlyParameters{},
	}

	pinwheelPattern := PinwheelPattern{
		BasePattern: BasePattern{
			Label: "Pinwheel",
		},
		pixelMap: pixelMap,
		Parameters: PinwheelParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.001),
				Max:   0.1,
				Value: 0.02,
				Type:  TYPE_FLOAT,
			},
			Divisions: IntParameter{
				Min:   intPointer(1),
				Max:   15,
				Value: 4,
				Type:  TYPE_INT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		currentSaturation: 0.0,
	}
	lightsOffPattern := LightsOffPattern{
		BasePattern: BasePattern{
			Label: "Lights Off",
		},
		pixelMap:   pixelMap,
		Parameters: LightsOffParameters{},
	}
	pulsePattern := PulsePattern{
		BasePattern: BasePattern{
			Label: "Pulse",
		},
		pixelMap: pixelMap,
		Parameters: PulseParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   3.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
			MinBrightness: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   100.0,
				Value: 10.0,
				Type:  TYPE_FLOAT,
			},
			MaxBrightness: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   100.0,
				Value: 100.0,
				Type:  TYPE_FLOAT,
			},
		},
	}
	sparklePattern := SparklePattern{
		BasePattern: BasePattern{
			Label: "Sparkle",
		},
		pixelMap:   pixelMap,
		Parameters: SparkleParameters{},
	}
	spiralPattern := SpiralPattern{
		BasePattern: BasePattern{
			Label: "Spiral",
		},
		pixelMap: pixelMap,
		Parameters: SpiralParameters{
			BackgroundColor: ColorParameter{
				Value: Color{R: 0, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Speed: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   20.0,
				Value: 8.0,
				Type:  TYPE_FLOAT,
			},
			MaxTurns: IntParameter{
				Min:   intPointer(1),
				Max:   12,
				Value: 4,
				Type:  TYPE_INT,
			},
			Width: FloatParameter{
				Min:   floatPointer(10.0),
				Max:   40.0,
				Value: 30.0,
				Type:  TYPE_FLOAT,
			},
		},
	}
	stripesPattern := StripesPattern{
		BasePattern: BasePattern{
			Label: "Stripes",
		},
		pixelMap: pixelMap,
		Parameters: StripesParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   50.0,
				Value: 10.0,
				Type:  TYPE_FLOAT,
			},
			Size: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   100.0,
				Value: 30.0,
				Type:  TYPE_FLOAT,
			},
			Rotation: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   360.0,
				Value: 0.0,
				Type:  TYPE_FLOAT,
			},
			Stripes: IntParameter{
				Min:   intPointer(0),
				Max:   10,
				Value: 1,
				Type:  TYPE_INT,
			},
		},
	}

	chaserPattern := ChaserPattern{
		BasePattern: BasePattern{
			Label: "Chaser",
		},
		pixelMap: pixelMap,
		Parameters: ChaserParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   5.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Size: IntParameter{
				Min:   intPointer(1),
				Max:   100,
				Value: 10,
				Type:  TYPE_INT,
			},
			Spacing: IntParameter{
				Min:   intPointer(1),
				Max:   100,
				Value: 10,
				Type:  TYPE_INT,
			},
			Reversed: BooleanParameter{
				Value: false,
				Type:  TYPE_BOOL,
			},
		},
	}

	// create and register the random pattern
	randomPattern := RandomPattern{
		BasePattern: BasePattern{
			Label: "Random",
		},
		pixelMap: pixelMap,
		patterns: patterns, // this will be empty initially, we'll fix it below
		Parameters: RandomPatternParameters{
			SwitchInterval: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   60.0,
				Value: 15.0,
				Type:  TYPE_FLOAT,
			},
			RandomizeColorMasks: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
			TransitionTime: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
		},
		lastSwitchTime: time.Now(),
		inTransition:   false,
	}

	// Register all patterns first
	patterns[maskOnlyPattern.GetName()] = &maskOnlyPattern
	patterns[pinwheelPattern.GetName()] = &pinwheelPattern
	patterns[lightsOffPattern.GetName()] = &lightsOffPattern
	patterns[stripesPattern.GetName()] = &stripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern
	patterns[pulsePattern.GetName()] = &pulsePattern
	patterns[spiralPattern.GetName()] = &spiralPattern
	patterns[sparklePattern.GetName()] = &sparklePattern

	// Now add the random pattern with access to all other patterns
	randomPattern.patterns = patterns
	patterns[randomPattern.GetName()] = &randomPattern

	return patterns
}
