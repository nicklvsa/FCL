package main

import (
	"fcl/parser"
	"flag"
	"fmt"
)

func main() {
	var inputFile string

	flag.StringVar(&inputFile, "input", "", "-input <config.fcl>")
	flag.Parse()

	config, err := parser.ParseInput(inputFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("VERSION: %s\n", config.Version)
	fmt.Printf("Shared global scripts?: %+v\n", config.ScriptData.Shared)
}
