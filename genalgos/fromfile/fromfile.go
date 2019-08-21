// Package full creates a grid with all walls in place
package fromfile

import (
	"time"

	"github.com/DanTulovsky/mazes/genalgos"
	"github.com/DanTulovsky/mazes/maze"

	"flag"

	"path"

	"io/ioutil"

	"github.com/tevino/abool"

	"bufio"
	pb "github.com/DanTulovsky/mazes/proto"
	"os"
)

var (
	SavedMazePath = flag.String("saved_maze_path", "", "path where exported mazes are stored")
)

type Fromfile struct {
	genalgos.Common
}

// MazeSizeFromFile returns the number of columns and rows in the input file
func MazeSizeFromFile(config *pb.MazeConfig) (c, r int, err error) {
	filename := path.Join(*SavedMazePath, config.GetFromFile())

	file, _ := os.Open(filename)
	fileScanner := bufio.NewScanner(file)
	lineCount := 0
	colCount := 0

	for fileScanner.Scan() {
		if lineCount == 0 {
			// grab row length
			colCount = len(fileScanner.Text())
		}
		lineCount++
	}

	return colCount, lineCount, nil
}

// Apply reads in the provided file and sets up the passages
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
