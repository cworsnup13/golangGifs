package main

import (
	"image"
	"image/color"
	"image/gif"
	"math"
	"os"
)

const (
	defaultSteps = 100
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
	Draw(step int, x, y float64) uint8
}

type Coordinates struct {
	X, Y float64
}

type Circle struct {
	X, Y, R float64
}

type Pattern struct {
	StepPoints [][]*Circle
}

func (p *Pattern) Draw(step int, x, y float64) uint8 {
	for _, v := range p.StepPoints[step] {
		if v.Brightness(x, y) == 0 {
			return 0
		}
	}
	return 255
}

func (c *Circle) Brightness(x, y float64) uint8 {
	if math.Sqrt(math.Pow(c.X-x, 2)+math.Pow(c.Y-y, 2))/c.R > 1 {
		return 255
	}
	return 0
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
	var points = make([][]*Circle, defaultSteps)
	for i := range points {
		points[i] = []*Circle{{X: pattern[i].X, Y: pattern[i].Y, R: 3,}}
	}
	return &Pattern{points}
}

func RotatingSquares(hw, hh float64) ShapePattern {
	var squares = make([][]Coordinates, 4)
	for i := range squares {
		squares[i] = GetSquarePattern(hw, hh, 50, 1, 25*i)
	}

	var squarePoints = make([][]*Circle, defaultSteps)
	for i := range squarePoints {
		squarePoints[i] = make([]*Circle, len(squares))
		for j := range squarePoints[i] {
			squarePoints[i][j] = &Circle{X: squares[j][i].X, Y: squares[j][i].Y, R: 3,}
		}
	}
	return &Pattern{squarePoints}
}

func DrawPalette(w, h, step int, patterns []ShapePattern, ) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, w, h), palette)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			px := uint8(255)
			for _, v := range patterns {
				tmp := v.Draw(step, float64(x), float64(y))
				if tmp < px {
					px = tmp
				}
			}
			img.Set(x, y, color.RGBA{px, px, px, 255})
		}
	}
	return img
}

func main() {
	var w, h = 240, 240
	var hw, hh = float64(w/2), float64(h/2)

	patterns := []ShapePattern{RotatingSquares(hw, hh), RotatingCircle(hw, hh)}

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
}

// Seven Pointed
//θ := 2.0 * math.Pi / float64(steps) * float64(step)
//θ0 := 2 * math.Pi / 3 * 0.0 // figure out phase later
//circles[0].X = hw - 100*math.Sin(θ0+θ)
//circles[0].Y = hh - 100*math.Cos(θ0+θ)
//circles[0].R = 3
//for i, point := range points {
//	point.X = hw - 100*math.Sin(float64(i) * (2.0 * math.Pi/float64(len(points))))
//	point.Y = hh - 100*math.Cos(float64(i) * (2.0 * math.Pi/float64(len(points))))
//	point.R = 3
//}
