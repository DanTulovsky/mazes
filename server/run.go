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
)

const (
	port = ":50051"
)

// For gui support
// brew install sdl2{_image,_ttf,_gfx}
// brew install sdl2_mixer --with-flac --with-fluid-synth --with-libmikmod --with-libmodplug --with-libvorbis --with-smpeg2
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}
// if slow compile, run: go install -a mazes/generate
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

// for proto: protoc -I ./mazes/proto/ ./mazes/proto/mazes.proto --go_out=plugins=grpc:mazes/proto/
// protoc -I ./proto/ ./proto/mazes.proto --go_out=plugins=grpc:proto/

var (
	winTitle         string = "Maze"
	fromCell, toCell *maze.Cell

	w      *sdl.Window
	r      *sdl.Renderer
	sdlErr error
	// runningMutex sync.Mutex

	// maze
	maskImage          = flag.String("mask_image", "", "file name of mask image")
	allowWeaving       = flag.Bool("weaving", false, "allow weaving")
	weavingProbability = flag.Float64("weaving_probability", 1, "controls the amount of weaving that happens, with 1 being the max")
	braidProbability   = flag.Float64("braid_probability", 0, "braid the maze with this probabily, 0 results in a perfect maze, 1 results in no deadends at all")
	randomFromTo       = flag.Bool("random_path", false, "show a random path through the maze")

	// dimensions
	rows    = flag.Int64("r", 30, "number of rows in the maze")
	columns = flag.Int64("c", 60, "number of rows in the maze")

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
	cellWidth = flag.Int64("w", 20, "cell width (best as multiple of 2)")
	pathWidth = flag.Int64("path_width", 2, "path width")
	wallWidth = flag.Int64("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	wallSpace = flag.Int64("wall_space", 0, "how much space between two side by side walls (min of 2)")

	// maze draw
	showGUI = flag.Bool("gui", true, "show gui maze")

	// display
	avatarImage        = flag.String("avatar_image", "", "file name of avatar image, the avatar should be facing to the left in the image")
	frameRate          = flag.Uint("frame_rate", 120, "frame rate for animation")
	genDrawDelay       = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	markVisitedCells   = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	showFromToColors   = flag.Bool("show_from_to_colors", false, "show from/to colors")
	showDistanceColors = flag.Bool("show_distance_colors", false, "show distance colors")
	showDistanceValues = flag.Bool("show_distance_values", false, "show distance values")

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	skipGridCheck = flag.Bool("skip_grid_check", false, "set to true to skip grid check (disable spanning tree check)")

	// misc
	bgMusic = flag.String("bg_music", "", "file name of background music to play")

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
	winWidth = int((*columns)**cellWidth + *wallWidth*2)
	winHeight = int((*rows)**cellWidth + *wallWidth*2)

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
				os.Exit(0)
			}
		}

	})
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
		*bgColor = "black"
		if *wallColor == "black" {
			*wallColor = "white"
		}
		log.Printf("Setting bgcolor to %v and adjusting wall color to %v since distance colors don't work with white right now.", *bgColor, *wallColor)

	}

	if *bgColor == "black" {
		if *wallColor == "black" {
			*wallColor = "white"
		}
	}

	if *cellWidth == 2 && *wallWidth == 2 {

		*wallWidth = 1
		log.Printf("cell_width and wall_width both 2, adjusting wall_width to %v", *wallWidth)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &pb.MazeConfig{
		Rows:                 *rows,
		Columns:              *columns,
		CellWidth:            *cellWidth,
		WallWidth:            *wallWidth,
		WallSpace:            *wallSpace,
		PathWidth:            *pathWidth,
		SkipGridCheck:        *skipGridCheck,
		BgColor:              *bgColor,
		BorderColor:          *borderColor,
		WallColor:            *wallColor,
		PathColor:            *pathColor,
		VisitedCellColor:     *visitedCellColor,
		AllowWeaving:         *allowWeaving,
		WeavingProbability:   *weavingProbability,
		MarkVisitedCells:     *markVisitedCells,
		CurrentLocationColor: *currentLocationColor,
		AvatarImage:          *avatarImage,
		ShowDistanceValues:   *showDistanceValues,
		ShowDistanceColors:   *showDistanceColors,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
	}

	//config := &maze.Config{
	//	Rows:                 *rows,
	//	Columns:              *columns,
	//	CellWidth:            *cellWidth,
	//	WallWidth:            *wallWidth,
	//	WallSpace:            *wallSpace,
	//	PathWidth:            *pathWidth,
	//	SkipGridCheck:        *skipGridCheck,
	//	BgColor:              colors.GetColor(*bgColor),
	//	BorderColor:          colors.GetColor(*borderColor),
	//	WallColor:            colors.GetColor(*wallColor),
	//	PathColor:            colors.GetColor(*pathColor),
	//	VisitedCellColor:     colors.GetColor(*visitedCellColor),
	//	AllowWeaving:         *allowWeaving,
	//	WeavingProbability:   *weavingProbability,
	//	MarkVisitedCells:     *markVisitedCells,
	//	CurrentLocationColor: colors.GetColor(*currentLocationColor),
	//	AvatarImage:          *avatarImage,
	//	ShowDistanceValues:   *showDistanceValues,
	//	ShowDistanceColors:   *showDistanceColors,
	//	FromCellColor:        colors.GetColor(*fromCellColor),
	//	ToCellColor:          colors.GetColor(*toCellColor),
	//}

	var m *maze.Maze
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
		*columns, *rows = m.Dimensions()
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
	setupSDL()

	defer func() {
		sdl.Do(func() {
			w.Destroy()
		})
	}()
	//defer func() {
	//	sdl.Do(func() {
	//		r.Destroy()
	//	})
	//}()
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
				x, _ := strconv.ParseInt(from[0], 10, 64)
				y, _ := strconv.ParseInt(from[1], 10, 64)
				fromCell, err = m.Cell(x, y, 0)
				if err != nil {
					log.Fatalf("invalid fromCell: %v", err)
				}
			}
		}

		if *toCellStr != "" {
			var x, y int64
			if *toCellStr == "max" {
				toCell = m.LargestCell()
			} else {
				from := strings.Split(*toCellStr, ",")
				if len(from) != 2 {
					log.Fatalf("%v is not a valid coordinate", *toCellStr)
				}
				x, _ = strconv.ParseInt(from[0], 10, 64)
				y, _ = strconv.ParseInt(from[1], 10, 64)
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

	if *showFromToColors {
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
	if *showGUI {
		// wd.Add(1)
		go func() {
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
		}()

		showMazeStats(m)
		log.Print("server ready...")

		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		s := grpc.NewServer()
		pb.RegisterMazerServer(s, &server{})
		// Register reflection service on gRPC server.
		reflection.Register(s)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}

		// wd.Wait()
	}

}

func main() {
	// must be run like this to keep drawing functions in main thread
	sdl.Main(run)
}

// server is used to implement MazerServer.
type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

// ShowMaze displays the maze specified by the config
func (s *server) ShowMaze(ctx context.Context, in *pb.ShowMazeRequest) (*pb.ShowMazeReply, error) {
	return &pb.ShowMazeReply{}, nil
}
