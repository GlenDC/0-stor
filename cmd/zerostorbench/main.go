package main

import (
	"fmt"
	"log"
    "os"
    "runtime/pprof"
    _ "runtime/trace"
)

func main() {
    configFile := "config.yaml"
    profileFile := "/opt/tmp/profile"

    fmt.Println("creating client")
    client, err := getClient(configFile)
    if err != nil {
		log.Fatal(err.Error())
	}

    f, err := os.Create(profileFile)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("uploading over the client")

    // trace.Start(f)
    pprof.StartCPUProfile(f)
    upload(client)
    pprof.StopCPUProfile()
    // trace.Stop()

    fmt.Println("all done")
}
