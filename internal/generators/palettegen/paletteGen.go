package palettegen

import (
	"fmt"
	"image"
	"image/color"
	"rice-paper/internal/config"
	"slices"
	"sort"

	"github.com/disintegration/imaging"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

// RicePalette represents a generated color palette consisting of a primary (main) palette
// and a secondary (accent) palette, along with a classification of the palette type.
type RicePalette struct {
	MainPalette   []Hsl       // The primary color palette extracted from the image.
	AccentPalette []Hsl       // The secondary or supporting color palette.
	PaletteType   PaletteType // The type classification of the palette (e.g., Greyscale, Monochromatic, Dynamic).
}

// HslPalette contains processed HSL color data used to construct a RicePalette.
// It holds statistical and range information useful for determining palette type and structure.
type HslPalette struct {
	HslSamples  []Hsl   // Raw HSL samples extracted from image clusters.
	TotalHue    float64 // Sum of hue values, weighted by saturation (used for average hue calculation).
	TotalWeight float64 // Sum of saturation weights used in weighted hue averaging.
	GreyGuard   float64 // Aggregate saturation used to assess if the image is greyscale.
	MonoMin     float64 // Minimum hue value observed (used for monochromatic range checks).
	MonoMax     float64 // Maximum hue value observed (used for monochromatic range checks).
}

// GenrateColorPalette generates a RicePalette from the provided image.Image.
//
// The function performs the following steps:
//   - Resizes the input image for consistent analysis.
//   - Extracts color data and clusters it using K-Means.
//   - Converts clustered RGB colors to HSL.
//   - Classifies the palette type (Greyscale, Monochromatic, or Dynamic).
//   - Builds and refines the main and accent palettes accordingly.
//
// Returns a pointer to a RicePalette containing the generated palettes,
// or an error if any step in the pipeline fails.
func GenrateColorPalette(img image.Image) (*RicePalette, error) {
	img, err := resizeImage(img)
	if err != nil {
		return nil, err
	}

	samples, err := extractColorObservations(img)
	if err != nil {
		return nil, err
	}

	dominatesColors, err := extractDominateClusters(samples)
	if err != nil {
		return nil, err
	}

	hslPalette, err := convertToHsl(dominatesColors)
	if err != nil {
		return nil, err
	}
	ricePalete, err := newRicePalete(hslPalette)
	if err != nil {
		return nil, err
	}

	switch ricePalete.PaletteType {
	case Unknown:
		return nil, fmt.Errorf("generator: failed to determine palette type.")
	case Greyscale:
		err = generateGreyScalePalette(ricePalete)
		if err != nil {
			return nil, err
		}
		return ricePalete, nil
	case Monochromatic:
		err = generateMonochromaticPalette(ricePalete)
		if err != nil {
			return nil, err
		}
		return ricePalete, nil
	case Dynamic:
		err = ensurePaletteVariation(ricePalete)
		if err != nil {
			return nil, err
		}
		err = vaildateAccentPalette(ricePalete)
		if err != nil {
			return nil, err
		}
		err = tightenPalette(&ricePalete.MainPalette, 0, 0, 0, 0)
		if err != nil {
			return nil, err
		}
		err = createMainPaletteGradient(&ricePalete.MainPalette)
		if err != nil {
			return nil, err
		}
		err = createAccentPaletteGradient(&ricePalete.AccentPalette)
		if err != nil {
			return nil, err
		}
	}
	return ricePalete, nil
}

// resizeImage will resize the given image to the size defined in the application config.
// See also: config.Config.Generator and git_root/application_config.yml
// Returns a new image.Image at the defined size and error if any, otherwise nil.
func resizeImage(img image.Image) (image.Image, error) {
	if img == nil {
		return nil, fmt.Errorf("generator: Expected type image.Image and received nil")
	}

	width := config.ImageCompression()
	height := config.ImageCompression()
	return imaging.Resize(img, width, height, imaging.NearestNeighbor), nil
}

// extractColorObservations extracts pixel data into Observations.
// Returns a slice of clusters.Observation containing normalized pixel data and error if any, otherwise nil.
func extractColorObservations(img image.Image) (*clusters.Observations, error) {
	if img == nil {
		return nil, fmt.Errorf("generator: invalid image or palette.")
	}

	var samples clusters.Observations

	imgBounds := img.Bounds()
	for y := imgBounds.Min.Y; y < imgBounds.Max.X; y++ {
		for x := imgBounds.Min.X; x < imgBounds.Max.X; x++ {
			pxColor := img.At(x, y)
			samples = append(samples, normalizeColorToCoordinate(pxColor))
		}
	}
	return &samples, nil
}

// normalizeColorToCoordinate will normalize a color.Color to a vector of float64 in range 0 - 1 and
// pack it into a coordinate to be used by kmeans.
// The color.Color package does not support Hsl values. Colors are converted to color.RGBA
// and the alpha is ignored.
// Returns a coordinate containing the normalized data.
// Note: See generator.normalizeHslToCoordinate for hsl normalization.
// TODO::generator Early conversion to Hsl to.
func normalizeColorToCoordinate(color color.Color) clusters.Coordinates {
	r, g, b, _ := color.RGBA()
	return clusters.Coordinates{
		float64(r>>8) / 255,
		float64(g>>8) / 255,
		float64(b>>8) / 255,
	}
}

// extractDominateClusters uses the kmeans algorithm to cluster pixels of similar color.
// See also: config.Config.Generator and git_root/application_config.yml for configuration.
// Returns a pointer to the partitioned clusters.Cluster and error if any.
func extractDominateClusters(samples *clusters.Observations) (*clusters.Clusters, error) {
	threshold := config.KmeansThreshold()
	km, err := kmeans.NewWithOptions(threshold, nil)
	if err != nil {
		return nil, err
	}

	result, err := km.Partition(*samples, config.KmeansPartionCount())
	return &result, err
}

// convertToHsl will convert a slice of cluster.Cluster to a HslPalette.
// Returns a pointer to a new HslPalette and error if any, otherwise nil.
func convertToHsl(samples *clusters.Clusters) (*HslPalette, error) {
	if samples == nil {
		return nil, fmt.Errorf("generator: samples and palette must be valid.")
	}
	result := HslPalette{
		MonoMin: 360,
		MonoMax: 0,
	}
	for _, sample := range *samples {
		centroid := sample.Center.Coordinates()
		r := int(centroid[0] * 255)
		g := int(centroid[1] * 255)
		b := int(centroid[2] * 255)
		hsl := RgbToHsl(r, g, b)

		if hsl.S > 0.1 {
			result.TotalHue += hsl.H
			result.TotalWeight += hsl.S

			if hsl.H < result.MonoMin {
				result.MonoMin = hsl.H
			}
			if hsl.H > result.MonoMax {
				result.MonoMax = hsl.H
			}
		}
		result.GreyGuard += hsl.S
		result.HslSamples = append(result.HslSamples, hsl)
	}
	sort.Slice(result.HslSamples, func(i, j int) bool { return result.HslSamples[i].L < result.HslSamples[j].L })
	return &result, nil
}

// newRicePalete creates a RicePalette form the provided HslPalette.
// Note: Main and accent palette distribution is based on histogram scanning
// Largest histogram color band wins main palette and remaining are assigned to accent.
// Returns pointer to new RicePalette and error if any, otherwise nil.
func newRicePalete(hslPalette *HslPalette) (*RicePalette, error) {
	if hslPalette == nil {
		return nil, fmt.Errorf("generator: expected vaild HslPalette.")
	}
	if len(hslPalette.HslSamples) < config.KmeansPartionCount() {
		return nil, fmt.Errorf("generator: expected a hsl sample size of %d, recived %d", config.KmeansPartionCount(), len(hslPalette.HslSamples))
	}
	var result RicePalette
	result.PaletteType = Unknown
	if hslPalette.GreyGuard == 0 {
		result.PaletteType = Greyscale
	} else if hslPalette.MonoMax-hslPalette.MonoMin < 5 && result.PaletteType != Greyscale {
		result.PaletteType = Monochromatic
	} else {
		result.PaletteType = Dynamic
	}

	println("Creating Histogram")
	histogramResult := CreateHistogram(hslPalette.HslSamples)
	largestBinIdx := 0
	for i := 0; i < config.HistogramBinCount(); i++ {
		if len(histogramResult[i]) > len(histogramResult[largestBinIdx]) {
			largestBinIdx = i
		}
	}

	result.MainPalette = histogramResult[largestBinIdx]
	for i, bin := range histogramResult {
		if i != largestBinIdx {
			result.AccentPalette = append(result.AccentPalette, bin...)
		}
	}

	println("Main Palette")
	PrintHslColors(result.MainPalette)
	println("Accent Palette")
	PrintHslColors(result.AccentPalette)
	return &result, nil
}

// ensurePaletteVariation ensures that there is variation of hue,saturation, and luminance.
// See also: config.Config.Generator and git_root/application_config.yml to configure
// tolerance values.
// Note: On some images if there is a vast amount of a single color it can influence
// what bin is selected as the main palette.
// example: an image that has a monochromatic background with simple geometric shapes of another color.
// ensurePaletteVariation aims to solve this by preforming a secondary pass if necessary.
// Returns error if any, otherwise nil.
func ensurePaletteVariation(palette *RicePalette) error {
	if palette == nil {
		return fmt.Errorf("generator: expected pointer to a RicePalette, received nil.")
	}
	minH, maxH := 360., 0.0
	minS, maxS := 1.0, 0.0
	minL, maxL := 1.0, 0.0

	for _, hsl := range palette.MainPalette {
		if hsl.H < minH {
			minH = hsl.H
		}
		if hsl.H > maxH {
			maxH = hsl.H
		}
		if hsl.S < minS {
			minS = hsl.S
		}
		if hsl.S > maxS {
			maxS = hsl.S
		}
		if hsl.L < minL {
			minL = hsl.L
		}
		if hsl.L > maxL {
			maxL = hsl.L
		}
	}

	hueVariation := maxH - minH
	saturationVariation := maxS - minS
	luminanceVaraition := maxL - minL

	hueTolerance := config.HueVariationTolerance()
	saturationTolerance := config.SaturationVariationTolerance()
	luminanceTolerance := config.LuminanceVariationTolerance()

	if hueVariation < hueTolerance && saturationVariation < saturationTolerance && luminanceVaraition < luminanceTolerance {
		fmt.Println("Low variation detected in palette. Performing secondary pass.")
		secondPass := slices.Clone(palette.AccentPalette)
		secondPass = append(secondPass, palette.MainPalette[0])
		palette.MainPalette = nil

		histogram := CreateHistogram(secondPass)
		maxBandIdx := 0
		for i := 0; i < 6; i++ {
			if len(histogram[i]) > len(histogram[maxBandIdx]) {
				maxBandIdx = i
			}
		}
		palette.MainPalette = histogram[maxBandIdx]
	} else {
		println("Palette has good variation.")
	}
	return nil
}

// vaildateAccentPalette ensures there is enough colors in the palette to
// create accent gradients. If less than configured amount will attempt to
// grab outliers from the main palette.
// Returns error if any, otherwise nil.
func vaildateAccentPalette(palette *RicePalette) error {
	if palette == nil {
		return fmt.Errorf("generator: Expected pointer to  RicePalette, received nil.")
	}
	if len(palette.AccentPalette) == 0 || len(palette.MainPalette) == 0 {
		return fmt.Errorf("generator: Hsl slice is empty.")
	}
	// TODO:generator Rip out to palette parsing.
	var colorsToRemove []int
	for i, hsl := range palette.AccentPalette {
		if hsl.L < config.LuminanceMin() || hsl.L > config.LuminanceMax() {
			colorsToRemove = append(colorsToRemove, i)
		}
	}
	palette.AccentPalette = removeColors(palette.AccentPalette, colorsToRemove)
	// TODO:generator Move accent swatch count to config.
	if len(palette.AccentPalette) < 4 {
		fmt.Println("Extending accent palette.")
		for _, h := range palette.MainPalette {
			// TODO::generator Move accent thresholds to config.
			if h.S >= 0.4 && h.L > 0.20 && h.L < 0.9 {
				palette.AccentPalette = append(palette.AccentPalette, h)
			}
		}
	} else {
		println("Valid Accent Palette.")
	}

	return nil
}

// tightenPalette aims to remove any colors that don't fit withing the
// configured specifications.
// See also: config.Config.Generator and git_root/application_config.yml
// Returns error if any, otherwise nil.
func tightenPalette(palette *[]Hsl, hueThreshold, saturationThreshold, luminanceMin, luminanceMax float64) error {
	if palette == nil {
		return fmt.Errorf("generator: Expected pointer to Hsl slice, received nil.")
	}

	ohFuck := false
	if hueThreshold != 0 || saturationThreshold != 0 || luminanceMin != 0 || luminanceMax != 0 {
		ohFuck = true
	}
	if hueThreshold == 0 {
		hueThreshold = config.HueShiftTolerance()
	}
	if saturationThreshold == 0 {
		saturationThreshold = config.SaturationShiftTolerance()
	}
	if luminanceMin == 0 {
		luminanceMin = config.LuminanceMin()
	}
	if luminanceMax == 0 {
		luminanceMax = config.LuminanceMax()
	}

	println("Tightening Palette")
	var averageSaturation float64 = 0.0
	var averageHue float64 = 0.0

	for _, hsl := range *palette {
		averageHue += hsl.H
		averageSaturation += hsl.S
	}

	averageHue = averageHue / float64(len(*palette))
	averageSaturation = averageSaturation / float64(len(*palette))
	var colorsToRemove []int

	for i, hsl := range *palette {
		if !HueInRange(hsl.H, averageHue, hueThreshold) {
			colorsToRemove = append(colorsToRemove, i)
			continue
		}
		if !SaturationInRange(hsl.S, averageSaturation, saturationThreshold) {
			colorsToRemove = append(colorsToRemove, i)
			continue
		}
		if hsl.L < luminanceMin || hsl.L > luminanceMax {
			colorsToRemove = append(colorsToRemove, i)
		}
	}

	desiredColors := removeColors(*palette, colorsToRemove)

	/*
	 * If the user's configured thresholds are too strict, the palette may be empty
	 * after filtering. Rather than failing immediately, we attempt a fallback pass
	 * with looser constraints to salvage usable colors.
	 *
	 * This fallback is primarily intended to preserve color information for
	 * gradient generation, even if it means allowing less ideal colors. The
	 * gradient system is expected to realign the final output to the user's
	 * preferences.
	 */
	if len(desiredColors) <= 0 {
		if ohFuck {
			return fmt.Errorf("generator: No colors within palette fit configured settings.")
		}
		// Fallback to looser constraints
		const (
			fallbackHueThreshold        = 0
			fallbackSaturationThreshold = 0.40
			fallbackLuminanceMin        = 0.10
			fallbackLuminanceMax        = 0.80 // was 0, which might be invalid
		)
		err := tightenPalette(palette, fallbackHueThreshold, fallbackSaturationThreshold, fallbackLuminanceMin, fallbackLuminanceMax)
		if err != nil {
			return err
		}
		return nil
	}
	*palette = desiredColors

	return nil
}

// createMainPaletteGradient will create a gradient from the provided hsl slice.
// Returns error if any, nil otherwise.
func createMainPaletteGradient(palette *[]Hsl) error {
	if palette == nil {
		return fmt.Errorf("generator: Expected pointer to Hsl slice, received nil.")
	}
	if len(*palette) < 1 {
		return fmt.Errorf("generator: Hsl slice is empty.")
	}
	start := (*palette)[0]
	var end Hsl
	if len(*palette) >= 2 {
		end = (*palette)[len(*palette)-1]
	} else {
		end = Hsl{
			start.H + config.HueShiftTolerance(),
			start.S,
			config.LuminanceMax(),
		}
	}

	// Ensure good coverage of saturation and luminance
	if start.L > .30 {
		start.L = .25
	}
	if end.L < .70 {
		end.L = .80
	}
	if start.S > .40 {
		end.S *= .50
	}
	println("Creating Main Palette gradient.")
	gradient, err := CreateGradient(start, end, config.PaletteSwatchCount())
	if err != nil {
		return err
	}
	*palette = gradient
	return nil
}

// createAccentPaletteGradient creates a gradient for each hsl color in the provided hsl slice.
// Returns error if any, otherwise nil.
// TODO:generator Add configuration support for accent gradients.
func createAccentPaletteGradient(accentPalette *[]Hsl) error {
	if accentPalette == nil {
		return fmt.Errorf("generator: Expected pointer to Hsl slice, received nil.")
	}
	if len(*accentPalette) < 4 {
		return fmt.Errorf("generator: Accent palette must contain at least four elements.")
	}

	var samples []clusters.Observation
	for _, hsl := range *accentPalette {
		samples = append(samples, NormalizeHslToCoordinate(hsl))
	}

	km, err := kmeans.NewWithOptions(config.KmeansThreshold(), nil)
	if err != nil {
		return err
	}

	partions, err := km.Partition(samples, 4)
	if err != nil {
		return err
	}
	var hslPalette []Hsl
	for _, partition := range partions {
		c := partition.Center.Coordinates()
		hslPalette = append(hslPalette, Hsl{
			c[0] * 360,
			c[1],
			c[2],
		})
	}
	println("Create accent pallet")
	// Create 4 accent gradients
	gradietnStarts := hslPalette
	hslPalette = nil
	for _, h := range gradietnStarts {
		luminance := h.L
		if luminance < .50 {
			luminance = .70
		} else {
			luminance = .30
		}
		saturation := h.S
		if saturation > .40 {
			saturation *= .50
		}
		// Wrap hue if needed.
		newHue := h.H + config.HueShiftTolerance()
		if newHue > 360 {
			newHue -= 360
		}
		end := Hsl{newHue, saturation, luminance}

		grad, err := CreateGradient(h, end, 4)
		if err != nil {
			return err
		}
		sort.Slice(grad, func(i, j int) bool { return grad[i].L < grad[j].L })
		hslPalette = append(hslPalette, grad...)
	}
	*accentPalette = hslPalette
	return nil
}

func generateGreyScalePalette(ricePalette *RicePalette) error {
	return nil
}

func generateMonochromaticPalette(ricePalette *RicePalette) error {
	return nil
}
