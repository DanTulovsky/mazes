// Package main runs the server component that creates the mazes and shows the GUI
package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/pkg/profile"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
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
// if slow compile, createMaze: go install -a mazes/generate
// for tests: go get -u gopkg.in/check.v1
// https://blog.jetbrains.com/idea/2015/08/experimental-zero-latency-typing-in-intellij-idea-15-eap/

// for proto: protoc -I ./mazes/proto/ ./mazes/proto/mazes.proto --go_out=plugins=grpc:mazes/proto/
// protoc -I ./proto/ ./proto/mazes.proto --go_out=plugins=grpc:proto/

var (
	winTitle string = "Maze"

	sdlErr error

	// maze
	maskImage = flag.String("mask_image", "", "file name of mask image")

	// maze draw
	showGUI = flag.Bool("gui", true, "show gui maze")

	// display
	frameRate = flag.Uint("frame_rate", 120, "frame rate for animation")

	// misc
	bgMusic = flag.String("bg_music", "", "file name of background music to play")

	// stats
	showStats = flag.Bool("maze_stats", false, "show maze stats")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	winWidth, winHeight int

	// keep track of mazes
	mazeMap = *safemap.NewSafeMap()
)

// Send returns true if it was able to send t on channel c.
// It returns false if c is closed.
// This isn't great, but for simplicity here.
//func Send(c chan commChannel, t string) (ok bool) {
//	defer func() { recover() }()
//	c <- t
//	return true
//}

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

