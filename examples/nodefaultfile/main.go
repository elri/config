package main

import (
	"flag"
	"fmt"

	"github.com/elri/config"
)

// -- Configuration struct
type Configuration struct {
	Nothing string `yaml:"nothing" toml:"nothing"`
}

var (
	_ = flag.String("nothing", "Hello World", "does nothing")
)

func main() {
	config.ParseFlags()

	cfg := new(Configuration)
	err := config.SetUpConfiguration(cfg)
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.Nothing)

}
