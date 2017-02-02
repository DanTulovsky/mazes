package grid

import (
	"mazes/utils"
	"testing"
)

var gridcreatetests = []struct {
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

func TestNewGrid(t *testing.T) {

	for _, tt := range gridcreatetests {
		g, err := NewGrid(tt.config)

		if err != nil {
			if !tt.wantErr {
				t.Errorf("invalid config: %v", err)
			} else {
				continue // skip the rest of the tests
			}

		}

		if g.Size() != tt.config.Rows*tt.config.Columns {
			t.Errorf("Expected size [%v], but have [%v]", tt.config.Rows*tt.config.Columns, g.Size())
		}
	}
}

func BenchmarkNewGrid(b *testing.B) {
	config := &Config{
		Rows:    10,
		Columns: 10,
	}

	_, err := NewGrid(config)
	if err != nil {
		b.Errorf("invalid config: %v", err)
	}

	for i := 0; i < b.N; i++ {
		NewGrid(config)
	}

}