func createMaze(config *pb.MazeConfig, comm chan commandData) {
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

	var m *maze.Maze
	var err error

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
		if m.Config().GetBraidProbability() > 0 {
			m.Braid(m.Config().GetBraidProbability())
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

	// this is the main maze thread that draws the maze and interacts with it via comm
	wd.Add(1)
	go func(r *sdl.Renderer) {
		defer wd.Done()
		running := abool.New()
		running.Set()

		// create background texture, it is saved and re-rendered as a picture
		mTexture, err := m.MakeBGTexture()
		if err != nil {
			log.Fatalf("failed to create background: %v", err)
		}
		m.SetBGTexture(mTexture)

		// Allow clients to connect, solvers can start running
		mazeReady.Set()

		for running.IsSet() {
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.loop.latency", nil)

			checkQuit(running)

			// check for client communications, they are serialized for one maze
			checkComm(m, comm)

			// Displays the maze
			sdl.Do(func() {
				// reset the clear color back to black
				// but it doesn't matter, as background texture takes up the entire view
				colors.SetDrawColor(colors.GetColor("black"), r)

				r.Clear()
				m.DrawMaze(r, m.BGTexture())

				r.Present()
				sdl.Delay(uint32(1000 / *frameRate))
			})
			t.UpdateSince(start)
		}
		mazeMap.Delete(m.Config().GetId())

		log.Printf("maze is done...")
	}(r)

	showMazeStats(m)
	wd.Wait()
}

func checkComm(m *maze.Maze, comm commChannel) {
	select {
	case in := <-comm: // type == commandData
		switch in.Action {
		case maze.CommandListClients:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.list-clients.latency", nil)

			answer := m.Clients()
			var clients []string
			for c := range answer {
				clients = append(clients, c)
			}
			// send reply via the reply channel
			in.Reply <- commandReply{answer: clients}

			t.UpdateSince(start)
		case maze.CommandAddClient:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.add-client.latency", nil)

			m.AddClient(in.ClientID, in.ClientConfig)

			// send reply via the reply channel
			in.Reply <- commandReply{}
			t.UpdateSince(start)
		case maze.CommandGetDirections:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.get-direction.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to get client links: %v", err)}

			} else {
				in.Reply <- commandReply{answer: client.CurrentLocation().DirectionLinks(in.ClientID)}
			}
			t.UpdateSince(start)
		case maze.CommandSetInitialClientLocation:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.set-initial-client-location.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to set initial client location: %v", err)}
			} else {
				m.Reset()

				client.SetCurrentLocation(m.FromCell(client))
				cell := client.CurrentLocation()

				// Add initial location to paths
				s := maze.NewSegment(cell, "north")
				cell.SetVisited(in.ClientID)
				client.TravelPath.AddSegement(s)
				client.SolvePath.AddSegement(s)
			}
			t.UpdateSince(start)
		case maze.CommandCurrentLocation:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.current-location.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				log.Printf("failed to get client location: %v", err)
			} else {
				cell := client.CurrentLocation()
				in.Reply <- commandReply{answer: cell.Location()}
			}
			t.UpdateSince(start)
		case maze.CommandLocationInfo:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.location-info.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				log.Printf("failed to get client location: %v", err)
			} else {
				info := &locationInfo{
					current: client.CurrentLocation().Location(),
					from:    m.FromCell(client).Location(),
					to:      m.ToCell(client).Location(),
				}
				in.Reply <- commandReply{answer: info}
			}
			t.UpdateSince(start)
		case maze.CommandMoveBack:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.moveback", nil)

			client, err := m.Client(in.ClientID)
			if err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to find client: %v", err)}
			}

			// remove from SolvePath if we are backtracking
			client.SolvePath.DelSegement()

			// previous cell
			currentCell := client.CurrentLocation()
			last := client.SolvePath.LastSegment().Cell()
			client.SetCurrentLocation(last)

			var facing string

			switch {
			case currentCell.North() == last:
				facing = "north"
			case currentCell.South() == last:
				facing = "south"
			case currentCell.West() == last:
				facing = "west"
			case currentCell.East() == last:
				facing = "east"
			}

			s := maze.NewSegment(client.CurrentLocation(), facing)
			client.TravelPath.AddSegement(s)
			m.SetPathFromTo(m.FromCell(client), client)

			in.Reply <- commandReply{
				answer: &moveReply{
					current:             client.CurrentLocation().Location(),
					availableDirections: client.CurrentLocation().DirectionLinks(in.ClientID),
					solved:              client.CurrentLocation().Location().String() == m.ToCell(client).Location().String(),
				},
			}
			t.UpdateSince(start)

		case maze.CommandMove:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.move", nil)

			r := in.Request
			direction := r.request.(string)

			client, err := m.Client(in.ClientID)
			if err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to find client: %v", err)}
			}

			switch direction {
			case "north":
				if client.CurrentLocation().Linked(client.CurrentLocation().North()) {
					client.SetCurrentLocation(client.CurrentLocation().North())
					s := maze.NewSegment(client.CurrentLocation(), "north")
					client.TravelPath.AddSegement(s)
					client.SolvePath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetPathFromTo(m.FromCell(client), client)

					in.Reply <- commandReply{
						answer: &moveReply{
							current:             client.CurrentLocation().Location(),
							availableDirections: client.CurrentLocation().DirectionLinks(in.ClientID),
							solved:              client.CurrentLocation().Location().String() == m.ToCell(client).Location().String(),
						},
					}
				} else {
					in.Reply <- commandReply{
						error: fmt.Errorf("cannot move 'north' from %v", client.CurrentLocation().String()),
					}
				}
			case "south":
				if client.CurrentLocation().Linked(client.CurrentLocation().South()) {
					client.SetCurrentLocation(client.CurrentLocation().South())
					s := maze.NewSegment(client.CurrentLocation(), "south")
					client.TravelPath.AddSegement(s)
					client.SolvePath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetPathFromTo(m.FromCell(client), client)

					in.Reply <- commandReply{
						answer: &moveReply{
							current:             client.CurrentLocation().Location(),
							availableDirections: client.CurrentLocation().DirectionLinks(in.ClientID),
							solved:              client.CurrentLocation().Location().String() == m.ToCell(client).Location().String(),
						},
					}
				} else {
					in.Reply <- commandReply{
						error: fmt.Errorf("cannot move 'south' from %v", client.CurrentLocation().String()),
					}
				}
			case "west":
				if client.CurrentLocation().Linked(client.CurrentLocation().West()) {
					client.SetCurrentLocation(client.CurrentLocation().West())
					s := maze.NewSegment(client.CurrentLocation(), "west")
					client.TravelPath.AddSegement(s)
					client.SolvePath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetPathFromTo(m.FromCell(client), client)

					in.Reply <- commandReply{
						answer: &moveReply{
							current:             client.CurrentLocation().Location(),
							availableDirections: client.CurrentLocation().DirectionLinks(in.ClientID),
							solved:              client.CurrentLocation().Location().String() == m.ToCell(client).Location().String(),
						},
					}
				} else {
					in.Reply <- commandReply{
						error: fmt.Errorf("cannot move 'west' from %v", client.CurrentLocation().String()),
					}
				}
			case "east":
				if client.CurrentLocation().Linked(client.CurrentLocation().East()) {
					client.SetCurrentLocation(client.CurrentLocation().East())
					s := maze.NewSegment(client.CurrentLocation(), "east")
					client.TravelPath.AddSegement(s)
					client.SolvePath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetPathFromTo(m.FromCell(client), client)

					in.Reply <- commandReply{
						answer: &moveReply{
							current:             client.CurrentLocation().Location(),
							availableDirections: client.CurrentLocation().DirectionLinks(in.ClientID),
							solved:              client.CurrentLocation().Location().String() == m.ToCell(client).Location().String(),
						},
					}
				} else {
					in.Reply <- commandReply{
						error: fmt.Errorf("cannot move 'east' from %v", client.CurrentLocation().String()),
					}
				}
			default:
				log.Printf("invalid direction: %v", direction)
				in.Reply <- commandReply{error: fmt.Errorf("invalid direction: %v", direction)}
			}
			t.UpdateSince(start)

		default:
			log.Printf("unknown command: %#v", in)
			in.Reply <- commandReply{error: fmt.Errorf("unknown command: %v", in)}
		}
	default:

	}
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

	log.Printf("starting metrics...")
	go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	exp.Exp(metrics.DefaultRegistry)
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2003")
	go graphite.Graphite(metrics.DefaultRegistry, 10e9, "metrics", addr)

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
	from    *pb.MazeLocation
	to      *pb.MazeLocation
}

