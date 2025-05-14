package config

type Config struct {
	Application ApplicationConfig `yaml:"application"`
	Generator   GeneratorSettings `yaml:"generator-settings"`
}
type ApplicationConfig struct {
	ApplicationName string `yaml:"application_name"`
}
type GeneratorSettings struct {
	ImageCompressionSize         int     `yaml:"img-compression-size"`
	KmeansThreshold              float64 `yaml:"kmeans-threshold"`
	KmeansPartitionCount         int     `yaml:"kmeans-partition-count"`
	HistogramBinCount            int     `yaml:"histogram-bin-count"`
	HueVariationTolerance        float64 `ymal:"hue-variation-tolerance"`
	SaturationVariationTolerance float64 `yaml:"saturation-variation-tolerance"`
	LuminanceVariationTolerance  float64 `yaml:"luminance-variation-tolerance"`
	HueShiftTolerance            float64 `yaml:"hue-shift-tolerance"`
	SaturationShiftTolerance     float64 `yaml:"saturation-dhift-tolerance"`
	LuminanceMax                 float64 `yaml:"luminance-max"`
	LuminanceMin                 float64 `yaml:"luminance-min"`
	PaletteSwatchCount           int     `yaml:"palette-swatch-count"`
}
