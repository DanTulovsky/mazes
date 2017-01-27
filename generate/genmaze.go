package main

import (
	"fmt"
	"mazes/genalgos/bintree"
	"mazes/grid"
	"os"
	"runtime"

	"github.com/veandco/go-sdl2/sdl"
)

// For gui support
// brew install sdl2{,_image,_ttf,_mixer}
// go get -v github.com/veandco/go-sdl2/sdl{,_mixer,_image,_ttf}

var winTitle string = "Maze"

func main() {

	// For https://github.com/veandco/go-sdl2#faq
	runtime.LockOSThread()

	// each cell is 10 pixels?
	rows := 10
	columns := 10

	g := grid.NewGrid(rows, columns)

	// apply algorithm
	g = bintree.Apply(g)

<<<<<<< HEAD
=======
	// ascii maze
>>>>>>> 289dac017d48a153981a447bdbde1b17fa5dea6d
	fmt.Printf("%v\n", g)

	// GUI maze
	sdl.Init(sdl.INIT_EVERYTHING)

	// window
	window, err := sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		rows*100, columns*100, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		os.Exit(1)
	}
	defer window.Destroy()

	// renderer
	r, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}
	defer r.Destroy()

	r.Clear() // call this before every Present()

	r = g.Draw(r) // adds maze to render

	fmt.Print("Rendering maze...\n")
	r.Present()

L:
	for {
		event := sdl.WaitEvent()
		switch event.(type) {
		case *sdl.QuitEvent:
			break L
		}

	}

	sdl.Quit()

}
