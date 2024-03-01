package main

func registerPatterns(pixelMap *PixelMap) map[string]Pattern {
	patterns := make(map[string]Pattern)

	rainbowDiagonalPattern := RainbowDiagonalPattern{
		pixelMap: pixelMap,
		Parameters: RainbowDiagonalParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   360.0,
				Value: 6.0,
			},
			Size: FloatParameter{
				Min:   0.1,
				Max:   360.0,
				Value: 0.5,
			},
			Reversed: BooleanParameter{
				Value: true,
			},
		},
		currentHue: 0.0,
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
			},
		},
	}
	rainbowPattern := RainbowPattern{
		pixelMap: pixelMap,
		Parameters: RainbowParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
			},
		},
	}
	solidColorFadePattern := SolidColorFadePattern{
		pixelMap: pixelMap,
		Parameters: SolidColorFadeParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   50.0,
				Value: 1.0,
			},
		},
	}

	verticalStripesPattern := VerticalStripesPattern{
		pixelMap: pixelMap,
		Parameters: VerticalStripesParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   100.0,
				Value: 3.0,
			},
			Size: FloatParameter{
				Min:   1.0,
				Max:   360.0,
				Value: 20.0,
			},
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
			},
		},
	}

	chaserPattern := ChaserPattern{
		pixelMap: pixelMap,
		Parameters: ChaserParameters{
			Speed: FloatParameter{
				Min:   0.1,
				Max:   360.0,
				Value: 1.0,
			},
			Size: IntParameter{
				Min:   0,
				Max:   500,
				Value: 5,
			},
			Spacing: IntParameter{
				Min:   0,
				Max:   500,
				Value: 5,
			},
			Color: ColorParameter{
				Value: Color{
					R: 255,
					G: 0,
					B: 0,
				},
			},
		},
	}

	patterns[rainbowDiagonalPattern.GetName()] = &rainbowDiagonalPattern
	patterns[rainbowPattern.GetName()] = &rainbowPattern
	patterns[solidColorPattern.GetName()] = &solidColorPattern
	patterns[solidColorFadePattern.GetName()] = &solidColorFadePattern
	patterns[verticalStripesPattern.GetName()] = &verticalStripesPattern
	patterns[chaserPattern.GetName()] = &chaserPattern

	return patterns
}
