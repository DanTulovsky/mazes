package main

import (
	"bufio"

	"github.com/veandco/go-sdl2/sdl"

	"flag"
	"fmt"
	"log"
	"mazes/algos"
	"mazes/colors"
	"mazes/grid"
	"os"

	"mazes/solvealgos"
	"sync"
)

// For gui support
// brew install sdl2{,_image,_ttf,_mixer}
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, run: go install -a mazes/generate
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

var (
	winTitle string                      = "Maze"
	actions  map[string]func(*grid.Grid) = map[string]func(*grid.Grid){
		"longestPath":        drawLongestPath,
		"shortestRandomPath": drawShortestPathRandomCells,
	}
	fromCell, toCell *grid.Cell

	w            *sdl.Window
	r            *sdl.Renderer
	sdlErr       error
	runningMutex sync.Mutex

	solver solvealgos.Algorithmer

	rows                 = flag.Int("r", 60, "number of rows in the maze")
	columns              = flag.Int("c", 60, "number of rows in the maze")
	bgColor              = flag.String("bgcolor", "white", "background color")
	wallColor            = flag.String("wall_color", "black", "wall color")
	borderColor          = flag.String("border_color", "black", "border color")
	currentLocationColor = flag.String("location_color", "yellow", "border color")
	pathColor            = flag.String("path_color", "red", "border color")
	visitedCellColor     = flag.String("visited_color", "red", "color of visited cell marker")
	cellWidth            = flag.Int("w", 10, "cell width")
	wallWidth            = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	pathWidth            = flag.Int("path_width", 2, "path width")
	showAscii            = flag.Bool("ascii", false, "show ascii maze")
	showGUI              = flag.Bool("gui", true, "show gui maze")
	showStats            = flag.Bool("stats", false, "show maze stats")
	markVisitedCells     = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	createAlgo           = flag.String("create_algo", "bintree", "algorithm used to create the maze")
	exportFile           = flag.String("export_file", "", "file to save maze to (does not work yet)")
	actionToRun          = flag.String("action", "", "action to run")
	solveAlgo            = flag.String("solve_algo", "recursive-backtracker", "algorithm to solve the maze")
	frameRate            = flag.Uint("frame_rate", 60, "frame rate for animation")
)

func setupSDL() {
	if !*showGUI {
		return
	}
	sdl.Do(func() {
		sdl.Init(sdl.INIT_EVERYTHING)
		sdl.EnableScreenSaver()
	})

	// window
	winWidth := (*columns)**cellWidth + *wallWidth*2
	winHeight := (*rows)**cellWidth + *wallWidth*2

	sdl.Do(func() {
		w, sdlErr = sdl.CreateWindow(winTitle, 0, 0,
			// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
			winWidth, winHeight, sdl.WINDOW_SHOWN)
	})
	if sdlErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", sdlErr)
		os.Exit(1)
	}

	// renderer
	sdl.Do(func() {
		r, sdlErr = sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	})
	if sdlErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", sdlErr)
		os.Exit(1)
	}

	// Set options
	// https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode#blendMode
	sdl.Do(func() {
		r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	})

	sdl.Do(func() {
		r.Clear()
	})
}

// checkCreateAlgo makes sure the passed in algorithm is valid
func checkCreateAlgo(a string) bool {
	for k := range algos.Algorithms {
		if k == a {
			return true
		}
	}
	return false
}

// checkSolveAlgo makes sure the passed in algorithm is valid
func checkSolveAlgo(a string) bool {
	for k := range algos.SolveAlgorithms {
		if k == a {
			return true
		}
	}
	return false
}

// drawShortestPathRandomCells draws the shortest distance between two random cells
func drawShortestPathRandomCells(g *grid.Grid) {
	fromCell = g.RandomCell()
	toCell = g.RandomCell()
	log.Printf("Finding shortest path: [%v] -> [%v]", fromCell, toCell)

	// For coloring
	g.SetDistanceColors(fromCell)

	// calculates and sets the path between cells
	g.SetPath(fromCell, toCell)
}

// drawLongestPath draws one possible longest path through the maze
func drawLongestPath(g *grid.Grid) {
	var dist int
	dist, fromCell, toCell, _ = g.LongestPath()
	g.SetDistanceColors(fromCell)
	g.SetPath(fromCell, toCell)
	log.Printf("Longest path from [%v]->[%v] = %v", fromCell, toCell, dist)

}