type moveReply struct {
	current             *pb.MazeLocation
	availableDirections []*pb.Direction
	solved              bool
}

// server is used to implement MazerServer.
type server struct{}

// CreateMaze creates and displays the maze specified by the config
func (s *server) CreateMaze(ctx context.Context, in *pb.CreateMazeRequest) (*pb.CreateMazeReply, error) {
	log.Printf("creating maze with config: %#v", in.Config)

	t := metrics.GetOrRegisterTimer("maze.rpc.create-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	mazeID := uuid.NewV4().String()
	in.GetConfig().Id = mazeID

	comm := make(chan commandData)
	mazeMap.Insert(mazeID, comm)

	go createMaze(in.Config, comm)

	return &pb.CreateMazeReply{MazeId: mazeID}, nil
}

// RegisterClient registers a new client with an existing maze
func (s *server) RegisterClient(ctx context.Context, in *pb.RegisterClientRequest) (*pb.RegisterClientReply, error) {
	log.Printf("associating new client with maze: %#v", in.GetMazeId())
	t := metrics.GetOrRegisterTimer("maze.rpc.register-client.latency", nil)
	defer t.UpdateSince(time.Now())

	clientID := uuid.NewV4().String()

	m, found := mazeMap.Find(in.GetMazeId())
	if !found {
		return &pb.RegisterClientReply{Success: false, Message: fmt.Sprintf("unable to lookup maze [%v]: %v", in.GetMazeId())}, nil
	}

	comm := m.(chan commandData)

	data := commandData{
		Action:       maze.CommandAddClient,
		ClientID:     clientID,
		ClientConfig: in.GetClientConfig(),
		Reply:        make(chan commandReply),
	}
	comm <- data
	// get response from maze
	reply := <-data.Reply

	if reply.error != nil {
		return &pb.RegisterClientReply{Success: false, Message: reply.error.(error).Error()}, nil
	}
	return &pb.RegisterClientReply{Success: true, ClientId: clientID}, nil

}

// ListMazes lists all the mazes
func (s *server) ListMazes(ctx context.Context, in *pb.ListMazeRequest) (*pb.ListMazeReply, error) {
	t := metrics.GetOrRegisterTimer("maze.rpc.list-mazes.latency", nil)
	defer t.UpdateSince(time.Now())

	response := &pb.ListMazeReply{}
	for _, k := range mazeMap.Keys() {
		m := &pb.Maze{MazeId: k}

		r, found := mazeMap.Find(k)
		if !found {
			// should never happen
			return nil, fmt.Errorf("failed to find maze with id %v", k)
		}

		// comm is the channel to talk to the maze
		comm := r.(chan commandData)
		replyChannel := make(chan commandReply)
		data := commandData{
			Action: maze.CommandListClients,
			Reply:  replyChannel,
		}
		// send request to maze
		comm <- data

		// receive reply from maze, blocking
		mazeReply := <-replyChannel
		// maze reply, in this case a []string, client IDs
		m.ClientIds = mazeReply.answer.([]string)

		response.Mazes = append(response.Mazes, m)
	}
	return response, nil
}

// SolveMaze is a streaming RPC to solve a maze
func (s *server) SolveMaze(stream pb.Mazer_SolveMazeServer) error {
	log.Printf("received initial client connect...")
	t := metrics.GetOrRegisterTimer("maze.rpc.solve-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	// initial connect
	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	m, found := mazeMap.Find(in.GetMazeId())
	if !found {
		return fmt.Errorf("unable to lookup maze [%v]: %v", in.GetMazeId(), err)
	}
	// comm is the communication channel with this maze
	comm := m.(chan commandData)

	// check that client is valid

	// set client initial location
	data := commandData{
		Action:   maze.CommandSetInitialClientLocation,
		ClientID: in.ClientId,
		Request:  commandRequest{},
	}
	comm <- data
	// send request into m for available directions, include client id
	data = commandData{
		Action:   maze.CommandGetDirections,
		ClientID: in.GetClientId(),
		Reply:    make(chan commandReply),
	}
	comm <- data
	// get response from maze
	dirReply := <-data.Reply

	// send request into m for current location
	data = commandData{
		Action:   maze.CommandLocationInfo,
		ClientID: in.GetClientId(),
		Reply:    make(chan commandReply),
	}
	comm <- data
	// get currentlocation from maze
	locationInfoReply := <-data.Reply
	locationInfo := locationInfoReply.answer.(*locationInfo)

	// return available directions and current location
	reply := &pb.SolveMazeResponse{
		Initial:             true,
		MazeId:              in.GetMazeId(),
		ClientId:            in.ClientId,
		AvailableDirections: dirReply.answer.([]*pb.Direction),
		CurrentLocation:     locationInfo.current,
		FromCell:            locationInfo.from,
		ToCell:              locationInfo.to,
	}
	if err := stream.Send(reply); err != nil {
		return err
	}

	// this is the main loop as the client tries to solve the maze
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if in.Initial {
			// client is mis-behaving!
			r := &pb.SolveMazeResponse{
				Error:        true,
				ErrorMessage: fmt.Sprintf("received initial on subsequent request"),
			}
			if err := stream.Send(r); err != nil {
				return err
			}
			continue
		}

		var action commandAction
		if in.GetMoveBack() {
			action = maze.CommandMoveBack
		} else {
			action = maze.CommandMove
		}

		data = commandData{
			Action:   action,
			ClientID: in.ClientId,
			Request: commandRequest{
				request: in.GetDirection(), // which way to move
			},
			Reply: make(chan commandReply),
		}

		comm <- data
		// get response from maze
		mazeReply := <-data.Reply
		if err := mazeReply.error; err != nil {
			r := &pb.SolveMazeResponse{
				Error:        true,
				ErrorMessage: err.(error).Error(),
			}
			if err := stream.Send(r); err != nil {
				return err
			}
			continue
		}

		moveReply := mazeReply.answer.(*moveReply)

		r := &pb.SolveMazeResponse{
			MazeId:              in.GetMazeId(),
			ClientId:            in.ClientId,
			CurrentLocation:     moveReply.current,
			AvailableDirections: moveReply.availableDirections,
			Solved:              moveReply.solved,
		}
		if err := stream.Send(r); err != nil {
			return err
		}
	}

	return nil
}
