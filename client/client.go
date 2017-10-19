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
	"mazes/colors"
	"mazes/ml"
	pb "mazes/proto"
	lsdl "mazes/sdl"

	"os"

	"mazes/maze"

	"mazes/ml/dp"
	"mazes/ml/mc"

	"mazes/genalgos/from_encoded_string"

	"mazes/ml/td"

	"mazes/solvealgos"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/nsf/termbox-go"
	"github.com/pkg/profile"
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/sasha-s/go-deadlock"
	"github.com/tevino/abool"
	"github.com/veandco/go-sdl2/sdl"
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

	// algo
	createAlgo    = flag.String("create_algo", "recursive-backtracker", "algorithm used to create the maze")
	solveAlgo     = flag.String("solve_algo", "recursive-backtracker", "algorithm to solve the maze")
	skipGridCheck = flag.Bool("skip_grid_check", true, "set to true to skip grid check (disable spanning tree check)")

	// solver
	mazeID        = flag.String("maze_id", "", "maze id")
	disableOffset = flag.Bool("disable_draw_offset", false, "disable path draw offset")
	fromCellStr   = flag.String("from_cell", "", "path from cell ('min' = minX, minY)")
	toCellStr     = flag.String("to_cell", "", "path to cell ('max' = maxX, maxY)")
	returnMaze    = flag.Bool("return_maze", false, "return the encoded maze in the create reply, used for ML DP algorithms")

	// ml params
	df                 = flag.Float64("df", 1, "discount factor [0-1], at one treats all steps equally")
	epsilon            = flag.Float64("epsilon", 1, "chance of picking random action [0-1], used to explore")
	epsilonDecayFactor = flag.Float64("epsilon_decay_factor", -0.001, "decay factor for epsilon, the closer to 0, the slow it decays")
	theta              = flag.Float64("theta", 0.000000001, "stops evaluation when value function change is less than this")
	alpha              = flag.Float64("alpha", 0.0001, "TD learning rate")
	gamma              = flag.Float64("gamma", 1, "discount rate of earlier steps")
	lambda             = flag.Float64("lambda", 0.99, "trace decay parameter (1 = monte carlo, 0 = TD(0)), used for eligibility traces")
	numEpisodes        = flag.Int64("num_episodes", 10000, "for episodic algorithms, run this many episodes")
	maxSteps           = flag.Int64("max_steps", 0, "run only this many steps per episode, 0 means set automatically")

	// misc
	exportMaze       = flag.Bool("export_maze", false, "save maze to a file on the server")
	bgMusic          = flag.String("bg_music", "", "file name of background music to play")
	enableMonitoring = flag.Bool("enable_monitoring", false, "enable monitoring")

	// debug
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
	enableProfile           = flag.Bool("enable_profile", false, "enable profiling")

	wd sync.WaitGroup
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
		ShowWeightValues:     *showWeightValues,
		SkipGridCheck:        *skipGridCheck,
		CurrentLocationColor: currentLocationColor,
		GenDrawDelay:         *genDrawDelay,
		BgColor:              *bgColor,
		BorderColor:          *borderColor,
		CreateAlgo:           createAlgo,
		BraidProbability:     *braidProbability,
		Gui:                  *showGUI,
		FromFile:             *mazeID,
		ReturnMaze:           *returnMaze,
		Title:                *title,
	}
	return config
}

// addClient creates a new client in the maze and runs the solver, the m value is the *local* maze for display
func addClient(ctx context.Context, mazeID string, config *pb.ClientConfig, m *maze.Maze, p *ml.Policy) error {
	log.Printf("registering and running new client in maze %v...", mazeID)
	c := solvealgos.NewClient()

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
	log.Printf("fromCell: %v; toCell: %v", r.GetFromCell(), r.GetToCell())

	// match from/to path on the server (if random, for example)
	config.FromCell = fmt.Sprintf("%d,%d", r.GetFromCell().GetX(), r.GetFromCell().GetY())
	config.ToCell = fmt.Sprintf("%d,%d", r.GetToCell().GetX(), r.GetToCell().GetY())

	log.Printf("path: %s -> %s", config.GetFromCell(), config.GetToCell())

	if m != nil {
		m.AddClient(r.GetClientId(), config)
	}

	return opSolve(mazeID, r.GetClientId(), config.GetSolveAlgo(), m, p)
}

