package from_encoded_string

import (
	"gogs.wetsnow.com/dant/mazes/genalgos"
	"gogs.wetsnow.com/dant/mazes/maze"
	"time"

	"github.com/tevino/abool"
)

type FromEncodedString struct {
	genalgos.Common
}

// Apply reads in the provided file and sets up the passages
func (a *FromEncodedString) Apply(m *maze.Maze, delay time.Duration, generating *abool.AtomicBool) error {

	defer genalgos.TimeTrack(m, time.Now())

	encoded := m.EncodedString()
	if err := m.Decode(encoded); err != nil {
		return err
	}

	a.Cleanup(m)
	return nil
}

func (a *FromEncodedString) CheckGrid(m *maze.Maze) error {
	return nil
}
