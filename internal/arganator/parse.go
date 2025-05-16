package arganator

import (
	"flag"
	"fmt"
)

type Request struct {
	ImagePath       string
	LogProgress     bool
	WriteDebugImage bool
	NeutralPalette  bool
}

func ParseUserArgs() (*Request, error) {
	req := &Request{}
	flag.BoolVar(&req.LogProgress, "v", false, "Enable verbose mode")
	flag.StringVar(&req.ImagePath, "i", "", "Path to input image")
	flag.BoolVar(&req.WriteDebugImage, "o", false, "Write debug image.")
	flag.BoolVar(&req.NeutralPalette, "n", false, "Remove neutral color alignment. If enabled, high saturaion images could produce vibrating colors.")

	flag.Parse()

	fmt.Println("Verbose?", req.LogProgress)
	fmt.Println("Image path:", req.ImagePath)

	// Ensure we can do something with the request.
	if req.ImagePath != "" {
		return req, nil
	}
	return nil, fmt.Errorf("invalid image path.")
}
