package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Auth `yaml:"auth"`
	Repo `yaml:"repo"`
}

type Auth struct {
	Token string `yaml:"personal_access_token"`
	Login string `yaml:"login"`
}

type Repo struct {
	Name  string `yaml:"name"`
	Owner string `yaml:"owner"`
}

func GetConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	return &cfg, err
}

func (r Repo) String() string {
	return fmt.Sprintf("%v/%v", r.Owner, r.Name)
}
