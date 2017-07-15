package maze

import (
	"strconv"
	"testing"

	"fmt"
	pb "mazes/proto"
	"mazes/utils"
)

var mazecreatetests = []struct {
	config  *pb.MazeConfig
	wantErr bool
}{
	{
		config: &pb.MazeConfig{
			Rows:    int64(utils.Random(5, 40)),
			Columns: int64(utils.Random(8, 33)),
		},
		wantErr: false,
	}, {
		config: &pb.MazeConfig{
			Rows:    10,
			Columns: 15,
		},
		wantErr: false,
	}, {
		config: &pb.MazeConfig{
			Rows:    55,
			Columns: 4,
		},
		wantErr: false,
	}, {
		config: &pb.MazeConfig{
			Rows:    0,
			Columns: 0,
		},
		wantErr: true,
	}, {
		config: &pb.MazeConfig{
			Rows:    -3,
			Columns: -3,
		},
		wantErr: true,
	}, {
		config:  &pb.MazeConfig{},
		wantErr: true,
	},
	{
		config: &pb.MazeConfig{
			Rows:    4,
			Columns: 3,
		},
		wantErr: false,
	},
}

var mazecreatefromimagetests = []struct {
	config  *pb.MazeConfig
	image   string
	wantErr bool
}{
	{
		config:  &pb.MazeConfig{},
		image:   "../masks/maze_text.png",
		wantErr: false,
	}, {
		config:  &pb.MazeConfig{},
		image:   "../masks/fail1.png",
		wantErr: true,
	},
}

func TestNewMaze(t *testing.T) {

	for _, tt := range mazecreatetests {
		m, err := NewMaze(tt.config, nil)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}

		}

		if m.Size() != tt.config.Rows*tt.config.Columns {
			t.Errorf("Expected size [%v], but have [%v]", tt.config.Rows*tt.config.Columns, m.Size())
		}

	}
}

func TestEncode(t *testing.T) {
	for _, tt := range mazecreatetests {
		m, err := NewMaze(tt.config, nil)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}

		}

		e, err := m.Encode()
		if err != nil {
			t.Errorf("failed to encode maze: %v", err)
		}
		if m.rows*m.columns+m.rows != int64(len(e)) {
			t.Errorf("expected encoding of length %v, but have %v.", m.rows*m.columns+m.rows, len(e))
		}
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range mazecreatetests {
		m, err := NewMaze(tt.config, nil)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}

		}

		encoded, err := m.Encode()
		if err != nil {
			t.Errorf("failed to encode maze: %v", err)
		}

		if err := m.Decode(encoded); err != nil {
			t.Errorf("error decoding: %v", err)
		}
	}
}

func TestNewMazeFromImage(t *testing.T) {

	for _, tt := range mazecreatefromimagetests {
		m, err := NewMazeFromImage(tt.config, tt.image, nil)

		if err != nil {
			if !tt.wantErr {
				t.Fatalf("unable to create maze from image (%v): %v", tt.image, err)
			} else {
				continue // skip the rest of the tests
			}

		}

		if m.Size() != tt.config.Rows*tt.config.Columns {
			t.Errorf("Expected size [%v], but have [%v]", tt.config.Rows*tt.config.Columns, m.Size())
		}
	}
}

