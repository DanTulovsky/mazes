package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pkg/profile"
	deadlock "github.com/sasha-s/go-deadlock"
	"google.golang.org/grpc"
	"mazes/algos"
	pb "mazes/proto"
	"mazes/solvealgos"
)

const (
	address = "localhost:50051"
)

var (
	winTitle string = "Maze"

	solver solvealgos.Algorithmer = algos.SolveAlgorithms[*solveAlgo]

	// operation
	op = flag.String("op", "list", "operation to run")

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

	// display
	avatarImage        = flag.String("avatar_image", "", "file name of avatar image, the avatar should be facing to the left in the image")
	genDrawDelay       = flag.String("gen_draw_delay", "0", "solver delay per step, used for animation")
	markVisitedCells   = flag.Bool("mark_visited", false, "mark visited cells (by solver)")
	showFromToColors   = flag.Bool("show_from_to_colors", false, "show from/to colors")
	showDistanceColors = flag.Bool("show_distance_colors", false, "show distance colors")
	showDistanceValues = flag.Bool("show_distance_values", false, "show distance values")
	solveDrawDelay     = flag.String("solve_draw_delay", "0", "solver delay per step, used for animation")

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	solveAlgo     = flag.String("solve_algo", "empty", "algorithm to solve the maze")
	skipGridCheck = flag.Bool("skip_grid_check", false, "set to true to skip grid check (disable spanning tree check)")

	// solver
	mazeID   = flag.String("maze_id", "", "maze id")
	clientID = flag.String("client_id", "", "client id")

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
		MarkVisitedCells:     *markVisitedCells,
		ShowDistanceColors:   *showDistanceColors,
		ShowDistanceValues:   *showDistanceValues,
		SkipGridCheck:        *skipGridCheck,
		CurrentLocationColor: currentLocationColor,
		GenDrawDelay:         *genDrawDelay,
		BgColor:              *bgColor,
		BorderColor:          *borderColor,
		CreateAlgo:           createAlgo,
		BraidProbability:     *braidProbability,
	}
	return config
}

func addClient(ctx context.Context, c pb.MazerClient, mazeID string, config *pb.ClientConfig) error {
	log.Printf("registering and running new client in maze %v...", mazeID)
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
	return opSolve(ctx, c, mazeID, r.GetClientId(), config.GetSolveAlgo())
}

// opCreateSolveMulti creates and solves the maze
func opCreateSolveMulti(ctx context.Context, c pb.MazerClient, config *pb.MazeConfig) error {
	log.Print("creating maze...")
	r, err := opCreate(ctx, c, config)
	if err != nil {
		return err
	}
	mazeId := r.GetMazeId()
	var wd sync.WaitGroup

	log.Printf("solving maze1 (client=%v; maze=%v)...", r.GetClientId(), mazeID)
	wd.Add(1)
	go addClient(context.Background(), c, mazeId, &pb.ClientConfig{
		SolveAlgo:        *solveAlgo,
		PathColor:        *pathColor,
		FromCell:         *fromCellStr,
		ToCell:           *toCellStr,
		FromCellColor:    *fromCellColor,
		ToCellColor:      *toCellColor,
		ShowFromToColors: *showFromToColors,
		VisitedCellColor: *visitedCellColor,
	})

	// register more clients
	wd.Add(1)
	go addClient(context.Background(), c, mazeId, &pb.ClientConfig{
		SolveAlgo:        "wall-follower",
		PathColor:        "blue",
		FromCell:         *fromCellStr,
		ToCell:           *toCellStr,
		FromCellColor:    *fromCellColor,
		ToCellColor:      *toCellColor,
		ShowFromToColors: *showFromToColors,
		VisitedCellColor: "blue",
	})
	wd.Add(1)
	go addClient(context.Background(), c, mazeId, &pb.ClientConfig{
		SolveAlgo:        "recursive-backtracker",
		PathColor:        "green",
		FromCell:         *fromCellStr,
		ToCell:           *toCellStr,
		FromCellColor:    *fromCellColor,
		ToCellColor:      *toCellColor,
		ShowFromToColors: *showFromToColors,
		VisitedCellColor: "green",
	})
	wd.Add(1)
	go addClient(context.Background(), c, mazeId, &pb.ClientConfig{
		SolveAlgo:        "recursive-backtracker",
		PathColor:        "purple",
		FromCell:         *fromCellStr,
		ToCell:           *toCellStr,
		FromCellColor:    *fromCellColor,
		ToCellColor:      *toCellColor,
		ShowFromToColors: *showFromToColors,
		VisitedCellColor: "purple",
	})

	log.Printf("waiting for clients...")
	wd.Wait()

	return nil
}

