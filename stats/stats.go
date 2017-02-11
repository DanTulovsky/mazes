package main

import (
	"flag"
	"fmt"
	"log"
	"mazes/colors"
	"mazes/genalgos"
	"os"

	"github.com/montanaflynn/stats"
	deadlock "github.com/sasha-s/go-deadlock"

	"sort"

	"mazes/algos"

	"time"

	"mazes/solvealgos"

	"net/http"

	"mazes/maze"
	_ "net/http/pprof"
)

var (
	// algo[stat] = value
	mazeStats map[string]*algostats = make(map[string]*algostats)

	rows                    = flag.Int("r", 10, "number of rows in the maze")
	columns                 = flag.Int("c", 10, "number of rows in the maze")
	bgColor                 = flag.String("bgcolor", "white", "background color")
	wallColor               = flag.String("wall_color", "black", "wall color")
	borderColor             = flag.String("border_color", "black", "border color")
	pathColor               = flag.String("path_color", "red", "border color")
	cellWidth               = flag.Int("w", 2, "cell width")
	wallWidth               = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	pathWidth               = flag.Int("path_width", 2, "path width")
	runs                    = flag.Int("runs", 20, "number of runs")
	showGenStats            = flag.Bool("gen_stats", true, "show generator stats")
	showSolverStats         = flag.Bool("solver_stats", true, "show solver stats")
	enableDeadlockDetection = flag.Bool("enable_deadlock_detection", false, "enable deadlock detection")
)

// per gen algorithm stats
type algostats struct {
	Name       string
	Deadends   []float64
	CreateTime []float64
	Solvers    map[string]*solverstat
}

type solverstat struct {
	Name          string
	TimeToSolve   []float64 // nanoseconds
	ShortestSteps []float64 // length of shortest solution (often the same as solveSteps)
	SolveSteps    []float64 // number of steps it took to find the solution
}

// setMazeStats sets stats about the maze
func setMazeStats(g *maze.Maze, algo string) {

	mazeStats[algo].Deadends = append(mazeStats[algo].Deadends, float64(len(g.DeadEnds())))
	mazeStats[algo].CreateTime = append(mazeStats[algo].CreateTime, float64(g.CreateTime().Nanoseconds()))
}

func printGenStats() {
	cells := float64(*rows * *columns)

	// dead ends
	fmt.Println("\nDead Ends (average)")
	for _, name := range keys(algos.Algorithms) {
		deadends, _ := stats.Mean(mazeStats[name].Deadends)
		fmt.Printf("  %-25s : %6.0f / %.0f (%5.2f%%)\n", name, deadends, cells, deadends/cells*100)
	}

	// create time
	fmt.Println("\nGenerators Create Time (min / avg / max)")
	for _, name := range keys(algos.Algorithms) {
		minTime, _ := stats.Min(mazeStats[name].CreateTime)
		meanTime, _ := stats.Mean(mazeStats[name].CreateTime)
		maxTime, _ := stats.Max(mazeStats[name].CreateTime)
		fmt.Printf("  %-25s : %12v / %12v / %12v\n", name,
			time.Duration(minTime),
			time.Duration(meanTime),
			time.Duration(maxTime))
	}
}

func printSolverStats() {
	// solvers
	fmt.Println("\nSolver Stats")
	for _, genAlgoName := range keys(algos.Algorithms) {
		fmt.Printf("\n  %-25s\n", genAlgoName)

		fmt.Println("      Time to Solve (min / avg / max)")
		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			minTime, _ := stats.Min(mazeStats[genAlgoName].Solvers[solverName].TimeToSolve)
			meanTime, _ := stats.Mean(mazeStats[genAlgoName].Solvers[solverName].TimeToSolve)
			maxTime, _ := stats.Max(mazeStats[genAlgoName].Solvers[solverName].TimeToSolve)
			fmt.Printf("          %-25s : %12v / %12v / %12v\n", solverName,
				time.Duration(minTime),
				time.Duration(meanTime),
				time.Duration(maxTime))
		}

		fmt.Println("      Length of Shortest Solution (min / avg / max)")
		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			min, _ := stats.Min(mazeStats[genAlgoName].Solvers[solverName].ShortestSteps)
			mean, _ := stats.Mean(mazeStats[genAlgoName].Solvers[solverName].ShortestSteps)
			max, _ := stats.Max(mazeStats[genAlgoName].Solvers[solverName].ShortestSteps)
			fmt.Printf("          %-25s : %12v / %12v / %12v\n", solverName,
				int(min),
				int(mean),
				int(max))
		}

		fmt.Println("      Travel Steps to find Solution (min / avg / max)")
		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			min, _ := stats.Min(mazeStats[genAlgoName].Solvers[solverName].SolveSteps)
			mean, _ := stats.Mean(mazeStats[genAlgoName].Solvers[solverName].SolveSteps)
			max, _ := stats.Max(mazeStats[genAlgoName].Solvers[solverName].SolveSteps)
			fmt.Printf("          %-25s : %12v / %12v / %12v\n", solverName,
				int(min),
				int(mean),
				int(max))
		}
	}
}

