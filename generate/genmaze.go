package main

import (
	"fmt"
	"mazes/bintree"
	"mazes/grid"
)

func main() {

	g := grid.NewGrid(10, 10)
	g = bintree.Apply(g)

	fmt.Printf("%v\n", g)

}
