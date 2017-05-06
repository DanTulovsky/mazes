// Package algos provides a all available algorithms
package algos

import (
	"mazes/genalgos"
	"mazes/genalgos/aldous_broder"
	"mazes/genalgos/bintree"
	"mazes/genalgos/hint_and_kill"
	"mazes/genalgos/kruskal"
	gen_rb "mazes/genalgos/recursive_backtracker"
	"mazes/genalgos/sidewinder"
	"mazes/genalgos/wilsons"
	"mazes/solvealgos"
	"mazes/solvealgos/dijkstra"
	"mazes/solvealgos/random"
	"mazes/solvealgos/random_unvisited"
	solve_rb "mazes/solvealgos/recursive_backtracker"
	"mazes/solvealgos/wall_follower"
)

var Algorithms map[string]genalgos.Algorithmer = map[string]genalgos.Algorithmer{
	"aldous-broder":         &aldous_broder.AldousBroder{},
	"bintree":               &bintree.Bintree{},
	"hunt-and-kill":         &hint_and_kill.HuntAndKill{},
	"randomized-kruskal":    &kruskal.Kruskal{},
	"recursive-backtracker": &gen_rb.RecursiveBacktracker{},
	"sidewinder":            &sidewinder.Sidewinder{},
	"wilsons":               &wilsons.Wilsons{},
}

var SolveAlgorithms map[string]solvealgos.Algorithmer = map[string]solvealgos.Algorithmer{
	"dijkstra":              &dijkstra.Dijkstra{},
	"random":                &random.Random{},
	"random-unvisited":      &random_unvisited.RandomUnvisited{},
	"recursive-backtracker": &solve_rb.RecursiveBacktracker{},
	"wall-follower":         &wall_follower.WallFollower{},
}
