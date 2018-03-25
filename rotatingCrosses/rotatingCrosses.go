package main

import (
	"image/color"
	"image"
	"os"
	"image/gif"
	"fmt"
	"time"
	"math"
)

const (
	defaultSteps = 1
)

var (
	defaultColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
)

var palette = []color.Color{
	color.RGBA{0x00, 0x00, 0x00, 0xff},
	color.RGBA{0x00, 0x00, 0xff, 0xff},
	color.RGBA{0x00, 0xff, 0x00, 0xff},
	color.RGBA{0x00, 0xff, 0xff, 0xff},
	color.RGBA{0xff, 0x00, 0x00, 0xff},
	color.RGBA{0xff, 0x00, 0xff, 0xff},
	color.RGBA{0xff, 0xff, 0x00, 0xff},
	color.RGBA{0xff, 0xff, 0xff, 0xff},
}

type ShapePattern interface {
	Draw(step int, x, y float64) (bool, color.Color)
}

type Shape interface {
	Brightness(x, y float64) (bool, color.Color)
}

type Coordinates struct {
	X, Y float64
}

type Cross struct {
	Center Coordinates
	Width, Height, Thickness float64
	Color color.Color
}

type CrossPattern struct {
	steps []Cross
}

type PatternComposite struct {
	Patterns []ShapePattern
}

func (c *Cross) Brightness(x, y float64) (bool, color.Color) {
	var drawn bool
	var retCol color.Color

	if x > c.Center.X - c.Width && x < c.Center.X + c.Width {
		if y > c.Center.Y - c.Thickness && y < c.Center.Y + c.Thickness{
			drawn = true
		}
	}
	if !drawn && y > c.Center.Y - c.Height && y < c.Center.Y + c.Height {
		if x > c.Center.X - c.Thickness && x < c.Center.X + c.Thickness{
			drawn = true
		}
	}

	if drawn {
		retCol = c.Color
	} else {
		retCol = defaultColor
	}
	return drawn, retCol
}

func (c *CrossPattern) Draw(step int, x, y float64) (bool, color.Color) {
	return c.steps[step].Brightness(x, y)
}

func (p *PatternComposite) Draw(step int, x, y float64) (bool, color.Color) {
	var drawn bool
	var col color.Color
	for _, v := range p.Patterns{
		d, tmpCol := v.Draw(step, x, y)
		if d {
			col = tmpCol
		}
	}
	return drawn, col
}

func SingleCross(xCenter, yCenter, width, height, thickness float64, col color.Color) CrossPattern {
	var steps = make([]Cross, defaultSteps)
	for i := range steps {
		steps[i] = Cross{
			Center: Coordinates{X: xCenter, Y: yCenter},
			Width: width,
			Height:height,
			Thickness: thickness,
			Color: col,
		}
	}
	return CrossPattern{steps:steps}
}


func DrawPalette(w, h, step int, patterns ShapePattern ) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			_, col := patterns.Draw(step, float64(x), float64(y))
			img.Set(x, y, col)
		}
	}
	return img
}

func main() {
	startTime := time.Now()
	var w, h = 240, 240
	var hw, hh = float64(w/2), float64(h/2)

	pattern := SingleCross(hw, hh, 15,15,5, palette[0])

	var images []*image.Paletted
	var delays []int
	steps := defaultSteps
	for step := 0; step < steps; step++ {
		img := DrawPalette(w, h, step, &pattern)
		images = append(images, img)
		delays = append(delays, 0)
	}

	f, _ := os.OpenFile("rgb.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	fmt.Printf("Built gif in %d seconds\n", time.Now().Unix() - startTime.Unix())
}