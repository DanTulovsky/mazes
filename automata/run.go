// Package main runs the server component that creates the mazes and shows the GUI
package main

import (
	"bufio"
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"mazes/algos"
	"mazes/automata/rules"
	"mazes/automata/states"
	"mazes/colors"
	"mazes/maze"
	pb "mazes/proto"
	lsdl "mazes/sdl"

	"mazes/genalgos/fromfile"

	"github.com/pkg/profile"
	"github.com/rcrowley/go-metrics"
	"github.com/sasha-s/go-deadlock"
	"github.com/satori/go.uuid"
	"github.com/tevino/abool"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	port = ":50051"
)

// For gui support
// brew install sdl2{_image,_ttf,_gfx}
// brew install sdl2_mixer --with-flac --with-fluid-synth --with-libmikmod --with-libmodplug --with-libvorbis --with-smpeg2
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, runMaze: go install -a mazes/server mazes/client
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

// brew install protobuf
// go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
// for proto: protoc -I ./mazes/proto/ ./mazes/proto/mazes.proto --go_out=plugins=grpc:mazes/proto/
//   protoc -I ./proto/ ./proto/mazes.proto --go_out=plugins=grpc:proto/
// python:
//   cd ~/python/src
//   python -m grpc_tools.protoc -I../../go/src/mazes/proto --python_out=mazes/protos/ --grpc_python_out=mazes/protos/ ../../go/src/mazes/proto/mazes.proto

var (
	// stats
	showStats = flag.Bool("maze_stats", false, "show maze stats")

	// maze
	maskImage = flag.String("mask_image", "", "file name of mask image")
	title     = flag.String("title", "", "maze title")

	// dimensions
	rows    = flag.Int64("r", 15, "number of rows in the maze")
	columns = flag.Int64("c", 15, "number of rows in the maze")

	// colors
	bgColor     = flag.String("bgcolor", "white", "background color")
	borderColor = flag.String("border_color", "black", "border color")
	wallColor   = flag.String("wall_color", "black", "wall color")

	// width
	cellWidth = flag.Int64("w", 30, "cell width (best as multiple of 2)")
	wallWidth = flag.Int64("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	wallSpace = flag.Int64("wall_space", 0, "how much space between two side by side walls (min of 2)")

	// display
	genDrawDelay     = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	showWeightValues = flag.Bool("show_weight_values", false, "show weight values")
	solveDrawDelay   = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")
	frameRate        = flag.Uint("frame_rate", 120, "frame rate for animation")
	delayMs          = flag.Uint("delay_ms", 100, "delay in milliseconds between updates (how fast each turn is)")

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	skipGridCheck = flag.Bool("skip_grid_check", true, "set to true to skip grid check (disable spanning tree check)")

	// solver
	mazeID = flag.String("maze_id", "", "maze id")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	wd sync.WaitGroup
)

// ResetFontCache is required to be able to call gfx.* functions on multiple windows.
func ResetFontCache() {
	gfx.SetFont(nil, 0, 0)
}

// showMazeStats shows some states about the maze
func showMazeStats(m *maze.Maze) {
	x, y := m.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(m.DeadEnds()))
}

