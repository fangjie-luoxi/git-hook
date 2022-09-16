package config

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func init() {
	cfg := config{
		Port: "11000",
	}
	Config = cfg
}

func Setup(path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(bytes, &Config); err != nil {
		return err
	}
	return nil
}

var Config config

type config struct {
	Port string `yaml:"port"`
	Git  struct {
		Token      string `yaml:"token"`
		Branch     string `yaml:"branch"`
		Dockerfile string `yaml:"dockerfile"`
	}
}
