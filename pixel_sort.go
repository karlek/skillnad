package skillnad

import (
	"image/color"
	"sort"
	"sync"
)

func (g Glitch) PixelSortX(threshold float64, src Glitcher) {
	var wg sync.WaitGroup
	wg.Add(src.Bounds().Max.Y)
	for row := 0; row < src.Bounds().Max.Y; row++ {
		go g.PerLine(sortXDraw, xat(src), xset(g), src, row, src.Bounds().Max.X, threshold, &wg)
	}
	wg.Wait()
}

func (g Glitch) PixelSortY(threshold float64, src Glitcher) {
	var wg sync.WaitGroup
	wg.Add(src.Bounds().Max.X)
	for col := 0; col < src.Bounds().Max.X; col++ {
		go g.PerLine(sortYDraw, yat(src), yset(g), src, col, src.Bounds().Max.Y, threshold, &wg)
	}
	wg.Wait()
}

func xat(src Glitcher) func(int, int) color.Color {
	return func(x, y int) color.Color { return src.At(x, y) }
}
func yat(src Glitcher) func(int, int) color.Color {
	return func(x, y int) color.Color { return src.At(y, x) }
}
func xset(g Glitch) func(int, int, color.Color) {
	return func(x, y int, c color.Color) { g.I.Set(x, y, c) }
}
func yset(g Glitch) func(int, int, color.Color) {
	return func(x, y int, c color.Color) { g.I.Set(y, x, c) }
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
			sort(g.I, &pixels, index, fixed)
			prev = c
			continue
		}
		// Otherwise add the pixel to pixels.
		pixels = append(pixels, c)
		prev = c
	}
	if len(pixels) > 0 {
		// Sort pixels and add to new picture.
		sort(g.I, &pixels, max, fixed)
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
	if r1*g1*b1 == 0 || r2*g2*b2 == 0 {
		return true
	}
	switch {
	case int(r1-r2) > int(threshold*65535.0):
		return true
	case int(g1-g2) > int(threshold*65535.0):
		return true
	case int(b1-b2) > int(threshold*65535.0):
		return true
	}
	return false
}
