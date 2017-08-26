package mc

import (
	"testing"

	"mazes/maze"
	"mazes/ml"

	"github.com/tevino/abool"

	"reflect"

	pb "mazes/proto"
)

var policytests = []struct {
	actions      []int
	config       *pb.MazeConfig
	clientConfig *pb.ClientConfig
	df           float64 // prefer more recent steps when calculating value (1 = prefer all)
	theta        float64 // when delta is smaller, eval stops
	clientID     string
}{
	{
		config: &pb.MazeConfig{
			Columns:    3,
			Rows:       3,
			CreateAlgo: "empty",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",               // doesn't matter
			ToCell:    "1,1",
		},
		clientID: "client-empty-dp-eval-only",
		df:       0.99,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	},
	//{
	//	config: &pb.MazeConfig{
	//		Columns:    3,
	//		Rows:       2,
	//		CreateAlgo: "full", // no passages, with df=1 does not converge
	//	},
	//	clientConfig: &pb.ClientConfig{
	//		SolveAlgo: "ml_dp_policy_eval", // no op yet
	//		FromCell:  "0,0",               // doesn't matter
	//		ToCell:    "2,1",
	//	},
	//	clientID: "client-full-dp-eval-only",
	//	df:       0.5,
	//	theta:    0.00001,
	//	actions:  ml.DefaultActions,
	//},
	{
		config: &pb.MazeConfig{
			Columns:    5,
			Rows:       4,
			CreateAlgo: "ellers",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "2,1",
		},
		clientID: "client-ellers-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	},
	{
		config: &pb.MazeConfig{
			Columns:    6,
			Rows:       3,
			CreateAlgo: "bintree",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "2,1",
		},
		clientID: "client-bintree-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	}, {
		config: &pb.MazeConfig{
			Columns:    4,
			Rows:       5,
			CreateAlgo: "prim",
		},
		clientConfig: &pb.ClientConfig{
			SolveAlgo: "ml_dp_policy_eval", // no op yet
			FromCell:  "0,0",
			ToCell:    "3,4",
		},
		clientID: "client-prim-dp-eval-only",
		df:       0.9,
		theta:    0.00001,
		actions:  ml.DefaultActions,
	},
}

func TestPolicy_Eval(t *testing.T) {
	for _, tt := range policytests {
		t.Logf("running maze size (%v, %v): %v (-> (%v)", tt.config.Columns, tt.config.Rows, tt.config.CreateAlgo, tt.clientConfig.ToCell)
		// create empty maze
		m, err := maze.NewMaze(tt.config, nil)
		if err != nil {
			t.Fatalf("error creating maze: %v", err)
		}

		// apply any algorithm to it
		algo := ml.Algorithms[tt.config.CreateAlgo]
		generating := abool.New()
		generating.Set()
		if err := algo.Apply(m, 0, generating); err != nil {
			generating.UnSet()
			t.Fatalf("error applying algorithm: %v", err)
		}
		// required to get the toCell
		_, toCell, err := m.AddClient(tt.clientID, tt.clientConfig)
		if err != nil {
			t.Fatalf("failed to add client: %v", err)
		}

		encoded, err := m.Encode()

		p := ml.NewRandomPolicy(int(tt.config.Rows*tt.config.Columns), tt.actions)
		numEpisodes := int64(1000)
		maxSteps := int64(10000) // max steps per run through maze
		vf, err := Evaluate(p, m, tt.clientID, numEpisodes, tt.df, toCell, maxSteps)
		if err != nil {
			t.Fatalf("error evaluating policy: %v", err)
		}
		t.Logf("maze:\n%v\nvalue function (%v):\n%v", encoded, tt.clientID, vf.Reshape(int(tt.config.Rows), int(tt.config.Columns)))

	}
}

var statesepisodetest = []struct {
	e        episode
	expected []int
}{
	{
		e:        episode{},
		expected: []int{},
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
			},
		},
		expected: []int{0, 1, 3},
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
				{1, ml.North, -1},
			},
		},
		expected: []int{0, 1, 3},
	},
}

