package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"sort"
)

var amountX float64
var amountY float64
var outFile string
var xy bool
var yx bool
var cpuprofile bool

func init() {
	flag.BoolVar(&xy, "xy", true, "sort the x-axis first, then the y-axis")
	flag.BoolVar(&yx, "yx", false, "sort the y-axis first, then the x-axis")
	flag.BoolVar(&cpuprofile, "cpuprofile", false, "pprof")
	flag.Float64Var(&amountX, "x", 0.1, "amount of pixel sort on the x-axis.")
	flag.Float64Var(&amountY, "y", 0.0, "amount of pixel sort on the y-axis.")
	flag.StringVar(&outFile, "o", "out.png", "filename of output.")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if cpuprofile {
		f, err := os.Create("skillnad.pprof")
		if err != nil {
			log.Fatalln(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	if flag.NArg() < 1 {
		flag.Usage()
	}
	err := play(flag.Arg(0))
	if err != nil {
		log.Fatalln(err)
	}
}

func play(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// _, t, err := image.Decode(f)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(t)

	im, err := png.Decode(f)
	if err != nil {
		return err
	}
	img := im.(*image.RGBA)
	glitch := image.NewRGBA(img.Bounds())

	if xy && !yx {
		pixelSortX(amountX, img, glitch)
		pixelSortY(amountY, glitch, glitch)
	} else if yx {
		pixelSortY(amountY, img, glitch)
		pixelSortX(amountX, glitch, glitch)
	}

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, glitch)
}

func pixelSortX(threshold float64, img, glitch *image.RGBA) {
	bound := img.Bounds()

	// Previous color used to measure if the
	var prev color.Color
	pixels := []color.Color{}
	for row := 0; row < bound.Max.Y; row++ {
		for col := 0; col < bound.Max.X; col++ {
			c := img.At(col, row)
			if col == 0 {
				prev = c
			}
			if Differ(threshold, c, prev) {
				// Sort pixels and add to new picture.
				glitch.Set(col, row, c)
				SortXDraw(glitch, &pixels, col, row)
				prev = c
				continue
			}
			// Otherwise add the pixel to pixels.
			pixels = append(pixels, c)
			prev = c
		}
		if len(pixels) > 0 {
			// Sort pixels and add to new picture.
			SortXDraw(glitch, &pixels, bound.Max.X, row)
		}
	}
}

func pixelSortY(threshold float64, img, glitch *image.RGBA) {
	bound := img.Bounds()

	var prev color.Color
	pixels := []color.Color{}
	for col := 0; col < bound.Max.X; col++ {
		for row := 0; row < bound.Max.Y; row++ {
			c := img.At(col, row)
			if col == 0 {
				prev = c
			}
			if Differ(threshold, c, prev) {
				// Sort pixels and add to new picture.
				glitch.Set(col, row, c)
				SortYDraw(glitch, &pixels, col, row)
				prev = c
				continue
			}
			// Otherwise add the pixel to pixels.
			pixels = append(pixels, c)
			prev = c
		}
		if len(pixels) > 0 {
			// Sort pixels and add to new picture.
			SortYDraw(glitch, &pixels, col, bound.Max.Y)
		}
	}
}

func SortXDraw(img *image.RGBA, pixels *[]color.Color, x, y int) {
	sort.Sort(ByLevel(*pixels))
	start := x - len(*pixels)
	for index := 0; start < x; start++ {
		img.Set(start, y, (*pixels)[index])
		index++
	}
	*pixels = []color.Color{}
}

func SortYDraw(img *image.RGBA, pixels *[]color.Color, x, y int) {
	sort.Sort(ByLevel(*pixels))
	start := y - len(*pixels)
	for index := 0; start < y; start++ {
		img.Set(x, start, (*pixels)[index])
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
	switch {
	case float64(r1-r2)/65535.0 >= threshold:
		return true
	case float64(g1-g2)/65535.0 >= threshold:
		return true
	case float64(b1-b2)/65535.0 >= threshold:
		return true
	}
	return false
}
