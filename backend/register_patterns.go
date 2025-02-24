package main

const TYPE_FLOAT = "float"
const TYPE_INT = "int"
const TYPE_COLOR = "color"
const TYPE_BOOL = "bool"

func registerPatterns(pixelMap *PixelMap) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		BasePattern: BasePattern{
			Label: "Rainbow Diagonal",
		},
		pixelMap: pixelMap,
		Parameters: RainbowDiagonalParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   20.0,
				Value: 6.0,
				Type:  TYPE_FLOAT,
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   1.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		currentHue: 0.0,
	}
	rainbowCirclePattern := RainbowCirclePattern{
		BasePattern: BasePattern{
			Label: "Rainbow Circle",
		},
		pixelMap: pixelMap,
		Parameters: RainbowCircleParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   25.0,
				Value: 6.0,
				Type:  TYPE_FLOAT,
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		currentHue: 0.0,
	}
	rainbowPinwheelPattern := RainbowPinwheelPattern{
		BasePattern: BasePattern{
			Label: "Rainbow Pinwheel",
		},
		pixelMap: pixelMap,
		Parameters: RainbowPinwheelParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   25.0,
				Value: 6.0,
				Type:  TYPE_FLOAT,
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  TYPE_BOOL,
			},
		},
		currentHue: 0.0,
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
			Hue: FloatParameter{
				Min:   floatPointer(0),
				Max:   360.0,
				Value: 120.0,
				Type:  TYPE_FLOAT,
			},
			Rainbow: BooleanParameter{
				Value: false,
				Type:  TYPE_BOOL,
			},
		},
		currentSaturation: 0.0,
	}
	gradientPattern := GradientPattern{
		BasePattern: BasePattern{
			Label: "Gradient",
		},
		pixelMap: pixelMap,
		Parameters: GradientParameters{
			Color1: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
			},
			Color2: ColorParameter{
				Value: Color{
					R: 0,
					G: 0,
					B: 255,
				},
				Type: TYPE_COLOR,
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   20.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Reversed: BooleanParameter{
				Value: false,
				Type:  TYPE_BOOL,
			},
		},
	}
	solidColorPattern := SolidColorPattern{
		BasePattern: BasePattern{
			Label: "Solid Color",
		},
		pixelMap: pixelMap,
		Parameters: SolidColorParameters{
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
		},
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
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   20.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
		},
	}
	sparklePattern := SparklePattern{
		BasePattern: BasePattern{
			Label: "Sparkle",
		},
		pixelMap: pixelMap,
		Parameters: SparkleParameters{
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
		},
	}
	spiralPattern := SpiralPattern{
		BasePattern: BasePattern{
			Label: "Spiral",
		},
		pixelMap: pixelMap,
		Parameters: SpiralParameters{
			Color1: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Color2: ColorParameter{
				Value: Color{R: 0, G: 0, B: 255},
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
	rainbowPattern := RainbowPattern{
		BasePattern: BasePattern{
			Label: "Rainbow",
		},
		pixelMap: pixelMap,
		Parameters: RainbowParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Brightness: FloatParameter{
				Min:   floatPointer(1),
				Max:   100,
				Value: 100,
				Type:  TYPE_FLOAT,
			},
		},
	}
	solidColorFadePattern := SolidColorFadePattern{
		BasePattern: BasePattern{
			Label: "Solid Color Fade",
		},
		pixelMap: pixelMap,
		Parameters: SolidColorFadeParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   25.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
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
				Max:   100.0,
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
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Reversed: BooleanParameter{
				Value: false,
				Type:  TYPE_BOOL,
			},
			Rainbow: BooleanParameter{
				Value: false,
				Type:  TYPE_BOOL,
			},
		},
	}

	patterns[rainbowCirclePattern.GetName()] = &rainbowCirclePattern
	patterns[rainbowPinwheelPattern.GetName()] = &rainbowPinwheelPattern
	patterns[pinwheelPattern.GetName()] = &pinwheelPattern
	patterns[gradientPattern.GetName()] = &gradientPattern
	patterns[rainbowDiagonalPattern.GetName()] = &rainbowDiagonalPattern
	patterns[rainbowPattern.GetName()] = &rainbowPattern
	patterns[solidColorPattern.GetName()] = &solidColorPattern
	patterns[lightsOffPattern.GetName()] = &lightsOffPattern
	patterns[solidColorFadePattern.GetName()] = &solidColorFadePattern
	patterns[stripesPattern.GetName()] = &stripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern
	patterns[pulsePattern.GetName()] = &pulsePattern
	patterns[spiralPattern.GetName()] = &spiralPattern
	patterns[sparklePattern.GetName()] = &sparklePattern

	return patterns
}
