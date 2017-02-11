package maze

import (
	"mazes/utils"
	"testing"
)

var mazecreatetests = []struct {
	config  *Config
	wantErr bool
}{
	{
		config: &Config{
			Rows:    utils.Random(5, 40),
			Columns: utils.Random(8, 33),
		},
		wantErr: false,
	}, {
		config: &Config{
			Rows:    10,
			Columns: 15,
		},
		wantErr: false,
	}, {
		config: &Config{
			Rows:    55,
			Columns: 4,
		},
		wantErr: false,
	}, {
		config: &Config{
			Rows:    0,
			Columns: 0,
		},
		wantErr: true,
	}, {
		config: &Config{
			Rows:    -3,
			Columns: -3,
		},
		wantErr: true,
	}, {
		config:  &Config{},
		wantErr: true,
	},
}

var mazecreatefromimagetests = []struct {
	config  *Config
	image   string
	wantErr bool
}{
	{
		config:  &Config{},
		image:   "../masks/maze_text.png",
		wantErr: false,
	}, {
		config:  &Config{},
		image:   "../masks/fail1.png",
		wantErr: true,
	},
}

func TestNewMaze(t *testing.T) {

	for _, tt := range mazecreatetests {
		m, err := NewMaze(tt.config)

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

func TestNewMazeFromImage(t *testing.T) {

	for _, tt := range mazecreatefromimagetests {
		m, err := NewMazeFromImage(tt.config, tt.image)

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

func BenchmarkNewMaze(b *testing.B) {
	config := &Config{
		Rows:    10,
		Columns: 10,
	}

	_, err := NewMaze(config)
	if err != nil {
		b.Errorf("invalid config: %v", err)
	}

	for i := 0; i < b.N; i++ {
		NewMaze(config)
	}

}
