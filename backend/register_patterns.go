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

	ripplePattern := RipplePattern{
		BasePattern: BasePattern{
			Label: "Ripple",
		},
		pixelMap: pixelMap,
		Parameters: RippleParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   5.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			RippleCount: IntParameter{
				Min:   intPointer(4),
				Max:   20,
				Value: 10,
				Type:  TYPE_INT,
			},
			RippleWidth: FloatParameter{
				Min:   floatPointer(5.0),
				Max:   50.0,
				Value: 20.0,
				Type:  TYPE_FLOAT,
			},
			RippleLifetime: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   15.0,
				Value: 8.0,
				Type:  TYPE_FLOAT,
			},
			BackgroundColor: ColorParameter{
				Value: Color{R: 0, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			AutoGenerate: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		lastUpdate: time.Time{},
	}

	matrixPattern := MatrixPattern{
		BasePattern: BasePattern{
			Label: "Matrix",
		},
		pixelMap: pixelMap,
		Parameters: MatrixParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 4.0,
				Type:  TYPE_FLOAT,
			},
			Density: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   5.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
			DropLength: FloatParameter{
				Min:   floatPointer(20.0),
				Max:   300.0,
				Value: 150.0,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
	}

	firePattern := FirePattern{
		BasePattern: BasePattern{
			Label: "Fire",
		},
		pixelMap: pixelMap,
		Parameters: FireParameters{
			Cooling: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 3.0,
				Type:  TYPE_FLOAT,
			},
			Sparking: FloatParameter{
				Min:   floatPointer(0.01),
				Max:   1.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 3.0,
				Type:  TYPE_FLOAT,
			},
			ColorScheme: IntParameter{
				Min:   intPointer(0),
				Max:   3,
				Value: 0, // 0=classic, 1=blue, 2=green, 3=purple
				Type:  TYPE_INT,
			},
			WindDirection: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   360.0,
				Value: 90.0,
				Type:  TYPE_FLOAT,
			},
			WindStrength: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   5.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
		},
	}

	plasmaPattern := PlasmaPattern{
		BasePattern: BasePattern{
			Label: "Plasma",
		},
		pixelMap: pixelMap,
		Parameters: PlasmaParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   5.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
			Scale: FloatParameter{
				Min:   floatPointer(0.5),
				Max:   10.0,
				Value: 4.0,
				Type:  TYPE_FLOAT,
			},
			Complexity: FloatParameter{
				Min:   floatPointer(0.5),
				Max:   10.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
			ColorShift: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   360.0,
				Value: 0.0,
				Type:  TYPE_FLOAT,
			},
		},
	}

	// audioReactivePattern := AudioReactivePattern{
	// 	BasePattern: BasePattern{
	// 		Label: "Audio Reactive",
	// 	},
	// 	pixelMap: pixelMap,
	// 	Parameters: AudioReactiveParameters{
	// 		Sensitivity: FloatParameter{
	// 			Min:   floatPointer(0.1),
	// 			Max:   5.0,
	// 			Value: 1.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		ColorSpeed: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   10.0,
	// 			Value: 2.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		BaseColor: ColorParameter{
	// 			Value: Color{R: 0, G: 0, B: 40, W: 0}, // Dark blue
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		AccentColor: ColorParameter{
	// 			Value: Color{R: 0, G: 200, B: 255, W: 0}, // Bright cyan
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		EffectType: IntParameter{
	// 			Min:   intPointer(0),
	// 			Max:   2,
	// 			Value: 0, // 0=pulse, 1=wave, 2=sparkle
	// 			Type:  TYPE_INT,
	// 		},
	// 		SmoothingTime: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   1.0,
	// 			Value: 0.2,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 	},
	// }

	// particlesPattern := ParticlePattern{
	// 	BasePattern: BasePattern{
	// 		Label: "Particles",
	// 	},
	// 	pixelMap: pixelMap,
	// 	Parameters: ParticleParameters{
	// 		EmissionRate: FloatParameter{
	// 			Min:   floatPointer(1.0),
	// 			Max:   100.0,
	// 			Value: 20.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		ParticleLife: FloatParameter{
	// 			Min:   floatPointer(0.5),
	// 			Max:   10.0,
	// 			Value: 3.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		Gravity: FloatParameter{
	// 			Min:   floatPointer(-50.0),
	// 			Max:   50.0,
	// 			Value: 20.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		InitialColor: ColorParameter{
	// 			Value: Color{R: 255, G: 200, B: 0, W: 0}, // Bright yellow
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		FinalColor: ColorParameter{
	// 			Value: Color{R: 255, G: 0, B: 0, W: 0}, // Red
	// 			Type:  TYPE_COLOR,
	// 		},
	// 		ParticleSize: FloatParameter{
	// 			Min:   floatPointer(1.0),
	// 			Max:   30.0,
	// 			Value: 10.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		EmitterX: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   1.0,
	// 			Value: 0.5, // Center of display horizontally
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		EmitterY: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   1.0,
	// 			Value: 0.8, // Near bottom of display
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		SpreadAngle: FloatParameter{
	// 			Min:   floatPointer(0.0),
	// 			Max:   360.0,
	// 			Value: 120.0, // 120 degree spread
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 		ParticleSpeed: FloatParameter{
	// 			Min:   floatPointer(10.0),
	// 			Max:   200.0,
	// 			Value: 50.0,
	// 			Type:  TYPE_FLOAT,
	// 		},
	// 	},
	// }

	// Register all patterns first
	patterns[maskOnlyPattern.GetName()] = &maskOnlyPattern
	patterns[pinwheelPattern.GetName()] = &pinwheelPattern
	patterns[lightsOffPattern.GetName()] = &lightsOffPattern
	patterns[stripesPattern.GetName()] = &stripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern
	patterns[pulsePattern.GetName()] = &pulsePattern
	patterns[spiralPattern.GetName()] = &spiralPattern
	patterns[sparklePattern.GetName()] = &sparklePattern
	patterns[ripplePattern.GetName()] = &ripplePattern
	patterns[matrixPattern.GetName()] = &matrixPattern
	patterns[firePattern.GetName()] = &firePattern
	patterns[plasmaPattern.GetName()] = &plasmaPattern
	// patterns[particlesPattern.GestName()] = &particlesPattern
	// patterns[audioReactivePattern.GetName()] = &audioReactivePattern

	// now add the random pattern with access to all other patterns
	randomPattern.patterns = patterns
	patterns[randomPattern.GetName()] = &randomPattern

	return patterns
}
