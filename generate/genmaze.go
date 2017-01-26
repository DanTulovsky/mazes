package main

import (
	"fmt"
	"mazes/grid"
)

func main() {

	g := grid.NewGrid(10, 10)
	fmt.Printf("%v\n", g)

	for _, cell := range g.Cells() {
		fmt.Printf("%v\n", cell)
	}

}
