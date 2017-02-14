package main

import (
	"github.com/pkg/profile"
	"github.com/veandco/go-sdl2/sdl"

	"flag"
	"fmt"
	"log"
	"mazes/algos"
	"mazes/colors"
	"os"

	"mazes/solvealgos"
	"time"

	"mazes/maze"

	"sync"

	"github.com/sasha-s/go-deadlock"
)

// For gui support
// brew install sdl2{,_image,_ttf,_mixer}
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, run: go install -a mazes/generate
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

var (
	winTitle         string = "Maze"
	fromCell, toCell *maze.Cell

	w      *sdl.Window
	r      *sdl.Renderer
	sdlErr error
	// runningMutex sync.Mutex

	solver solvealgos.Algorithmer

	rows                    = flag.Int("r", 30, "number of rows in the maze")
	columns                 = flag.Int("c", 60, "number of rows in the maze")
	bgColor                 = flag.String("bgcolor", "white", "background color")
	wallColor               = flag.String("wall_color", "black", "wall color")
	borderColor             = flag.String("border_color", "black", "border color")
	currentLocationColor    = flag.String("location_color", "lime", "border color")
	pathColor               = flag.String("path_color", "red", "border color")
	visitedCellColor        = flag.String("visited_color", "red", "color of visited cell marker")
	cellWidth               = flag.Int("w", 20, "cell width (best as multiple of 2)")
	wallWidth               = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	pathWidth               = flag.Int("path_width", 2, "path width")
	showAscii               = flag.Bool("ascii", false, "show ascii maze")
	darkMode                = flag.Bool("dark_mode", false, "only show cells solver has seen")
	showGUI                 = flag.Bool("gui", true, "show gui maze")
	showStats               = flag.Bool("stats", false, "show maze stats")
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")
	markVisitedCells        = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	createAlgo              = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	maskImage               = flag.String("mask_image", "", "file name of mask image")
	// exportFile           = flag.String("export_file", "", "file to save maze to (does not work yet)")
	solveAlgo      = flag.String("solve_algo", "recursive-backtracker", "algorithm to solve the maze")
	frameRate      = flag.Uint("frame_rate", 120, "frame rate for animation")
	genDrawDelay   = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	solveDrawDelay = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")

	winWidth, winHeight int
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
	winWidth = (*columns)**cellWidth + *wallWidth*2
	winHeight = (*rows)**cellWidth + *wallWidth*2

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
//func drawShortestPathRandomCells(g *grid.Grid) {
//	fromCell = g.RandomCell()
//	toCell = g.RandomCell()
//	log.Printf("Finding shortest path: [%v] -> [%v]", fromCell, toCell)
//
//	// For coloring
//	g.SetDistanceColors(fromCell)
//
//	// calculates and sets the path between cells
//	g.SetPath(fromCell, toCell)
//}

// drawLongestPath draws one possible longest path through the maze
//func drawLongestPath(g *grid.Grid) {
//	var dist int
//	dist, fromCell, toCell, _ = g.LongestPath()
//	g.SetDistanceColors(fromCell)
//	g.SetPath(fromCell, toCell)
//	log.Printf("Longest path from [%v]->[%v] = %v", fromCell, toCell, dist)
//
//}

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
func showMazeStats(m *maze.Maze) {
	x, y := m.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(m.DeadEnds()))
}

//func waitGUI() {
//L:
//	for {
//		event := sdl.WaitEvent()
//		switch event.(type) {
//		case *sdl.QuitEvent:
//			break L
//		}
//
//	}
//
//	sdl.Quit()
//}

// Solve runs the solvers against the grid.
func Solve(m *maze.Maze) (solvealgos.Algorithmer, error) {
	var err error

	if !checkSolveAlgo(*solveAlgo) {
		return nil, fmt.Errorf("invalid solve algorithm: %v", *solveAlgo)
	}

	log.Printf("running solver %v", *solveAlgo)

	m.Reset()

	solver = algos.SolveAlgorithms[*solveAlgo]
	delay, err := time.ParseDuration(*solveDrawDelay)
	if err != nil {
		return nil, err
	}
	m, err = solver.Solve(m, fromCell, toCell, delay)
	if err != nil {
		return nil, fmt.Errorf("error running solver: %v", err)
	}
	log.Printf("time to solve: %v", solver.SolveTime())
	log.Printf("steps taken to solve:   %v", solver.SolveSteps())
	log.Printf("steps in shortest path: %v", solver.SolvePath().Length())

	return solver, nil

}

func main() {
	//filename, _ := osext.Executable()
	//fmt.Println(filename)

	os.Exit(run())
}

