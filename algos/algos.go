// Package algos provides a all available algorithms
package algos

import (
	"github.com/DanTulovsky/mazes/solvealgos/dijkstra"
	"log"

	"github.com/DanTulovsky/mazes/genalgos"
	"github.com/DanTulovsky/mazes/genalgos/aldous_broder"
	"github.com/DanTulovsky/mazes/genalgos/bintree"
	"github.com/DanTulovsky/mazes/genalgos/ellers"
	gen_empty "github.com/DanTulovsky/mazes/genalgos/empty"
	"github.com/DanTulovsky/mazes/genalgos/from_encoded_string"
	"github.com/DanTulovsky/mazes/genalgos/fromfile"
	"github.com/DanTulovsky/mazes/genalgos/full"
	"github.com/DanTulovsky/mazes/genalgos/hunt_and_kill"
	"github.com/DanTulovsky/mazes/genalgos/kruskal"
	"github.com/DanTulovsky/mazes/genalgos/prim"
	gen_rb "github.com/DanTulovsky/mazes/genalgos/recursive_backtracker"
	"github.com/DanTulovsky/mazes/genalgos/recursive_division"
	"github.com/DanTulovsky/mazes/genalgos/sidewinder"
	"github.com/DanTulovsky/mazes/genalgos/wilsons"
	pb "github.com/DanTulovsky/mazes/proto"
	"github.com/DanTulovsky/mazes/solvealgos"
	"github.com/DanTulovsky/mazes/solvealgos/empty"
	"github.com/DanTulovsky/mazes/solvealgos/manual"
	ml_follow_policy "github.com/DanTulovsky/mazes/solvealgos/ml/follow_policy"
	"github.com/DanTulovsky/mazes/solvealgos/ml/td/one_step_sarsa"
	"github.com/DanTulovsky/mazes/solvealgos/ml/td/sarsa_lambda"
	"github.com/DanTulovsky/mazes/solvealgos/random"
	"github.com/DanTulovsky/mazes/solvealgos/random_unvisited"
	solve_rb "github.com/DanTulovsky/mazes/solvealgos/recursive_backtracker"
	"github.com/DanTulovsky/mazes/solvealgos/wall_follower"
)

var Algorithms map[string]genalgos.Algorithmer = map[string]genalgos.Algorithmer{
	"aldous-broder":         &aldous_broder.AldousBroder{},
	"bintree":               &bintree.Bintree{},
	"ellers":                &ellers.Ellers{},
	"empty":                 &gen_empty.Empty{},
	"from-encoded-string":   &from_encoded_string.FromEncodedString{},
	"fromfile":              &fromfile.Fromfile{},
	"full":                  &full.Full{},
	"hunt-and-kill":         &hunt_and_kill.HuntAndKill{},
	"kruskal":               &kruskal.Kruskal{},
	"prim":                  &prim.Prim{},
	"recursive-backtracker": &gen_rb.RecursiveBacktracker{},
	"recursive-division":    &recursive_division.RecursiveDivision{},
	"sidewinder":            &sidewinder.Sidewinder{},
	"wilsons":               &wilsons.Wilsons{},
}

var SolveAlgorithms map[string]func() solvealgos.Algorithmer = map[string]func() solvealgos.Algorithmer{
	"dijkstra":              NewDijkstra,
	"manual":                NewManual,
	"follow-policy":         NewFollowPolicy,
	"random":                NewRandom,
	"random-unvisited":      NewRandomUnvisited,
	"recursive-backtracker": NewRecursiveBacktracker,
	"wall-follower":         NewWallFollower,
	"empty":                 NewEmpty,
	"ml-td-one-step-sarsa":  NewMLTDOneStepSarsa,
	"ml-td-sarsa-lambda":    NewMLTDSarsaLambda,
}

func NewMLTDSarsaLambda() solvealgos.Algorithmer {
	return &sarsa_lambda.MLTDSarsaLambda{}
}

func NewMLTDOneStepSarsa() solvealgos.Algorithmer {
	return &one_step_sarsa.MLTDOneStepSarsa{}
}

func NewFollowPolicy() solvealgos.Algorithmer {
	return &ml_follow_policy.MLFollowPolicy{}
}

func NewDijkstra() solvealgos.Algorithmer {
	return &dijkstra.Dijkstra{}
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

func NewManual() solvealgos.Algorithmer {
	return &manual.Manual{}
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

// checkCreateAlgo makes sure the passed in algorithm is valid
func CheckCreateAlgo(a string) bool {
	for k := range Algorithms {
		if k == a {
			return true
		}
	}
	return false
}
