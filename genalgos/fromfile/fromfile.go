// Package full creates a grid with all walls in place
package fromfile

import (
	"time"

	"mazes/genalgos"
	"mazes/maze"

	"flag"

	"path"

	"io/ioutil"

	"github.com/tevino/abool"
)

var (
	SavedMazePath = flag.String("saved_maze_path", "", "path where exported mazes are stored")
)

type Fromfile struct {
	genalgos.Common
}

// Apply reads in the provided file sets up the passages
func (a *Fromfile) Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error {

	defer genalgos.TimeTrack(m, time.Now())

	filename := path.Join(*SavedMazePath, m.Config().FromFile)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := m.Decode(string(data)); err != nil {
		return err
	}

	a.Cleanup(m)
	return nil
}

func (a *Fromfile) CheckGrid(m *maze.Maze) error {
	return nil
}
