package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"mazes/algos"
	"mazes/colors"
	"mazes/maze"
	"mazes/solvealgos"

	"strings"

	"github.com/pkg/profile"
	"github.com/sasha-s/go-deadlock"
	"github.com/tevino/abool"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_image"
	"github.com/veandco/go-sdl2/sdl_mixer"
)

// For gui support
// brew install sdl2{_image,_ttf,_gfx}
// brew install sdl2_mixer --with-flac --with-fluid-synth --with-libmikmod --with-libmodplug --with-libvorbis --with-smpeg2
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

	// maze
	maskImage          = flag.String("mask_image", "", "file name of mask image")
	allowWeaving       = flag.Bool("weaving", false, "allow weaving")
	weavingProbability = flag.Float64("weaving_probability", 1, "controls the amount of weaving that happens, with 1 being the max")
	braidProbability   = flag.Float64("braid_probability", 0, "braid the maze with this probabily, 0 results in a perfect maze, 1 results in no deadends at all")
	randomFromTo       = flag.Bool("random_path", false, "show a random path through the maze")

	// dimensions
	rows    = flag.Int("r", 30, "number of rows in the maze")
	columns = flag.Int("c", 60, "number of rows in the maze")

	// colors
	bgColor              = flag.String("bgcolor", "white", "background color")
	borderColor          = flag.String("border_color", "black", "border color")
	currentLocationColor = flag.String("location_color", "lime", "border color")
	pathColor            = flag.String("path_color", "red", "border color")
	visitedCellColor     = flag.String("visited_color", "red", "color of visited cell marker")
	wallColor            = flag.String("wall_color", "black", "wall color")
	fromCellColor        = flag.String("from_cell_color", "gold", "from cell color")
	toCellColor          = flag.String("to_cell_color", "yellow", "to cell color")

	// width
	cellWidth = flag.Int("w", 20, "cell width (best as multiple of 2)")
	pathWidth = flag.Int("path_width", 2, "path width")
	wallWidth = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	wallSpace = flag.Int("wall_space", 0, "how much space between two side by side walls (min of 2)")

	// maze draw
	showAscii = flag.Bool("ascii", false, "show ascii maze")
	showGUI   = flag.Bool("gui", true, "show gui maze")

	// display
	avatarImage        = flag.String("avatar_image", "", "file name of avatar image, the avatar should be facing to the left in the image")
	frameRate          = flag.Uint("frame_rate", 120, "frame rate for animation")
	genDrawDelay       = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	markVisitedCells   = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	showFromToColors   = flag.Bool("show_from_to_colors", false, "show from/to colors")
	showDistanceColors = flag.Bool("show_distance_colors", false, "show distance colors")
	showDistanceValues = flag.Bool("show_distance_values", false, "show distance values")
	solveDrawDelay     = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")

	// algo
	createAlgo = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	solveAlgo  = flag.String("solve_algo", "", "algorithm to solve the maze")

	// misc
	exportFile = flag.String("export_file", "", "file to save maze to (does not work yet)")
	bgMusic    = flag.String("bg_music", "", "file name of background music to play")

	// stats
	showStats = flag.Bool("stats", false, "show maze stats")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	fromCellStr = flag.String("from_cell", "", "path from cell ('min' = minX, minY)")
	toCellStr   = flag.String("to_cell", "", "path to cell ('max' = maxX, maxY)")

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
			winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL)
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

func SaveImage(r *sdl.Renderer, window *sdl.Window, path string) error {
	return errors.New("exporting to file doesn't work yet...")

	log.Printf("exporting maze to: %v", path)
	if path == "" {
		return errors.New("path to file is required!")
	}

	w, h, err := r.GetRendererOutputSize()
	if err != nil {
		return err
	}

	s, err := sdl.CreateRGBSurface(0, int32(w), int32(h), 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0xff000000)
	if err != nil {
		return err
	}

	pixelFormat, err := window.GetPixelFormat()
	if err != nil {
		return err
	}
	pixels := s.Pixels()
	if err := r.ReadPixels(nil, pixelFormat, unsafe.Pointer(&pixels), int(s.Pitch)); err != nil {
		return err
	}

	img.SavePNG(s, path)
	s.Free()
	return nil
}

// showMazeStats shows some states about the maze
func showMazeStats(m *maze.Maze) {
	x, y := m.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(m.DeadEnds()))
}

// Solve runs the solvers against the grid.
func Solve(m *maze.Maze, keyInput chan string) (solvealgos.Algorithmer, error) {
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
	m, err = solver.Solve(m, fromCell, toCell, delay, keyInput)
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

	// must be run like this to keep drawing functions in main thread
	sdl.Main(run)
}