// opCreateSolveMulti creates and solves the maze
func opCreateSolveMulti() error {
	log.Print("creating maze...")

	r, _, err := opCreate()
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
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	// register more clients
	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "recursive-backtracker",
		PathColor:              "blue",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "blue",
		CurrentLocationColor:   "blue",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "recursive-backtracker",
		PathColor:              "green",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "green",
		CurrentLocationColor:   "green",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "recursive-backtracker",
		PathColor:              "purple",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "purple",
		CurrentLocationColor:   "purple",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "random",
		PathColor:              "pink",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "pink",
		CurrentLocationColor:   "pink",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "random-unvisited",
		PathColor:              "gold",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "gold",
		CurrentLocationColor:   "gold",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)

	wd.Add(1)
	go addClient(context.Background(), mazeId, &pb.ClientConfig{
		SolveAlgo:              "wall-follower",
		PathColor:              "teal",
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       "teal",
		CurrentLocationColor:   "teal",
		DisableDrawOffset:      *disableOffset,
		MarkVisitedCells:       *markVisitedCells,
		DrawPathLength:         *drawPathLength,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, nil, nil)
	log.Printf("waiting for clients...")
	wd.Wait()

	return nil
}

// opCreateSolve creates and solves the maze
func opCreateSolve() error {
	log.Print("creating maze...")

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	log.Print("solving maze...")
	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}
	return addClient(context.Background(), r.GetMazeId(), &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         *drawPathLength,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}, m, nil)
}

// opCreateSolveMlDpPolicyIteration creates and solves a maze using DP Policy Iteration
func opCreateSolveMlDpPolicyIteration() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         *drawPathLength,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	log.Print("figuring out optimal policy using dp policy iteration...")
	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	if *df == 1.0 {
		*df = 0.99 // does not complete with df = 1
		log.Printf("df == 1 with dp policy evaluation, this often does not converge, adjusting to %v", *df)

	}
	theta := *theta
	clientID := "clientID"
	log.Printf("Determining optimal policy using policy iteration..")
	clientConfig.DrawPathLength = 0 // don't draw on the client
	m.AddClient(clientID, clientConfig)
	// runs through the *local* maze to find optimal path
	policy, _, err := dp.PolicyImprovement(m, clientID, *df, theta, ml.DefaultActions)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	return addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)

}

// opCreateSolveMlDpValueIteration creates and solves a maze using DP Value Iteration
func opCreateSolveMlDpValueIteration() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         0, // don't draw on the client
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	log.Print("figuring out optimal policy using dp value iteration...")
	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	df := *df
	theta := *theta
	clientID := "clientID"
	log.Printf("Determining optimal policy using value iteration...")
	m.AddClient(clientID, clientConfig)
	// runs through the *local* maze to find optimal path
	policy, _, err := dp.ValueIteration(m, clientID, df, theta, ml.DefaultActions)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server

	return addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
}

