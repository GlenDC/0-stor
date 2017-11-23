package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/zero-os/0-stor/client"
	"github.com/zero-os/0-stor/proxy"
)

func main() {
	var (
		config  string
		address string
	)
	flag.StringVar(&config, "config", "./config.yaml", "configuration file")
	flag.StringVar(&address, "address", ":12121", "listen address")

	flag.Parse()

	f, err := os.Open(config)
	if err != nil {
		log.Fatalf("failed to open config: %v", err)
	}
	defer f.Close()

	pol, err := client.NewPolicyFromReader(f)
	if err != nil {
		log.Fatalf("failed to read policy: %v", err)
	}

	proxy, err := proxy.New(pol)
	if err != nil {
		log.Fatalf("failed to create proxy: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	_, err = proxy.Listen(address)
	if err != nil {
		log.Fatalf("proxy failed to listen: %v", err)
	}

	log.Infof("ztorproxy listening at %v", address)

	<-sigChan
}
