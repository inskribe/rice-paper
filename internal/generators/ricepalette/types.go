package ricepalette

import (
	"image"

	"github.com/inskribe/rice-paper.git/internal/hslx"
)

type ColorPalette struct {
	DarkValues   hslx.HslCollection
	LightValues  hslx.HslCollection
	AccentValues hslx.HslCollection
	StatusValues hslx.HslCollection
}

type PaletteRequest struct {
	Image image.Image
}
