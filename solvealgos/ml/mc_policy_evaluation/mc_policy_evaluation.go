package mc_policy_evaluation

import (
	"mazes/maze"
	"mazes/solvealgos"
	"time"

	"math"
	"mazes/ml/dp"
	pb "mazes/proto"
	"mazes/utils"

	"github.com/buger/goterm"
)

func printProgress(e, numEpisodes int) {
	goterm.Clear()
	if math.Mod(float64(e), 1000) == 0 {
		goterm.Printf("\nEpisode %d of %d\n", e, numEpisodes)
		goterm.Flush()
	}
}

type episode struct {
	state  int
	action int
	reward float64
}

type MCPolicyEvaluation struct {
	solvealgos.Common
}

func (a *MCPolicyEvaluation) Solve(mazeID, clientID string, fromCell, toCell *pb.MazeLocation, delay time.Duration,
	directions []*pb.Direction, m *maze.Maze) error {
	defer solvealgos.TimeTrack(a, time.Now())

	numStates := int(m.Config().Columns * m.Config().Rows)
	numEpisodes := 1000
	// each action has the same weight
	policy := dp.NewRandomPolicy(numStates, dp.DefaultActions)

	// map of state -> sum of all returns in that state
	returnsSum := make(map[int]float64)
	// map of state -> count of all returns
	returnsCount := make(map[int]float64)

	vf := dp.NewValueFunction(numStates)
	episodes := []episode{}

	// run through the policy this many times
	// each run is a walk through the maze until the end (or limit?)
	for e := 1; e <= numEpisodes; e++ {
		printProgress(e, numEpisodes)

		// generate an episode (wonder through the maze following policy)
		// An episode is an array of (state, action, reward) tuples

		// pick a random state to start at (fromCell and action)
		state := utils.Random(0, numStates)
		fromCell, err := utils.LocationFromState(m.Config().Rows, m.Config().Columns, int64(state))
		if err != nil {
			return nil, err
		}

		// get the action, according to policy, for this state
		action := policy.BestRandomActionsForState(state)

	}

	return nil
}
