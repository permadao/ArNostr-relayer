package relayer

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type RelayConfig struct {
	ArweaveConfig `yaml:"arweave"`
}

type ArweaveConfig struct {
	Enable    bool `yaml:"enable"`
	BatchSize int8 `yaml:"batchSize"`
}

func NewConfig() (*RelayConfig, error) {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}
	var conf RelayConfig
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