func run() int {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *enableDeadlockDetection {
		log.Println("enabling deadlock detection, this slows things down considerably!")
		deadlock.Opts.Disable = false
	} else {
		deadlock.Opts.Disable = true
	}

	if *enableProfile {
		log.Println("enabling profiling...")
		defer profile.Start().Stop()
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &maze.Config{
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
		DarkMode:             *darkMode,
	}

	var m *maze.Maze
	var err error

	// Mask image if provided.
	// If the mask image is provided, use that as the dimensions of the grid
	if *maskImage != "" {
		log.Printf("Using %v as grid mask", *maskImage)
		m, err = maze.NewMazeFromImage(config, *maskImage)
		if err != nil {
			fmt.Printf("invalid config: %v", err)
			os.Exit(1)
		}
		// Set these for correct window size
		*columns, *rows = m.Dimensions()
	} else {
		m, err = maze.NewMaze(config)
		if err != nil {
			fmt.Printf("invalid config: %v", err)
			os.Exit(1)
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	setupSDL()

	defer func() { sdl.Do(func() { w.Destroy() }) }()
	defer func() { sdl.Do(func() { r.Destroy() }) }()
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////

	if !checkCreateAlgo(*createAlgo) {
		log.Fatalf("invalid create algorithm: %v", *createAlgo)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// {Predefined actions to run}
	//////////////////////////////////////////////////////////////////////////////////////////////
	//if *actionToRun != "" {
	//	if action, ok := actions[*actionToRun]; !ok {
	//		log.Fatalf("no such action [%v]", *actionToRun)
	//	} else {
	//		action(g)
	//	}
	//
	//}

	///////////////////////////////////////////////////////////////////////////
	// Generator
	///////////////////////////////////////////////////////////////////////////
	// apply algorithm
	algo := algos.Algorithms[*createAlgo]

	delay, err := time.ParseDuration(*genDrawDelay)
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Printf("running generator %v", *createAlgo)
	m, err = algo.Apply(m, delay)
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := algo.CheckGrid(m); err != nil {
		log.Fatalf("maze is not valid: %v", err)
	}

	if *showStats {
		showMazeStats(m)
	}

	// solve the longest path
	if fromCell == nil || toCell == nil {
		log.Print("No fromCella and toCell set, defaulting to longestPath.")
		_, fromCell, toCell, _ = m.LongestPath()
	}

	m.SetDistanceColors(fromCell)
	m.SetFromToColors(fromCell, toCell)

	///////////////////////////////////////////////////////////////////////////
	// End Generator
	///////////////////////////////////////////////////////////////////////////

	runSolver := false
	var wd sync.WaitGroup
	///////////////////////////////////////////////////////////////////////////
	// Generators/Solvers
	///////////////////////////////////////////////////////////////////////////
	wd.Add(1)
	go func() {
		for !runSolver {
			log.Println("Maze not yet ready, sleeping 1s...")
			time.Sleep(time.Second)
		}

		if *solveAlgo != "" {
			solver, err = Solve(m)
			if err != nil {
				log.Print(err)
			}
		}
		wd.Done()
	}()
	///////////////////////////////////////////////////////////////////////////
	// End Generators/Solvers
	///////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// ascii maze
	if *showAscii {
		fmt.Printf("%v\n", m)
	}

	// gui maze
	if *showGUI {

		running := true

		// create background texture, it is saved and re-rendered as a picture
		mTexture, err := r.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET, winWidth, winHeight)
		if err != nil {
			log.Fatalf("failed to create background: %v", err)
		}

		// draw on the texture
		r.SetRenderTarget(mTexture)
		r.Clear()
		m.DrawMazeBackground(r)
		r.Present()

		// Reset to drawing on the screen
		r.SetRenderTarget(nil)
		r.Copy(mTexture, nil, nil)
		r.Present()

		runSolver = true
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
			//})

			// Displays the main maze, no paths or other markers
			sdl.Do(func() {
				// reset the clear color back to black
				colors.SetDrawColor(colors.GetColor("black"), r)

				r.Clear()
				m.DrawMaze(r, mTexture)
			})

			//	// wg := sync.WaitGroup{}
			//	// wg.Add(1)
			//
			// Do things between drawing here
			//
			//	// wg.Wait()

			sdl.Do(func() {
				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
				// fmt.Print("Press 'Enter' to continue...")
				// bufio.NewReader(os.Stdin).ReadBytes('\n')
			})
		}

	} else {

		// wait for solver thread here, used if gui not shown
		runSolver = true
		wd.Wait()
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
