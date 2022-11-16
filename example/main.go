package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/elri/config"
)

// -- Configuration struct
type Configuration struct {
	Welcome string `yaml:"welcome" toml:"welcome"`
	Hints   bool   `yaml:"hints" toml:"hints"`

	Limits struct {
		Min int `yaml:"min" toml:"min"`
		Max int `yaml:"max" toml:"max"`
	} `yaml:"limits" toml:"limits"`

	Correct int `yaml:"answer" toml:"answer"`
	Guess   int `yaml:"guess" toml:"guess"`
}

var (
	_         = flag.String("welcome", "Hello and Welcome to 'Guess a Number'", "welcome phrase")
	_         = flag.Bool("hints", false, "give hints like 'higher' or 'lower'")
	fConfFile = flag.String("config", "", "config file")
	_         = flag.Int("correct", -1, "correct guess (is set at runtime if not by flag)")
	_         = flag.Int("min", 0, "range min")
	_         = flag.Int("max", 100, "range max")
	_         = flag.Int("guess", 5, "number of guesses")
)

//print out defaults to file
//correct needs a default made from min & max

func main() {
	config.SetDefaultFile("default_conf.yml")
	config.ParseFlags()

	cfg, err := getConfig()
	if err != nil {
		panic(err)
	}
	fmt.Println(cfg.Welcome)

	var guess int
	for i := 0; i < cfg.Guess; i++ {
		fmt.Printf("Guess a number between %d and %d >  ", cfg.Limits.Min, cfg.Limits.Max)
		_, err = fmt.Scan(&guess)
		if err == nil {
			if guess == cfg.Correct {
				fmt.Println("Correct! :D")
				os.Exit(0)
			}
			if cfg.Hints {
				higher := cfg.Correct > guess

				if higher {
					fmt.Println("Incorrect. Try a higher number.")
				} else {
					fmt.Println("Incorrect. Try a lower number.")
				}
			} else {
				fmt.Println("Incorrect. Try again.")
			}
		}
	}

}

func getConfig() (*Configuration, error) {
	cfg := new(Configuration)
	err := config.SetUpConfigurationWithConfigFile(cfg, *fConfFile)
	if err == nil {
		min := cfg.Limits.Min
		max := cfg.Limits.Max
		if cfg.Correct == -1 {
			cfg.Correct = min + rand.Intn(2)*(max-min)
		}
	}

	return cfg, err
}
