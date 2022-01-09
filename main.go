package main

import (
	"flag"
	"fmt"
	"project-finder/process"
	"project-finder/shared"
)

func main() {
	var rootDir string
	var chunkSize int
	var dig bool

	flag.BoolVar(&dig, "dig", false, "--dig")
	flag.StringVar(&rootDir, "root", "", "--root {root directory}")
	flag.IntVar(&chunkSize, "chunk-size", 3, "--chunk-size {size}")

	flag.Parse()

	if ok := shared.ValidateCLI(rootDir, chunkSize, dig); !ok {
		panic("Please check your input arguments and try again!")
	}

	proc := process.NewProcessor(chunkSize, rootDir, dig)

	results, err := proc.Begin()
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		fmt.Printf("Project: %s\n", result.FullPath)
		if result.Info != nil {
			fmt.Printf("Author: %s\n", result.Info.Author)
			fmt.Printf("CreatedAt: %s\n", result.Info.CreatedAt.String())
		}
		fmt.Printf("\n\n")
	}
}
