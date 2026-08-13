package main

import (
	"os"

	"github.com/hackelia-micrantha/calathea-community/internal/application"
)

func main() {
	os.Exit(application.Run(os.Stdout, os.Stderr, os.Args[1:]))
}
