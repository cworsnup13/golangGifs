package main

import (
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
	"fmt"
	"time"
)

const (
	defaultSteps = 120
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
	Name() string
	Brightness(x, y float64) (bool, color.Color)
}

type Coordinates struct {
	X, Y float64
}

type Circle struct {
	X, Y, R float64
}

type Line struct {
	Start, End Coordinates
	Color color.Color
}

type Pattern struct {
	StepPoints [][]Shape
}

func (p *Pattern) Draw(step int, x, y float64) (bool, color.Color) {
	for _, v := range p.StepPoints[step] {
		draw, col := v.Brightness(x, y)
		if draw {
			return true, col
		}
	}
	return false, nil
}

func (c *Circle) Name() string {
	return "circle"
}

func (l *Line) Name() string {
	return "line"
}

func (c *Circle) Brightness(x, y float64) (bool, color.Color) {
	if math.Sqrt(math.Pow(c.X-x, 2)+math.Pow(c.Y-y, 2))/c.R > 1 {
		return false, nil
	}
	return true, palette[0]
}

func (l *Line) Brightness(x, y float64) (bool, color.Color) {
	if x < math.Min(l.End.X, l.Start.X){
		return false, nil
	}
	if x > math.Max(l.End.X, l.Start.X){
		return false, nil
	}

	if y < math.Min(l.End.Y, l.Start.Y){
		return false, nil
	}
	if y > math.Max(l.End.Y, l.Start.Y){
		return false, nil
	}

	slope := (l.End.Y - l.Start.Y) / (l.End.X - l.Start.X)
	intercept := l.Start.Y - (slope * l.Start.X)
	if math.Abs(y - (slope * x + intercept)) < 1{
		return true, l.Color
	}

	return false, nil
}

func GetSquarePattern(xCenter, yCenter, radius, speed float64, phase int) []Coordinates {
	diameter := radius * 2
	sideLength := diameter / (math.Sqrt(2))
	origin := Coordinates{
		X: xCenter - (sideLength / 2),
		Y: yCenter - (sideLength / 2),
	}
	points := []Coordinates{
		origin,
		{X: origin.X + sideLength, Y: origin.Y},
		{X: origin.X + sideLength, Y: origin.Y + sideLength},
		{X: origin.X, Y: origin.Y + sideLength},
	}

	var coords = make([]Coordinates, defaultSteps)
	var oneQuarter = int(float64(len(coords)) / 4.0 / speed)
	var start = points[0]
	var end = points[1]
	for i := range coords {
		mod := i % oneQuarter
		if mod == 0 {
			start = points[(i/oneQuarter)%4]
			end = points[((i/oneQuarter)+1)%4]
		}
		coords[i].X = float64(mod)*((end.X-start.X)/float64(oneQuarter)) + start.X
		coords[i].Y = float64(mod)*(end.Y-start.Y)/float64(oneQuarter) + start.Y
	}
	coordsPhase := append(coords[phase:], coords[:phase]...)
	return coordsPhase
}

func GetHeptagramPattern(xCenter, yCenter, radius, speed float64, phase int) []Coordinates {

	var points = make([]Coordinates, 7)
	for i := range points {
		points[i] = Coordinates{
			X: xCenter - radius * math.Sin(float64(i) * (2.0 * math.Pi/float64(len(points)))),
			Y: yCenter - radius * math.Cos(float64(i) * (2.0 * math.Pi/float64(len(points)))),
		}
	}

	var coords = make([]Coordinates, defaultSteps)
	var onePart = int(float64(len(coords)) / float64(len(points)) / speed)
	var start, end Coordinates
	end = points[0]
	var startIdx int
	for i := range coords {
		mod := i % onePart
		if mod == 0 {
			start = end
			startIdx = (startIdx+3) % len(points)
			end = points[startIdx]
		}
		coords[i].X = float64(mod)*((end.X-start.X)/float64(onePart)) + start.X
		coords[i].Y = float64(mod)*(end.Y-start.Y)/float64(onePart) + start.Y
	}
	coordsPhase := append(coords[phase:], coords[:phase]...)
	return coordsPhase
}

func GetCirclePattern(xCenter, yCenter, radius, speed float64, phase int) []Coordinates {
	var coords = make([]Coordinates, defaultSteps)
	for i := range coords {
		θ := 2.0 * math.Pi / float64(len(coords)) * float64(i)
		θ0 := 2 * math.Pi / 3 * 0.0 // figure out phase later
		coords[i].X = xCenter - radius*math.Sin(θ0+θ)
		coords[i].Y = yCenter - radius*math.Cos(θ0+θ)
	}
	coordsPhase := append(coords[phase:], coords[:phase]...)
	return coordsPhase
}

func RotatingCircle(hw, hh float64) ShapePattern {
	pattern := GetCirclePattern(hw, hh, 50, 1, 0)
	var points = make([][]Shape, defaultSteps)
	for i := range points {
		points[i] = []Shape{&Circle{X: pattern[i].X, Y: pattern[i].Y, R: 3,}}
	}
	return &Pattern{points}
}

func RotatingSquares(hw, hh float64) ShapePattern {
	var squares = make([][]Coordinates, 4)
	for i := range squares {
		squares[i] = GetSquarePattern(hw, hh, 50, 1, 25*i)
	}

	var squarePoints = make([][]Shape, defaultSteps)
	for i := range squarePoints {
		squarePoints[i] = make([]Shape, len(squares))
		for j := range squarePoints[i] {
			squarePoints[i][j] = &Circle{X: squares[j][i].X, Y: squares[j][i].Y, R: 3,}
		}
	}
	return &Pattern{squarePoints}
}

func RotatingHeptagram(hw, hh float64) ShapePattern {
	var set = make([][]Coordinates, 12)

	for i := range set {
		set[i] = GetHeptagramPattern(hw, hh, 100, 1, i * (defaultSteps/len(set)))
	}

	var steps = make([][]Shape, defaultSteps)
	for i := range steps {
		steps[i] = make([]Shape, 3 * len(set))
		for j := range steps[i] {
			if j < len(set) {
				steps[i][j] = &Circle{X: set[j][i].X, Y: set[j][i].Y, R: 3,}
			} else if j < 2 * len(set) {
				steps[i][j] = &Line{Start: set[j%len(set)][i], End: set[(j+3)%len(set)][i], Color:palette[1]}
			} else {
				steps[i][j] = &Line{Start: set[j%len(set)][i], End: set[(j+4)%len(set)][i], Color:palette[2]}
			}
		}
	}
	return &Pattern{steps}
}

func DrawPalette(w, h, step int, patterns []ShapePattern, ) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			var col color.Color
			col = color.RGBA{255,255,255,255}
			for _, v := range patterns {
				drawn, tmpColor := v.Draw(step, float64(x), float64(y))
				if drawn {
					col = tmpColor
				}
			}
			img.Set(x, y, col)
		}
	}
	return img
}

func main() {
	startTime := time.Now()
	var w, h = 240, 240
	var hw, hh = float64(w/2), float64(h/2)

	//patterns := []ShapePattern{RotatingSquares(hw, hh), RotatingCircle(hw, hh)}
	patterns := []ShapePattern{RotatingHeptagram(hw, hh)}

	var images []*image.Paletted
	var delays []int
	steps := defaultSteps
	for step := 0; step < steps; step++ {
		img := DrawPalette(w, h, step, patterns)
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
