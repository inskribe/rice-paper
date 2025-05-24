package hslx

import (
	"fmt"
	"math"

	"github.com/muesli/clusters"
)

// Hsl represents a color in the HSL (Hue, Saturation, Lightness) color space.
//
// H is the hue angle in degrees [0, 360).
// S is the saturation as a fraction [0.0, 1.0].
// L is the lightness as a fraction [0.0, 1.0].
type Hsl struct {
	H, S, L float64
}

type (
	// A slice of Hsl values.
	HslCollection []Hsl
	// A slice of slice containing groups of Hsl values.
	// Mainly for use with hslx.HslCollection.Partition but exposed for convince.
	HslPartitions []HslCollection
)

// Partition groups HSL colors by hue into 6 bins representing segments of the color wheel.
//
// Each bin covers a 60° range of hue values (0–60, 60–120, ..., 300–360).
// The result is a slice of 6 HslCollections, each containing colors that fall into its hue range.
//
// Returns an error if the collection has fewer than 2 elements.
func (collection HslCollection) Partition() (*HslPartitions, error) {
	if len(collection) < 2 {
		return nil, fmt.Errorf("func hslx.Partition: receiver must contain more than one element.")
	}

	result := make(HslPartitions, 6)
	for _, hsl := range collection {
		normalizedHue := math.Mod(hsl.H, 360)
		insertIndex := int(normalizedHue / 60)
		result[insertIndex] = append(result[insertIndex], hsl)
	}
	return &result, nil
}

// Converts a Rgb value to Hsl.
func RgbToHsl(red, green, blue uint8) Hsl {
	r := float64(red) / 255.0
	g := float64(green) / 255.0
	b := float64(blue) / 255.0

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
		panic("unable to determine hue")
	}

	hue *= 60
	if hue < 0 {
		hue += 360
	}

	return Hsl{hue, saturation, luminace}
}

// Rgb will return the Hsl value as Rgb.
func (hsl Hsl) Rgb() (int, int, int) {
	var r, g, b float64
	hue := hsl.H
	saturation := hsl.S
	luminance := hsl.L

	c := (1 - abs(2*luminance-1)) * saturation
	x := c * (1 - abs(float64(math.Mod(float64(hue/60), 2))-1))
	m := luminance - c/2

	switch {
	case hue >= 0 && hue < 60:
		r, g, b = c, x, 0
	case hue >= 60 && hue < 120:
		r, g, b = x, c, 0
	case hue >= 120 && hue < 180:
		r, g, b = 0, c, x
	case hue >= 180 && hue < 240:
		r, g, b = 0, x, c
	case hue >= 240 && hue < 300:
		r, g, b = x, 0, c
	case hue >= 300 && hue < 360:
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

// Hsl to Rgb helper.
func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// Hex will return the hsl value in hex format.
func (hsl Hsl) Hex() string {
	r, g, b := hsl.Rgb()
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// HexList will return a list of hex values from the contents of the HslCollection.
func (collection HslCollection) HexList() []string {
	var hexColors []string

	for _, hsl := range collection {
		hexColors = append(hexColors, hsl.Hex())
	}
	return hexColors
}

// Print will output the contents of a HslCollection.
func (collection HslCollection) Print() {
	for _, hsl := range collection {
		r, g, b := hsl.Rgb()

		// Print out swatch. May not be supported on all terminals.
		fmt.Printf("\x1b[48;2;%d;%d;%dm  \x1b[0m  ", r, g, b)
		// Print out hex and hsl values.
		fmt.Printf(
			"Hex: #%02X%02X%02X HSL H: %.1f S: %.2f L: %.2f \n",
			r, g, b,
			hsl.H, hsl.S, hsl.L,
		)

	}
}

func EnsureWrap(hue float64) float64 {
	hue = math.Mod(hue, 360)
	if hue < 0 {
		hue += 360
	}
	return hue
}

func (hsl *Hsl) Print() {
	r, g, b := hsl.Rgb()

	// Print out swatch. May not be supported on all terminals.
	fmt.Printf("\x1b[48;2;%d;%d;%dm  \x1b[0m  ", r, g, b)
	// Print out hex and hsl values.
	fmt.Printf(
		"Hex: #%02X%02X%02X HSL H: %.1f S: %.2f L: %.2f \n",
		r, g, b,
		hsl.H, hsl.S, hsl.L,
	)
}

// HueInRange returns true if the angular distance from h to center is within tolerance range.
func HueInRange(h, center, tolerance float64) bool {
	dist := HueDistance(h, center)
	return dist <= tolerance
}

// HueDistance returns the shortest distance between two hue angles.
//
// Hue values a and b are in degrees (0–360).
// The result accounts for wraparound on the color wheel.
// Example: HueDistance(30, 357) = 33
func HueDistance(a, b float64) float64 {
	distance := math.Abs(a - b)
	if distance > 180 {
		distance = 360 - distance
	}
	return float64(distance)
}

// SaturationInRange returns true if the distance from s to center is within tolerance range.
func SaturationInRange(s, center, tolerance float64) bool {
	return math.Abs(s-center) <= tolerance
}

func FindDirection(start, destination float64) int {
	distance := math.Abs(start - destination)
	if distance < 180 {
		return 1
	}
	return -1
}

func (collection *HslCollection) RemoveColors(idxSlice []int) {
	var result []Hsl

	for i := range *collection {
		found := false
		for _, idx := range idxSlice {
			if i == idx {
				found = true
			}
		}
		if !found {
			result = append(result, (*collection)[i])
		}
	}
	*collection = result
}

func CreateGradient(start, end Hsl, stepNum int) ([]Hsl, error) {
	if stepNum < 2 {
		return nil, fmt.Errorf("must have two or more steps. stepNum: %d", stepNum)
	}
	result := make([]Hsl, stepNum)
	for i := range stepNum {
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

// Helper to convert Hsl values to coordinate to work with kmeans pkg.
func NormalizeHslToCoordinate(hsl Hsl) clusters.Coordinates {
	return clusters.Coordinates{
		hsl.H / 360,
		hsl.S,
		hsl.L,
	}
}
