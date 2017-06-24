package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"mazes/algos"
	pb "mazes/proto"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/pkg/profile"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/sasha-s/go-deadlock"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

var (
	// operation
	op = flag.String("op", "list", "operation to run")

	// maze
	maskImage          = flag.String("mask_image", "", "file name of mask image")
	allowWeaving       = flag.Bool("weaving", false, "allow weaving")
	weavingProbability = flag.Float64("weaving_probability", 1, "controls the amount of weaving that happens, with 1 being the max")
	braidProbability   = flag.Float64("braid_probability", 0, "braid the maze with this probabily, 0 results in a perfect maze, 1 results in no deadends at all")
	randomFromTo       = flag.Bool("random_path", false, "show a random path through the maze")
	showGUI            = flag.Bool("gui", true, "show gui maze")

	// dimensions
	rows    = flag.Int64("r", 30, "number of rows in the maze")
	columns = flag.Int64("c", 60, "number of rows in the maze")

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
	cellWidth = flag.Int64("w", 20, "cell width (best as multiple of 2)")
	pathWidth = flag.Int64("path_width", 2, "path width")
	wallWidth = flag.Int64("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	wallSpace = flag.Int64("wall_space", 0, "how much space between two side by side walls (min of 2)")

	// display
	avatarImage        = flag.String("avatar_image", "", "file name of avatar image, the avatar should be facing to the left in the image")
	genDrawDelay       = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	markVisitedCells   = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	showFromToColors   = flag.Bool("show_from_to_colors", false, "show from/to colors")
	showDistanceColors = flag.Bool("show_distance_colors", false, "show distance colors")
	showDistanceValues = flag.Bool("show_distance_values", false, "show distance values")
	drawPathLength     = flag.Int64("draw_path_length", -1, "draw client path length, -1 = all, 0 = none")
	solveDrawDelay     = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	solveAlgo     = flag.String("solve_algo", "recursive-backtracker", "algorithm to solve the maze")
	skipGridCheck = flag.Bool("skip_grid_check", false, "set to true to skip grid check (disable spanning tree check)")

	// solver
	mazeID        = flag.String("maze_id", "", "maze id")
	disableOffset = flag.Bool("disable_draw_offset", false, "disable path draw offset")

	// misc
	exportFile = flag.String("export_file", "", "file to save maze to (does not work yet)")
	bgMusic    = flag.String("bg_music", "", "file name of background music to play")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	fromCellStr = flag.String("from_cell", "", "path from cell ('min' = minX, minY)")
	toCellStr   = flag.String("to_cell", "", "path to cell ('max' = maxX, maxY)")
)

func newMazeConfig(createAlgo, currentLocationColor string) *pb.MazeConfig {
	config := &pb.MazeConfig{
		Rows:                 *rows,
		Columns:              *columns,
		AllowWeaving:         *allowWeaving,
		WeavingProbability:   *weavingProbability,
		CellWidth:            *cellWidth,
		WallWidth:            *wallWidth,
		WallSpace:            *wallSpace,
		WallColor:            *wallColor,
		PathWidth:            *pathWidth,
		ShowDistanceColors:   *showDistanceColors,
		ShowDistanceValues:   *showDistanceValues,
		SkipGridCheck:        *skipGridCheck,
		CurrentLocationColor: currentLocationColor,
		GenDrawDelay:         *genDrawDelay,
		BgColor:              *bgColor,
		BorderColor:          *borderColor,
		CreateAlgo:           createAlgo,
		BraidProbability:     *braidProbability,
		Gui:                  *showGUI,
	}
	return config
}

func addClient(ctx context.Context, mazeID string, config *pb.ClientConfig) error {
	log.Printf("registering and running new client in maze %v...", mazeID)
	c := NewClient()

	r, err := c.RegisterClient(ctx,
		&pb.RegisterClientRequest{
			MazeId:       mazeID,
			ClientConfig: config,
		})
	if err != nil {
		return err
	}

	if !r.GetSuccess() {
		return fmt.Errorf("failed to register second client: %v", r.GetMessage())
	}

	log.Printf("solving maze (client=%v)...", r.GetClientId())
	return opSolve(mazeID, r.GetClientId(), config.GetSolveAlgo())
}

// opCreateSolveMulti creates and solves the maze
func opCreateSolveMulti() error {
	log.Print("creating maze...")

	r, err := opCreate()
	if err != nil {
		return err
	}
	mazeId := r.GetMazeId()
	var wd sync.WaitGroup

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            *solveAlgo,
		PathColor:            *pathColor,
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     *visitedCellColor,
		CurrentLocationColor: *currentLocationColor,
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	// register more clients
	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "recursive-backtracker",
		PathColor:            "blue",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "blue",
		CurrentLocationColor: "blue",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "recursive-backtracker",
		PathColor:            "green",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "green",
		CurrentLocationColor: "green",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "recursive-backtracker",
		PathColor:            "purple",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "purple",
		CurrentLocationColor: "purple",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "random",
		PathColor:            "pink",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "pink",
		CurrentLocationColor: "pink",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "random-unvisited",
		PathColor:            "gold",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "gold",
		CurrentLocationColor: "gold",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:            "wall-follower",
		PathColor:            "teal",
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     "teal",
		CurrentLocationColor: "teal",
		DisableDrawOffset:    *disableOffset,
		MarkVisitedCells:     *markVisitedCells,
		DrawPathLength:       *drawPathLength,
	})
	log.Printf("waiting for clients...")
	wd.Wait()

	return nil
}

