package generator

//go:generate stringer -type=PaletteType

type PaletteType int

const (
	Unknown PaletteType = iota
	Monochromatic
	Greyscale
	Dynamic
)
