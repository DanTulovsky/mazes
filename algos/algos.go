// Package algos provides a all available algorithms
package algos

import (
	"mazes/genalgos"
	"mazes/genalgos/aldous_broder"
	"mazes/genalgos/bintree"
	"mazes/genalgos/hint_and_kill"
	"mazes/genalgos/recursive_backtracker"
	"mazes/genalgos/sidewinder"
	"mazes/genalgos/wilsons"
)

var Algorithms map[string]genalgos.Algorithmer = map[string]genalgos.Algorithmer{
	"aldous-broder":         &aldous_broder.AldousBroder{},
	"bintree":               &bintree.Bintree{},
	"hunt-and-kill":         &hint_and_kill.HuntAndKill{},
	"recursive-backtracker": &recursive_backtracker.RecursiveBacktracker{},
	"sidewinder":            &sidewinder.Sidewinder{},
	"wilsons":               &wilsons.Wilsons{},
}