// opCreateSolve creates and solves the maze
func opCreateSolve() error {
	log.Print("creating maze...")

	r, err := opCreate()
	if err != nil {
		return err
	}

	log.Print("solving maze...")
	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}
	return addClient(context.Background(), r.GetMazeId(), &pb.ClientConfig{
		SolveAlgo:            *solveAlgo,
		PathColor:            *pathColor,
		FromCell:             *fromCellStr,
		ToCell:               *toCellStr,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
		ShowFromToColors:     *showFromToColors,
		VisitedCellColor:     *visitedCellColor,
		CurrentLocationColor: *currentLocationColor,
		DrawPathLength:       *drawPathLength,
	})
}

// opCreate creates a new maze
func opCreate() (*pb.CreateMazeReply, error) {
	config := newMazeConfig(*createAlgo, *currentLocationColor)
	c := NewClient()

	r, err := c.CreateMaze(context.Background(), &pb.CreateMazeRequest{Config: config})
	if err != nil {
		log.Fatalf("could not show maze: %v", err)
	}
	return r, nil
}

// opList lists available mazes by their id
func opList() (*pb.ListMazeReply, error) {
	c := NewClient()

	r, err := c.ListMazes(context.Background(), &pb.ListMazeRequest{})
	if err != nil {
		log.Fatalf("could not list mazes: %v", err)
	}
	return r, nil
}

func opSolve(mazeID, clientID, solveAlgo string) error {
	log.Printf("in opSolve, client: %v", clientID)
	c := NewClient()

	stream, err := c.SolveMaze(context.Background())
	if err != nil {
		return err
	}

	if !checkSolveAlgo(solveAlgo) {
		return fmt.Errorf("invalid solve algorithm: %v", solveAlgo)
	}

	// initial connect to server to get the maze and client id
	r := &pb.SolveMazeRequest{
		Initial:  true,
		MazeId:   mazeID,
		ClientId: clientID,
	}
	if err := stream.Send(r); err != nil {
		log.Fatalf("talking to server: %v", err)
	}

	in, err := stream.Recv()
	if err != nil {
		log.Fatalf("error talking to server: %v", err)
	}
	for _, d := range in.GetAvailableDirections() {
		log.Printf("  can go: %#v", d)
	}

	log.Printf("maze id: %v; client id: %v", mazeID, clientID)

	solver := algos.NewSolver(solveAlgo, stream)
	delay, err := time.ParseDuration(*solveDrawDelay)
	if err != nil {
		return err
	}

	log.Printf("running solver %v", solveAlgo)

	if err := solver.Solve(mazeID, clientID, in.GetFromCell(), in.GetToCell(), delay, in.GetAvailableDirections()); err != nil {
		return fmt.Errorf("error running solver: %v", err)
	}

	return nil
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

func setFlags() {
	if *currentLocationColor == "" {
		*currentLocationColor = *pathColor
	}
}

// NewClient creates a server connection and returns a new SoleMazeClient
func NewClient() pb.MazerClient {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	// defer conn.Close()
	return pb.NewMazerClient(conn)
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	exp.Exp(metrics.DefaultRegistry)

	addr, _ := net.ResolveTCPAddr("tcp", "192.168.99.100:2003")
	go graphite.Graphite(metrics.DefaultRegistry, 10e9, "metrics", addr)

	// run http server for expvars
	sock, err := net.Listen("tcp", "localhost:8124")
	if err != nil {
		log.Fatalf(err.Error())
	}
	go func() {
		fmt.Println("metrics now available at http://localhost:8124/debug/metrics")
		http.Serve(sock, nil)
	}()

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

	setFlags()

	ctx := context.Background()

	//// Set up a connection to the server.
	//conn, err := grpc.Dial(address, grpc.WithInsecure())
	//if err != nil {
	//	log.Fatalf("did not connect: %v", err)
	//}
	//defer conn.Close()
	//c := pb.NewMazerClient(conn)

	//config := newMazeConfig(*createAlgo, *currentLocationColor)

	log.Printf("running: %v", *op)

	switch *op {
	case "create":
		if r, err := opCreate(); err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("%#v", r)
		}
	case "list":
		if r, err := opList(); err != nil {
			log.Fatalf(err.Error())
		} else {
			for _, m := range r.GetMazes() {
				log.Printf("maze: %v", m.GetMazeId())
				for _, c := range m.GetClientIds() {
					log.Printf("  client: %v", c)
				}
			}
		}
	case "solve":
		if *randomFromTo {
			*fromCellStr = "random"
			*toCellStr = "random"
		}

		if err := addClient(ctx, *mazeID, &pb.ClientConfig{
			SolveAlgo:        *solveAlgo,
			PathColor:        *pathColor,
			FromCell:         *fromCellStr,
			ToCell:           *toCellStr,
			FromCellColor:    *fromCellColor,
			ToCellColor:      *toCellColor,
			ShowFromToColors: *showFromToColors,
			VisitedCellColor: *visitedCellColor,
			DrawPathLength:   *drawPathLength,
		}); err != nil {
			log.Fatalf(err.Error())
		}
	case "create_solve":
		if err := opCreateSolve(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_multi":
		if err := opCreateSolveMulti(); err != nil {
			log.Print(err.Error())
		}
	}

}
