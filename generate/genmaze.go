package main

import (
	"flag"
	"fmt"
	"mazes/genalgos/bintree"
	"mazes/grid"
	"os"
	"runtime"

	"log"

	"mazes/colors"

	"github.com/veandco/go-sdl2/sdl"
)

// For gui support
// brew install sdl2{,_image,_ttf,_mixer}
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}

var (
	winTitle    string = "Maze"
	rows               = flag.Int("r", 30, "number of rows in the maze")
	columns            = flag.Int("c", 30, "number of rows in the maze")
	bgColor            = flag.String("bgcolor", "white", "background color")
	wallColor          = flag.String("wall_color", "black", "wall color")
	borderColor        = flag.String("border_color", "red", "border color")
	cellWidth          = flag.Int("w", 20, "cell width")
	showAscii          = flag.Bool("ascii", false, "show ascii maze")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// For https://github.com/veandco/go-sdl2#faq
	runtime.LockOSThread()

	g := grid.NewGrid(*rows, *columns, *cellWidth, colors.GetColor(*bgColor), colors.GetColor(*borderColor), colors.GetColor(*wallColor))

	// apply algorithm
	g = bintree.Apply(g)

	// apply Dijkstra’s to record distance information
	x, y := *rows/2, *columns/2
	source, err := g.Cell(x, y)
	if err != nil {
		log.Fatalf("error getting cell: %v", err)
	}
	d := source.Distances()
	// log.Printf("%#v", d)
	for _, c := range d.Cells() {
		dist, _ := d.Get(c)
		log.Printf("%v: %v", c, dist)
	}

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// ascii maze
	if *showAscii {
		fmt.Printf("%v\n", g)
	}

	// GUI maze
	sdl.Init(sdl.INIT_EVERYTHING)

	// window
	window, err := sdl.CreateWindow(winTitle, 0, 0,
		// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
		(*rows)**cellWidth, (*columns)**cellWidth, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		os.Exit(1)
	}
	defer window.Destroy()

	// renderer
	r, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}
	defer r.Destroy()

	r.Clear() // call this before every Present()

	r = g.Draw(r) // adds maze to render

	fmt.Print("Rendering maze...\n")
	r.Present()

L:
	for {
		event := sdl.WaitEvent()
		switch event.(type) {
		case *sdl.QuitEvent:
			break L
		}

	}

	sdl.Quit()

}
