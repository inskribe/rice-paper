package ricepalette

import (
	"fmt"
	"image"
	"sort"

	"github.com/inskribe/rice-paper.git/internal/hslx"
)

type DeafultExtractor struct {
	GroupingFactor int
}
type HueBin struct {
	Count int
	Hsl   hslx.Hsl
}

func (extractor DeafultExtractor) Extract(img *image.NRGBA) (*PaletteBase, error) {
	base := NewPaletteBase()
	bounds := img.Bounds()
	bins := make([]HueBin, extractor.GroupingFactor)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := img.NRGBAAt(x, y)
			hsl := hslx.RgbToHsl(pixel.R, pixel.G, pixel.B)
			hue := int(hsl.H)
			bins[hue].Count++
			bins[hue].Hsl = hsl
		}
	}

	sort.Slice(bins, func(i, j int) bool { return bins[i].Count > bins[j].Count })
	dominateHue, err := findSuitableHue(bins)
	if err != nil {
		return nil, err
	}
	base.DarkHue = dominateHue
	base.LightHue = dominateHue

	var shiftTolerance float64 = 20

	for _, bin := range bins {
		hsl := bin.Hsl
		if !hslx.HueInRange(hsl.H, bins[0].Hsl.H, shiftTolerance) {
			if hsl.S < 0.3 || hsl.L < 0.2 || hsl.L > 0.8 {
				continue
			}
			if hslx.HueInRange(hsl.H, float64(base.DarkHue), 5) {
				continue
			}
			base.AccentCenterHue = int(hsl.H)
			break
		}
	}
	if base.AccentCenterHue == -1 {
		return nil, fmt.Errorf("unable to find accent hue center")
	}
	return base, nil
}

func findSuitableHue(bins []HueBin) (int, error) {
	for _, bin := range bins {
		if bin.Hsl.L < 0.10 {
			continue
		}
		return int(bin.Hsl.H), nil
	}
	return -1, fmt.Errorf("failed to find suitable dominate hue")
}
