// Generate a large number of random mazes and save them to disk
package main

import (
	"flag"
	"log"

	"mazes/algos"
	"mazes/maze"

	"os"

	pb "mazes/proto"

	"fmt"
	"io/ioutil"
	"path"
	"reflect"

	"mazes/utils"

	"sync"

	"github.com/sasha-s/go-deadlock"
	"github.com/tevino/abool"
)

var (
	numMazes  = flag.Int("num_mazes", 10, "generate this number of random mazes")
	outputDir = flag.String("output_dir", "/tmp/mazes", "output directory for mazes")
)

func newMazeConfig(createAlgo string, rows, columns int64) *pb.MazeConfig {
	config := &pb.MazeConfig{
		Rows:               rows,
		Columns:            columns,
		AllowWeaving:       false,
		WeavingProbability: 0,
		SkipGridCheck:      true,
		CreateAlgo:         createAlgo,
		BraidProbability:   0,
	}
	return config
}

func createMaze(config *pb.MazeConfig) (*maze.Maze, error) {

	if !algos.CheckCreateAlgo(config.CreateAlgo) {
		log.Fatalf("invalid create algorithm: %v", config.CreateAlgo)
	}

	m, err := maze.NewMaze(config, nil)
	if err != nil {
		log.Printf("invalid maze config: %v", err)
		os.Exit(1)
	}

	// create empty maze
	algo := algos.Algorithms[config.CreateAlgo]

	running := abool.New() // not used
	running.Set()
	// apply generator
	if err := algo.Apply(m, 0, running); err != nil {
		return nil, err
	}

	return m, nil
}

func cleanupAlgos(algos []reflect.Value) []string {
	var allAlgos []string
	var r []string

	for _, a := range algos {
		allAlgos = append(allAlgos, a.String())
	}

	excluded := []string{"aldous-broder", "empty", "from-encoded-string", "fromfile", "full"}
	for _, a := range allAlgos {
		if utils.StrInList(excluded, a) {
			continue
		}
		r = append(r, a)
	}
	return r
}

func writeMaze(e string, count int) error {
	filename := path.Join(*outputDir, fmt.Sprintf("%d", count))
	return ioutil.WriteFile(filename, []byte(e), 0644)
}

func makeOneMaze(algos []string, rows, columns int64, count int) {
	createAlgo := algos[utils.Random(0, len(algos))]

	log.Printf("Generating maze %d using %s", count, createAlgo)

	config := newMazeConfig(createAlgo, rows, columns)

	m, err := createMaze(config)
	if err != nil {
		log.Fatalf("failed to create maze: %v", err)
	}

	e, err := m.Encode()
	if err != nil {
		log.Fatalf("failed to encode maze to ascii: %v", err)
	}

	if err := writeMaze(e, count); err != nil {
		log.Fatalf("error writing file: %v", err)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	deadlock.Opts.Disable = true

	al := cleanupAlgos(reflect.ValueOf(algos.Algorithms).MapKeys())

	var rows, columns int64
	rows, columns = 10, 20

	log.Printf("output dir: %v", *outputDir)

	maxGoroutines := 100000
	guard := make(chan struct{}, maxGoroutines)
	var wg sync.WaitGroup

	for i := 0; i < *numMazes; i++ {
		guard <- struct{}{} // would block if guard channel is already filled
		wg.Add(1)
		go func(algos []string, rows, columns int64, i int) {
			makeOneMaze(algos, rows, columns, i)
			<-guard
		}(al, rows, columns, i)
	}
	wg.Wait()
}
