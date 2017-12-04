package config

import (
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Config represents a 0-stor server benchmarker config
type Config struct {
	// Servers represents the addresses of 0-stor servers that can be used for testing
	Servers []string `yaml:"servers"`

	// Methodes defines which methodes should be benchmarked
	Methodes []string `yaml:"methodes"`

	// KeySize represents the size of the 0-stor keys in bytes
	KeySize int `yaml:"key_size"`

	// ValueSize represents the size of the 0-stor Value in bytes
	ValueSize int `yaml:"value_size"`

	// GRPCVersion represents the API version of the 0-stor server
	GRPCVersion string `yaml:"grpc_version"`

	// JWTToken represents the JWT token used for authenticating to the server
	JWTToken string `yaml:"jwt_token"`

	// Count represents how many times a method should be called for the benchmark
	Count int `yaml:"count"`

	// Duration represents the time in milliseconds of how long the test should run for
	Duration int `yaml:"duration"`
}

// Validate validates the config
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if len(c.Servers) == 0 {
		return fmt.Errorf("no 0-stor servers were provided")
	}

	if len(c.Methodes) == 0 {
		return fmt.Errorf("no benchmarking methodes were provided")
	}

	return nil
}

// FromReader returns a Config from a reader
func FromReader(r io.Reader) (*Config, error) {
	c := &Config{}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// unmarshal
	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}

	// validate
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}
