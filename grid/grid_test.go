package grid

import "testing"

var gridcreatetests = []struct {
	config *Config
}{
	{config: &Config{
		Rows:    10,
		Columns: 10,
	},
	},
	{config: &Config{
		Rows:    10,
		Columns: 15,
	},
	},
}

func TestNewGrid(t *testing.T) {

	for _, tt := range gridcreatetests {
		g := NewGrid(tt.config)
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

	NewGrid(config)

	for i := 0; i < b.N; i++ {
		NewGrid(config)
	}

}
