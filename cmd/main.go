package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"rice-paper/internal/arganator"
	"rice-paper/internal/config"
	"rice-paper/internal/generators/palettegen"
	"rice-paper/internal/generators/templategen"

	"github.com/disintegration/imaging"
)

type Request struct {
	ImagePath   string
	LogProgress bool
}

func main() {
	err := config.LoadApplicationConfig()
	if err != nil {
		log.Fatal(err)
	}

	req, err := arganator.ParseUserArgs()

	file, err := os.Open(req.ImagePath)
	if err != nil {
		fmt.Printf("Failed to read file at %s", req.ImagePath)
		return
	}
	config, conf_fmt, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Printf("%s", err)
		return
	}
	fmt.Printf("format: %s, height: %d, width: %d\n", conf_fmt, config.Height, config.Width)
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal("seek failed")
	}

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	file.Close()

	ricePalette, err := palettegen.GenrateColorPalette(img)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	fmt.Println("Main Palette")
	palettegen.PrintHslColors(ricePalette.MainPalette)
	fmt.Println("Accent Palette")
	palettegen.PrintHslColors(ricePalette.AccentPalette)

	if req.WriteDebugImage {
		compressedImage := imaging.Resize(img, 100, 0, imaging.NearestNeighbor)
		file, err := os.Create("./debug-output/debug.png")
		if err != nil {
			fmt.Printf("failed to create debug image file. %v", err)
			file.Close()
		}
		err = png.Encode(file, compressedImage)
		if err != nil {
			fmt.Printf("failed to encode debug image. %v", err)
			err = file.Close()
			if err != nil {
				fmt.Printf("Failed to close debug file. %v", err)
			}
		}
	}

	templategen.WritePalettes(*ricePalette)
}