// opCreateSolveMlMCEpsilonGreedyControl creates and solves a maze using MC Epsilon Greedy Control
func opCreateSolveMlMCEpsilonGreedyControl() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         100,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	// max steps per run through maze
	if *maxSteps == 0 {
		*maxSteps = m.Config().Rows * m.Config().Columns * 10000
	}
	clientID := "clientID"
	m.AddClient(clientID, clientConfig)

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()

	var wd sync.WaitGroup

	comparison := false
	if comparison {
		// for comparison
		mazeId := r.GetMazeId()
		wd.Add(1)

		go addClient(context.Background(), mazeId, &pb.ClientConfig{
			SolveAlgo:              "recursive-backtracker",
			PathColor:              "purple",
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       "purple",
			CurrentLocationColor:   "purple",
			DisableDrawOffset:      *disableOffset,
			MarkVisitedCells:       *markVisitedCells,
			DrawPathLength:         *drawPathLength,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil)
	}

	// runs through the *local* maze to find optimal path
	log.Print("figuring out optimal policy using mc epsilon greedy control...")
	_, policy, err := mc.ControlEpsilonGreedy(m, clientID, *numEpisodes, *theta, *df, c.FromCell().Location(),
		c.ToCell(), *maxSteps, *epsilon, *epsilonDecayFactor)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	log.Printf("policy:\n%v", policy)
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	policy.SetType("deterministic")               // follow this policy exactly

	wd.Add(1)
	go addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
	wd.Wait()

	return nil
}

func opCreateSolveMlMCOffPolicy() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         100,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	// max steps per run through maze
	if *maxSteps == 0 {
		*maxSteps = m.Config().Rows * m.Config().Columns * 10000
	}
	clientID := "clientID"
	m.AddClient(clientID, clientConfig)

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()

	var wd sync.WaitGroup

	comparison := false
	if comparison {
		// for comparison
		mazeId := r.GetMazeId()
		wd.Add(1)

		go addClient(context.Background(), mazeId, &pb.ClientConfig{
			SolveAlgo:              "recursive-backtracker",
			PathColor:              "purple",
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       "purple",
			CurrentLocationColor:   "purple",
			DisableDrawOffset:      *disableOffset,
			MarkVisitedCells:       *markVisitedCells,
			DrawPathLength:         *drawPathLength,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil)
	}

	// runs through the *local* maze to find optimal path
	log.Print("figuring out optimal policy using mc off policy importance sampling...")
	_, policy, err := mc.OffPolicyControl(m, clientID, *numEpisodes, *theta, *df, c.FromCell().Location(),
		c.ToCell(), *maxSteps)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	log.Printf("policy:\n%v", policy)
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	policy.SetType("deterministic")               // follow this policy exactly

	wd.Add(1)
	go addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
	wd.Wait()

	return nil
}

func opCreateSolveMlTDOneStepSarsa() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         100,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	// max steps per run through maze
	if *maxSteps == 0 {
		*maxSteps = m.Config().Rows * m.Config().Columns * 10000
	}
	clientID := "clientID"
	m.AddClient(clientID, clientConfig)

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()

	var wd sync.WaitGroup

	comparison := false
	if comparison {
		// for comparison
		mazeId := r.GetMazeId()
		wd.Add(1)

		go addClient(context.Background(), mazeId, &pb.ClientConfig{
			SolveAlgo:              "recursive-backtracker",
			PathColor:              "purple",
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       "purple",
			CurrentLocationColor:   "purple",
			DisableDrawOffset:      *disableOffset,
			MarkVisitedCells:       *markVisitedCells,
			DrawPathLength:         *drawPathLength,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil)
	}

	// runs through the *local* maze to find optimal path
	log.Print("figuring out optimal policy using td one step sarsa...")
	_, policy, err := td.OneStepSarsa(m, clientID, *numEpisodes, *alpha, *df, c.FromCell().Location(),
		c.ToCell(), *maxSteps, *epsilon, *epsilonDecayFactor)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	log.Printf("policy:\n%v", policy)
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	policy.SetType("deterministic")               // follow this policy exactly

	wd.Add(1)
	go addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
	wd.Wait()

	return nil
}

