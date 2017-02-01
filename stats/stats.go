package main

import (
	"flag"
	"fmt"
	"log"
	"mazes/colors"
	"mazes/genalgos"
	"mazes/grid"
	"os"

	"sort"

	"mazes/algos"

	"time"

	"mazes/solvealgos"

	"net/http"

	"github.com/montanaflynn/stats"

	_ "net/http/pprof"
)

var (
	// algo[stat] = value
	mazeStats map[string]map[string][]float64 = make(map[string]map[string][]float64)

	rows            = flag.Int("r", 20, "number of rows in the maze")
	columns         = flag.Int("c", 20, "number of rows in the maze")
	bgColor         = flag.String("bgcolor", "white", "background color")
	wallColor       = flag.String("wall_color", "black", "wall color")
	borderColor     = flag.String("border_color", "black", "border color")
	pathColor       = flag.String("path_color", "red", "border color")
	cellWidth       = flag.Int("w", 2, "cell width")
	wallWidth       = flag.Int("wall_width", 2, "wall width (min of 2 to have walls - half on each side")
	pathWidth       = flag.Int("path_width", 2, "path width")
	runs            = flag.Int("runs", 20, "number of runs")
	showGenStats    = flag.Bool("gen_stats", true, "show generator stats")
	showSolverStats = flag.Bool("solver_stats", true, "show solver stats")
)

// setMazeStats sets stats about the maze
func setMazeStats(g *grid.Grid, algo string) {
	mazeStats[algo]["deadends"] = append(mazeStats[algo]["deadends"], float64(len(g.DeadEnds())))
	mazeStats[algo]["createtime"] = append(mazeStats[algo]["createtime"], float64(g.CreateTime().Nanoseconds()))
}

func printGenStats() {
	cells := float64(*rows * *columns)

	// dead ends
	fmt.Println("\nDead Ends (average)")
	for _, name := range keys(algos.Algorithms) {
		deadends, _ := stats.Mean(mazeStats[name]["deadends"])
		fmt.Printf("  %-25s : %6.2f / %.0f (%5.2f%%)\n", name, deadends, cells, deadends/cells*100)
	}

	// create time
	fmt.Println("\nGenerators (average create time)")
	for _, name := range keys(algos.Algorithms) {
		ctime, _ := stats.Mean(mazeStats[name]["createtime"])
		t := time.Duration(ctime)
		fmt.Printf("  %-25s : %6v\n", name, t)
	}
}

func printSolverStats() {
	// solvers
	fmt.Println("\nSolver Stats")
	for _, name := range keys(algos.Algorithms) {
		fmt.Printf("\n  %-25s\n", name)
		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			fmt.Println("      Time to Solve (average)")
			key := fmt.Sprintf("%v_solve_time", solverName)
			t, _ := stats.Mean(mazeStats[name][key])
			fmt.Printf("          %-25s : %6v\n", solverName, time.Duration(t))

			fmt.Println("      Length of Solution (average)")
			key = fmt.Sprintf("%v_solve_path_length", solverName)
			l, _ := stats.Mean(mazeStats[name][key])
			fmt.Printf("          %-25s : %6v\n", solverName, l)

			fmt.Println("      Steps to find Solution (average)")
			key = fmt.Sprintf("%v_solve_steps", solverName)
			s, _ := stats.Mean(mazeStats[name][key])
			fmt.Printf("          %-25s : %6v\n", solverName, s)

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

func RunAll(config *grid.Config) {
	// Loop over all algos and collect stats
	for _, name := range keys(algos.Algorithms) {
		algo := algos.Algorithms[name]

		if _, ok := mazeStats[name]; !ok {
			mazeStats[name] = make(map[string][]float64)
		}
		log.Printf("running (gen): %v", name)

		g, err := grid.NewGrid(config)
		if err != nil {
			fmt.Printf("invalid config: %v", err)
			os.Exit(1)
		}

		g, err = algo.Apply(g)
		if err != nil {
			log.Fatalf(err.Error())
		}
		if err := algo.CheckGrid(g); err != nil {
			log.Fatalf("maze is not valid: %v", err)
		}

		// solve using all available solvers, use longest path in maze
		_, fromCell, toCell, _ := g.LongestPath()
		g.ResetVisited()

		for _, solverName := range solveKeys(algos.SolveAlgorithms) {
			solver := algos.SolveAlgorithms[solverName]
			log.Printf("running (solver): %v", solverName)
			g, err = solver.Solve(g, fromCell, toCell)

			key := fmt.Sprintf("%v_solve_time", solverName)
			mazeStats[name][key] = append(mazeStats[name][key], float64(solver.SolveTime().Nanoseconds()))

			key = fmt.Sprintf("%v_solve_path_length", solverName)
			mazeStats[name][key] = append(mazeStats[name][key], float64(len(solver.SolvePath())))

			key = fmt.Sprintf("%v_solve_steps", solverName)
			mazeStats[name][key] = append(mazeStats[name][key], float64(solver.SolveSteps()))

		}

		// shows some stats about the maze
		setMazeStats(g, name)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// defer profile.Start().Stop()

	//////////////////////////////////////////////////////////////////////////////////////////////
	// Configure new grid
	//////////////////////////////////////////////////////////////////////////////////////////////
	config := &grid.Config{
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
		RunAll(config)
	}
	showMazeStats()

}
