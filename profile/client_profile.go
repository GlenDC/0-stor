package main

import (
	"flag"

	"github.com/pkg/profile"
)

var (
	profileOutDir = flag.String("profileOutDir", "", "")
	profileMode   = flag.String("profile.mode", "", "enable profiling mode, one of [cpu, mem, trace, block]")
)

func main() {
	flag.Parse()

	switch *profileMode {
	case "cpu":
		defer profile.Start(profile.ProfilePath(*profileOutDir), profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.ProfilePath(*profileOutDir), profile.MemProfile).Stop()
	case "trace":
		defer profile.Start(profile.ProfilePath(*profileOutDir), profile.TraceProfile).Stop()
	case "block":
		defer profile.Start(profile.ProfilePath(*profileOutDir), profile.BlockProfile).Stop()
	default:
		return
	}

	// code for client
}
