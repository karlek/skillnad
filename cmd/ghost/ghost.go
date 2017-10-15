package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"sort"

	"github.com/karlek/skillnad"
)

var amount float64

func init() {
	flag.Float64Var(&amount, "a", 0, "amount edge detection.")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()

	for _, filename := range flag.Args() {
		err := ghost(filename)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func ghost(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Make sure the image type is glitchable (i.e. we can change the pixels).
	img, err := skillnad.NewGlitcher(f)
	if err != nil {
		return err
	}

	var g skillnad.Glitch
	g.I = image.NewRGBA(img.Bounds())

	var prev color.Color
	pixels := []color.Color{}
	for row := 0; row < img.Bounds().Max.Y; row++ {
		for col := 0; col < img.Bounds().Max.X; col++ {
			c := img.At(col, row)
			if row == 0 && col == 0 {
				prev = c
			}
			if Differ(amount, c, prev) {
				// Sort pixels and add to new picture.
				SortDraw(g.I, &pixels, col, row)
				g.I.Set(col, row, c)
				prev = c
				continue
			}
			// Otherwise add the pixel to pixels.
			pixels = append(pixels, c)
			prev = c
		}
		if len(pixels) > 0 {
			// Sort pixels and add to new picture.
			SortDraw(g.I, &pixels, img.Bounds().Max.X, row)
		}
	}
	out, err := os.Create(skillnad.RemoveExt(filename) + "-ghost.png")
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, g.I)
}

func SortDraw(img *image.RGBA, pixels *[]color.Color, x, y int) {
	sort.Sort(ByLevel(*pixels))
	start := x - len(*pixels)

	for index := 0; start < len(*pixels); start++ {
		img.Set(start, y, (*pixels)[index])
		index++
	}
	// Sort pixels and add to new picture.
	*pixels = []color.Color{}
}

type ByLevel []color.Color

func (b ByLevel) Len() int           { return len(b) }
func (b ByLevel) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByLevel) Less(i, j int) bool { return Level(b[i]) > Level(b[j]) }

func Level(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	return float64(r+g+b) / 65535.0
}

func Differ(threshold float64, c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	switch {
	case math.Abs(float64(r1-r2)/65535.0) >= threshold:
		return true
	case math.Abs(float64(g1-g2)/65535.0) >= threshold:
		return true
	case math.Abs(float64(b1-b2)/65535.0) >= threshold:
		return true
	}
	return false
}
