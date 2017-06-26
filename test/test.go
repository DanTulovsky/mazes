package main

import (
	"fmt"
	"log"
	"os"

	"sync"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_gfx"
)

func setupSDL(n int) (*sdl.Window, *sdl.Renderer) {
	log.Print("setting up sdl window and renderer")

	w := new(sdl.Window)
	r := new(sdl.Renderer)

	sdl.Do(func() {
		if sdl.WasInit(0) == 0 {
			if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
				fmt.Fprintf(os.Stderr, "init error: %v", err)
				os.Exit(1)
			}
			sdl.EnableScreenSaver()
		}

		winWidth := 300
		winHeight := 300

		var err error

		w, err = sdl.CreateWindow(fmt.Sprintf("win_%d", n), n*winWidth, n*winHeight,
			winWidth, winHeight, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL|sdl.WINDOW_RESIZABLE)
		if err != nil {
			log.Printf("Failed to create window: %s\n", err)
			os.Exit(1)
		}

		// renderer
		log.Printf("window: %#v", w.GetID())
		r, err = sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC|sdl.RENDERER_TARGETTEXTURE)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
			os.Exit(1)
		}
	})

	log.Print("done SDL setup")
	return w, r
}

func win(n int) {
	w, r := setupSDL(n)
	defer func() {
		sdl.Do(func() {
			r.Destroy()
			w.Destroy()
		})
	}()

	log.Printf("window pointer (%d): %p", n, w)
	log.Printf("renderer pointer (%d): %p", n, r)

	x, _ := w.GetRenderer()
	log.Printf("window renderer pointer (%d): %p", n, x)

	for {

		sdl.Do(func() {
			if err := r.Clear(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to clear: %s\n", err)
				os.Exit(1)
			}

			if e := gfx.StringRGBA(r, int(0), int(0), fmt.Sprint("testing"), 255, 255, 255, 255); e != true {
				log.Printf("error (%d): %v", n, sdl.GetError())
			}

			r.Present()
			sdl.Delay(uint32(1000 / 120))
		})
	}
}

func run() {
	var wd sync.WaitGroup

	wd.Add(1)
	go win(1)

	wd.Add(1)
	go win(2)

	wd.Wait()

}

func main() {
	sdl.Main(run)
}