func showMazeStats() {
	// general stats
	fmt.Println("\nAbout Maze")
	fmt.Printf("  Size: %d x %d\n", *rows, *columns)
	fmt.Printf("  Number of Runs: %d\n", *runs)

	if *showGenStats {
		printGenStats()
	}

	if *showSolverStats {
		printSolverStats()
	}

}

func keys(m map[string]genalgos.Algorithmer) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func solveKeys(m map[string]solvealgos.Algorithmer) []string {
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func RunAll(config *maze.Config) {
	// Loop over all algos and collect stats
	for _, genAlgoName := range keys(algos.Algorithms) {

		algo := algos.Algorithms[genAlgoName]

		if _, ok := mazeStats[genAlgoName]; !ok {
			mazeStats[genAlgoName] = &algostats{
				Name:    genAlgoName,
				Solvers: make(map[string]*solverstat),
			}
		}
		log.Printf("running (gen): %v", genAlgoName)

		m, err := maze.NewMaze(config)
		if err != nil {
			fmt.Printf("invalid config: %v", err)
			os.Exit(1)
		}

		m, err = algo.Apply(m, 0)
		if err != nil {
			log.Fatalf(err.Error())
		}
		if err := algo.CheckGrid(m); err != nil {
			log.Fatalf("maze is not valid: %v", err)
		}

		// solve using all available solvers, use longest path in maze
		_, fromCell, toCell, _ := m.LongestPath()

		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			if _, ok := mazeStats[genAlgoName].Solvers[solverName]; !ok {
				mazeStats[genAlgoName].Solvers[solverName] = &solverstat{
					Name: solverName,
				}
			}
			m.Reset()

			solver := algos.SolveAlgorithms[solverName]
			log.Printf("  running (solver): %v", solverName)
			m, err = solver.Solve(m, fromCell, toCell, 0)
			if err != nil {
				log.Fatalf("failed to run solver [%v]: %v", solverName, err)
			}

			mazeStats[genAlgoName].Solvers[solverName].Name = solverName
			mazeStats[genAlgoName].Solvers[solverName].TimeToSolve = append(mazeStats[genAlgoName].Solvers[solverName].TimeToSolve, float64(solver.SolveTime().Nanoseconds()))

			// Solution path length (not travel path)
			mazeStats[genAlgoName].Solvers[solverName].ShortestSteps = append(mazeStats[genAlgoName].Solvers[solverName].ShortestSteps, float64(solver.SolvePath().Length()))

			// Travel path length while finding the end
			mazeStats[genAlgoName].Solvers[solverName].SolveSteps = append(mazeStats[genAlgoName].Solvers[solverName].SolveSteps, float64(solver.SolveSteps()))
		}

		// shows some stats about the maze
		setMazeStats(m, genAlgoName)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if *enableDeadlockDetection {
		deadlock.Opts.Disable = false
	} else {
		deadlock.Opts.Disable = true
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// defer profile.Start().Stop()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &maze.Config{
		Rows:        *rows,
		Columns:     *columns,
		CellWidth:   *cellWidth,
		WallWidth:   *wallWidth,
		PathWidth:   *pathWidth,
		BgColor:     colors.GetColor(*bgColor),
		BorderColor: colors.GetColor(*borderColor),
		WallColor:   colors.GetColor(*wallColor),
		PathColor:   colors.GetColor(*pathColor),
	}
	//////////////////////////////////////////////////////////////////////////////////////////////
	// End Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////

	for x := 0; x < *runs; x++ {
		log.Printf("Run: %v", x)
		RunAll(config)
	}
	showMazeStats()

}
