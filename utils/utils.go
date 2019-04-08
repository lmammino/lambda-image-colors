package utils

import (
	chromath "github.com/jkl1337/go-chromath"
	"github.com/jkl1337/go-chromath/deltae"
	"math"
)

type Palette map[string][]uint32

func Rgb2lab(r, g, b uint32) chromath.Lab {
	src := chromath.RGB{float64(r), float64(g), float64(b)}

	targetIlluminant := &chromath.IlluminantRefD50
	rgb2xyz := chromath.NewRGBTransformer(&chromath.SpaceSRGB, &chromath.AdaptationBradford, targetIlluminant, &chromath.Scaler8bClamping, 1.0, nil)
	lab2xyz := chromath.NewLabTransformer(targetIlluminant)

	colorXyz := rgb2xyz.Convert(src)
	colorLab := lab2xyz.Invert(colorXyz)

	return colorLab
}

func GetClosestColor(r, g, b uint32, p Palette) string {
	minDiff := math.MaxFloat64
	minColor := ""
	colorLab := Rgb2lab(r, g, b)
	for colorName, color := range p {
		currLab := Rgb2lab(color[0], color[1], color[2])
		currDiff := deltae.CIE2000(colorLab, currLab, &deltae.KLChDefault)

		if currDiff < minDiff {
			minDiff = currDiff
			minColor = colorName
		}
	}

	return minColor
}
