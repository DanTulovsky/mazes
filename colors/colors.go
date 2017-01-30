package colors

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

const DEFAULTCOLOR = "black"

var (
	ColorMap = map[string]Color{
		"black":    {R: 0, G: 0, B: 0, A: 255, Name: "black"},
		"blue":     {R: 0, G: 0, B: 255, A: 255, Name: "blue"},
		"cyan":     {R: 0, G: 255, B: 255, A: 255, Name: "cyan"},
		"deeppink": {R: 255, G: 20, B: 147, A: 255, Name: "deeppink"},
		"gold":     {R: 212, G: 175, B: 55, A: 255, Name: "gold"},
		"gray":     {R: 128, G: 128, B: 128, A: 255, Name: "gray"},
		"green":    {R: 0, G: 128, B: 0, A: 255, Name: "green"},
		"lime":     {R: 0, G: 255, B: 0, A: 255, Name: "lime"},
		"magenta":  {R: 255, G: 0, B: 255, A: 255, Name: "magenta"},
		"maroon":   {R: 128, G: 0, B: 0, A: 255, Name: "maroon"},
		"navy":     {R: 0, G: 0, B: 128, A: 255, Name: "navy"},
		"olive":    {R: 128, G: 128, B: 0, A: 255, Name: "olive"},
		"purple":   {R: 128, G: 0, B: 128, A: 255, Name: "purple"},
		"red":      {R: 255, G: 0, B: 0, A: 255, Name: "red"},
		"silver":   {R: 192, G: 192, B: 192, A: 255, Name: "silver"},
		"teal":     {R: 0, G: 128, B: 128, A: 255, Name: "teal"},
		"white":    {R: 255, G: 255, B: 255, A: 255, Name: "white"},
		"yellow":   {R: 255, G: 255, B: 0, A: 255, Name: "yellow"},
	}
)

type Color struct {
	R, G, B uint8  // colors
	A       uint8  // alpha mask
	Name    string // english name
}

// OpacityAdjust returns a Color object with its opacity adjust by "i". i can be negative.
func OpacityAdjust(c Color, i int) Color {
	a := int(c.A) - i

	c.A = uint8(a)
	c.Name = "custom"
	return c
}

// GetColor returns *Color based on the name
func GetColor(c string) Color {
	color, ok := ColorMap[c]
	if !ok {
		log.Printf("invalid color [%v], returning default", c)
		return ColorMap[DEFAULTCOLOR]
	}
	return color
}

// SetDrawColor Sets the drawing color on the renderer
func SetDrawColor(c Color, r *sdl.Renderer) *sdl.Renderer {
	if err := r.SetDrawColor(c.R, c.G, c.B, c.A); err != nil {
		log.Printf("error setting color [%#v]: %v", c, err)
	}
	return r
}
