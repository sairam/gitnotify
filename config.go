package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

func loadConfig(configFile string) (*config, error) {

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := ioutil.ReadFile(configFile)

	if os.IsNotExist(err) {
		return nil, nil
	}

	conf := &config{}

	err = yaml.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

// read config from .s3deploy.yml if found.
type config struct {
	Repos []*repo `yaml:"repos"`
}

func (c *config) String() string {
	arr := make([]string, len(c.Repos))
	for i, repo := range c.Repos {
		arr[i] = fmt.Sprint(repo)
	}
	return strings.Join(arr, "\n")
}

type repo struct {
	Repo            string       `yaml:"repo"`
	NamedReferences []*reference `yaml:"commits"`
	Branches        bool         `yaml:"new_branches"`
	Tags            bool         `yaml:"new_tags"`
}

func (r *repo) String() string {
	return fmt.Sprintf("repo: %s, references: %v, branches: %t, tags: %t", r.Repo, r.NamedReferences, r.Branches, r.Tags)
}

type reference string

func (x reference) String() string { return fmt.Sprintf("<%s>", string(x)) }
