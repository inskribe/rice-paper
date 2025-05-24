package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"

	"github.com/disintegration/imaging"
	"github.com/inskribe/rice-paper.git/internal/arganator"
	"github.com/inskribe/rice-paper.git/internal/config"
	"github.com/inskribe/rice-paper.git/internal/generators/ricepalette"
	"github.com/inskribe/rice-paper.git/internal/generators/templategen"
)

func main() {
	err := config.LoadApplicationConfig()
	if err != nil {
		log.Fatal(err)
	}

	PrintRicePaper()
	req, err := arganator.ParseUserArgs()

	file, err := os.Open(req.ImagePath)
	if err != nil {
		fmt.Printf("Failed to read file at %s", req.ImagePath)
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	file.Close()
	palReq := ricepalette.PaletteRequest{
		Image:  img,
		Silent: req.Silent,
	}
	// colorPalette, err := palReq.CreatePalette(ricepalette.KmeansExtractor{})
	colorPalette, err := palReq.CreatePalette(ricepalette.DeafultExtractor{GroupingFactor: 360})
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

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

	templategen.WritePalettes(colorPalette, req)
}

func PrintRicePaper() {
	ascii := "       __                                                                  \n" +
		" _ __ /\\_\\    ___     __           _____      __     _____      __   _ __  \n" +
		"/\\`'__\\/\\ \\  /'___\\ /'__`\\ _______/\\ '__`\\  /'__`\\  /\\ '__`\\  /'__`\\/\\`'__\\\n" +
		"\\ \\ \\/ \\ \\ \\/\\ \\__//\\  __//\\______\\ \\ \\L\\ \\/\\ \\L\\.\\_\\ \\ \\L\\ \\/\\  __/\\ \\ \\/ \n" +
		" \\ \\_\\  \\ \\_\\ \\____\\ \\____\\/______/\\ \\ ,__/\\ \\__/.\\_\\\\ \\ ,__/\\ \\____\\\\ \\_\\ \n" +
		"  \\/_/   \\/_/\\/____/\\/____/         \\ \\ \\/  \\/__/\\/_/ \\ \\ \\/  \\/____/ \\/_/ \n" +
		"                                     \\ \\_\\             \\ \\_\\               \n" +
		"                                      \\/_/              \\/_/               \n"

	fmt.Print(ascii)
}
