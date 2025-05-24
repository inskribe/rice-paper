package ricepalette

import (
	"image"

	"github.com/disintegration/imaging"
	"github.com/inskribe/rice-paper.git/internal/config"
	"github.com/inskribe/rice-paper.git/internal/hslx"
)

func (request PaletteRequest) CreatePalette(extractor Extractor) (*ColorPalette, error) {
	img := ResizeImage(request.Image)
	base, err := extractor.Extract(img)
	if err != nil {
		return nil, err
	}

	colorPalette := newColorPalette(*base)

	if !request.Silent {
		println("Dark Values")
		colorPalette.DarkValues.Print()
		println("Light Values")
		colorPalette.LightValues.Print()
		println("Accent Values")
		colorPalette.AccentValues.Print()
		println("Status Values")
		colorPalette.StatusValues.Print()
	}
	return colorPalette, nil
}

func ResizeImage(img image.Image) *image.NRGBA {
	desiredSize := config.ImageCompression()
	return imaging.Resize(img, desiredSize, desiredSize, imaging.NearestNeighbor)
}

func newColorPalette(base PaletteBase) *ColorPalette {
	return &ColorPalette{
		DarkValues:   createDarkValues(float64(base.DarkHue)),
		LightValues:  createLightValues(float64(base.LightHue)),
		AccentValues: createAccentValues(float64(base.AccentCenterHue), float64(base.DarkHue)),
		// TODO::Palette
		// Have not decided how to handle status values yet.
		// For now just hard coded values.
		StatusValues: StatusHslValues{
			Info:    hslx.Hsl{H: 40.0, S: 0.74, L: 0.74},
			Hint:    hslx.Hsl{H: hslx.EnsureWrap(float64(base.DarkHue)), S: 0.80, L: 0.80},
			Warn:    hslx.Hsl{H: 17.4, S: 0.50, L: 0.64},
			Error:   hslx.Hsl{H: 355.2, S: 0.45, L: 0.56},
			Success: hslx.Hsl{H: 95.0, S: 0.26, L: 0.65},
		},
	}
}

func createDarkValues(baseHue float64) hslx.HslCollection {
	var result hslx.HslCollection

	for i := range 4 {
		result = append(result, hslx.Hsl{
			H: baseHue + 1.0*float64(i),
			S: 0.17 - 0.01*float64(i),
			L: 0.11 + 0.08*float64(i),
		})
	}
	return result
}

func createLightValues(baseHue float64) hslx.HslCollection {
	var result hslx.HslCollection

	for i := range 4 {
		result = append(result, hslx.Hsl{
			H: hslx.EnsureWrap(baseHue + 1.0*float64(i)),
			S: 0.20 - .01*float64(i),
			L: 0.65 + 0.085*float64(i),
		})
	}
	return result
}

func createAccentValues(baseHue, DominateHue float64) hslx.HslCollection {
	direction := hslx.FindDirection(baseHue, DominateHue)
	if direction == -1 {
		return hslx.HslCollection{
			{H: hslx.EnsureWrap(baseHue + 30), S: 0.45, L: 0.65},
			{H: hslx.EnsureWrap(baseHue + 25), S: 0.40, L: 0.68},
			{H: hslx.EnsureWrap(baseHue + 10), S: 0.35, L: 0.63},
			{H: hslx.EnsureWrap(baseHue), S: 0.30, L: 0.53},
		}
	}

	return hslx.HslCollection{
		{H: hslx.EnsureWrap(baseHue - 30), S: 0.45, L: 0.65},
		{H: hslx.EnsureWrap(baseHue - 25), S: 0.40, L: 0.68},
		{H: hslx.EnsureWrap(baseHue - 10), S: 0.35, L: 0.63},
		{H: hslx.EnsureWrap(baseHue), S: 0.30, L: 0.53},
	}
}
