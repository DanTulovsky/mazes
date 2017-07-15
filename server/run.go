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

	"mazes/algos"
	"mazes/colors"
	"mazes/maze"
	pb "mazes/proto"
	lsdl "mazes/sdl"
	"safemap"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/pkg/profile"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/sasha-s/go-deadlock"
	"github.com/satori/go.uuid"
	"github.com/tevino/abool"
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

// for proto: protoc -I ./mazes/proto/ ./mazes/proto/mazes.proto --go_out=plugins=grpc:mazes/proto/
//   protoc -I ./proto/ ./proto/mazes.proto --go_out=plugins=grpc:proto/
// python:
//   cd ~/python/src
//   python -m grpc_tools.protoc -I../../go/src/mazes/proto --python_out=mazes/protos/ --grpc_python_out=mazes/protos/ ../../go/src/mazes/proto/mazes.proto

var (
	// maze
	maskImage = flag.String("mask_image", "", "file name of mask image")

	// display
	frameRate = flag.Uint("frame_rate", 120, "frame rate for animation")

	// misc
	bgMusic = flag.String("bg_music", "", "file name of background music to play")

	// stats
	showStats = flag.Bool("maze_stats", false, "show maze stats")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	exportPath = flag.String("export_path", "", "path where exported mazes are stored")

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

// required to be able to call gfx.* functions on multiple windows.
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
	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	title := fmt.Sprintf("Server View")
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

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Background Music
	//////////////////////////////////////////////////////////////////////////////////////////////
	if *bgMusic != "" {
		if err := mix.Init(mix.INIT_MP3); err != nil {
			return nil, nil, nil, fmt.Errorf("error initialing mp3: %v", err)
		}

		if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 2048); err != nil {
			return nil, nil, nil, fmt.Errorf("cannot initialize audio: %v", err)
		}

		music, err := mix.LoadMUS(*bgMusic)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load music file %v: %v", *bgMusic, err)
		}

		music.Play(-1) // loop forever
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Background Music
	//////////////////////////////////////////////////////////////////////////////////////////////

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

		//for x := 0; x < *columns; x++ {
		//	if x == *columns-1 {
		//		continue
		//	}
		//	c, _ := m.Cell(x, *rows/2)
		//	c.SetWeight(1000)
		//}

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
			})
		}
	}
	wd.Wait()
	///////////////////////////////////////////////////////////////////////////
	// End Generator
	///////////////////////////////////////////////////////////////////////////
	log.Printf("finished creating maze...")
	return m, r, w, nil
}

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

	wd.Add(1)
	go func() {
		defer wd.Done()
		log.Print("starting client comm thread...")
		for running.IsSet() {
			// check for client communications, they are serialized for one maze
			checkComm(m, comm, updateBG)
		}
		log.Printf("client comm thread died...")
	}()

	for running.IsSet() {
		start := time.Now()
		t := metrics.GetOrRegisterTimer("maze.loop.latency", nil)

		lsdl.CheckQuit(running)

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
		t.UpdateSince(start)
	}
	mazeMap.Delete(m.Config().GetId())

	log.Printf("maze is done...")

	showMazeStats(m)
	wd.Wait()

}

func checkComm(m *maze.Maze, comm commChannel, updateBG *abool.AtomicBool) {
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
		case maze.CommandExportMaze:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.export-maze.latency", nil)

			if *exportPath == "" {
				in.Reply <- commandReply{error: fmt.Errorf("export_path not set on server")}
				return
			}
			encoded := m.Encode()
			if encoded == "" {
				in.Reply <- commandReply{error: fmt.Errorf("failed to encode maze")}
				return
			}

			if err := m.Export(*exportPath); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to export maze: %v", err)}
				return
			}

			in.Reply <- commandReply{error: nil}

			t.UpdateSince(start)
		case maze.CommandAddClient:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.add-client.latency", nil)

			// TODO(dan): Is this needed?
			m.Reset()

			fromCell, toCell, err := m.AddClient(in.ClientID, in.ClientConfig)
			if err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to add client: %v", err)}
				return
			}
			updateBG.Set()

			l := &locationInfo{
				From: fromCell.Location(),
				To:   toCell.Location(),
			}

			// send reply via the reply channel
			in.Reply <- commandReply{answer: l}
			t.UpdateSince(start)
		case maze.CommandGetDirections:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.get-direction.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to get client links: %v", err)}
				return

			} else {
				in.Reply <- commandReply{answer: client.CurrentLocation().DirectionLinks(in.ClientID)}
			}
			t.UpdateSince(start)
		case maze.CommandSetInitialClientLocation:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.set-initial-client-location.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to set initial client location: %v", err)}
				return
			} else {
				client.SetCurrentLocation(m.FromCell(client))
				cell := client.CurrentLocation()

				// Add initial location to paths
				s := maze.NewSegment(cell, "north", true)
				cell.SetVisited(in.ClientID)
				client.TravelPath.AddSegement(s)
				in.Reply <- commandReply{error: nil}
			}
			t.UpdateSince(start)
		case maze.CommandCurrentLocation:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.current-location.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to get client location: %v", err)}
				return
			} else {
				cell := client.CurrentLocation()
				in.Reply <- commandReply{answer: cell.Location()}
			}
			t.UpdateSince(start)
		case maze.CommandLocationInfo:
			start := time.Now()
			t := metrics.GetOrRegisterTimer("maze.command.location-info.latency", nil)

			if client, err := m.Client(in.ClientID); err != nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to get client location: %v", err)}
				return
			} else {
				info := &locationInfo{
					current: client.CurrentLocation().Location(),
					From:    m.FromCell(client).Location(),
					To:      m.ToCell(client).Location(),
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
				return
			}

			// remove from solution if we are backtracking
			client.TravelPath.LastSegment().RemoveFromSolution()

			// previous cell
			currentCell := client.CurrentLocation()
			lastSegment := client.TravelPath.PreviousSegmentinSolution()
			if lastSegment == nil {
				in.Reply <- commandReply{error: fmt.Errorf("failed to find previous unvisited cell, path is: %v", client.TravelPath)}
				return

			}
			last := lastSegment.Cell()
			last.SetVisited(in.ClientID)
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

			// moving back, don't set this cell where we've been as solution
			s := maze.NewSegment(client.CurrentLocation(), facing, false)
			client.TravelPath.AddSegement(s)
			m.SetClientPath(client)

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
					s := maze.NewSegment(client.CurrentLocation(), "north", true)
					client.TravelPath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetClientPath(client)

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
					s := maze.NewSegment(client.CurrentLocation(), "south", true)
					client.TravelPath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetClientPath(client)

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
					s := maze.NewSegment(client.CurrentLocation(), "west", true)
					client.TravelPath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetClientPath(client)

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
					s := maze.NewSegment(client.CurrentLocation(), "east", true)
					client.TravelPath.AddSegement(s)
					client.CurrentLocation().SetVisited(in.ClientID)
					m.SetClientPath(client)

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
		// when the client disconnects, this will block until the timer fires
		// if this is just a 'default' fall through, much cpu is used as multiple mazes are run
	case <-time.After(5 * time.Second):

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
	// go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	exp.Exp(metrics.DefaultRegistry)
	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2003")
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
	From    *pb.MazeLocation
	To      *pb.MazeLocation
}

