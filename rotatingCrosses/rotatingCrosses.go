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
	defaultSteps = 20
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

type EqualCross struct {
	Center                      Coordinates
	Radius, Rotation, Thickness float64
	xMin, xMax, yMin, yMax      float64
	Color                       color.Color
}

type Cross struct {
	Center                   Coordinates
	Width, Height, Thickness float64
	Color                    color.Color
}

type CrossPattern struct {
	steps []Cross
}

type EqualCrossPattern struct {
	steps []EqualCross
}

type PatternComposite struct {
	Patterns []ShapePattern
}

func (p *PatternComposite) AddChild(pattern ShapePattern) {
	p.Patterns = append(p.Patterns, pattern)
}

func BetweenLines(c1, c2, c3, c4 Coordinates, xCheck, yCheck float64) bool {
	line1Slope := (c2.Y - c1.Y) / (c2.X - c1.X)
	line1Intercept := c1.Y - (line1Slope * c1.X)

	line2Slope := (c4.Y - c3.Y) / (c4.X - c3.X)
	line2Intercept := c3.Y - (line2Slope * c3.X)

	infCheck := math.IsInf(line1Slope, 1) || math.IsInf(line1Slope, -1) || math.IsInf(line2Slope, 1) || math.IsInf(line2Slope, -1)
	approxInfCheck := line1Slope > 1e6 || line1Slope < -1e6 || line2Slope > 1e6 || line2Slope < -1e6
	if infCheck || approxInfCheck {
		xmin := math.Min(c1.X, math.Min(c2.X, math.Min(c3.X, c3.X)))
		xmax := math.Max(c1.X, math.Max(c2.X, math.Max(c3.X, c3.X)))
		if xCheck > xmin && xCheck < xmax {
			return true
		}
		return false
	}
	parallel := yCheck - line1Slope*xCheck
	pmin := math.Min(line1Intercept, line2Intercept)
	pmax := math.Max(line1Intercept, line2Intercept)

	if parallel > pmin && parallel < pmax {
		return true
	}
	return false
}

func (c *EqualCross) Brightness(x, y float64) (bool, color.Color) {
	if x < c.xMin || x > c.xMax || y < c.yMin || y > c.yMax {
		return false, defaultColor
	}

	var points = make([]Coordinates, 8)

	for i := range points {
		even := i%2 == 0
		thickness := c.Thickness
		if even {
			thickness *= -1
		}
		points[i] = Coordinates{
			X: c.Center.X - c.Radius*math.Sin(float64(i/2)*(math.Pi/2)+thickness+c.Rotation),
			Y: c.Center.Y - c.Radius*math.Cos(float64(i/2)*(math.Pi/2)+thickness+c.Rotation),
		}
	}
	bw1 := BetweenLines(points[2], points[7], points[3], points[6], x, y)
	bw2 := BetweenLines(points[0], points[5], points[1], points[4], x, y)
	bw3 := BetweenLines(points[0], points[1], points[4], points[5], x, y)
	bw4 := BetweenLines(points[2], points[3], points[6], points[7], x, y)
	if bw1 && bw4 || bw2 && bw3 {
		return true, c.Color
	}
	return false, defaultColor
}

func (c *EqualCrossPattern) Draw(step int, x, y float64) (bool, color.Color) {
	return c.steps[step].Brightness(x, y)
}

func (p *PatternComposite) Draw(step int, x, y float64) (bool, color.Color) {
	var drawn bool
	var col color.Color
	col = defaultColor
	for _, v := range p.Patterns {
		d, tmpCol := v.Draw(step, x, y)
		if d {
			drawn = d
			col = tmpCol
		}
	}
	return drawn, col
}

func SingleEqualCross(xCenter, yCenter, radius, thickness, rotDir float64, col color.Color) EqualCrossPattern {
	var steps = make([]EqualCross, defaultSteps)
	for i := range steps {
		steps[i] = EqualCross{
			Center:    Coordinates{X: xCenter, Y: yCenter},
			Radius:    radius,
			Thickness: thickness,
			Color:     col,
			Rotation:  (math.Pi * 2) * (float64(i) * rotDir / float64(defaultSteps)),
			xMin:      xCenter - radius,
			xMax:      xCenter + radius,
			yMin:      yCenter - radius,
			yMax:      yCenter + radius,
		}
	}
	return EqualCrossPattern{steps: steps}
}

func RowEqualCross(xCenter, yCenter, radius, thickness, rotDir float64, col color.Color) PatternComposite {
	var children = make([]ShapePattern, 6)

	var xThickness = radius * (math.Cos(0) - math.Cos(thickness))
	var yThickness = radius * (math.Sin(0) - math.Sin(thickness))
	xStart := xCenter + radius - xThickness
	yStart := yCenter + yThickness
	var xSpacer = 2*radius - 2*xThickness
	var ySpacer = 2 * yThickness
	for i := range children {
		x := xStart + float64(i)*xSpacer
		y := yStart + float64(i)*ySpacer
		cross := SingleEqualCross(x, y, radius, thickness, rotDir, col)
		children[i] = &cross
	}
	return PatternComposite{Patterns: children}
}

func GridEqualCross(xStart, yStart, radius, thickness, rotDir float64, col color.Color) PatternComposite {
	var cosThickness = math.Abs(radius*math.Cos(0) - radius*math.Cos(thickness))
	var sinThickness = math.Abs(radius*math.Sin(0) - radius*math.Sin(thickness))
	var patterns PatternComposite
	xNext := xStart - radius - sinThickness
	yNext := yStart + 2*sinThickness - 1
	for i := 1; i < 8; i++ {
		if i != 0 && i%4 == 0 {
			xNext -= 2*radius - 2*cosThickness
			yNext += 2 * sinThickness
		}
		pattern := RowEqualCross(xNext, yNext, radius, thickness, rotDir, col)
		pattern.Draw(0, 0, 0)

		patterns.AddChild(&pattern)
		xNext += 2 * sinThickness
		yNext += 2*radius - 2*cosThickness
	}
	return patterns
}

func DrawPalette(w, h, step int, patterns ShapePattern) *image.Paletted {
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
	var w, h = 235, 235
	//var hw, hh = float64(w/2), float64(h/2)
	var thickness = math.Pi / 10
	var radius = 25.0

	//var cosThickness = math.Abs(radius*math.Cos(0) - radius*math.Cos(thickness))
	var sinThickness = math.Abs(radius*math.Sin(0) - radius*math.Sin(thickness))

	var mainPattern PatternComposite

	pattern1 := GridEqualCross(0, 0, radius, thickness, 1, palette[0])
	mainPattern.AddChild(&pattern1)
	pattern2 := GridEqualCross(-2*sinThickness, 4*sinThickness, radius, thickness, -1, palette[5])
	mainPattern.AddChild(&pattern2)

	var images []*image.Paletted
	var delays []int
	steps := defaultSteps
	for step := 0; step < steps; step++ {
		fmt.Println(step)
		img := DrawPalette(w, h, step, &mainPattern)
		images = append(images, img)
		delays = append(delays, 0)
	}

	f, _ := os.OpenFile("rgb.gif", os.O_WRONLY|os.O_CREATE, 0600)
	defer f.Close()
	gif.EncodeAll(f, &gif.GIF{
		Image: images,
		Delay: delays,
	})
	fmt.Printf("Built gif in %d seconds\n", time.Now().Unix()-startTime.Unix())
}