func opCreateSolveMlTDSarsaLambda() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         100,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	// max steps per run through maze
	if *maxSteps == 0 {
		*maxSteps = m.Config().Rows * m.Config().Columns * 10000
	}
	clientID := "clientID"
	m.AddClient(clientID, clientConfig)

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()

	var wd sync.WaitGroup

	comparison := false
	if comparison {
		// for comparison
		mazeId := r.GetMazeId()
		wd.Add(1)

		go addClient(context.Background(), mazeId, &pb.ClientConfig{
			SolveAlgo:              "recursive-backtracker",
			PathColor:              "purple",
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       "purple",
			CurrentLocationColor:   "purple",
			DisableDrawOffset:      *disableOffset,
			MarkVisitedCells:       *markVisitedCells,
			DrawPathLength:         *drawPathLength,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil)
	}

	// runs through the *local* maze to find optimal path
	log.Print("figuring out optimal policy using td sarsa lambda...")
	_, policy, err := td.SarsaLambda(m, clientID, *numEpisodes, *gamma, *lambda, *alpha, *epsilon, *epsilonDecayFactor,
		c.FromCell().Location(), c.ToCell(), *maxSteps)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	log.Printf("policy:\n%v", policy)
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	policy.SetType("deterministic")               // follow this policy exactly

	wd.Add(1)
	go addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
	wd.Wait()

	return nil
}

