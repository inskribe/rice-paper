package ricepalette

import (
	"image"

	"github.com/inskribe/rice-paper.git/internal/hslx"
)

type ColorPalette struct {
	DarkValues   hslx.HslCollection
	LightValues  hslx.HslCollection
	AccentValues hslx.HslCollection
	StatusValues StatusHslValues
}

type StatusHslValues struct {
	Info    hslx.Hsl
	Hint    hslx.Hsl
	Warn    hslx.Hsl
	Error   hslx.Hsl
	Success hslx.Hsl
}

func (values StatusHslValues) Print() {
	values.Info.Print()
	values.Hint.Print()
	values.Warn.Print()
	values.Error.Print()
	values.Success.Print()
}

type PaletteRequest struct {
	Image  image.Image
	Silent bool
}
