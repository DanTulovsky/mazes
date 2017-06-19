// Package algos provides a all available algorithms
package algos

import (
	"log"

	"mazes/genalgos"
	"mazes/genalgos/aldous_broder"
	"mazes/genalgos/bintree"
	"mazes/genalgos/ellers"
	"mazes/genalgos/hunt_and_kill"
	"mazes/genalgos/kruskal"
	"mazes/genalgos/prim"
	gen_rb "mazes/genalgos/recursive_backtracker"
	"mazes/genalgos/recursive_division"
	"mazes/genalgos/sidewinder"
	"mazes/genalgos/wilsons"
	pb "mazes/proto"
	"mazes/solvealgos"
	"mazes/solvealgos/empty"
	"mazes/solvealgos/random"
	"mazes/solvealgos/random_unvisited"
	solve_rb "mazes/solvealgos/recursive_backtracker"
	"mazes/solvealgos/wall_follower"
)

var Algorithms map[string]genalgos.Algorithmer = map[string]genalgos.Algorithmer{
	"aldous-broder":         &aldous_broder.AldousBroder{},
	"bintree":               &bintree.Bintree{},
	"ellers":                &ellers.Ellers{},
	"hunt-and-kill":         &hunt_and_kill.HuntAndKill{},
	"kruskal":               &kruskal.Kruskal{},
	"prim":                  &prim.Prim{},
	"recursive-backtracker": &gen_rb.RecursiveBacktracker{},
	"recursive-division":    &recursive_division.RecursiveDivision{},
	"sidewinder":            &sidewinder.Sidewinder{},
	"wilsons":               &wilsons.Wilsons{},
}

//var SolveAlgorithms map[string]solvealgos.Algorithmer = map[string]solvealgos.Algorithmer{
//	//"dijkstra":              &dijkstra.Dijkstra{},
//	//"manual":                &manual.Manual{},
//	"random":                &random.Random{},
//	"random-unvisited":      &random_unvisited.RandomUnvisited{},
//	"recursive-backtracker": &solve_rb.RecursiveBacktracker{},
//	"wall-follower":         &wall_follower.WallFollower{},
//	"empty":                 &empty.Empty{},
//}

var SolveAlgorithms map[string]func() solvealgos.Algorithmer = map[string]func() solvealgos.Algorithmer{
	//"dijkstra":              &dijkstra.Dijkstra{},
	//"manual":                &manual.Manual{},
	"random":                NewRandom,
	"random-unvisited":      NewRandomUnvisited,
	"recursive-backtracker": NewRecursiveBacktracker,
	"wall-follower":         NewWallFollower,
	"empty":                 NewEmpty,
}

func NewEmpty() solvealgos.Algorithmer {
	return &empty.Empty{}
}

func NewWallFollower() solvealgos.Algorithmer {
	return &wall_follower.WallFollower{}
}

func NewRandom() solvealgos.Algorithmer {
	return &random.Random{}
}

func NewRandomUnvisited() solvealgos.Algorithmer {
	return &random_unvisited.RandomUnvisited{}
}

func NewRecursiveBacktracker() solvealgos.Algorithmer {
	return &solve_rb.RecursiveBacktracker{}
}

// NewSolver returns a new solver
func NewSolver(n string, stream pb.Mazer_SolveMazeClient) solvealgos.Algorithmer {
	af, ok := SolveAlgorithms[n]
	if !ok {
		log.Fatalf("invalid solve algorithm: %s", n)
	}

	a := af()
	a.SetStream(stream)

	return a
}

// NewGenerator returns a new generator
func NewGenerator(n string) genalgos.Algorithmer {
	a := Algorithms[n]
	return a
}
