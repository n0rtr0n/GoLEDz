package main

func registerPatterns(pixelMap *PixelMap) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap: pixelMap,
		Parameters: RainbowDiagonalParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   100.0,
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
	rainbowPattern := RainbowPattern{
		pixelMap: pixelMap,
		Parameters: RainbowParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
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
		},
		Label: "Chaser",
	}

	patterns[rainbowDiagonalPattern.GetName()] = &rainbowDiagonalPattern
	patterns[rainbowPattern.GetName()] = &rainbowPattern
	patterns[solidColorPattern.GetName()] = &solidColorPattern
	patterns[solidColorFadePattern.GetName()] = &solidColorFadePattern
	patterns[verticalStripesPattern.GetName()] = &verticalStripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern

	return patterns
}
