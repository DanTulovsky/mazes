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
	pathColor          = flag.String("path_color", "red", "border color")
	cellWidth          = flag.Int("w", 20, "cell width")
	wallWidth          = flag.Int("wall_width", 4, "wall width (min of 2 to have walls - half on each side")
	pathWidth          = flag.Int("path_width", 2, "path width")
	showAscii          = flag.Bool("ascii", false, "show ascii maze")
)

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// For https://github.com/veandco/go-sdl2#faq
	runtime.LockOSThread()

	config := &grid.Config{
		Rows:        *rows,
		Columns:     *columns,
		CellWidth:   *cellWidth,
		WallWidth:   *wallWidth,
		PathWidth:   *pathWidth,
		BgColor:     colors.GetColor(*bgColor),
		BorderColor: colors.GetColor(*borderColor),
		WallColor:   colors.GetColor(*wallColor),
		PathColor:   colors.GetColor(*pathColor),
	}

	g, err := grid.NewGrid(config)
	if err != nil {
		fmt.Printf("invalid config: %v", err)
		os.Exit(1)
	}

	// apply algorithm
	g = bintree.Apply(g)

	// find the longest path in the maze automatically
	dist, fromCell, toCell, _ := g.LongestPath()
	log.Printf("Longest path from [%v]->[%v] = %v", fromCell, toCell, dist)

	//x, y := *rows/2, *columns/2
	//fromCell, err := g.Cell(x, y)
	//if err != nil {
	//	log.Fatalf("error getting cell: %v", err)
	//}
	//toCell, err := g.Cell(x+10, y)
	//if err != nil {
	//	log.Fatalf("error getting cell: %v", err)
	//}
	//// Find shortests distance between fromCell and toCell
	// dist, path = g.ShortestPath(fromCell, toCell)
	// log.Printf("Shortest path from [%v]->[%v] = %v > %v", fromCell, toCell, dist, path)

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// ascii maze
	if *showAscii {
		fmt.Printf("%v\n", g)
	}

	// GUI maze
	sdl.Init(sdl.INIT_EVERYTHING)
	sdl.EnableScreenSaver()

	// window
	window, err := sdl.CreateWindow(winTitle, 0, 0,
		// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
		(*columns)**cellWidth, (*rows)**cellWidth, sdl.WINDOW_SHOWN)
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

	// https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode#blendMode
	r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
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
