// Package main runs the server component that creates the mazes and shows the GUI
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/profile"
	"github.com/sasha-s/go-deadlock"
	"github.com/satori/go.uuid"
	"github.com/tevino/abool"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_mixer"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"mazes/algos"
	"mazes/colors"
	"mazes/maze"
	pb "mazes/proto"
	"safemap"
)

const (
	port = ":50051"
)

// For gui support
// brew install sdl2{_image,_ttf,_gfx}
// brew install sdl2_mixer --with-flac --with-fluid-synth --with-libmikmod --with-libmodplug --with-libvorbis --with-smpeg2
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, showMaze: go install -a mazes/generate
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

// for proto: protoc -I ./mazes/proto/ ./mazes/proto/mazes.proto --go_out=plugins=grpc:mazes/proto/
// protoc -I ./proto/ ./proto/mazes.proto --go_out=plugins=grpc:proto/

var (
	winTitle string = "Maze"

	sdlErr error

	// maze
	maskImage        = flag.String("mask_image", "", "file name of mask image")
	braidProbability = flag.Float64("braid_probability", 0, "braid the maze with this probabily, 0 results in a perfect maze, 1 results in no deadends at all")

	// maze draw
	showGUI = flag.Bool("gui", true, "show gui maze")

	// display
	frameRate        = flag.Uint("frame_rate", 120, "frame rate for animation")
	showFromToColors = flag.Bool("show_from_to_colors", false, "show from/to colors")

	// misc
	bgMusic = flag.String("bg_music", "", "file name of background music to play")

	// stats
	showStats = flag.Bool("stats", false, "show maze stats")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	winWidth, winHeight int

	// keep track of mazes
	mazeMap = *safemap.NewSafeMap()
)

func setupSDL(config *pb.MazeConfig, w *sdl.Window, r *sdl.Renderer) (*sdl.Window, *sdl.Renderer) {
	if !*showGUI {
		return nil, nil
	}
	sdl.Do(func() {
		sdl.Init(sdl.INIT_EVERYTHING)
		sdl.EnableScreenSaver()
	})

	// window
	winWidth = int((config.Columns)*config.CellWidth + config.WallWidth*2)
	winHeight = int((config.Rows)*config.CellWidth + config.WallWidth*2)

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

	return w, r
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

// showMazeStats shows some states about the maze
func showMazeStats(m *maze.Maze) {
	x, y := m.Dimensions()
	log.Printf(">> Dimensions: [%v, %v]", x, y)
	log.Printf(">> Dead Ends: %v", len(m.DeadEnds()))
}

func checkQuit(running *abool.AtomicBool) {
	sdl.Do(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				log.Print("received quit request, exiting...")
				running.UnSet()
			}
		}

	})
}

func configToCell(m *maze.Maze, config *pb.MazeConfig, c string) (*maze.Cell, error) {

	switch c {
	case "min":
		return m.SmallestCell(), nil
	case "max":
		return m.LargestCell(), nil
	case "random":
		return m.RandomCell(), nil
	default:
		from := strings.Split(c, ",")
		if len(from) != 2 {
			log.Fatalf("%v is not a valid coordinate", config.FromCell)
		}
		x, _ := strconv.ParseInt(from[0], 10, 64)
		y, _ := strconv.ParseInt(from[1], 10, 64)
		cell, err := m.Cell(x, y, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid fromCell: %v", err)
		}
		return cell, nil
	}

}

func showMaze(config *pb.MazeConfig, comm chan commChannel) {
	var (
		w *sdl.Window
		r *sdl.Renderer
	)

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

	var m *maze.Maze
	var fromCell, toCell *maze.Cell
	var err error

	// Mask image if provided.
	// If the mask image is provided, use that as the dimensions of the grid
	if *maskImage != "" {
		log.Printf("Using %v as grid mask", *maskImage)
		m, err = maze.NewMazeFromImage(config, *maskImage)
		if err != nil {
			log.Printf("invalid config: %v", err)
			os.Exit(1)
		}
		// Set these for correct window size
		config.Columns, config.Rows = m.Dimensions()
	} else {
		m, err = maze.NewMaze(config)
		if err != nil {
			log.Printf("invalid config: %v", err)
			os.Exit(1)
		}
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	w, r = setupSDL(config, w, r)

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

	if !checkCreateAlgo(config.CreateAlgo) {
		log.Fatalf("invalid create algorithm: %v", config.CreateAlgo)
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
	go func() {
		defer wd.Done()
		log.Printf("running generator %v", config.CreateAlgo)

		if err := algo.Apply(m, delay, generating); err != nil {
			log.Printf(err.Error())
			generating.UnSet()
			return
		}
		if err := algo.CheckGrid(m); err != nil {
			log.Printf("maze is not valid: %v", err)
			generating.UnSet()
			return
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

		if config.FromCell != "" {
			fromCell, err = configToCell(m, config, config.FromCell)
		}

		if config.ToCell != "" {
			toCell, err = configToCell(m, config, config.ToCell)
		}

		// solve the longest path
		if fromCell == nil || toCell == nil {
			log.Print("No fromCella and/or toCell set, defaulting to longestPath.")
			_, fromCell, toCell, _ = m.LongestPath()
		}

		log.Printf("Path: %v -> %v", fromCell, toCell)

		m.SetDistanceInfo(fromCell)

		generating.UnSet()
	}()

	if *showGUI {
		for generating.IsSet() {
			checkQuit(generating)
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

	if config.ShowFromToColors {
		// Set the colors for the from and to cells
		m.SetFromToColors(fromCell, toCell)
	}
	///////////////////////////////////////////////////////////////////////////
	// End Generator
	///////////////////////////////////////////////////////////////////////////

	mazeReady := abool.New()

	///////////////////////////////////////////////////////////////////////////
	// DISPLAY
	///////////////////////////////////////////////////////////////////////////
	// gui maze
	if !*showGUI {
		return
	}

	wd.Add(1)
	go func(r *sdl.Renderer) {
		defer wd.Done()
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

		// Allow clients to connect
		mazeReady.Set()

		for running.IsSet() {
			checkQuit(running)

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
		mazeMap.Delete(m.Config().GetId())

		log.Printf("maze is done...")
	}(r)

	showMazeStats(m)
	wd.Wait()
}

func runServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterMazerServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("server ready on port %v", port)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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

type commChannel chan commandData

type commandAction int

type commandData struct {
	action commandAction
	key    string
}

// server is used to implement MazerServer.
type server struct{}

// CreateMaze creates and displays the maze specified by the config
func (s *server) CreateMaze(ctx context.Context, in *pb.CreateMazeRequest) (*pb.CreateMazeReply, error) {

	log.Printf("running maze with config: %#v", in.Config)

	id := uuid.NewV4().String()
	in.Config.Id = id

	comm := make(chan commChannel)
	mazeMap.Insert(id, comm)

	go showMaze(in.Config, comm)

	return &pb.CreateMazeReply{MazeId: id}, nil
}

// ListMazes lists all the mazes
func (s *server) ListMazes(ctx context.Context, in *pb.ListMazeRequest) (*pb.ListMazeReply, error) {
	keys := []string{}
	for _, k := range mazeMap.Keys() {
		keys = append(keys, k)
	}
	return &pb.ListMazeReply{keys}, nil
}
