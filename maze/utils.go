package maze

import (
	"log"
	"os"
	"runtime/debug"
)

// Fail fails the process due to an unrecoverable error
func Fail(err error) {
	log.Println(err)
	debug.PrintStack()
	os.Exit(1)

}
