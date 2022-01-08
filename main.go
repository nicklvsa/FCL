package main

import (
	"fcl/parser"
	"fcl/shared"
	"flag"
	"fmt"
	"strings"
)

func main() {
	var inputFile string

	flag.StringVar(&inputFile, "input", "", "-input <config.fcl>")
	flag.Parse()

	if ok, errs := shared.ValidateArgs(inputFile); !ok {
		fmt.Println("The following errors occurred: ")
		panic(strings.Join(errs, ", "))
	}

	config, err := parser.ParseInput(inputFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("VERSION: %s\n", config.Version)
	fmt.Printf("Shared global scripts?: %+v\n", config.ScriptData.Shared)
}