func opCreateSolveMlTDQLearning() error {
	log.Print("creating maze...")

	// always return maze from server
	*returnMaze = true
	*solveAlgo = "follow-policy"

	if *randomFromTo {
		*fromCellStr = "random"
		*toCellStr = "random"
	}

	clientConfig := &pb.ClientConfig{
		SolveAlgo:              *solveAlgo,
		PathColor:              *pathColor,
		FromCell:               *fromCellStr,
		ToCell:                 *toCellStr,
		FromCellColor:          *fromCellColor,
		ToCellColor:            *toCellColor,
		ShowFromToColors:       *showFromToColors,
		VisitedCellColor:       *visitedCellColor,
		CurrentLocationColor:   *currentLocationColor,
		DrawPathLength:         100,
		MarkVisitedCells:       *markVisitedCells,
		NumberMarkVisitedCells: *numberMarkVisitedCells,
	}

	r, m, err := opCreate()
	if err != nil {
		return err
	}

	algo := &from_encoded_string.FromEncodedString{}
	if err := algo.Apply(m, 0, nil); err != nil {
		return fmt.Errorf("error applying algorithm: %v", err)
	}

	// max steps per run through maze
	if *maxSteps == 0 {
		*maxSteps = m.Config().Rows * m.Config().Columns * 10000
	}
	clientID := "clientID"
	m.AddClient(clientID, clientConfig)

	// we must pass the same from/to to the server as we used locally
	c, err := m.Client(clientID)
	if err != nil {
		return err
	}
	clientConfig.FromCell = c.FromCell().StringXY()
	clientConfig.ToCell = c.ToCell().StringXY()

	var wd sync.WaitGroup

	comparison := false
	if comparison {
		// for comparison
		mazeId := r.GetMazeId()
		wd.Add(1)

		go addClient(context.Background(), mazeId, &pb.ClientConfig{
			SolveAlgo:              "recursive-backtracker",
			PathColor:              "purple",
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       "purple",
			CurrentLocationColor:   "purple",
			DisableDrawOffset:      *disableOffset,
			MarkVisitedCells:       *markVisitedCells,
			DrawPathLength:         *drawPathLength,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil)
	}

	// runs through the *local* maze to find optimal path
	log.Print("figuring out optimal policy using td sarsa...")
	_, policy, err := td.QLearning(m, clientID, *numEpisodes, *alpha, *df, c.FromCell().Location(),
		c.ToCell(), *maxSteps, *epsilon, *epsilonDecayFactor)
	if err != nil {
		return fmt.Errorf("error calculating optimal policy: %v", err)
	}

	log.Printf("policy:\n%v", policy)
	m.Reset()

	clientConfig.DrawPathLength = *drawPathLength // restore for the server
	policy.SetType("deterministic")               // follow this policy exactly

	wd.Add(1)
	go addClient(context.Background(), r.GetMazeId(), clientConfig, m, policy)
	wd.Wait()

	return nil
}

// opCreate creates a new maze
func opCreate() (*pb.CreateMazeReply, *maze.Maze, error) {
	config := newMazeConfig(*createAlgo, *currentLocationColor)
	c := solvealgos.NewClient()
	ctx := context.Background()

	resp, err := c.CreateMaze(ctx, &pb.CreateMazeRequest{Config: config})
	if err != nil {
		log.Fatalf("could not create maze: %v", err)
	}

	var m *maze.Maze
	var r *sdl.Renderer
	var w *sdl.Window

	// create local maze for DP algorithms or local gui
	if *showLocalGUI || *solveAlgo == "follow-policy" || *solveAlgo == "ml-td-one-step-sarsa" || *solveAlgo == "ml-td-sarsa-lambda" {
		if *showLocalGUI {
			// if server gui is off, enable this so the client gui works
			config.Gui = true
		}

		encodedMaze := resp.GetEncodedMaze()
		if m, r, w, err = createMaze(config, encodedMaze); err != nil {
			log.Fatalf("could not create local client view of maze: %v", err)
		}
		if *showLocalGUI {
			wd.Add(1)
			go showMaze(m, r, w)
		}

	}

	if *exportMaze {
		log.Printf("requesting server export maze...")
		r, err := c.ExportMaze(ctx, &pb.ExportMazeRequest{MazeId: resp.GetMazeId()})
		if err != nil || !r.GetSuccess() {
			log.Printf("could not save maze on the server: %v (%v)", err, r.GetMessage())
		}
	}
	return resp, m, nil
}

// opList lists available mazes by their id
func opList() (*pb.ListMazeReply, error) {
	c := solvealgos.NewClient()

	r, err := c.ListMazes(context.Background(), &pb.ListMazeRequest{})
	if err != nil {
		log.Fatalf("could not list mazes: %v", err)
	}
	return r, nil
}

// opSolve solves the maze with mazeID, m is the *local* maze for display only
func opSolve(mazeID, clientID, solveAlgo string, m *maze.Maze, p *ml.Policy) error {
	log.Printf("in opSolve, client: %v", clientID)
	c := solvealgos.NewClient()

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
	solver.SetPolicy(p)

	log.Printf("running solver %v", solveAlgo)

	if err := solver.Solve(mazeID, clientID, in.GetFromCell(), in.GetToCell(), delay, in.GetAvailableDirections(), m); err != nil {
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

func newMaze(config *pb.MazeConfig, r *sdl.Renderer, encodedMaze string) (*maze.Maze, error) {
	m, err := maze.NewMaze(config, r)
	if err != nil {
		log.Printf("invalid maze config: %v", err)
		os.Exit(1)
	}

	m.SetEncodedString(encodedMaze)

	// create empty maze
	algo := algos.Algorithms[config.CreateAlgo]
	delay, err := time.ParseDuration(config.GenDrawDelay)
	if err != nil {
		return nil, err
	}

	running := abool.New() // no used
	running.Set()
	if err := algo.Apply(m, delay, running); err != nil {
		return nil, err
	}

	if *showLocalGUI {
		// create background texture, it is saved and re-rendered as a picture
		mTexture, err := m.MakeBGTexture()
		if err != nil {
			return nil, err
		}
		m.SetBGTexture(mTexture)
	}

	return m, nil
}

// createMaze sets up the new maze
func createMaze(config *pb.MazeConfig, encodedMaze string) (*maze.Maze, *sdl.Renderer, *sdl.Window, error) {
	log.Print("showing client's view of the maze...")

	if encodedMaze != "" {
		config.CreateAlgo = "from-encoded-string"
	} else {
		// client always starts with an full view
		config.CreateAlgo = "full"
	}

	if !algos.CheckCreateAlgo(config.CreateAlgo) {
		log.Fatalf("invalid create algorithm: %v", config.CreateAlgo)
	}

	var w *sdl.Window
	var r *sdl.Renderer
	//////////////////////////////////////////////////////////////////////////////////////////////
	// Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	if *showLocalGUI {
		// offset this window one to the right so it shows up next to the server one
		w, r = lsdl.SetupSDL(config, "Client View", 1, 0)
	}

	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Setup SDL
	//////////////////////////////////////////////////////////////////////////////////////////////
	var m *maze.Maze
	var err error
	if m, err = newMaze(config, r, encodedMaze); err != nil {
		return nil, nil, nil, err
	}
	return m, r, w, nil
}

func showMaze(m *maze.Maze, r *sdl.Renderer, w *sdl.Window) {
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

	running := abool.New()
	running.Set()

	for running.IsSet() {
		lsdl.CheckQuit(running)

		sdl.Do(func() {
			colors.SetDrawColor(colors.GetColor("gray"), r)
			r.Clear()
			m.DrawMaze(r, nil)

			for _, c := range m.ClientsSorted() {
				cell := c.CurrentLocation()
				if cell == nil {
					continue
				}
				cell.DrawCurrentLocation(r, c, nil, "")
			}

			r.Present()
			sdl.Delay(uint32(1000 / *frameRate))
		})
	}
	wd.Done()
}

func run() {
	setFlags()

	log.Printf("running: %v", *op)

	switch *op {
	case "create":
		if r, _, err := opCreate(); err != nil {
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

		if err := addClient(context.Background(), *mazeID, &pb.ClientConfig{
			SolveAlgo:              *solveAlgo,
			PathColor:              *pathColor,
			FromCell:               *fromCellStr,
			ToCell:                 *toCellStr,
			FromCellColor:          *fromCellColor,
			ToCellColor:            *toCellColor,
			ShowFromToColors:       *showFromToColors,
			VisitedCellColor:       *visitedCellColor,
			CurrentLocationColor:   *currentLocationColor,
			DrawPathLength:         *drawPathLength,
			MarkVisitedCells:       *markVisitedCells,
			NumberMarkVisitedCells: *numberMarkVisitedCells,
		}, nil, nil); err != nil {
			log.Fatalf(err.Error())
		}
	case "create_solve":
		if err := opCreateSolve(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_dp_value_iteration":
		if err := opCreateSolveMlDpValueIteration(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_dp_policy_iteration":
		if err := opCreateSolveMlDpPolicyIteration(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_mc_epsilon_greedy_control":
		if err := opCreateSolveMlMCEpsilonGreedyControl(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_mc_off_policy":
		if err := opCreateSolveMlMCOffPolicy(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_td_one_step_sarsa":
		if err := opCreateSolveMlTDOneStepSarsa(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_td_sarsa_lamba":
		if err := opCreateSolveMlTDSarsaLambda(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_ml_td_q_learning":
		if err := opCreateSolveMlTDQLearning(); err != nil {
			log.Print(err.Error())
		}
	case "create_solve_multi":
		if err := opCreateSolveMulti(); err != nil {
			log.Print(err.Error())
		}
	}

	log.Print("waiting for background draw thread...")
	wd.Wait()
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *solveAlgo == "manual" {
		err := termbox.Init()
		if err != nil {
			panic(err)
		}
		// defer termbox.Close()

	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	if *enableMonitoring {
		// go metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
		exp.Exp(metrics.DefaultRegistry)

		addr, _ := net.ResolveTCPAddr("tcp", "localhost:2003")
		go graphite.Graphite(metrics.DefaultRegistry, 10e9, "metrics", addr)
	}

	// run http server for expvars
	sock, err := net.Listen("tcp", "localhost:0")
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

	// must be like this to keep drawing functions in main thread]
	sdl.Main(run)
}
