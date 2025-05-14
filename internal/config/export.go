package config

func ImageCompression() int {
	return AppConfig.Generator.ImageCompressionSize
}

func KmeansThreshold() float64 {
	return AppConfig.Generator.KmeansThreshold
}

func KmeansPartionCount() int {
	return AppConfig.Generator.KmeansPartitionCount
}

func HistogramBinCount() int {
	return AppConfig.Generator.HistogramBinCount
}

func HueVariationTolerance() float64 {
	return AppConfig.Generator.HueVariationTolerance
}

func SaturationVariationTolerance() float64 {
	return AppConfig.Generator.SaturationVariationTolerance
}

func LuminanceVariationTolerance() float64 {
	return AppConfig.Generator.LuminanceVariationTolerance
}

func HueShiftTolerance() float64 {
	return AppConfig.Generator.HueShiftTolerance
}

func SaturationShiftTolerance() float64 {
	return AppConfig.Generator.SaturationShiftTolerance
}

func LuminanceMax() float64 {
	return AppConfig.Generator.LuminanceMax
}

func LuminanceMin() float64 {
	return AppConfig.Generator.LuminanceMin
}

func PaletteSwatchCount() int {
	return AppConfig.Generator.PaletteSwatchCount
}
