package templategen

type TemplateArgs struct {
	Dark0 string
	Dark1 string
	Dark2 string
	Dark3 string

	Light0 string
	Light1 string
	Light2 string
	Light3 string

	Accent0 string
	Accent1 string
	Accent2 string
	Accent3 string

	StatusInfo    string
	StatusHint    string
	StatusWarn    string
	StatusError   string
	StatusSuccess string

	Delimiter   string
	ImagePath   string
	PaletteMode string

	DarkMode          bool
	DarkModeIntAsBool int
}
