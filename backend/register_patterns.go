package main

func registerPatterns(pixelMap *PixelMap, controller *PixelController) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap: pixelMap,
		Parameters: RainbowDiagonalParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   25.0,
				Value: 6.0,
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  "float",
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  "bool",
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
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  "float",
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  "bool",
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
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   100.0,
				Value: 0.5,
				Type:  "float",
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  "bool",
			},
		},
		currentHue: 0.0,
		Label:      "Rainbow Pinwheel",
	}
	gradientPinwheelPattern := GradientPinwheelPattern{
		pixelMap: pixelMap,
		Parameters: GradientPinwheelParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.001),
				Max:   0.1,
				Value: 0.02,
				Type:  "float",
			},
			Divisions: IntParameter{
				Min:   intPointer(1),
				Max:   15,
				Value: 4,
				Type:  "int",
			},
			Reversed: BooleanParameter{
				Value: true,
				Type:  "bool",
			},
		},
		currentSaturation: 0.0,
		Label:             "Gradient Pinwheel",
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
				Type: "color",
			},
			Color2: ColorParameter{
				Value: Color{
					R: 0,
					G: 0,
					B: 255,
				},
				Type: "color",
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   20.0,
				Value: 1.0,
				Type:  "float",
			},
			Reversed: BooleanParameter{
				Value: false,
				Type:  "bool",
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
				Type: "color",
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
				Type: "color",
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   20.0,
				Value: 1.0,
				Type:  "float",
			},
		},
		Label: "Pulse",
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
				Type: "color",
			},
			Color2: ColorParameter{
				Value: Color{
					R: 0,
					G: 0,
					B: 255,
				},
				Type: "color",
			},
			Speed: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   20.0,
				Value: 8.0,
				Type:  "float",
			},
			MaxTurns: IntParameter{
				Min:   intPointer(1),
				Max:   15,
				Value: 3,
				Type:  "int",
			},
			Width: FloatParameter{
				Min:   floatPointer(10.0),
				Max:   60.0,
				Value: 40.0,
				Type:  "float",
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
				Type:  "float",
			},
			Brightness: FloatParameter{
				Min:   floatPointer(1),
				Max:   100,
				Value: 100,
				Type:  "float",
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
				Type: "color",
			},
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   50.0,
				Value: 1.0,
				Type:  "float",
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
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   100.0,
				Value: 30.0,
				Type:  "float",
			},
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: "color",
			},
			Rotation: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   360.0,
				Value: 0.0,
				Type:  "float",
			},
			Stripes: IntParameter{
				Min:   intPointer(0),
				Max:   10,
				Value: 1,
				Type:  "int",
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
				Type:  "float",
			},
			Size: IntParameter{
				Min:   intPointer(1),
				Max:   100,
				Value: 10,
				Type:  "int",
			},
			Spacing: IntParameter{
				Min:   intPointer(1),
				Max:   100,
				Value: 10,
				Type:  "int",
			},
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
				Type: "color",
			},
			Reversed: BooleanParameter{
				Value: false,
				Type:  "bool",
			},
			Rainbow: BooleanParameter{
				Value: false,
				Type:  "bool",
			},
		},
		Label: "Chaser",
	}

	patterns[rainbowCirclePattern.GetName()] = &rainbowCirclePattern
	patterns[rainbowPinwheelPattern.GetName()] = &rainbowPinwheelPattern
	patterns[gradientPinwheelPattern.GetName()] = &gradientPinwheelPattern
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

	return patterns
}
