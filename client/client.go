package main

import (
	"context"
	"flag"
	"log"
	"mazes/maze"
	"mazes/solvealgos"

	"google.golang.org/grpc"

	pb "mazes/proto"
)

const (
	address = "localhost:50051"
)

var (
	winTitle         string = "Maze"
	fromCell, toCell *maze.Cell

	solver solvealgos.Algorithmer

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
	showAscii = flag.Bool("ascii", false, "show ascii maze")
	showGUI   = flag.Bool("gui", true, "show gui maze")

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
	solveAlgo     = flag.String("solve_algo", "", "algorithm to solve the maze")
	skipGridCheck = flag.Bool("skip_grid_check", false, "set to true to skip grid check (disable spanning tree check)")

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

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewMazerClient(conn)

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
		PathColor:            *pathColor,
		MarkVisitedCells:     *markVisitedCells,
		ShowDistanceColors:   *showDistanceColors,
		ShowDistanceValues:   *showDistanceValues,
		SkipGridCheck:        *skipGridCheck,
		AvatarImage:          *avatarImage,
		VisitedCellColor:     *visitedCellColor,
		BgColor:              *bgColor,
		BorderColor:          *borderColor,
		CurrentLocationColor: *currentLocationColor,
		FromCellColor:        *fromCellColor,
		ToCellColor:          *toCellColor,
	}

	log.Printf("%#v", config)
	// Contact the server and print out its response.
	r, err := c.ShowMaze(ctx, &pb.ShowMazeRequest{Config: config})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("> %v", r)
}
