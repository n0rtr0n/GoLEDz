package main

func registerPatterns(pixelMap *PixelMap) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap: pixelMap,
		Parameters: RainbowDiagonalParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   25.0,
				Value: 6.0,
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   0.1,
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
				Min:   0.1,
				Max:   25.0,
				Value: 6.0,
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   0.1,
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
				Min:   0.1,
				Max:   25.0,
				Value: 6.0,
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   0.1,
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
				Min:   0.001,
				Max:   0.1,
				Value: 0.01,
				Type:  "float",
			},
			Divisions: IntParameter{
				Min:   1,
				Max:   30,
				Value: 1,
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
				Min:   0.1,
				Max:   20.0,
				Value: 1.0,
				Type:  "float",
			},
		},
		Label: "Pulse",
	}
	rainbowPattern := RainbowPattern{
		pixelMap: pixelMap,
		Parameters: RainbowParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
				Type:  "float",
			},
			Brightness: FloatParameter{
				Min:   1,
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
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
				Type:  "float",
			},
		},
		Label: "Solid Color Fade",
	}

	verticalStripesPattern := VerticalStripesPattern{
		pixelMap: pixelMap,
		Parameters: VerticalStripesParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   100.0,
				Value: 3.0,
				Type:  "float",
			},
			Size: FloatParameter{
				Min:   1.0,
				Max:   100.0,
				Value: 20.0,
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
		},
		Label: "Vertical Stripes",
	}

	chaserPattern := ChaserPattern{
		pixelMap: pixelMap,
		Parameters: ChaserParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
				Type:  "float",
			},
			Size: IntParameter{
				Min:   1,
				Max:   100,
				Value: 5,
				Type:  "int",
			},
			Spacing: IntParameter{
				Min:   1,
				Max:   100,
				Value: 5,
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
	patterns[rainbowDiagonalPattern.GetName()] = &rainbowDiagonalPattern
	patterns[rainbowPattern.GetName()] = &rainbowPattern
	patterns[solidColorPattern.GetName()] = &solidColorPattern
	patterns[solidColorFadePattern.GetName()] = &solidColorFadePattern
	patterns[verticalStripesPattern.GetName()] = &verticalStripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern
	patterns[pulsePattern.GetName()] = &pulsePattern

	return patterns
}
