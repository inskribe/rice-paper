package ricepalette

import "image"

type PaletteBase struct {
	DarkHue         int
	LightHue        int
	AccentCenterHue int

	InfoHue    int
	WarnHue    int
	ErrorHue   int
	SuccessHue int
}

type Extractor interface {
	Extract(img *image.NRGBA) (*PaletteBase, error)
}

func NewPaletteBase() *PaletteBase {
	return &PaletteBase{
		DarkHue:         -1,
		LightHue:        -1,
		AccentCenterHue: -1,
		InfoHue:         -1,
		WarnHue:         -1,
		ErrorHue:        -1,
		SuccessHue:      -1,
	}
}
