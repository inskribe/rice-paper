package ricepalette

import (
	"fmt"
	"image"
	"math"

	"github.com/disintegration/imaging"
	"github.com/inskribe/rice-paper.git/internal/config"
	"github.com/inskribe/rice-paper.git/internal/hslx"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

func (request PaletteRequest) CreatePalette() (*ColorPalette, error) {
	img := ResizeImage(request.Image)

	pixelObservations, err := extractPixelObservations(img)
	if err != nil {
		return nil, err
	}

	clusters, err := extractClusters(pixelObservations)
	if err != nil {
		return nil, err
	}

	collection, err := extractHslCollection(clusters)
	if err != nil {
		return nil, err
	}

	hslPartitions, err := collection.Partition()
	if err != nil {
		return nil, err
	}
	avgHue, err := findAverageHue(hslPartitions)
	if err != nil {
		return nil, err
	}

	colorPalette := newColorPalette(avgHue)
	println("Dark Values")
	colorPalette.DarkValues.Print()
	println("Light Values")
	colorPalette.LightValues.Print()

	println("Accent Values")
	colorPalette.AccentValues.Print()

	println("Status Values")
	colorPalette.StatusValues.Print()
	return colorPalette, nil
}

func ResizeImage(img image.Image) *image.NRGBA {
	desiredSize := config.ImageCompression()
	return imaging.Resize(img, desiredSize, desiredSize, imaging.NearestNeighbor)
}

func extractPixelObservations(img *image.NRGBA) (*clusters.Observations, error) {
	if img == nil {
		return nil, fmt.Errorf("func extractPixelObservations: expected pointer to image.NRGBA received nil.")
	}

	var result clusters.Observations
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			color := img.NRGBAAt(x, y)
			hsl := hslx.RgbToHsl(color.R, color.G, color.B)
			result = append(result, hslx.NormalizeHslToCoordinate(hsl))
		}
	}
	return &result, nil
}

func extractClusters(samples *clusters.Observations) (*clusters.Clusters, error) {
	if samples == nil {
		return nil, fmt.Errorf("func extractClusters: expected pointer to clusters.Observations, received nil")
	}
	if len(*samples) < 2 {
		return nil, fmt.Errorf("func extractClusters: parameter samples must contain more that one clusters.Observation.")
	}

	threshold := config.KmeansThreshold()
	km, err := kmeans.NewWithOptions(threshold, nil)
	if err != nil {
		return nil, fmt.Errorf("func extractClusters: %w", err)
	}
	partitions, err := km.Partition(*samples, config.KmeansPartionCount())
	if err != nil {
		return nil, fmt.Errorf("func extractClusters: %w", err)
	}
	return &partitions, nil
}

func extractHslCollection(pixelClusters *clusters.Clusters) (*hslx.HslCollection, error) {
	if pixelClusters == nil {
		return nil, fmt.Errorf("func extractHslCollection: expected pointer to clusters.Clusters, received nil.")
	}
	if len(*pixelClusters) < 2 {
		return nil, fmt.Errorf("func extractHslCollection: parameter pixelClusters must contain more than one element.")
	}

	var collection hslx.HslCollection
	for _, cluster := range *pixelClusters {
		coordinate := cluster.Center.Coordinates()
		hsl := hslx.Hsl{
			H: coordinate[0] * 360,
			S: coordinate[1],
			L: coordinate[2],
		}
		collection = append(collection, hsl)
	}
	return &collection, nil
}

func findAverageHue(collection *hslx.HslPartitions) (float64, error) {
	var largestPartitionIndex int = 0
	for i, partition := range *collection {
		if len(partition) > len((*collection)[largestPartitionIndex]) {
			largestPartitionIndex = i
		}
		fmt.Printf("Printing partition: %d\n", i)
		partition.Print()
	}

	hueTotal := float64(0)
	for _, hsl := range (*collection)[largestPartitionIndex] {
		hueTotal += hsl.H
	}
	return hueTotal / float64(len((*collection)[largestPartitionIndex])), nil
}

func newColorPalette(averageHue float64) *ColorPalette {
	fmt.Printf("average hue: %.2f\n", averageHue)

	return &ColorPalette{
		DarkValues:   createDarkValues(averageHue),
		LightValues:  createLightValues(averageHue),
		AccentValues: createAccentValues(averageHue),
		// TODO::Palette
		// Have not decided how to handle StatusValues yet.
		// For now just hard coded values.
		StatusValues: hslx.HslCollection{
			{H: 355.2, S: 0.45, L: 0.56}, // error
			{H: 17.4, S: 0.50, L: 0.64},  // warning
			{H: 40.0, S: 0.74, L: 0.74},  // info
			{H: 95.0, S: 0.26, L: 0.65},  // success
			{H: 320.0, S: 0.20, L: 0.66}, // hint
		},
	}
}

func createDarkValues(baseHue float64) hslx.HslCollection {
	var result hslx.HslCollection

	for i := range 4 {
		result = append(result, hslx.Hsl{
			H: baseHue + 1.0*float64(i),
			S: 0.17 - 0.01*float64(i),
			L: 0.21 + 0.05*float64(i),
		})
	}

	return result
}

func createLightValues(baseHue float64) hslx.HslCollection {
	var result hslx.HslCollection

	for i := range 4 {
		result = append(result, hslx.Hsl{
			H: baseHue,
			S: 0.20 - .01*float64(i),
			L: 0.65 + 0.085*float64(i),
		})
	}
	return result
}

func createAccentValues(baseH float64) hslx.HslCollection {
	return hslx.HslCollection{
		{H: Hue(baseH - 40), S: 0.45, L: 0.65},
		{H: Hue(baseH - 25), S: 0.40, L: 0.68},
		{H: Hue(baseH - 10), S: 0.35, L: 0.63},
		{H: Hue(baseH), S: 0.30, L: 0.53},
	}
}

func Hue(h float64) float64 {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}
