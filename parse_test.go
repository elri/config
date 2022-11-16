package config

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO
/*type InnerTestConfig struct {
	Pim  string   `toml:"pim" yaml:"pim"`
	Age  int      `toml:"age" yaml:"age"`
	Cats []string `toml:"cats" yaml:"cats"`
}*/

type TestConfig struct {
	Pim        string    `toml:"pim" yaml:"pim"`
	Age        int       `toml:"age" yaml:"age"`
	Cats       []string  `toml:"cats" yaml:"cats"`
	Pi         float64   `toml:"pi" yaml:"pi"`
	Perfection []int     `toml:"perfection" yaml:"perfection"`
	Dreams     bool      `toml:"dreams" yaml:"dreams"`
	DOB        time.Time `toml:"dob" yaml:"dob"`
	Piglet     struct {
		Name string `toml:"name" yaml:"name"`
		Age  int    `toml:"age" yaml:"age"`
	} `toml:"piglet" yaml:"piglet"`
}

func fullTestConfigToml() *TestConfig {
	dob, _ := time.Parse("2006-01-02 15:04:05", "1987-07-05 05:45:00")
	fullToml := TestConfig{
		Dreams:     true,
		Pi:         3.14,
		Perfection: []int{6, 28, 496, 8128},
		DOB:        dob,
	}
	fullToml.Pim = "sweet candy"
	fullToml.Age = 25
	fullToml.Cats = []string{"Pella", "Hjördis"}
	fullToml.Piglet.Name = "Yim"
	fullToml.Piglet.Age = 10
	return &fullToml
}

func fullTestConfigYml() *TestConfig {
	dob, _ := time.Parse("2006-01-02 15:04:05", "1987-07-07 07:47:00")
	fullYml := TestConfig{
		Dreams:     true,
		Pi:         3.1415,
		Perfection: []int{8128, 496, 28, 6},
		DOB:        dob,
	}
	fullYml.Pim = "sour candy"
	fullYml.Age = 27
	fullYml.Cats = []string{"Kajsa", "Meja"}
	fullYml.Piglet.Name = "Milt"
	fullYml.Piglet.Age = 5
	return &fullYml
}

func partialYmlOverwritesToml() *TestConfig {
	fullYml := fullTestConfigYml()
	fullToml := fullTestConfigToml()
	partialYmlOverwritesToml := &TestConfig{
		Pi:         fullToml.Pi,
		Dreams:     fullToml.Dreams,
		Perfection: fullToml.Perfection,
		DOB:        fullYml.DOB,
		Piglet:     fullToml.Piglet,
	}
	partialYmlOverwritesToml.Pim = fullYml.Pim
	partialYmlOverwritesToml.Age = fullYml.Age
	partialYmlOverwritesToml.Cats = fullYml.Cats
	return partialYmlOverwritesToml
}

func Test_SetAndGetDefaultFIle(t *testing.T) {
	set := "test/test.yml"
	SetDefaultFile(set)
	get := GetDefaultFile()
	assert.Equal(t, set, get)
}

func Test_ParseDefaultConfigFile(t *testing.T) {
	var err error
	var cfg *TestConfig
	tests := []struct {
		name           string
		defaultFile    string
		expectedConfig *TestConfig
		expectedErr    error
	}{
		{
			name:        "No default config",
			expectedErr: ErrNoDefaultConfig,
		},
		{
			name:           "default config is toml file",
			defaultFile:    "test/test.toml",
			expectedConfig: fullTestConfigToml(),
		},
		{
			name:           "default config is yaml file",
			defaultFile:    "test/test.yml",
			expectedConfig: fullTestConfigYml(),
		},
		{
			name:        "default config is incorrect file",
			defaultFile: "test/test.json",
			expectedErr: ErrInvalidConfigFile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = new(TestConfig)
			SetDefaultFile(tt.defaultFile)

			err = ParseDefaultConfigFile(cfg)
			if tt.expectedErr != nil {
				assert.NotNil(t, err)
				if err != nil {
					assert.True(t, strings.Contains(err.Error(), tt.expectedErr.Error()))
				}
			} else {
				assert.Nil(t, err)
			}
			if tt.expectedConfig != nil {
				assert.Equal(t, *tt.expectedConfig, *cfg)
			}

		})
	}

}

func Test_ParseConfigFile(t *testing.T) {
	var err error
	var cfg *TestConfig

	fullYml := fullTestConfigYml()
	fullToml := fullTestConfigToml()
	partialYmlOverwritesToml := partialYmlOverwritesToml()

	tests := []struct {
		name           string
		defaultFile    string
		configFile     string
		expectedConfig *TestConfig
		expectedErr    error
	}{
		{
			name:        "No config file given",
			expectedErr: ErrNoConfigFileToParse,
		},
		{
			name:        "Given config file do not exist",
			configFile:  "fake.toml",
			expectedErr: ErrNoFileFound,
		},
		{
			name:        "Given config file is of type that cannot be handled",
			configFile:  "test.json",
			expectedErr: ErrInvalidConfigFile,
		},
		{
			name:           "Given config file is yml (no default)",
			configFile:     "test.yml",
			expectedConfig: fullYml,
		},
		{
			name:           "Given config file is toml (no default)",
			configFile:     "test.toml",
			expectedConfig: fullToml,
		},
		{
			name:           "Given config file overwrites default completely",
			defaultFile:    "test.toml",
			configFile:     "test.yml",
			expectedConfig: fullYml,
		},
		{
			name:           "Given config file overwrites default partially",
			defaultFile:    "test.toml",
			configFile:     "test_partial.yml",
			expectedConfig: partialYmlOverwritesToml,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg = new(TestConfig)
			if tt.defaultFile != "" {
				SetDefaultFile("test/" + tt.defaultFile)
			}
			err = ParseConfigFile(cfg, tt.configFile, "test")
			if tt.expectedErr != nil {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedErr.Error()))
			}
			if tt.expectedConfig != nil {
				assert.Equal(t, *tt.expectedConfig, *cfg)
			}

		})
	}
}

func Test_ParseConfigTypes(t *testing.T) {
	var err error
	var i int
	var m map[string]interface{}

	SetDefaultFile(DEFAULT_TEST_CONFIG)

	tests := []struct {
		name  string
		cfg   interface{}
		valid bool
	}{
		{
			name: "Simple type (int)",
			cfg:  &i,
		},
		{
			name:  "Complex type (map)",
			cfg:   &m,
			valid: true,
		},
		{
			name: "Not a pointer",
			cfg:  TestConfig{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = ParseDefaultConfigFile(tt.cfg)
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}

			err = ParseConfigFile(tt.cfg, "test/test.yml")
			if tt.valid {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}

		})
	}

	//reset default config
	SetDefaultFile("")
}

func Test_encode(t *testing.T) {
	ymlCfg := fullTestConfigYml()

	buf, err := encode(ymlCfg, "test/wtest.yml")
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	tomlCfg := fullTestConfigToml()
	buf, err = encode(tomlCfg, "test/wtest.toml")
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	buf, err = encode(tomlCfg, "fakefile.json")
	assert.NotNil(t, err)
	assert.Equal(t, &bytes.Buffer{}, buf)
}
