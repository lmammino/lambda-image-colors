package main

import (
	"fmt"
	"os"

	"github.com/EdlinOrg/prominentcolor"
	"image"
	_ "image/jpeg"

	"loige.co/image-colors/utils"
)

func main() {
	palette := utils.Palette{
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
		"gray":      {128, 128, 128},
		"white":     {255, 255, 255},
	}

	fmt.Println(palette)

	fmt.Println(os.Args[1])

	f, _ := os.Open(os.Args[1])
	defer f.Close()

	img, _, _ := image.Decode(f)
	res, _ := prominentcolor.Kmeans(img)

	fmt.Println(res)
	for _, color := range res {
		fmt.Println(utils.GetClosestColor(color.Color.R, color.Color.G, color.Color.B, palette))
	}
}