type moveReply struct {
	current             *pb.MazeLocation
	availableDirections []*pb.Direction
	solved              bool
}

// server is used to implement MazerServer.
type server struct{}

// ExportMaze exports the given maze to disk, only the structure is preserved
func (s *server) ExportMaze(ctx context.Context, in *pb.ExportMazeRequest) (*pb.ExportMazeReply, error) {
	log.Printf("exporting maze with id: %v", in.GetMazeId())
	if in.GetMazeId() == "" {
		return nil, fmt.Errorf("maze id cannot be empty")
	}

	t := metrics.GetOrRegisterTimer("maze.rpc.export-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	m, found := mazeMap.Find(in.GetMazeId())
	if !found {
		return &pb.ExportMazeReply{Success: false, Message: fmt.Sprintf("unable to lookup maze [%v]: %v", in.GetMazeId())}, nil
	}

	comm := m.(chan commandData)

	data := commandData{
		Action: maze.CommandExportMaze,
		Reply:  make(chan commandReply),
	}
	comm <- data
	// get response from maze
	reply := <-data.Reply
	if reply.error != nil {
		return &pb.ExportMazeReply{Success: false, Message: reply.error.(error).Error()}, nil
	}

	return &pb.ExportMazeReply{Success: true}, nil
}

// CreateMaze creates and displays the maze specified by the config
func (s *server) CreateMaze(ctx context.Context, in *pb.CreateMazeRequest) (*pb.CreateMazeReply, error) {
	log.Printf("creating maze with config: %#v", in.Config)
	if in.Config == nil {
		return nil, fmt.Errorf("maze config cannot be nil")
	}

	t := metrics.GetOrRegisterTimer("maze.rpc.create-maze.latency", nil)
	defer t.UpdateSince(time.Now())

	mazeID := uuid.NewV4().String()
	in.GetConfig().Id = mazeID

	comm := make(chan commandData)
	mazeMap.Insert(mazeID, comm)

	m, r, w, err := createMaze(in.Config)
	if err != nil {
		return nil, err
	}
	go runMaze(m, r, w, comm)

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

	locationInfo := reply.answer.(*locationInfo)
	return &pb.RegisterClientReply{Success: true, ClientId: clientID,
		FromCell: locationInfo.From, ToCell: locationInfo.To}, nil

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
		Reply:    make(chan commandReply),
		Request:  commandRequest{},
	}
	comm <- data
	initialLocationReply := <-data.Reply
	if initialLocationReply.error != nil {
		return initialLocationReply.error.(error)
	}

	// send request into m for available directions, include client id
	data = commandData{
		Action:   maze.CommandGetDirections,
		ClientID: in.GetClientId(),
		Reply:    make(chan commandReply),
	}
	comm <- data
	// get response from maze
	dirReply := <-data.Reply
	if dirReply.error != nil {
		return dirReply.error.(error)
	}

	// send request into m for current location
	data = commandData{
		Action:   maze.CommandLocationInfo,
		ClientID: in.GetClientId(),
		Reply:    make(chan commandReply),
	}
	comm <- data
	// get currentlocation from maze
	locationInfoReply := <-data.Reply
	if locationInfoReply.error != nil {
		return locationInfoReply.error.(error)
	}
	locationInfo := locationInfoReply.answer.(*locationInfo)

	// return available directions and current location
	reply := &pb.SolveMazeResponse{
		Initial:             true,
		MazeId:              in.GetMazeId(),
		ClientId:            in.ClientId,
		AvailableDirections: dirReply.answer.([]*pb.Direction),
		CurrentLocation:     locationInfo.current,
		FromCell:            locationInfo.From,
		ToCell:              locationInfo.To,
	}
	if err := stream.Send(reply); err != nil {
		return err
	}

	trpc := metrics.GetOrRegisterTimer("maze.rpc.solve-maze-loop.latency", nil)

	// this is the main loop as the client tries to solve the maze
	for {
		start := time.Now()

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

		commStart := time.Now()
		tcomm := metrics.GetOrRegisterTimer("maze.rpc.solve-maze-loop-comm.latency", nil)
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
		tcomm.UpdateSince(commStart)

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

		trpc.UpdateSince(start)
	}

	return nil
}
