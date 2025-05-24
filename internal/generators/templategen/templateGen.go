package templategen

import (
	"embed"
	"os"
	"path/filepath"
	"text/template"

	"github.com/inskribe/rice-paper.git/internal/arganator"
	"github.com/inskribe/rice-paper.git/internal/generators/ricepalette"
)

//go:embed templates/*
var templates embed.FS

func WritePalettes(colorPalette *ricepalette.ColorPalette, request *arganator.Request) error {
	args := NewTemplateArgs(colorPalette, request)
	cacheDir, err := getCacheDir()
	if err != nil {
		return err
	}

	templatesOutputDir := filepath.Join(cacheDir, "rice-paper")

	err = os.MkdirAll(templatesOutputDir, 0775)
	if err != nil {
		return err
	}

	riceTemplates, err := template.ParseFS(templates, "templates/*")
	if err != nil {
		return err
	}

	for _, template := range riceTemplates.Templates() {
		file, err := os.Create(filepath.Join(templatesOutputDir, template.Name()))
		if err != nil {
			return err
		}
		args.UpdateStatementArgs(template.Name(), request)
		err = riceTemplates.ExecuteTemplate(file, template.Name(), args)
		if err != nil {
			return err
		}
	}

	return nil
}

func getCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func NewTemplateArgs(colorPalette *ricepalette.ColorPalette, request *arganator.Request) TemplateArgs {
	return TemplateArgs{
		// Dark Colors
		Dark0: colorPalette.DarkValues[0].Hex(),
		Dark1: colorPalette.DarkValues[1].Hex(),
		Dark2: colorPalette.DarkValues[2].Hex(),
		Dark3: colorPalette.DarkValues[3].Hex(),

		// Light colors
		Light0: colorPalette.LightValues[0].Hex(),
		Light1: colorPalette.LightValues[1].Hex(),
		Light2: colorPalette.LightValues[2].Hex(),
		Light3: colorPalette.LightValues[3].Hex(),

		// Accent colors
		Accent0: colorPalette.AccentValues[0].Hex(),
		Accent1: colorPalette.AccentValues[1].Hex(),
		Accent2: colorPalette.AccentValues[2].Hex(),
		Accent3: colorPalette.AccentValues[3].Hex(),

		// Status colors
		StatusInfo:    colorPalette.StatusValues.Info.Hex(),
		StatusHint:    colorPalette.StatusValues.Hint.Hex(),
		StatusWarn:    colorPalette.StatusValues.Warn.Hex(),
		StatusError:   colorPalette.StatusValues.Error.Hex(),
		StatusSuccess: colorPalette.StatusValues.Success.Hex(),
	}
}

func (args *TemplateArgs) UpdateStatementArgs(templateName string, request *arganator.Request) {
	args.Delimiter = GetCommentDelimiter(templateName)
	args.ImagePath = request.ImagePath
	args.PaletteMode = "Light Mode"
	args.DarkMode = false
	args.DarkModeIntAsBool = 0
	if request.DarkMode {
		args.PaletteMode = "Dark Mode"
		args.DarkMode = true
		args.DarkModeIntAsBool = 1
	}
}
