package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/zero-os/0-stor/benchmark/server/config"
)

func main() {
	configPath := flag.String("cfg", "config.yaml", "YAML config file location for the benchmark")
	outputPath := flag.String("o", "out.txt", "Output file of the benchmarking results")
	flag.Parse()

	file, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := config.FromReader(file)
	if err != nil {
		log.Fatal(err)
	}

	_ = cfg

	// create grpc client

	// run benchmark

	// output results
	fmt.Println(outputPath)
}
