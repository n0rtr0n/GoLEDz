package main

const TYPE_FLOAT = "float"
const TYPE_INT = "int"
const TYPE_COLOR = "color"
const TYPE_BOOL = "bool"

func registerPatterns(pixelMap *PixelMap, controller *PixelController) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
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
		Label:      "Rainbow Diagonal",
	}
	rainbowCirclePattern := RainbowCirclePattern{
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
		Label:      "Rainbow Circle",
	}
	rainbowPinwheelPattern := RainbowPinwheelPattern{
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
		Label:      "Rainbow Pinwheel",
	}
	pinwheelPattern := PinwheelPattern{
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
		Label:             "Pinwheel",
	}
	gradientPattern := GradientPattern{
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
		Label: "Gradient",
	}
	solidColorPattern := SolidColorPattern{
		pixelMap: pixelMap,
		Parameters: SolidColorParameters{
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
			},
		},
		Label: "Solid Color",
	}
	lightsOffPattern := LightsOffPattern{
		pixelMap:   pixelMap,
		Label:      "Lights Off",
		Parameters: LightsOffParameters{},
	}
	pulsePattern := PulsePattern{
		pixelMap: pixelMap,
		Parameters: PulseParameters{
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   20.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
		},
		Label: "Pulse",
	}
	sparklePattern := SparklePattern{
		pixelMap: pixelMap,
		Parameters: SparkleParameters{
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
			},
		},
		Label: "Sparkle",
	}
	spiralPattern := SpiralPattern{
		pixelMap: pixelMap,
		Parameters: SpiralParameters{
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
		Label: "Spiral",
	}
	rainbowPattern := RainbowPattern{
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
		Label: "Rainbow",
	}
	solidColorFadePattern := SolidColorFadePattern{
		pixelMap: pixelMap,
		Parameters: SolidColorFadeParameters{
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   50.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
		},
		Label: "Solid Color Fade",
	}
	stripesPattern := StripesPattern{
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
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
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
		Label: "Stripes",
	}

	chaserPattern := ChaserPattern{
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
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: TYPE_COLOR,
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
		Label: "Chaser",
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
