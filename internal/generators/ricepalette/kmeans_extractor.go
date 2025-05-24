package ricepalette

import (
	"fmt"
	"image"

	"github.com/inskribe/rice-paper.git/internal/config"
	"github.com/inskribe/rice-paper.git/internal/hslx"
	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
)

type KmeansExtractor struct{}

func (extractor KmeansExtractor) Extract(img *image.NRGBA) (*PaletteBase, error) {
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
	base := &PaletteBase{}
	base.DarkHue = int(avgHue)
	base.LightHue = int(avgHue)
	return base, nil
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
	}

	hueTotal := float64(0)
	for _, hsl := range (*collection)[largestPartitionIndex] {
		hueTotal += hsl.H
	}
	return hueTotal / float64(len((*collection)[largestPartitionIndex])), nil
}
