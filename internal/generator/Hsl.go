package generator

import (
	"fmt"
	"math"

	"github.com/muesli/clusters"
)

func NormalizeHslToCoordinate(hsl Hsl) clusters.Coordinates {
	return clusters.Coordinates{
		hsl.H / 360,
		hsl.S,
		hsl.L,
	}
}

type Hsl struct {
	H, S, L float64
}

type Rgb struct {
	R, G, B int
}

func RgbToHsl(red, green, blue int) Hsl {
	r := float64(red) / 255
	g := float64(green) / 255
	b := float64(blue) / 255

	minRgb := min(min(r, g), b)
	maxRgb := max(max(r, g), b)

	luminace := (maxRgb + minRgb) / 2

	desaturated := minRgb == maxRgb
	if desaturated {
		return Hsl{0, 0, luminace}
	}

	var saturation float64
	if luminace <= 0.5 {
		saturation = (maxRgb - minRgb) / (maxRgb + minRgb)
	} else {
		saturation = (maxRgb - minRgb) / (2.0 - maxRgb - minRgb)
	}

	var hue float64
	if r == maxRgb {
		hue = (g - b) / (maxRgb - minRgb)
	} else if g == maxRgb {
		hue = 2.0 + (b-r)/(maxRgb-minRgb)
	} else if b == maxRgb {
		hue = 4.0 + (r-g)/(maxRgb-minRgb)
	} else {
		panic("unable to determin hue")
	}

	hue *= 60
	if hue < 0 {
		hue += 360
	}

	return Hsl{hue, saturation, luminace}
}

func HslToRgb(h, s, l float64) (int, int, int) {
	var r, g, b float64

	c := (1 - abs(2*l-1)) * s
	x := c * (1 - abs(float64(math.Mod(float64(h/60), 2))-1))
	m := l - c/2

	switch {
	case h >= 0 && h < 60:
		r, g, b = c, x, 0
	case h >= 60 && h < 120:
		r, g, b = x, c, 0
	case h >= 120 && h < 180:
		r, g, b = 0, c, x
	case h >= 180 && h < 240:
		r, g, b = 0, x, c
	case h >= 240 && h < 300:
		r, g, b = x, 0, c
	case h >= 300 && h < 360:
		r, g, b = c, 0, x
	default:
		r, g, b = 0, 0, 0
	}

	// Scale and clamp to 0–255
	to255 := func(v float64) int {
		return int(math.Round(float64((v + m) * 255)))
	}

	return to255(r), to255(g), to255(b)
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func HslToHex(h Hsl) string {
	r, g, b := HslToRgb(h.H, h.S, h.L)
	return fmt.Sprintf("%02X%02X%02X", r, g, b)
}

func hslToHexList(hslSamples []Hsl) []string {
	var hexColors []string

	for _, hsl := range hslSamples {
		r, g, b := HslToRgb(hsl.H, hsl.S, hsl.L)
		hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
		hexColors = append(hexColors, hex)
	}

	return hexColors
}

func PrintHslColors(c []Hsl) {
	for _, hsl := range c {
		r, g, b := HslToRgb(hsl.H, hsl.S, hsl.L)

		fmt.Printf("\x1b[48;2;%d;%d;%dm  \x1b[0m  ", r, g, b)
		fmt.Printf(
			"Hex=#%02X%02X%02X HSL H=%.1f S=%.2f L=%.2f \n",
			r, g, b,
			hsl.H, hsl.S, hsl.L,
		)

	}
}

func HueInRange(h, center, tolerance float64) bool {
	dist := HueDistance(h, center)
	return dist <= tolerance
}

func HueDistance(a, b float64) float64 {
	angularDist := math.Abs(a - b)
	if angularDist > 180 {
		angularDist = 360 - angularDist
	}
	return float64(angularDist)
}

func SaturationInRange(s, center, tolerance float64) bool {
	return math.Abs(s-center) <= tolerance
}

func removeColors(s []Hsl, idxSlice []int) []Hsl {
	var result []Hsl

	for i := range s {
		found := false
		for _, idx := range idxSlice {
			if i == idx {
				found = true
			}
		}
		if !found {
			result = append(result, s[i])
		}
	}
	return result
}

func CreateGradient(start, end Hsl, stepNum int) ([]Hsl, error) {
	if stepNum < 2 {
		return nil, fmt.Errorf("must have two or more steps. stepNum: %d", stepNum)
	}
	result := make([]Hsl, stepNum)
	for i := 0; i < stepNum; i++ {
		dist := float64(i) / float64(stepNum-1)
		hueDiff := end.H - start.H
		if math.Abs(hueDiff) > 180 {
			if hueDiff > 0 {
				hueDiff -= 360
			} else {
				hueDiff += 360
			}
		}
		hue := start.H + dist*hueDiff
		// Allow for hue to wrap.
		if hue < 0 {
			hue += 360
		} else if hue >= 360 {
			hue -= 360
		}
		// TODO::Need to find a good way to desaturate as I step up in luminance.
		// This will help ensure color vibration is minimized
		saturation := start.S + dist*(end.S-start.S)
		// saturation *= .50
		luminance := start.L + dist*(end.L-start.L)

		result[i] = Hsl{hue, saturation, luminance}
	}
	return result, nil
}

func CreateHistogram(samples []Hsl) [][]Hsl {
	histogram := make([][]Hsl, 6)
	hueErrorNum := 0
	for _, c := range samples {
		idx := int(c.H / 60)
		if idx < 6 {
			histogram[idx] = append(histogram[idx], c)
		} else {
			hueErrorNum++
			if hueErrorNum > 2 {
				/*  TODO::
				* Well I guess I fucked up somewhere.
				* Kick out a safe palette and log out to inform user.
				 */
				panic("Encountered a unknown hue")
			}
		}

	}
	return histogram
}
