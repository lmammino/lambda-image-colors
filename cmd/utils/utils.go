package utils

import (
	"github.com/EdlinOrg/prominentcolor"
	chromath "github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"math"

	"image"
	_ "image/jpeg" // enables decoding of jpegs
	"io"
)

// Palette defines a palette of colors.
// It's essentially a map where keys are color names and values is a sequence of
// 3 RGB uint32 values
type Palette map[string][]uint32

// GetDefaultPalette returns a default palette
func GetDefaultPalette() *Palette {
	return &Palette{
		"red":       {255, 0, 0},
		"orange":    {255, 165, 0},
		"yellow":    {255, 255, 0},
		"green":     {0, 255, 0},
		"turquoise": {0, 222, 222},
		"blue":      {0, 0, 255},
		"violet":    {128, 0, 255},
		"pink":      {255, 0, 255},
		"brown":     {160, 82, 45},
		"black":     {0, 0, 0},
		"white":     {255, 255, 255},
	}
}

func rgb2lab(r, g, b uint32) chromath.Lab {
	src := chromath.RGB{float64(r), float64(g), float64(b)}

	targetIlluminant := &chromath.IlluminantRefD50
	rgb2xyz := chromath.NewRGBTransformer(&chromath.SpaceSRGB, &chromath.AdaptationBradford, targetIlluminant, &chromath.Scaler8bClamping, 1.0, nil)
	lab2xyz := chromath.NewLabTransformer(targetIlluminant)

	colorXyz := rgb2xyz.Convert(src)
	colorLab := lab2xyz.Invert(colorXyz)

	return colorLab
}

func getClosestColorName(color prominentcolor.ColorRGB, p Palette) string {
	minDiff := math.MaxFloat64
	minColor := ""
	colorLab := rgb2lab(color.R, color.G, color.B)
	for colorName, color := range p {
		currLab := rgb2lab(color[0], color[1], color[2])
		currDiff := deltae.CIE2000(colorLab, currLab, &deltae.KLChDefault)

		if currDiff < minDiff {
			minDiff = currDiff
			minColor = colorName
		}
	}

	return minColor
}

func appendIfMissing(slice []string, value string) []string {
	for _, ele := range slice {
		if ele == value {
			return slice
		}
	}
	return append(slice, value)
}

// GetProminentColors takes a Reader and a palette. It reads the image content
// using the reader and identifies the prominent colors in it which are returned
// as a slice of string values.
func GetProminentColors(imageContent io.Reader, palette Palette) ([]string, error) {
	img, _, err := image.Decode(imageContent)
	if err != nil {
		return nil, err
	}

	res, err := prominentcolor.Kmeans(img)
	if err != nil {
		return nil, err
	}

	colors := []string{}

	for _, match := range res {
		colors = appendIfMissing(colors, getClosestColorName(match.Color, palette))
	}

	return colors, nil
}
