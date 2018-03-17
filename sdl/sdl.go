// Package sdl provides helper functions for using sdl
package sdl

import (
	"fmt"
	"log"
	"math"
	"mazes/automata/rules"
	"mazes/colors"
	"mazes/maze"
	"mazes/utils"
	"os"

	"github.com/veandco/go-sdl2/sdl"

	pb "mazes/proto"

	"github.com/tevino/abool"
)

// SetupSDL initializes SDL and returns the window and renderer object
// xOffset and yOffset are offset to position window in full windows
func SetupSDL(config *pb.MazeConfig, winTitle string, xOffset, yOffset int32) (*sdl.Window, *sdl.Renderer) {
	w := new(sdl.Window)
	r := new(sdl.Renderer)

	log.Print("setting up sdl window and renderer")
	if !config.GetGui() {
		return nil, nil
	}
	sdl.Do(func() {
		if sdl.WasInit(0) == 0 {
			if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
				fmt.Fprintf(os.Stderr, "init error: %v", err)
				os.Exit(1)
			}
			sdl.EnableScreenSaver()
		}
	})

	// window
	winWidth := int32((config.Columns)*config.CellWidth + config.WallWidth*2)
	winHeight := int32((config.Rows)*config.CellWidth + config.WallWidth*2)

	if xOffset != 0 {
		xOffset = xOffset * winWidth
	}
	if yOffset != 0 {
		yOffset = yOffset * winHeight
	}

	var err error

	sdl.Do(func() {
		w, err = sdl.CreateWindow(winTitle, xOffset, yOffset,
			// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
			winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE)
	})
	if err != nil {
		log.Printf("Failed to create window: %s\n", err)
		os.Exit(1)
	}

	// renderer
	sdl.Do(func() {
		r, err = sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC|sdl.RENDERER_TARGETTEXTURE)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	// Set options
	// https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode#blendMode
	sdl.Do(func() {
		if err := r.SetDrawBlendMode(sdl.BLENDMODE_BLEND); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
			os.Exit(1)
		}
	})

	sdl.Do(func() {
		if err := r.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clear: %s\n", err)
			os.Exit(1)
		}
	})

	log.Print("done SDL setup")
	return w, r
}

// CheckQuit catches quit events n the gui and cleansup
func CheckQuit(running *abool.AtomicBool) {
	sdl.Do(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				log.Print("received quit request, exiting...")
				running.UnSet()
			case *sdl.WindowEvent:
				e := event.(*sdl.WindowEvent)
				if e.Event == sdl.WINDOWEVENT_RESIZED {
					// TODO(imlement redraw based on this)
					log.Printf("window resized: %#v", e)
				}
			}

		}
	})
}

// swapCellColor changes the cell from live to dead, and vice versa
func swapCellColor(cell *maze.Cell) {
	if cell.BGColor() == colors.GetColor(rules.AliveColor) {
		cell.SetBGColor(colors.GetColor(rules.DeadColor))
	} else {
		cell.SetBGColor(colors.GetColor(rules.AliveColor))
	}
}

func coordsToCell(m *maze.Maze, x, y, w int64) (*maze.Cell, error) {
	c := int64(math.Floor(float64(x) / float64(w)))
	r := int64(math.Floor(float64(y) / float64(w)))

	cell, err := m.Cell(c, r, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting cell at coordinates (%v, %v)", x, y)
	}
	return cell, nil
}

// CheckEvents checks for events and updates the maze
func CheckEvents(m *maze.Maze, running *abool.AtomicBool, updateBG *abool.AtomicBool) {
	sdl.Do(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				log.Print("received quit request, exiting...")
				running.UnSet()
			case *sdl.MouseMotionEvent:
				if t.State == sdl.BUTTON_LEFT {
					cell, err := coordsToCell(m, int64(t.X), int64(t.Y), m.GetCellWidth())
					if err != nil {
						log.Print(err)
						break
					}

					if m.LastUpdatedCell() != nil && !utils.LocsSame(cell.Location(), m.LastUpdatedCell().Location()) {
						swapCellColor(cell)
					}

					m.SetLastUpdatedCell(cell)
					updateBG.Set()
				}
			case *sdl.MouseButtonEvent:
				switch t.Type {
				case sdl.MOUSEBUTTONDOWN:
					cell, err := coordsToCell(m, int64(t.X), int64(t.Y), m.GetCellWidth())
					if err != nil {
						log.Print(err)
						break
					}
					swapCellColor(cell)
					updateBG.Set()
				}
			}
		}
	})
}
