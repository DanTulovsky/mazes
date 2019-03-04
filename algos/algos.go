// Package algos provides a all available algorithms
package algos

import (
	"log"

	"gogs.wetsnow.com/dant/mazes/genalgos"
	"gogs.wetsnow.com/dant/mazes/genalgos/aldous_broder"
	"gogs.wetsnow.com/dant/mazes/genalgos/bintree"
	"gogs.wetsnow.com/dant/mazes/genalgos/ellers"
	gen_empty "gogs.wetsnow.com/dant/mazes/genalgos/empty"
	"gogs.wetsnow.com/dant/mazes/genalgos/from_encoded_string"
	"gogs.wetsnow.com/dant/mazes/genalgos/fromfile"
	"gogs.wetsnow.com/dant/mazes/genalgos/full"
	"gogs.wetsnow.com/dant/mazes/genalgos/hunt_and_kill"
	"gogs.wetsnow.com/dant/mazes/genalgos/kruskal"
	"gogs.wetsnow.com/dant/mazes/genalgos/prim"
	gen_rb "gogs.wetsnow.com/dant/mazes/genalgos/recursive_backtracker"
	"gogs.wetsnow.com/dant/mazes/genalgos/recursive_division"
	"gogs.wetsnow.com/dant/mazes/genalgos/sidewinder"
	"gogs.wetsnow.com/dant/mazes/genalgos/wilsons"
	pb "gogs.wetsnow.com/dant/mazes/proto"
	"gogs.wetsnow.com/dant/mazes/solvealgos"
	"gogs.wetsnow.com/dant/mazes/solvealgos/empty"
	"gogs.wetsnow.com/dant/mazes/solvealgos/manual"
	ml_follow_policy "gogs.wetsnow.com/dant/mazes/solvealgos/ml/follow_policy"
	"gogs.wetsnow.com/dant/mazes/solvealgos/ml/td/one_step_sarsa"
	"gogs.wetsnow.com/dant/mazes/solvealgos/ml/td/sarsa_lambda"
	"gogs.wetsnow.com/dant/mazes/solvealgos/random"
	"gogs.wetsnow.com/dant/mazes/solvealgos/random_unvisited"
	solve_rb "gogs.wetsnow.com/dant/mazes/solvealgos/recursive_backtracker"
	"gogs.wetsnow.com/dant/mazes/solvealgos/wall_follower"
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
	//"dijkstra":              &dijkstra.Dijkstra{},
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
