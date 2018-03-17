// Package main runs the server component that creates the mazes and shows the GUI
package main

import (
	"bufio"
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"mazes/algos"
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
	showStats        = flag.Bool("maze_stats", false, "show maze stats")
	enableMonitoring = flag.Bool("enable_monitoring", false, "enable monitoring")

	// maze
	maskImage          = flag.String("mask_image", "", "file name of mask image")
	allowWeaving       = flag.Bool("weaving", false, "allow weaving")
	weavingProbability = flag.Float64("weaving_probability", 1, "controls the amount of weaving that happens, with 1 being the max")
	braidProbability   = flag.Float64("braid_probability", 0, "braid the maze with this probabily, 0 results in a perfect maze, 1 results in no deadends at all")
	randomFromTo       = flag.Bool("random_path", false, "show a random path through the maze")
	showGUI            = flag.Bool("gui", true, "show gui maze")
	showLocalGUI       = flag.Bool("local_gui", false, "show client's view of the maze")
	title              = flag.String("title", "", "maze title")

	// dimensions
	rows    = flag.Int64("r", 15, "number of rows in the maze")
	columns = flag.Int64("c", 15, "number of rows in the maze")

	// colors
	bgColor              = flag.String("bgcolor", "white", "background color")
	borderColor          = flag.String("border_color", "black", "border color")
	currentLocationColor = flag.String("location_color", "", "current location color, if empty, path color is used")
	pathColor            = flag.String("path_color", "red", "border color")
	visitedCellColor     = flag.String("visited_color", "red", "color of visited cell marker")
	wallColor            = flag.String("wall_color", "black", "wall color")
	fromCellColor        = flag.String("from_cell_color", "", "from cell color, based on path if empty")
	toCellColor          = flag.String("to_cell_color", "", "to cell color, based on path if empty")

	// width
	cellWidth = flag.Int64("w", 30, "cell width (best as multiple of 2)")
	pathWidth = flag.Int64("path_width", 2, "path width")
	wallWidth = flag.Int64("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	wallSpace = flag.Int64("wall_space", 0, "how much space between two side by side walls (min of 2)")

	// display
	avatarImage            = flag.String("avatar_image", "", "file name of avatar image, the avatar should be facing to the left in the image")
	genDrawDelay           = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	markVisitedCells       = flag.Bool("mark_visited", false, "mark visited cells (by solver) with a properly sized square")
	numberMarkVisitedCells = flag.Bool("mark_visited_number", false, "mark visited cells (by solver) with a number")
	showFromToColors       = flag.Bool("show_from_to_colors", false, "show from/to colors")
	showDistanceColors     = flag.Bool("show_distance_colors", false, "show distance colors")
	showDistanceValues     = flag.Bool("show_distance_values", false, "show distance values")
	showWeightValues       = flag.Bool("show_weight_values", false, "show weight values")
	drawPathLength         = flag.Int64("draw_path_length", -1, "draw client path length, -1 = all, 0 = none")
	solveDrawDelay         = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")
	frameRate              = flag.Uint("frame_rate", 120, "frame rate for animation")
	delayMs                = flag.Uint("delay_ms", 500, "delay in milliseconds between updates")

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	solveAlgo     = flag.String("solve_algo", "recursive-backtracker", "algorithm to solve the maze")
	skipGridCheck = flag.Bool("skip_grid_check", true, "set to true to skip grid check (disable spanning tree check)")

	// solver
	mazeID = flag.String("maze_id", "", "maze id")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	wd sync.WaitGroup
)

// Send returns true if it was able to send t on channel c.
// It returns false if c is closed.
// This isn't great, but for simplicity here.
//func Send(c chan commChannel, t string) (ok bool) {
//	defer func() { recover() }()
//	c <- t
//	return true
//}

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

	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	if config.AllowWeaving && config.WallSpace == 0 {
		// weaving requires some wall space to look nice
		config.WallSpace = 4
		log.Printf("weaving enabled, setting wall_space to non-zero value (%d)", config.WallSpace)

	}

	if config.ShowDistanceColors && config.BgColor == "white" {
		config.BgColor = "black"
		if config.WallColor == "black" {
			config.WallColor = "white"
		}
		log.Printf("Setting bgcolor to %v and adjusting wall color to %v since distance colors don't work with white right now.", config.BgColor, config.WallColor)

	}

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
	// TODO(dan): redo error return as a channel to catch problems here
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
func runMaze(m *maze.Maze, r *sdl.Renderer, w *sdl.Window, comm chan commandData) {
	defer func() {
		sdl.Do(func() {
			r.Destroy()
			w.Destroy()
		})
	}()
	var wd sync.WaitGroup

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// this is the main maze thread that draws the maze and interacts with it via comm
	running := abool.New()
	running.Set()

	// when this is set to true, a redraw of the background texture is triggered
	updateBG := abool.New()

	if m.Config().GetGui() {
		// create background texture, it is saved and re-rendered as a picture
		mTexture, err := m.MakeBGTexture()
		if err != nil {
			log.Fatalf("failed to create background: %v", err)
		}
		m.SetBGTexture(mTexture)
	}

	showMazeStats(m)

	// process background events in the maze
	ticker := time.NewTicker(time.Millisecond * time.Duration(*delayMs))
	go func() {
		for range ticker.C {
			processMazeEvents(m, r, updateBG)
			if !running.IsSet() {
				return
			}
		}
	}()

	// main loop processing and updating the maze
	for running.IsSet() {
		start := time.Now()
		t := metrics.GetOrRegisterTimer("maze.loop.latency", nil)

		lsdl.CheckQuit(running)
		updateMazeBackground(m, updateBG)
		displaymaze(m, r)

		t.UpdateSince(start)
	}

	wd.Wait()
}

// processMazeEvents takes care of periodic events happening in the maze
// we defer setting the colors so that they all get processed at once
func processMazeEvents(m *maze.Maze, r *sdl.Renderer, updateBG *abool.AtomicBool) {
	log.Printf("[%v] processing events...", time.Now())

	for c := range m.Cells() {
		liveNeighbors := 0
		for _, n := range c.AllNeighbors() {
			if n.BGColor() == colors.GetColor("black") {
				liveNeighbors++
			}
		}

		if liveNeighbors < 2 {
			// die, lonely
			defer c.SetBGColor(colors.GetColor("white"))
			continue
		}

		if liveNeighbors > 3 {
			// die, overcrowded
			defer c.SetBGColor(colors.GetColor("white"))
			continue
		}

		if liveNeighbors == 3 && c.BGColor() == colors.GetColor("white") {
			// Dead cell with 3 live neighbors becomes alive
			defer c.SetBGColor(colors.GetColor("black"))

		}
	}

	// redraw the background
	updateBG.Set()

}

func updateMazeBackground(m *maze.Maze, updateBG *abool.AtomicBool) {
	if updateBG.IsSet() {
		if m.Config().GetGui() {
			log.Printf("setting background")
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

func runServer() {
	config := &pb.MazeConfig{
		Rows:               *rows,
		Columns:            *columns,
		AllowWeaving:       *allowWeaving,
		WeavingProbability: *weavingProbability,
		CellWidth:          *cellWidth,
		WallWidth:          *wallWidth,
		WallSpace:          *wallSpace,
		WallColor:          *wallColor,
		PathWidth:          *pathWidth,
		ShowDistanceColors: *showDistanceColors,
		ShowDistanceValues: *showDistanceValues,
		ShowWeightValues:   *showWeightValues,
		SkipGridCheck:      *skipGridCheck,
		GenDrawDelay:       *genDrawDelay,
		BgColor:            *bgColor,
		BorderColor:        *borderColor,
		CreateAlgo:         *createAlgo,
		BraidProbability:   *braidProbability,
		Gui:                *showGUI,
		FromFile:           *mazeID,
		Title:              *title,
	}
	CreateMaze(config)

	log.Print("presee <enter> to quit...")
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

	// run http server for expvars
	sock, err := net.Listen("tcp", "localhost:8123")
	if err != nil {
		log.Fatalf(err.Error())
	}
	go func() {
		fmt.Println("metrics now available at http://localhost:8123/debug/metrics")
		http.Serve(sock, nil)
	}()

	// must be like this to keep drawing functions in main thread
	sdl.Main(runServer)

}

type commChannel chan commandData

type commandAction int

// request parameters sent in
type commandRequest struct {
	request interface{}
}

type commandReply struct {
	answer interface{}
	error  error
}

type commandData struct {
	Action       commandAction
	ClientConfig *pb.ClientConfig
	ClientID     string
	Request      commandRequest
	Reply        chan commandReply // reply from the maze is sent over this channel
}

type locationInfo struct {
	current *pb.MazeLocation
	From    *pb.MazeLocation
	To      *pb.MazeLocation
}

type moveReply struct {
	current             *pb.MazeLocation
	availableDirections []*pb.Direction
	solved              bool
	reward              float64
}

// server is used to implement MazerServer.
type server struct{}

// setInitialStates sets the initial state of the squares
func setInitialStates(m *maze.Maze) *maze.Maze {
	// maxX, maxY := m.Dimensions()

	liveCells := []*maze.Cell{
		// block
		m.CellBeSure(1, 1, 0),
		m.CellBeSure(1, 2, 0),
		m.CellBeSure(2, 1, 0),
		m.CellBeSure(2, 2, 0),
		// line
		m.CellBeSure(5, 4, 0),
		m.CellBeSure(5, 5, 0),
		m.CellBeSure(5, 6, 0),
		// two diagonal blocks
		m.CellBeSure(8, 1, 0),
		m.CellBeSure(8, 2, 0),
		m.CellBeSure(9, 1, 0),
		m.CellBeSure(9, 2, 0),
		m.CellBeSure(10, 3, 0),
		m.CellBeSure(10, 4, 0),
		m.CellBeSure(11, 3, 0),
		m.CellBeSure(11, 4, 0),
	}

	for _, c := range liveCells {
		c.SetBGColor(colors.GetColor("black"))
	}
	return m

}

// CreateMaze creates and displays the maze specified by the config
func CreateMaze(config *pb.MazeConfig) error {
	log.Printf("creating maze with config: %#v", config)
	if config == nil {
		return fmt.Errorf("maze config cannot be nil")
	}

	t := metrics.GetOrRegisterTimer("maze.rpc.create-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	var mazeID string
	mazeID = uuid.NewV4().String()
	config.Id = mazeID

	comm := make(chan commandData)

	m, r, w, err := createMaze(config)
	if err != nil {
		return err
	}

	m = setInitialStates(m)

	log.Print("running maze...")
	go runMaze(m, r, w, comm)

	return nil
}
