package arganator

import (
	"flag"
	"fmt"
)

type Request struct {
	ImagePath       string
	LogProgress     bool
	WriteDebugImage bool
	DarkMode        bool
	Silent          bool
}

func ParseUserArgs() (*Request, error) {
	req := &Request{}
	flag.BoolVar(&req.LogProgress, "v", false, "Enable verbose mode")
	flag.StringVar(&req.ImagePath, "i", "", "Path to input image")
	flag.BoolVar(&req.WriteDebugImage, "o", false, "Write debug image.")
	flag.BoolVar(&req.DarkMode, "d", false, "Generate a dark mode color palette.")
	flag.BoolVar(&req.Silent, "s", false, "Disable terminal output. Errors will still be displayed.")

	flag.Parse()

	if req.ImagePath != "" {
		return req, fmt.Errorf("must provide a path to an image.")
	}
	return nil, fmt.Errorf("invalid image path.")
}
