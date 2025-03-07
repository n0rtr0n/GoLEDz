package main

func registerColorMasks() map[string]ColorMaskPattern {
	masks := make(map[string]ColorMaskPattern)
	gradientMask := GradientColorMask{
		BasePattern: BasePattern{
			Label: "Gradient",
		},
		Parameters: GradientParameters{
			Color1: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Color2: ColorParameter{
				Value: Color{R: 0, G: 0, B: 255},
				Type:  TYPE_COLOR,
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
			BlendSize: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   1.0,
				Value: 0.50,
				Type:  TYPE_FLOAT,
			},
		},
	}
	solidColorMask := SolidColorMask{
		BasePattern: BasePattern{
			Label: "Solid Color",
		},
		Parameters: SolidColorParameters{
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
		},
	}
	solidColorFadeMask := SolidColorFadeMask{
		BasePattern: BasePattern{
			Label: "Solid Color Fade",
		},
		Parameters: SolidColorFadeParameters{
			Speed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   15.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Color: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
		},
	}
	rainbowDiagonalMask := RainbowDiagonalMask{
		BasePattern: BasePattern{
			Label: "Rainbow Diagonal",
		},
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
	}
	rainbowCircleMask := RainbowCircleMask{
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
	rainbowPinwheelMask := RainbowPinwheelMask{
		BasePattern: BasePattern{
			Label: "Rainbow Pinwheel",
		},
		Parameters: RainbowPinwheelParameters{
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

	waveMask := WaveColorMask{
		BasePattern: BasePattern{
			Label: "Wave Interference",
		},
		Parameters: WaveParameters{
			Color1: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Color2: ColorParameter{
				Value: Color{R: 0, G: 0, B: 255},
				Type:  TYPE_COLOR,
			},
			WaveSpeed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 2.0,
				Type:  TYPE_FLOAT,
			},
			WaveFrequency: FloatParameter{
				Min:   floatPointer(1.0),
				Max:   20.0,
				Value: 5.0,
				Type:  TYPE_FLOAT,
			},
			WaveCount: IntParameter{
				Min:   intPointer(1),
				Max:   10,
				Value: 3,
				Type:  TYPE_INT,
			},
			WaveDirection: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   360.0,
				Value: 45.0,
				Type:  TYPE_FLOAT,
			},
			InterferenceMode: IntParameter{
				Min:   intPointer(0),
				Max:   2,
				Value: 0,
				Type:  TYPE_INT,
			},
			Amplitude: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   2.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
		},
	}

	kaleidoscopeMask := KaleidoscopeColorMask{
		BasePattern: BasePattern{
			Label: "Kaleidoscope",
		},
		Parameters: KaleidoscopeParameters{
			Color1: ColorParameter{
				Value: Color{R: 255, G: 0, B: 0},
				Type:  TYPE_COLOR,
			},
			Color2: ColorParameter{
				Value: Color{R: 0, G: 255, B: 0},
				Type:  TYPE_COLOR,
			},
			Color3: ColorParameter{
				Value: Color{R: 0, G: 0, B: 255},
				Type:  TYPE_COLOR,
			},
			RotationSpeed: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   10.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Segments: IntParameter{
				Min:   intPointer(2),
				Max:   16,
				Value: 6,
				Type:  TYPE_INT,
			},
			ZoomLevel: FloatParameter{
				Min:   floatPointer(0.1),
				Max:   5.0,
				Value: 1.0,
				Type:  TYPE_FLOAT,
			},
			Distortion: FloatParameter{
				Min:   floatPointer(0.0),
				Max:   50.0,
				Value: 10.0,
				Type:  TYPE_FLOAT,
			},
			ColorBlendMode: IntParameter{
				Min:   intPointer(0),
				Max:   3,
				Value: 0,
				Type:  TYPE_INT,
			},
		},
	}

	masks[gradientMask.GetName()] = &gradientMask
	masks[solidColorMask.GetName()] = &solidColorMask
	masks[solidColorFadeMask.GetName()] = &solidColorFadeMask
	masks[rainbowDiagonalMask.GetName()] = &rainbowDiagonalMask
	masks[rainbowCircleMask.GetName()] = &rainbowCircleMask
	masks[rainbowPinwheelMask.GetName()] = &rainbowPinwheelMask
	masks[waveMask.GetName()] = &waveMask
	masks[kaleidoscopeMask.GetName()] = &kaleidoscopeMask

	return masks
}
