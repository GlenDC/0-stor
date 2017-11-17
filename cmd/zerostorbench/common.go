package main

import (
	"fmt"
	"os"

	"github.com/zero-os/0-stor/client"
)

func getClient(configFile string) (*client.Client, error) {
	policy, err := readPolicy(configFile)
	if err != nil {
		return nil, err
	}
	// create client
	cl, err := client.New(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return cl, nil
}

func readPolicy(configFile string) (client.Policy, error) {
	// read config
	f, err := os.Open(configFile)
	if err != nil {
		return client.Policy{}, err
	}

	return client.NewPolicyFromReader(f)
}