func createMaze(config *pb.MazeConfig) (m *maze.Maze, r *sdl.Renderer, w *sdl.Window, err error) {
	if config.GetCreateAlgo() == "fromfile" {
		if c, r, err := fromfile.MazeSizeFromFile(config); err == nil {
			config.Columns, config.Rows = int64(c), int64(r)
		} else {
			return nil, nil, nil, err
		}
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	title := fmt.Sprintf("Server: %v", config.GetTitle())
	w, r = lsdl.SetupSDL(config, title, 0, 0)
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////

	if config.BgColor == "black" {
		if config.WallColor == "black" {
			config.WallColor = "white"
		}
	}

	if config.CellWidth == 2 && config.WallWidth == 2 {
		config.WallWidth = 1
		log.Printf("cell_width and wall_width both 2, adjusting wall_width to %v", config.WallWidth)
	}

	// Mask image if provided.
	// If the mask image is provided, use that as the dimensions of the grid
	if *maskImage != "" {
		log.Printf("Using %v as grid mask", *maskImage)
		m, err = maze.NewMazeFromImage(config, *maskImage, r)
		if err != nil {
			log.Printf("invalid config: %v", err)
			os.Exit(1)
		}
		// Set these for correct window size
		config.Columns, config.Rows = m.Dimensions()
	} else {
		m, err = maze.NewMaze(config, r)
		if err != nil {
			log.Printf("invalid config: %v", err)
			os.Exit(1)
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	if !algos.CheckCreateAlgo(config.CreateAlgo) {
		return nil, nil, nil, fmt.Errorf("invalid create algorithm: %v", config.CreateAlgo)
	}

	///////////////////////////////////////////////////////////////////////////
	// Generator
	///////////////////////////////////////////////////////////////////////////
	// apply algorithm
	algo := algos.Algorithms[config.CreateAlgo]

	delay, err := time.ParseDuration(config.GenDrawDelay)
	if err != nil {
		log.Printf(err.Error())
	}

	// Display generator while building
	generating := abool.New()
	generating.Set()
	var wd sync.WaitGroup

	wd.Add(1)
	generate := func() error {
		defer wd.Done()
		log.Printf("running generator %v", config.CreateAlgo)

		if err := algo.Apply(m, delay, generating); err != nil {
			log.Printf(err.Error())
			generating.UnSet()
			return fmt.Errorf("error applying algorithm: %v", err)
		}
		if err := algo.CheckGrid(m); err != nil {
			generating.UnSet()
			return fmt.Errorf("maze is not valid: %v", err)
		}

		if *showStats {
			showMazeStats(m)
		}

		// braid if requested
		if m.Config().GetBraidProbability() > 0 {
			m.Braid(m.Config().GetBraidProbability())
		}

		if *showStats {
			showMazeStats(m)
		}

		generating.UnSet()
		return nil
	}
	go generate()

	if m.Config().GetGui() {
		for generating.IsSet() {
			lsdl.CheckQuit(generating)
			// Displays the main maze while generating it
			sdl.Do(func() {
				// reset the clear color back to white
				colors.SetDrawColor(colors.GetColor("white"), r)

				r.Clear()
				m.DrawMazeBackground(r)
				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
				ResetFontCache()
			})
		}
	}
	wd.Wait()
	///////////////////////////////////////////////////////////////////////////
	// End Generator
	///////////////////////////////////////////////////////////////////////////
	log.Printf("finished creating maze...")

	encoded, err := m.Encode()
	if err != nil {
		encoded = err.Error()
	}

	if m.Config().GetReturnMaze() {
		m.SetEncodedString(encoded)
	}

	return m, r, w, nil
}

// runMaze runs the maze
func runMaze(m *maze.Maze, r *sdl.Renderer, w *sdl.Window) {
	defer func() {
		sdl.Do(func() {
			r.Destroy()
			w.Destroy()
		})
	}()

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// this is the main maze thread that draws the maze

	// when this is set to true, a redraw of the background texture is triggered
	updateBG := abool.New()

	// create background texture, it is saved and re-rendered as a picture
	mTexture, err := m.MakeBGTexture()
	if err != nil {
		log.Fatalf("failed to create background: %v", err)
	}
	m.SetBGTexture(mTexture)

	// process background events in the maze
	// this is where the work happens
	ticker := time.NewTicker(time.Millisecond * time.Duration(*delayMs))
	go func() {
		for range ticker.C {
			processMazeEvents(m, r, updateBG)
		}
	}()

	// main loop re-drawing the maze
	for {
		updateMazeBackground(m, updateBG)
		displaymaze(m, r)
	}
}

func updateMazeBackground(m *maze.Maze, updateBG *abool.AtomicBool) {
	if updateBG.IsSet() {
		if m.Config().GetGui() {
			mTexture, err := m.MakeBGTexture()
			if err != nil {
				log.Fatalf("failed to create background: %v", err)
			}
			m.SetBGTexture(mTexture)
		}
		updateBG.UnSet()
	}
}

// displaymaze draws the maze
func displaymaze(m *maze.Maze, r *sdl.Renderer) {
	if m.Config().GetGui() {
		// Displays the maze
		sdl.Do(func() {
			if err := r.Clear(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to clear: %s\n", err)
				os.Exit(1)
			}

			m.DrawMaze(r, m.BGTexture())

			r.Present()
			sdl.Delay(uint32(1000 / *frameRate))
			ResetFontCache()

		})
	}
}

// setInitialStates sets the initial state of the squares
func setInitialStates(m *maze.Maze) *maze.Maze {
	//m = states.Concrete(m)
	m = states.Random(m)
	return m
}

// processMazeEvents takes care of periodic events happening in the maze
// we defer setting the colors so that they all get processed at once
func processMazeEvents(m *maze.Maze, r *sdl.Renderer, updateBG *abool.AtomicBool) {

	// Use classic rules; http://web.stanford.edu/~cdebs/GameOfLife/
	rules.Classic(m)

	// Use Play1 rules; random chance cell with 0 neighbors to come alive
	// rules.Play1(m)

	// Use Play2 rules; experimental
	// rules.Play2(m)

	// redraw the background
	updateBG.Set()

}

// CreateMaze creates and displays the maze specified by the config
func CreateMaze(config *pb.MazeConfig) error {
	log.Printf("creating maze with config: %#v", config)
	if config == nil {
		return fmt.Errorf("maze config cannot be nil")
	}

	t := metrics.GetOrRegisterTimer("maze.rpc.create-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	config.Id = uuid.NewV4().String()

	m, r, w, err := createMaze(config)
	if err != nil {
		return err
	}

	m = setInitialStates(m)

	log.Print("running maze...")
	go runMaze(m, r, w)

	return nil
}

func runServer() {
	config := &pb.MazeConfig{
		Rows:             *rows,
		Columns:          *columns,
		CellWidth:        *cellWidth,
		WallWidth:        *wallWidth,
		WallSpace:        *wallSpace,
		WallColor:        *wallColor,
		ShowWeightValues: *showWeightValues,
		SkipGridCheck:    *skipGridCheck,
		GenDrawDelay:     *genDrawDelay,
		BgColor:          *bgColor,
		BorderColor:      *borderColor,
		CreateAlgo:       *createAlgo,
		Gui:              true,
		FromFile:         *mazeID,
		Title:            *title,
	}
	CreateMaze(config)

	log.Print("presee <ctr>+c to quit...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func main() {
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

	// must be like this to keep drawing functions in main thread
	sdl.Main(runServer)

}
