// Package sdl provides helper functions for using sdl
package sdl

import (
	"fmt"
	"log"
	"os"

	"github.com/veandco/go-sdl2/sdl"

	pb "mazes/proto"

	"github.com/tevino/abool"
)

// SetupSDL initializes SDL and returns the window and renderer object
func SetupSDL(config *pb.MazeConfig, w *sdl.Window, r *sdl.Renderer, winTitle string) (*sdl.Window, *sdl.Renderer) {
	log.Print("setting up sdl window and renderer")
	if !config.GetGui() {
		return nil, nil
	}
	sdl.Do(func() {
		sdl.Init(sdl.INIT_EVERYTHING)
		sdl.EnableScreenSaver()
	})

	// window
	winWidth := int((config.Columns)*config.CellWidth + config.WallWidth*2)
	winHeight := int((config.Rows)*config.CellWidth + config.WallWidth*2)

	var err error
	sdl.Do(func() {
		w, err = sdl.CreateWindow(winTitle, 0, 0,
			// TODO(dan): consider sdl.WINDOW_ALLOW_HIGHDPI; https://goo.gl/k9Ak0B
			winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL)
	})
	if err != nil {
		log.Printf("Failed to create window: %s\n", err)
		os.Exit(1)
	}

	// renderer
	sdl.Do(func() {
		r, err = sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	// Set options
	// https://wiki.libsdl.org/SDL_SetRenderDrawBlendMode#blendMode
	sdl.Do(func() {
		r.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	})

	sdl.Do(func() {
		r.Clear()
	})

	log.Print("done SDL setup")
	return w, r
}

func CheckQuit(running *abool.AtomicBool) {
	sdl.Do(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				log.Print("received quit request, exiting...")
				running.UnSet()
			}
		}
	})
}