func run() {
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

	if *allowWeaving && *wallSpace == 0 {
		// weaving requires some wall space to look nice
		*wallSpace = 4
		log.Printf("weaving enabled, setting wall_space to non-zero value (%d)", *wallSpace)

	}

	if *showDistanceColors && *bgColor == "white" {
		log.Print("Setting bgcolor to 'black' and adjusting wall color since distance colors don't work with white right now.")
		*bgColor = "black"
		if *wallColor == "black" {
			*wallColor = "white"
		}
	}

	if *bgColor == "black" {
		if *wallColor == "black" {
			*wallColor = "white"
		}
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &maze.Config{
		Rows:                 *rows,
		Columns:              *columns,
		CellWidth:            *cellWidth,
		WallWidth:            *wallWidth,
		WallSpace:            *wallSpace,
		PathWidth:            *pathWidth,
		BgColor:              colors.GetColor(*bgColor),
		BorderColor:          colors.GetColor(*borderColor),
		WallColor:            colors.GetColor(*wallColor),
		PathColor:            colors.GetColor(*pathColor),
		VisitedCellColor:     colors.GetColor(*visitedCellColor),
		AllowWeaving:         *allowWeaving,
		WeavingProbability:   *weavingProbability,
		MarkVisitedCells:     *markVisitedCells,
		CurrentLocationColor: colors.GetColor(*currentLocationColor),
		AvatarImage:          *avatarImage,
		ShowDistanceValues:   *showDistanceValues,
		ShowDistanceColors:   *showDistanceColors,
		FromCellColor:        colors.GetColor(*fromCellColor),
		ToCellColor:          colors.GetColor(*toCellColor),
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

	defer func() {
		sdl.Do(func() {
			w.Destroy()
		})
	}()
	defer func() {
		sdl.Do(func() {
			r.Destroy()
		})
	}()
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////

	if !checkCreateAlgo(*createAlgo) {
		log.Fatalf("invalid create algorithm: %v", *createAlgo)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Background Music
	//////////////////////////////////////////////////////////////////////////////////////////////
	if *bgMusic != "" {

		if err := mix.Init(mix.INIT_MP3); err != nil {
			log.Fatalf("error initialing mp3: %v", err)
		}

		if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 2048); err != nil {
			log.Fatalf("cannot initialize audio: %v", err)
		}

		music, err := mix.LoadMUS(*bgMusic)
		if err != nil {
			log.Fatalf("cannot load music file %v: %v", *bgMusic, err)
		}

		music.Play(-1) // loop forever
	}

	///////////////////////////////////////////////////////////////////////////
	// Generator
	///////////////////////////////////////////////////////////////////////////
	// apply algorithm
	algo := algos.Algorithms[*createAlgo]

	delay, err := time.ParseDuration(*genDrawDelay)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Display generator while building
	generating := abool.New()
	generating.Set()
	var wd sync.WaitGroup

	wd.Add(1)
	go func() {
		log.Printf("running generator %v", *createAlgo)

		if err := algo.Apply(m, delay); err != nil {
			log.Fatalf(err.Error())
		}
		if err := algo.CheckGrid(m); err != nil {
			log.Fatalf("maze is not valid: %v", err)
		}

		if *showStats {
			showMazeStats(m)
		}

		// braid if requested
		if *braidProbability > 0 {
			m.Braid(*braidProbability)
		}

		if *showStats {
			showMazeStats(m)
		}

		//for x := 0; x < *columns; x++ {
		//	if x == *columns-1 {
		//		continue
		//	}
		//	c, _ := m.Cell(x, *rows/2)
		//	c.SetWeight(1000)
		//}

		if *fromCellStr != "" {
			if *fromCellStr == "min" {
				fromCell = m.SmallestCell()
			} else {
				from := strings.Split(*fromCellStr, ",")
				if len(from) != 2 {
					log.Fatalf("%v is not a valid coordinate", *fromCellStr)
				}
				x, _ := strconv.Atoi(from[0])
				y, _ := strconv.Atoi(from[1])
				fromCell, err = m.Cell(x, y, 0)
				if err != nil {
					log.Fatalf("invalid fromCell: %v", err)
				}
			}
		}

		if *toCellStr != "" {
			var x, y int
			if *toCellStr == "max" {
				toCell = m.LargestCell()
			} else {
				from := strings.Split(*toCellStr, ",")
				if len(from) != 2 {
					log.Fatalf("%v is not a valid coordinate", *toCellStr)
				}
				x, _ = strconv.Atoi(from[0])
				y, _ = strconv.Atoi(from[1])
				toCell, err = m.Cell(x, y, 0)
				if err != nil {
					log.Fatalf("invalid toCell: %v", err)
				}
			}
		}

		if *randomFromTo {
			if fromCell == nil {
				fromCell = m.RandomCell()
			}
			if toCell == nil {
				toCell = m.RandomCell()
			}
		}

		// solve the longest path
		if fromCell == nil || toCell == nil {
			log.Print("No fromCella and/or toCell set, defaulting to longestPath.")
			_, fromCell, toCell, _ = m.LongestPath()
		}

		log.Printf("Path: %v -> %v", fromCell, toCell)

		m.SetDistanceInfo(fromCell)

		generating.UnSet()
		wd.Done()
	}()

	if *showGUI {
		for generating.IsSet() {
			// Displays the main maze while generating it
			sdl.Do(func() {
				// reset the clear color back to white
				colors.SetDrawColor(colors.GetColor("white"), r)

				r.Clear()
				m.DrawMazeBackground(r)
				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
			})
		}
	}
	wd.Wait()

	if *showFromToColors || *solveAlgo != "" {
		// Set the colors for the from and to cells
		m.SetFromToColors(fromCell, toCell)
	}
	///////////////////////////////////////////////////////////////////////////
	// End Generator
	///////////////////////////////////////////////////////////////////////////

	///////////////////////////////////////////////////////////////////////////
	// Solvers
	///////////////////////////////////////////////////////////////////////////
	runSolver := abool.New()

	// channel to pass key presses into solvers
	keyInput := make(chan string)

	wd.Add(1)
	go func() {
		for !runSolver.IsSet() {
			log.Println("Maze not yet ready, sleeping 1s...")
			time.Sleep(time.Second)
		}

		if *solveAlgo != "" {
			solver, err = Solve(m, keyInput)
			if err != nil {
				log.Print(err)
			}
		}

		// Save picture of solved maze
		if *exportFile != "" {
			if err := SaveImage(r, w, *exportFile); err != nil {
				log.Printf("error exporting image: %v", err.Error())
			}
		}

		wd.Done()
	}()
	///////////////////////////////////////////////////////////////////////////
	// End Solvers
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
		running := abool.New()
		running.Set()

		// create background texture, it is saved and re-rendered as a picture
		mTexture, err := r.CreateTexture(sdl.PIXELFORMAT_RGB24, sdl.TEXTUREACCESS_TARGET, winWidth, winHeight)
		if err != nil {
			log.Fatalf("failed to create background: %v", err)
		}

		// draw on the texture
		sdl.Do(func() {
			r.SetRenderTarget(mTexture)
			// background is black so that transparency works
			colors.SetDrawColor(colors.GetColor("white"), r)
			r.Clear()
		})
		m.DrawMazeBackground(r)
		sdl.Do(func() {
			r.Present()
		})

		// Reset to drawing on the screen
		sdl.Do(func() {
			r.SetRenderTarget(nil)
			r.Copy(mTexture, nil, nil)
			r.Present()
		})

		// Allow solver to start
		runSolver.Set()

		for running.IsSet() {
			sdl.Do(func() {
				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch event.(type) {
					case *sdl.QuitEvent:
						running.UnSet()

					case *sdl.KeyDownEvent:
						// use event.(*sdl.KeyDownEvent).Keysym.Sym
						// arrow keys are
						// sdl.SDLK_DOWN; down
						// sdl.SDLK_UP; up
						// sdl.SDLK_RIGHT; right
						// sdl.SDLK_LEFT; left
						key := sdl.GetKeyName(event.(*sdl.KeyDownEvent).Keysym.Sym)
						keyInput <- key
					}
				}
			})

			// Displays the main maze, no paths or other markers
			sdl.Do(func() {
				// reset the clear color back to white
				// but it doesn't matter, as background texture takes up the entire view
				colors.SetDrawColor(colors.GetColor("black"), r)

				r.Clear()
				m.DrawMaze(r, mTexture)

				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
			})
		}
	} else {
		// wait for solver thread here, used if gui not shown
		runSolver.Set()
		wd.Wait()
	}

	//// Save to file
	if *exportFile != "" {
		log.Printf("saving image to: %v", *exportFile)
		if err := SaveImage(r, w, *exportFile); err != nil {
			log.Printf("error saving file: %v", err)
		}
	}
}
