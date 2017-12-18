package client

import (
	"io/ioutil"

	"github.com/zero-os/0-stor/client/itsyouonline"
	"github.com/zero-os/0-stor/client/pipeline"
	yaml "gopkg.in/yaml.v2"
)

// TODO: handle configuration using https://github.com/spf13/viper

// ReadConfig reads the configuration from a file.
// NOTE that it isn't validated, this will be done automatically,
// when you use the config to create a 0-stor client.
func ReadConfig(path string) (Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config

	// for now we only support YAML
	if yaml.Unmarshal(bytes, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Config defines the configuration of the 0-stor client.
// It configures everything from namespaces, permissions,
// storage clusters, as well as the entire read/write pipeline,
// used to read and write data.
type Config struct {
	// IYO defines the IYO (itsyou.online) configuration,
	// used for this 0-stor client.
	IYO itsyouonline.Config `yaml:"iyo" json:"iyo"`
	// Namespace defines the label (ID of namespace),
	// to be used for all read/write/delete operations.
	Namespace string `yaml:"namespace" json:"namespace"`

	// DataStor defines the configuration for the zstordb data shards (servers),
	// at least one zstordb shard is given, but more might be required,
	// if you define a distributed storage configuration in the pipeline config.
	DataStor DataStorConfig `yaml:"datastor" json:"datastor"`

	// MetaStor defines the configuration for the metadata shards (servers).
	// For now only an ETCD cluster is supported using this config.
	MetaStor MetaStorConfig `yaml:"metastor" json:"metastor"`

	// Pipeline defines the object read/write pipeline configuration
	// for this 0-stor client. It defines how to structure,
	// process, identify and store all data to be written,
	// and that same configuration is required to read the data back.
	Pipeline pipeline.Config `yaml:"pipeline" json:"pipeline"`
}

// DataStorConfig is used to configure a zstordb cluster.
type DataStorConfig struct {
	Shards []string `yaml:"shards" json:"shards"`
}

// MetaStorConfig is used to configure the metastor client.
// TODO: remove this fixed/non-dynamic struct
type MetaStorConfig struct {
	Shards []string `yaml:"shards" json:"shards"`
}
