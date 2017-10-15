package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/karlek/skillnad"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
	}

	for _, filename := range flag.Args() {
		err := markov(filename)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func markov(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	img, err := skillnad.NewGlitcher(f)
	if err != nil {
		return err
	}

	var g skillnad.Glitch
	g.I = image.NewRGBA(img.Bounds())

	rect := img.Bounds()
	width := rect.Max.X
	height := rect.Max.Y

	yindexes := rand.Perm(height)
	for y, yindex := range yindexes {
		xindexes := rand.Perm(width)
		for x, xindex := range xindexes {
			g.I.Set(x, y, img.At(xindex, yindex))
			// proccessPixel(x, y, xindex, yindex, img, g.I)
		}
	}

	// Start at width/2 height/2

	// Filename of the glitched image.
	outFile := skillnad.RemoveExt(filepath.Base(filename)) + "-markov.png"

	// Create our glitched image.
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Println("[!]    Done:", outFile)
	return png.Encode(out, g.I)
}

// var pixels = []Pixel{}

// type Pixel struct {
// 	p image.Pt
// 	c color.Color
// }

// func proccessPixel(x, y, xindex, yindex, int, img, g.I, skillnad.Glitcher) {
// 	neighbourIndexes := []image.Pt{
// 		{X: 1, Y: 0},
// 		{X: -1, Y: 1},
// 		{X: 0, Y: 1},
// 		{X: 0, Y: -1},
// 	}
// 	for _, pt := range neighbourIndexes {
// 		if x+pt.X < 0 ||
// 			x+pt.X >= width ||
// 			y+pt.Y < 0 ||
// 			y+pt.Y >= height {
// 			continue
// 		}
// 		pixels = append(Pixel{Pt: {X: x + pt.X, Y: y + pt.Y}, img.At(xindex, yindex)})
// 	}
// }