// opCreateSolve creates and solves the maze
func opCreateSolve(ctx context.Context, c pb.MazerClient, config *pb.MazeConfig) error {
	log.Print("creating maze...")
	r, err := opCreate(ctx, c, config)
	if err != nil {
		return err
	}

	log.Print("solving maze...")
	return addClient(context.Background(), c, r.GetMazeId(), &pb.ClientConfig{
		SolveAlgo:        *solveAlgo,
		PathColor:        *pathColor,
		FromCell:         *fromCellStr,
		ToCell:           *toCellStr,
		FromCellColor:    *fromCellColor,
		ToCellColor:      *toCellColor,
		ShowFromToColors: *showFromToColors,
		VisitedCellColor: *visitedCellColor,
	})
}

// opCreate creates a new maze
func opCreate(ctx context.Context, c pb.MazerClient, config *pb.MazeConfig) (*pb.CreateMazeReply, error) {
	r, err := c.CreateMaze(ctx, &pb.CreateMazeRequest{Config: config})
	if err != nil {
		log.Fatalf("could not show maze: %v", err)
	}
	return r, nil
}

// opList lists available mazes by their id
func opList(ctx context.Context, c pb.MazerClient) (*pb.ListMazeReply, error) {
	r, err := c.ListMazes(ctx, &pb.ListMazeRequest{})
	if err != nil {
		log.Fatalf("could not list mazes: %v", err)
	}
	log.Printf("> %v", r)
	return r, nil
}

func opSolve(ctx context.Context, c pb.MazerClient, mazeID, clientID, solveAlgo string) error {
	log.Printf("in opSolve, client: %v", clientID)
	stream, err := c.SolveMaze(ctx)
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
	log.Print("initial send to server")
	if err := stream.Send(r); err != nil {
		log.Fatalf("talking to server: %v", err)
	}
	log.Print("sent...")

	log.Print("waiting for reply")
	in, err := stream.Recv()
	if err != nil {
		log.Fatalf("error talking to server: %v", err)
	}
	log.Printf("have reply: %#v", in)
	log.Printf("current_location: %#v", in.GetCurrentLocation())
	for _, d := range in.GetAvailableDirections() {
		log.Printf("  can go: %#v", d)
	}

	log.Printf("maze id: %v; client id: %v", mazeID, clientID)

	solver = algos.NewSolver(solveAlgo, stream)
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

	ctx := context.Background()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMazerClient(conn)

	config := newMazeConfig(*createAlgo, *currentLocationColor)

	log.Printf("running: %v", *op)

	switch *op {
	case "create":
		if r, err := opCreate(ctx, c, config); err != nil {
			log.Printf(err.Error())
		} else {
			log.Printf("%#v", r)
		}
	case "list":
		if r, err := opList(ctx, c); err != nil {
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
		if err := opSolve(ctx, c, *mazeID, *clientID, *solveAlgo); err != nil {
			log.Fatalf(err.Error())
		}

	case "create_solve":
		if err := opCreateSolve(ctx, c, config); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_multi":
		if err := opCreateSolveMulti(ctx, c, config); err != nil {
			log.Print(err.Error())
		}
	}

}
