package maze

import (
	"fmt"
	"os"
)

// LocInLocList returns true if lo is in locList
func LocInLocList(l Location, locList []Location) bool {
	for _, loc := range locList {
		if l.X == loc.X && l.Y == loc.Y {
			return true
		}
	}
	return false
}

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	fmt.Println(err)
	os.Exit(1)

}
