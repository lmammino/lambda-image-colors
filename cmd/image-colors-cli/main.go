package main

import (
	"fmt"
	"os"

	"github.com/lmammino/lambda-image-colors/cmd/utils"
)

func main() {
	palette := utils.GetDefaultPalette()

	for _, filename := range os.Args[1:] {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while opening %s: %s\n", filename, err.Error())
			os.Exit(1)
		}

		defer file.Close()

		colors, err := utils.GetProminentColors(file, *palette)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while extrapolating prominent colors from %s: %s\n", filename, err.Error())
			os.Exit(1)
		}

		fmt.Printf("%s: %v\n", filename, colors)
	}
}
