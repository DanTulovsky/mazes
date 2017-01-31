package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"mazes/algos"
	"mazes/colors"
	"mazes/grid"
	"os"
	"runtime"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
)

// For gui support
// brew install sdl2{,_image,_ttf,_mixer}
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, run: go install -a mazes/generate

var (
	winTitle string                      = "Maze"
	actions  map[string]func(*grid.Grid) = map[string]func(*grid.Grid){
		"longestPath":        drawLongestPath,
		"shortestRandomPath": drawShortestPathRandomCells,
	}

	rows        = flag.Int("r", 60, "number of rows in the maze")
	columns     = flag.Int("c", 60, "number of rows in the maze")
	bgColor     = flag.String("bgcolor", "white", "background color")
	wallColor   = flag.String("wall_color", "black", "wall color")
	borderColor = flag.String("border_color", "black", "border color")
	pathColor   = flag.String("path_color", "red", "border color")
	cellWidth   = flag.Int("w", 10, "cell width")
	wallWidth   = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	pathWidth   = flag.Int("path_width", 2, "path width")
	showAscii   = flag.Bool("ascii", false, "show ascii maze")
	showGUI     = flag.Bool("gui", true, "show gui maze")
	showStats   = flag.Bool("stats", false, "show maze stats")
	createAlgo  = flag.String("create_algo", "bintree", "algorithm used to create the maze")
	exportFile  = flag.String("export_file", "", "file to save maze to (does not work yet)")
	actionToRun = flag.String("action", "longestPath", "action to run")
)

func setupSDL() (*sdl.Window, *sdl.Renderer) {
	sdl.Init(sdl.INIT_EVERYTHING)
	sdl.EnableScreenSaver()

	// window
	winWidth := (*columns)**cellWidth + *wallWidth*2
	winHeight := (*rows)**cellWidth + *wallWidth*2

	w, err := sdl.CreateWindow(winTitle, 0, 0,
		// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
		winWidth, winHeight, sdl.WINDOW_HIDDEN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		os.Exit(1)
	}

	// renderer
	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	// r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_SOFTWARE)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	// Set options
	// https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode#blendMode
	r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	return w, r
}

// checkAlgo makes sure the passed in algorithm is valid
func checkAlgo(a string) bool {
	for k := range algos.Algorithms {
		if k == a {
			return true
		}
	}
	return false
}

// drawShortestPathRandomCells draws the shortest distance between two random cells
func drawShortestPathRandomCells(g *grid.Grid) {
	fromCell := g.RandomCell()
	toCell := g.RandomCell()
	log.Printf("Finding shortest path: [%v] -> [%v]", fromCell, toCell)

	// For coloring
	g.SetDistanceColors(fromCell)

	// calculates and sets the path between cells
	g.SetPath(fromCell, toCell)
}

// drawLongestPath draws one possible longest path through the maze
func drawLongestPath(g *grid.Grid) {
	dist, fromCell, toCell, _ := g.LongestPath()
	g.SetDistanceColors(fromCell)
	g.SetPath(fromCell, toCell)
	log.Printf("Longest path from [%v]->[%v] = %v", fromCell, toCell, dist)

}

func SaveImage(r *sdl.Renderer, window *sdl.Window, path string) error {
	if path == "" {
		return errors.New("path to file is required!")
	}

	w, h := window.GetSize()
	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0xff000000)
	if err != nil {
		return err
	}

	rect := &sdl.Rect{0, 0, int32(w) - 1, int32(h) - 1}
	pixelFormat, err := window.GetPixelFormat()
	if err != nil {
		return err
	}
	pixels := s.Pixels()
	if err := r.ReadPixels(rect, pixelFormat, unsafe.Pointer(&pixels), int(s.Pitch)); err != nil {
		return err
	}

	img.SavePNG(s, path)
	s.Free()
	return nil
}

// showMazeStats shows some states about the maze
func showMazeStats(g *grid.Grid) {
	x, y := g.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(g.DeadEnds()))
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// profiling
	// defer profile.Start().Stop()

	// For https://github.com/veandco/go-sdl2#faq
	runtime.LockOSThread()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	w, r := setupSDL()
	defer w.Destroy()
	defer r.Destroy()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
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
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	if !checkAlgo(*createAlgo) {
		log.Fatalf("invalid algorithm: %v", *createAlgo)
	}
	// apply algorithm
	algo := algos.Algorithms[*createAlgo]

	g, err = algo.Apply(g)
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := algo.CheckGrid(g); err != nil {
		log.Fatalf("maze is not valid: %v", err)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// {Predefined actions to run}
	//////////////////////////////////////////////////////////////////////////////////////////////
	if action, ok := actions[*actionToRun]; !ok {
		log.Fatalf("no such action [%v]", *actionToRun)
	} else {
		action(g)
	}

	if *showStats {
		showMazeStats(g)
	}

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// ascii maze
	if *showAscii {
		fmt.Printf("%v\n", g)
	}

	// gui maze
	if *showGUI {
		// show window here, otherwise if we just display it, it sits as a white
		// blob while Apply() runs
		w.Show()
		g.ClearDrawPresent(r, w)

		// Save to file
		if *exportFile != "" {
			log.Printf("saving image to: %v", *exportFile)
			if err := SaveImage(r, w, *exportFile); err != nil {
				log.Printf("error saving file: %v", err)
			}
		}

		// wait for GUI to be closed
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

}