//func SaveImage(r *sdl.Renderer, window *sdl.Window, path string) error {
//	if path == "" {
//		return errors.New("path to file is required!")
//	}
//
//	w, h := window.GetSize()
//	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0xff000000)
//	if err != nil {
//		return err
//	}
//
//	rect := &sdl.Rect{0, 0, int32(w) - 1, int32(h) - 1}
//	pixelFormat, err := window.GetPixelFormat()
//	if err != nil {
//		return err
//	}
//	pixels := s.Pixels()
//	if err := r.ReadPixels(rect, pixelFormat, unsafe.Pointer(&pixels), int(s.Pitch)); err != nil {
//		return err
//	}
//
//	img.SavePNG(s, path)
//	s.Free()
//	return nil
//}

// showMazeStats shows some states about the maze
func showMazeStats(g *grid.Grid) {
	x, y := g.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(g.DeadEnds()))
}

func waitGUI() {
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

func main() {
	os.Exit(run())
}

func run() int {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// profiling
	// defer profile.Start().Stop()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	setupSDL()

	defer func() { sdl.Do(func() { w.Destroy() }) }()
	defer func() { sdl.Do(func() { r.Destroy() }) }()
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &grid.Config{
		Rows:                 *rows,
		Columns:              *columns,
		CellWidth:            *cellWidth,
		WallWidth:            *wallWidth,
		PathWidth:            *pathWidth,
		BgColor:              colors.GetColor(*bgColor),
		BorderColor:          colors.GetColor(*borderColor),
		WallColor:            colors.GetColor(*wallColor),
		PathColor:            colors.GetColor(*pathColor),
		VisitedCellColor:     colors.GetColor(*visitedCellColor),
		MarkVisitedCells:     *markVisitedCells,
		CurrentLocationColor: colors.GetColor(*currentLocationColor),
	}

	g, err := grid.NewGrid(config)
	if err != nil {
		fmt.Printf("invalid config: %v", err)
		os.Exit(1)
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	if !checkCreateAlgo(*createAlgo) {
		log.Fatalf("invalid create algorithm: %v", *createAlgo)
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
	if *actionToRun != "" {
		if action, ok := actions[*actionToRun]; !ok {
			log.Fatalf("no such action [%v]", *actionToRun)
		} else {
			action(g)
		}

		if *showStats {
			showMazeStats(g)
		}
	}

	///////////////////////////////////////////////////////////////////////////
	// Solvers
	///////////////////////////////////////////////////////////////////////////
	if *solveAlgo != "" {
		if !checkSolveAlgo(*solveAlgo) {
			log.Fatalf("invalid solve algorithm: %v", *solveAlgo)
		}

		// solve the longest path
		if fromCell == nil || toCell == nil {
			log.Printf("No fromCella and toCell set, defaulting to longestPath.")
			_, fromCell, toCell, _ = g.LongestPath()
		}

		g.SetDistanceColors(fromCell)
		g.SetFromToColors(fromCell, toCell)
		g.ResetVisited()

		solver = algos.SolveAlgorithms[*solveAlgo]
		g, err = solver.Solve(g, fromCell, toCell)
		if err != nil {
			log.Fatalf("error running solver: %v", err)
		}
		log.Printf("time to solve: %v", solver.SolveTime())
		log.Printf("steps taken to solve: %v", solver.SolveSteps())
		log.Printf("steps in shortest path: %v", len(solver.SolvePath().List()))
	}

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// ascii maze
	if *showAscii {
		fmt.Printf("%v\n", g)
	}

	animation := 0
	// gui maze
	if *showGUI {
		running := true

		for running {

			//sdl.Do(func() {
			//	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			//		switch event.(type) {
			//		case *sdl.QuitEvent:
			//			runningMutex.Lock()
			//			running = false
			//			runningMutex.Unlock()
			//		}
			//	}

			// Displays the main maze, no paths or other markers
			sdl.Do(func() {
				// reset the clear color back to black
				colors.SetDrawColor(colors.GetColor("black"), r)

				r.Clear()
				g.DrawMaze(r)
			})

			// For solvers, to animate the path.
			if *solveAlgo != "" {
				// update display
				// wg := sync.WaitGroup{}
				// wg.Add(1)

				// used to draw only a part of the path
				x := animation
				if x > len(solver.TravelPath().List()) {
					x = len(solver.TravelPath().List()) - 1
				}
				sdl.Do(func() {
					path := grid.NewPath()
					path.AddSegements(solver.TravelPath().List()[0:x])
					g.DrawPath(r, path, *markVisitedCells)
				})
				animation++

				// wg.Wait()

			}

			sdl.Do(func() {
				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
				// fmt.Print("Press 'Enter' to continue...")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			})
		}

	}
	return 0
}

//// Save to file
//if *exportFile != "" {
//	log.Printf("saving image to: %v", *exportFile)
//	if err := SaveImage(r, w, *exportFile); err != nil {
//		log.Printf("error saving file: %v", err)
//	}
//}
