package main

import (
	"fmt"

	"github.com/elri/config"
)

type SuperConfiguration struct {
	Welcome string `yaml:"welcome" toml:"welcome"`

	Owner struct {
		Name string `yaml:"name" toml:"name"`
	} `yaml:"owner" toml:"owner"`

	Type string `yaml:"type" toml:"type"`
}

type Bottle struct {
	Name       string `yaml:"name" toml:"name"`
	Age        int    `yaml:"age" toml:"age"`
	Distillery string `yaml:"distillery" toml:"distillery"`
	District   string `yaml:"district" toml:"disctric"`
	Country    string `yaml:"country" toml:"country"`
	Tasty      bool   `yaml:"tasty" toml:"tasty"`
}

func (b *Bottle) String() string {
	return config.StringIgnoreZeroValues(b)
}

type Configuration struct {
	SuperConfiguration

	Bottles []Bottle `yaml:"bottles" toml:"bottles"`
}

func main() {
	config.SetDefaultFile("default_conf.yml")

	cfg := new(Configuration)
	err := config.SetUpConfiguration(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Welcome)
	fmt.Println("Now printing", cfg.Owner.Name+"'s bottles of "+cfg.Type)
	for _, b := range cfg.Bottles {
		fmt.Println(b.String())
	}

}
