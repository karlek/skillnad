package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/karlek/profile"
	"github.com/karlek/skillnad"
)

var amountX float64
var amountY float64
var outFile string
var xy bool
var yx bool
var cpuprof bool
var fname string
var n int

func init() {
	flag.BoolVar(&xy, "xy", true, "sort the x-axis first, then the y-axis")
	flag.BoolVar(&yx, "yx", false, "sort the y-axis first, then the x-axis")
	flag.BoolVar(&cpuprof, "profile", false, "cpu profile")
	flag.Float64Var(&amountX, "x", 0, "amount of pixel sort on the x-axis.")
	flag.Float64Var(&amountY, "y", 0, "amount of pixel sort on the y-axis.")
	flag.StringVar(&fname, "o", "", "output filename")
	flag.IntVar(&n, "n", 1, "number of sorts")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.Parse()
	if cpuprof {
		defer profile.Start().Stop()
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	if flag.NArg() < 1 {
		flag.Usage()
	}

	errChan := make(chan error)
	var wg sync.WaitGroup
	go func(w *sync.WaitGroup, ec chan error) {
		w.Wait()
		close(ec)
	}(&wg, errChan)

	wg.Add(flag.NArg())
	for _, filename := range flag.Args() {
		go func(filename string, errChan chan<- error) {
			errChan <- pixelSort(filename, n)
		}(filename, errChan)
	}
	for err := range errChan {
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}
}

func pixelSort(filename string, n int) (err error) {
	fmt.Println("[!]    Glitching:", filename)

	// Open file.
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	// Don't forget to close.
	defer f.Close()

	// Make sure the image type is glitchable (i.e. we can change the pixels).
	img, err := skillnad.NewGlitcher(f)
	if err != nil {
		return err
	}

	var g skillnad.Glitch
	g.I = image.NewRGBA(img.Bounds())
	for i := 0; i < n; i++ {
		xy, yx = yx, xy
		// Which axis should we pixel sort first? X then y or reversed?
		if xy && !yx {
			if amountX != 0 {
				g.PixelSortX(amountX, img)
				if amountY != 0 {
					g.PixelSortY(amountY, g.I)
				}
			} else if amountY != 0 {
				g.PixelSortY(amountY, img)
			}
		} else if yx {
			if amountY != 0 {
				g.PixelSortY(amountY, img)
				if amountX != 0 {
					g.PixelSortX(amountX, g.I)
				}
			} else if amountX != 0 {
				g.PixelSortX(amountX, img)
			}
		}
	}

	// Filename of the glitched image.
	outFile := skillnad.RemoveExt(filepath.Base(filename)) + "-glitch.png"
	if fname != "" {
		outFile = fname
	}
	// Create our glitched image.
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Println("[!]    Done:", outFile)
	return png.Encode(out, g.I)
}