func Test_statesInEpisode(t *testing.T) {
	for _, tt := range statesepisodetest {
		states := statesInEpisode(tt.e)

		if !reflect.DeepEqual(tt.expected, states) {
			t.Errorf("expected: %v; returned: %v", tt.expected, states)
		}
	}
}

var statesepisodeindextest = []struct {
	e        episode
	s        int
	expected int
	wantErr  bool
}{
	{
		e:        episode{},
		s:        0,
		expected: -1,
		wantErr:  true,
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
			},
		},
		s:        1,
		expected: 1,
		wantErr:  false,
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
				{1, ml.North, -1},
			},
		},
		s:        3,
		expected: 2,
		wantErr:  false,
	},
}

func Test_firstStateInEpisodeIndex(t *testing.T) {
	for _, tt := range statesepisodeindextest {
		idx, err := firstStateInEpisodeIdx(tt.e, tt.s)
		if err != nil {
			if tt.wantErr {
				continue
			}
			t.Errorf("dit not expect error, received: %v", err)
			continue
		}

		if idx != tt.expected {
			t.Errorf("expected: %v; returned: %v", tt.expected, idx)
		}
	}
}

var sumrewardsinceidxtest = []struct {
	e        episode
	idx      int
	df       float64
	expected float64
	wantErr  bool
}{
	{
		e:        episode{},
		df:       1,
		idx:      0,
		expected: -1,
		wantErr:  true,
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
			},
		},
		idx:      1,
		df:       1,
		expected: -2,
		wantErr:  false,
	},
	{
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.North, -1},
				{1, ml.North, -1},
			},
		},
		idx:      1,
		df:       1,
		expected: -3,
		wantErr:  false,
	},
}

func Test_sumRewardsSinceIdx(t *testing.T) {
	for _, tt := range sumrewardsinceidxtest {
		sum, err := sumRewardsSinceIdx(tt.e, tt.idx, tt.df)
		if err != nil {
			if tt.wantErr {
				continue
			}
			t.Errorf("dit not expect error, received: %v", err)
			continue
		}

		if sum != tt.expected {
			t.Errorf("expected: %v; returned: %v; e: %v", tt.expected, sum, tt.e)
		}
	}
}

var stateactionsepisodetest = []struct {
	e        episode
	expected []StateAction
}{
	{
		e:        episode{},
		expected: []StateAction{},
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.South, -1},
				{3, ml.North, -1},
			},
		},
		expected: []StateAction{
			{0, ml.North},
			{1, ml.South},
			{3, ml.North},
		},
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.South, -1},
				{3, ml.East, -1},
				{1, ml.South, -1},
			},
		},
		expected: []StateAction{
			{0, ml.North},
			{1, ml.South},
			{3, ml.East},
		},
	},
}

func Test_stateActionsInEpisode(t *testing.T) {
	for _, tt := range stateactionsepisodetest {
		stateActions := stateActionsInEpisode(tt.e)

		if !reflect.DeepEqual(tt.expected, stateActions) {
			t.Errorf("expected: %#v; returned: %#v; episode: %#v", tt.expected, stateActions, tt.e)
		}
	}
}

var stateactionsepisodeindextest = []struct {
	e        episode
	s        int
	a        int
	expected int
	wantErr  bool
}{
	{
		e:        episode{},
		s:        0,
		a:        ml.North,
		expected: -1,
		wantErr:  true,
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.East, -1},
			},
		},
		s:        1,
		a:        ml.North,
		expected: 1,
		wantErr:  false,
	}, {
		e: episode{
			sr: []stateReturn{
				{0, ml.North, -1},
				{1, ml.North, -1},
				{3, ml.East, -1},
				{1, ml.North, -1},
			},
		},
		s:        3,
		a:        ml.East,
		expected: 2,
		wantErr:  false,
	},
}

func Test_firstStateActionInEpisodeIdx(t *testing.T) {
	for _, tt := range stateactionsepisodeindextest {
		idx, err := firstStateActionInEpisodeIdx(tt.e, tt.s, tt.a, 0)
		if err != nil {
			if tt.wantErr {
				continue
			}
			t.Errorf("dit not expect error, received: %v", err)
			continue
		}

		if idx != tt.expected {
			t.Errorf("expected: %v; returned: %v", tt.expected, idx)
		}
	}
}
