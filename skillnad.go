package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"sync"
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
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [FILE],,,\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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

	errChan := make(chan error)
	var wg sync.WaitGroup
	go func(w *sync.WaitGroup, ec chan error) {
		w.Wait()
		close(ec)
	}(&wg, errChan)

	for _, filename := range flag.Args() {
		go skillnadErrWrap(filename, &wg, errChan)
		wg.Add(1)
	}
	for err := range errChan {
		if err != nil {
			log.Println(err)
		}
		wg.Done()
	}
}

func skillnadErrWrap(filename string, wg *sync.WaitGroup, errChan chan error) {
	errChan <- skillnad(filename)
}

type Glitcher interface {
	At(x, y int) color.Color
	Bounds() image.Rectangle
	ColorModel() color.Model
	Set(x, y int, c color.Color)
}

type Glitch struct {
	Glitcher
	i *image.RGBA
}

func skillnad(filename string) (err error) {
	fmt.Println("[!]    Glitching:", filename)

	// Open file.
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	// Don't forget to close.
	defer f.Close()

	// Decode image and get the type.
	im, t, err := image.Decode(f)
	if err != nil {
		return err
	}
	// If the file isn't png convert it to png.
	if t != "png" {
		var buf []byte
		b := bytes.NewBuffer(buf)
		err = png.Encode(b, im)
		if err != nil {
			return err
		}
		im, err = png.Decode(b)
		if err != nil {
			return err
		}
	}
	// Make sure the image type is glitchable (i.e. we can change the pixels).
	img, err := NewGlitcher(im)
	if err != nil {
		return err
	}

	var g Glitch
	g.i = image.NewRGBA(img.Bounds())

	// Which axis should we pixel sort first? X then y or reversed?
	if xy && !yx {
		g.PixelSortX(amountX, img)
		g.PixelSortY(amountY, g.i)
	} else if yx {
		g.PixelSortY(amountY, img)
		g.PixelSortX(amountX, g.i)
	}

	// Filename of the glitched image.
	outFile := removeExt(filepath.Base(filename)) + "-glitch.png"

	// Create our glitched image.
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	fmt.Println("[!]    Done:", outFile)
	return png.Encode(out, g.i)
}

func removeExt(path string) string {
	index := strings.LastIndex(path, filepath.Ext(path))
	if index == -1 {
		return path
	}
	return path[:index]
}

// NewGlitcher takes an image and returns an interface which is glitchable.
func NewGlitcher(im image.Image) (img Glitcher, err error) {
	img, ok := im.(*image.RGBA)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.NRGBA)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.RGBA64)
	if ok {
		return img, nil
	}
	return nil, fmt.Errorf("Unknown or unsupported image type: %T", im)
}

func (g Glitch) PixelSortX(threshold float64, src Glitcher) {
	var wg sync.WaitGroup
	wg.Add(src.Bounds().Max.Y)
	for row := 0; row < src.Bounds().Max.Y; row++ {
		go g.PerLine(sortXDraw, xat(src), xset(src), src, row, src.Bounds().Max.X, threshold, &wg)
	}
	wg.Wait()
}

func (g Glitch) PixelSortY(threshold float64, src Glitcher) {
	var wg sync.WaitGroup
	wg.Add(src.Bounds().Max.X)
	for col := 0; col < src.Bounds().Max.X; col++ {
		go g.PerLine(sortYDraw, yat(src), yset(src), src, col, src.Bounds().Max.Y, threshold, &wg)
	}
	wg.Wait()
}

func xat(src Glitcher) func(int, int) color.Color {
	return func(x, y int) color.Color { return src.At(x, y) }
}
func yat(src Glitcher) func(int, int) color.Color {
	return func(x, y int) color.Color { return src.At(y, x) }
}
func xset(src Glitcher) func(int, int, color.Color) {
	return func(x, y int, c color.Color) { src.Set(x, y, c) }
}
func yset(src Glitcher) func(int, int, color.Color) {
	return func(x, y int, c color.Color) { src.Set(y, x, c) }
}

func (g Glitch) PerLine(
	sort func(Glitcher, *[]color.Color, int, int),
	at func(int, int) color.Color,
	set func(int, int, color.Color),
	src Glitcher,
	fixed, max int,
	threshold float64,
	wg *sync.WaitGroup) {
	// Previous color used to measure if the
	var prev color.Color
	pixels := []color.Color{}

	for index := 0; index < max; index++ {
		c := at(index, fixed)
		if index == 0 {
			prev = c
		}
		if Differ(threshold, c, prev) {
			// Sort pixels and add to new picture.
			set(index, fixed, c)
			sort(g.i, &pixels, index, fixed)
			prev = c
			continue
		}
		// Otherwise add the pixel to pixels.
		pixels = append(pixels, c)
		prev = c
	}
	if len(pixels) > 0 {
		// Sort pixels and add to new picture.
		sort(g.i, &pixels, max, fixed)
	}
	wg.Done()
}

func sortXDraw(img Glitcher, pixels *[]color.Color, x, y int) {
	sort.Sort(ByLevel(*pixels))
	start := x - len(*pixels)
	for index := 0; start < x; start++ {
		img.Set(start, y, (*pixels)[index])
		index++
	}
	*pixels = []color.Color{}
}

func sortYDraw(img Glitcher, pixels *[]color.Color, y, x int) {
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

// Level returns the level of a color.
func Level(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return r + g + b
}

// Differ returns true if c1 and c2 differ more then the allowed threshold.
func Differ(threshold float64, c1, c2 color.Color) bool {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Prevents overflow. Disable these to get a glitchy effect. However pixel
	// sort will stop to work.
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
