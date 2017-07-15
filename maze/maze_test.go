package maze

import (
	"log"
	"mazes/utils"
	"strconv"
	"testing"

	"fmt"
	pb "mazes/proto"
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

		e := m.Encode()
		if m.rows*m.columns+m.rows != int64(len(e)) {
			t.Errorf("expected encoding of length %v, but have %v.", m.rows*m.columns+m.rows, len(e))
		}
		log.Printf("\n%v\n", e)

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

		var encoded string
		// create encoding string
		for _, row := range m.Rows() {
			for _ = range row {
				encoded = encoded + "F"
			}
			encoded = encoded + "\n"
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
		linkSouth: false,
		linkEast:  true,
		linkWest:  true,
		encoded:   "1011",
	},
}

func TestCellEncode(t *testing.T) {
	for _, tt := range cellencodetests {

		tt.c.SetNorth(nil)
		tt.c.SetSouth(nil)
		tt.c.SetEast(nil)
		tt.c.SetWest(nil)

		if tt.linkNorth {
			tt.c.Link(tt.cNorth)
			tt.c.SetNorth(tt.cNorth)
		}
		if tt.linkSouth {
			tt.c.Link(tt.cSouth)
			tt.c.SetSouth(tt.cSouth)
		}
		if tt.linkEast {
			tt.c.Link(tt.cEast)
			tt.c.SetEast(tt.cEast)
		}
		if tt.linkWest {
			tt.c.Link(tt.cWest)
			tt.c.SetWest(tt.cWest)
		}

		e := tt.c.Encode()
		expected, _ := strconv.ParseInt(tt.encoded, 2, 0)
		if e != fmt.Sprintf("%X", expected) {
			t.Errorf("expected: %X; received: %s", expected, e)
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