var cellencodetests = []struct {
	config    *pb.MazeConfig
	c         *Cell
	cNorth    *Cell
	cSouth    *Cell
	cEast     *Cell
	cWest     *Cell
	linkNorth bool
	linkSouth bool
	linkEast  bool
	linkWest  bool
	encoded   string
}{
	{
		config:    &pb.MazeConfig{},
		c:         NewCell(5, 5, 0, &pb.MazeConfig{}),
		cNorth:    NewCell(5, 4, 0, &pb.MazeConfig{}),
		cEast:     NewCell(6, 5, 0, &pb.MazeConfig{}),
		cSouth:    NewCell(5, 6, 0, &pb.MazeConfig{}),
		cWest:     NewCell(4, 5, 0, &pb.MazeConfig{}),
		linkNorth: true,
		linkSouth: true,
		linkEast:  true,
		linkWest:  true,
		encoded:   "1111",
	}, {
		config:    &pb.MazeConfig{},
		c:         NewCell(5, 5, 0, &pb.MazeConfig{}),
		cNorth:    NewCell(5, 4, 0, &pb.MazeConfig{}),
		cEast:     NewCell(6, 5, 0, &pb.MazeConfig{}),
		cSouth:    NewCell(5, 6, 0, &pb.MazeConfig{}),
		cWest:     NewCell(4, 5, 0, &pb.MazeConfig{}),
		linkNorth: true,
		linkSouth: true,
		linkEast:  false,
		linkWest:  false,
		encoded:   "1100",
	}, {
		config:    &pb.MazeConfig{},
		c:         NewCell(5, 5, 0, &pb.MazeConfig{}),
		cNorth:    NewCell(5, 4, 0, &pb.MazeConfig{}),
		cEast:     NewCell(6, 5, 0, &pb.MazeConfig{}),
		cSouth:    NewCell(5, 6, 0, &pb.MazeConfig{}),
		cWest:     NewCell(4, 5, 0, &pb.MazeConfig{}),
		linkNorth: true,
		linkSouth: false,
		linkEast:  true,
		linkWest:  true,
		encoded:   "1011",
	},
}

func TestCellEncode(t *testing.T) {
	for _, tt := range cellencodetests {

		if tt.linkNorth {
			if err := tt.c.Link(tt.cNorth); err != nil {
				t.Error(err)
			}
			tt.c.SetNorth(tt.cNorth)
		}
		if tt.linkSouth {
			if err := tt.c.Link(tt.cSouth); err != nil {
				t.Error(err)
			}
			tt.c.SetSouth(tt.cSouth)
		}
		if tt.linkEast {
			if err := tt.c.Link(tt.cEast); err != nil {
				t.Error(err)
			}
			tt.c.SetEast(tt.cEast)
		}
		if tt.linkWest {
			if err := tt.c.Link(tt.cWest); err != nil {
				t.Error(err)
			}
			tt.c.SetWest(tt.cWest)
		}

		e := tt.c.Encode()
		expected, _ := strconv.ParseInt(tt.encoded, 2, 0)
		if e != fmt.Sprintf("%X", expected) {
			t.Errorf("expected: %X; received: %s", expected, e)
		}
	}
}

func TestCellDecode(t *testing.T) {
	for _, tt := range cellencodetests {

		tt.c.SetNorth(tt.cNorth)
		tt.c.SetSouth(tt.cSouth)
		tt.c.SetEast(tt.cEast)
		tt.c.SetWest(tt.cWest)

		// the test uses binary string representation for human understanding
		i, _ := strconv.ParseInt(tt.encoded, 2, 0)
		if err := tt.c.Decode(fmt.Sprintf("%X", i)); err != nil {
			t.Errorf("error decoding into cell: %v: i: %b (%X)", err, i, i)
		}

		if tt.linkNorth {
			if !tt.c.Linked(tt.c.North()) {
				t.Errorf("expected link north, but did not have it: i: %b (%X)", i, i)
			}
		}
		if tt.linkSouth {
			if !tt.c.Linked(tt.c.South()) {
				t.Errorf("expected link south, but did not have it: i: %b (%X)", i, i)
			}
		}
		if tt.linkEast {
			if !tt.c.Linked(tt.c.East()) {
				t.Errorf("expected link east, but did not have it: i: %b (%X)", i, i)
			}
		}
		if tt.linkWest {
			if !tt.c.Linked(tt.c.West()) {
				t.Errorf("expected link west, but did not have it: i: %b (%X)", i, i)
			}
		}
	}
}
func BenchmarkNewMaze(b *testing.B) {
	config := &pb.MazeConfig{
		Rows:    10,
		Columns: 10,
	}

	_, err := NewMaze(config, nil)
	if err != nil {
		b.Errorf("invalid config: %v", err)
	}

	for i := 0; i < b.N; i++ {
		NewMaze(config, nil)
	}

}
