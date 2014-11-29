package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	// "image/draw"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"sort"
)

func main() {
	flag.Parse()

	err := play(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
}

func play(filename string) (err error) {
	fmt.Println(filename)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	im, err := png.Decode(f)
	if err != nil {
		return err
	}
	img := im.(*image.NRGBA)
	out := image.NewNRGBA(img.Bounds())
	bound := img.Bounds()
	var prev color.Color
	pixels := []color.Color{}
	num := 0
	for row := 0; row < bound.Max.Y; row++ {
		for col := 0; col < bound.Max.X; col++ {
			c := img.At(col, row)
			if col == 0 {
				prev = c
			}
			if Differ(0.1, c, prev) {
				num++
				// Sort pixels and add to new picture.
				out.Set(col, row, c)
				SortDraw(out, &pixels, col, row)
				prev = c
				continue
			}
			// Otherwise add the pixel to pixels.
			pixels = append(pixels, c)
			prev = c
		}
		if len(pixels) > 0 {
			// Sort pixels and add to new picture.
			SortDraw(out, &pixels, bound.Max.X, row)
		}
	}
	fmt.Println(num)
	fmt.Println(bound.Max.X * bound.Max.Y)
	f2, err := os.Create("b.png")
	if err != nil {
		return err
	}
	defer f2.Close()

	return png.Encode(f2, out)
}

func SortDraw(img *image.NRGBA, pixels *[]color.Color, x, y int) {
	sort.Sort(ByLevel(*pixels))
	start := x - len(*pixels)
	for index := 0; start < x; start++ {
		img.Set(start, y, (*pixels)[index])
		index++
	}
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
	if r1 < r2 {
		r1, r2 = r2, r1
	}
	if g1 < g2 {
		g1, g2 = g2, g1
	}
	if b1 < b2 {
		b1, b2 = b2, b1
	}
	// fmt.Println(float64(r1))
	// fmt.Println(float64(r2))
	// fmt.Println(float64(r1 - r2))
	// fmt.Println(float64(r1-r2) / 65535.0)
	// fmt.Println(float64(r1-r2)/65535.0 >= threshold)
	// fmt.Println()
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
