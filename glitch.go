package skillnad

import (
	"bytes"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"io"

	"github.com/mewkiz/pkg/errutil"
)

type Glitcher interface {
	At(x, y int) color.Color
	Bounds() image.Rectangle
	ColorModel() color.Model
	Set(x, y int, c color.Color)
}

type Glitch struct {
	Glitcher
	I *image.RGBA
}

// NewGlitcher takes an image and returns an interface which is glitchable.
func NewGlitcher(r io.Reader) (img Glitcher, err error) {
	// Decode image and get the type.
	im, t, err := image.Decode(r)
	if err != nil {
		return nil, err
	}

	// If the file isn't png convert it to png.
	if t != "png" {
		var buf []byte
		b := bytes.NewBuffer(buf)
		err = png.Encode(b, im)
		if err != nil {
			return nil, err
		}
		im, err = png.Decode(b)
		if err != nil {
			return nil, err
		}
	}

	img, ok := im.(*image.RGBA)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.NRGBA)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.NRGBA64)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.RGBA64)
	if ok {
		return img, nil
	}
	img, ok = im.(*image.Gray)
	if ok {
		return img, nil
	}
	return nil, errutil.NewNoPosf("Unknown or unsupported image type: %T", im)
}
