package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dobYml, _  = time.Parse("2006-01-02 15:04:05", "1987-07-07 07:47:00")
	dobToml, _ = time.Parse("2006-01-02 15:04:05", "1985-05-05 05:45:00")
	dobJson, _ = time.Parse("2006-01-02 15:04:05", "1981-01-01 01:41:00")
)

//TODO
type InnerTestConfig struct {
	Pim  string   `toml:"pim" yaml:"pim" json:"pim"`
	Age  int      `toml:"age" yaml:"age" json:"age"`
	Cats []string `toml:"cats" yaml:"cats" json:"cats"`
}

type TestConfig struct {
	//InnerTestConfig
	Pim        string    `toml:"pim" yaml:"pim" json:"pim"`
	Age        int       `toml:"age" yaml:"age" json:"age"`
	Cats       []string  `toml:"cats" yaml:"cats" json:"cats"`
	Pi         float64   `toml:"pi" yaml:"pi" json:"pi"`
	Perfection []int     `toml:"perfection" yaml:"perfection" json:"perfection"`
	Dreams     bool      `toml:"dreams" yaml:"dreams" json:"dreams"`
	DOB        time.Time `toml:"dob" yaml:"dob" json:"dob"`
	Piglet     struct {
		Name string `toml:"name" yaml:"name" json:"name"`
		Age  int    `toml:"age" yaml:"age" json:"age"`
	} `toml:"piglet" yaml:"piglet" json:"piglet"`
}

func fullTestConfigToml() *TestConfig {
	cfg := new(TestConfig)
	cfg.Dreams = true
	cfg.Pi = 3.14
	cfg.Perfection = []int{6, 28, 496, 8128}
	cfg.DOB = dobToml
	cfg.Pim = "sweet candy"
	cfg.Age = 25
	cfg.Cats = []string{"Pella", "Hjördis"}
	cfg.Piglet.Name = "Yim"
	cfg.Piglet.Age = 10
	return cfg
}

func fullTestConfigYml() *TestConfig {
	cfg := new(TestConfig)
	cfg.Dreams = true
	cfg.Pi = 3.1415
	cfg.Perfection = []int{8128, 496, 28, 6}
	cfg.DOB = dobYml
	cfg.Pim = "sour candy"
	cfg.Age = 27
	cfg.Cats = []string{"Kajsa", "Meja"}
	cfg.Piglet.Name = "Milt"
	cfg.Piglet.Age = 5
	return cfg
}

func fullTestConfigJson() *TestConfig {
	cfg := new(TestConfig)
	cfg.Dreams = true
	cfg.Pi = 3.14159
	cfg.Perfection = []int{1, 1, 2, 3, 5}
	cfg.DOB = dobJson
	cfg.Pim = "salmiak"
	cfg.Age = 2
	cfg.Cats = []string{"Lucifer", "Felix"}
	cfg.Piglet.Name = "Noef"
	cfg.Piglet.Age = 5
	return cfg
}

func partialYmlOverwritesToml() *TestConfig {
	fullYml := fullTestConfigYml()
	fullToml := fullTestConfigToml()
	cfg := new(TestConfig)
	cfg.Pi = fullToml.Pi
	cfg.Dreams = fullToml.Dreams
	cfg.Perfection = fullToml.Perfection
	cfg.DOB = fullYml.DOB
	cfg.Piglet = fullToml.Piglet
	cfg.Pim = fullYml.Pim
	cfg.Age = fullYml.Age
	cfg.Cats = fullYml.Cats
	return cfg
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
			name:           "default config is json file",
			defaultFile:    "test/test.json",
			expectedConfig: fullTestConfigJson(),
		},
		{
			name:        "default config is incorrect file",
			defaultFile: "test/test.fake",
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
	fullJson := fullTestConfigJson()
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
			configFile:  "test.fake",
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
			name:           "Given config file is json (no default)",
			configFile:     "test.json",
			expectedConfig: fullJson,
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
	var err error
	var parsed *TestConfig

	tests := []struct {
		name          string
		cfg           *TestConfig
		filename      string
		expectedError error
	}{
		{
			name:     "Encode TOML",
			cfg:      fullTestConfigToml(),
			filename: "test/wtest.toml",
		},
		{
			name:     "Encode YAML",
			cfg:      fullTestConfigYml(),
			filename: "test/wtest.yml",
		},
		{
			name:     "Encode JSON",
			cfg:      fullTestConfigJson(),
			filename: "test/wtest.json",
		},
		{
			name:          "Fail to encode invalid file",
			filename:      "test/test.fake",
			expectedError: ErrInvalidConfigFile,
		},
		{
			name:          "Non-existing file",
			filename:      "test/none.json",
			expectedError: ErrNoFileFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed = new(TestConfig)
			_, err = encode(tt.cfg, tt.filename)
			if tt.expectedError == nil {
				err = ParseConfigFile(parsed, tt.filename)
				assert.Equal(t, ErrNoDefaultConfig, err)
			} else {
				assert.NotNil(t, err)
				assert.True(t, strings.Contains(err.Error(), tt.expectedError.Error()))
			}

			if tt.cfg != nil {
				assert.Equal(t, *tt.cfg, *parsed)
			}
		})
	}
}

type TestConf struct {
	Version string
}

func newTestConf(v string) *TestConf {
	return &TestConf{Version: v}
}

func Test_writeToDefaultFile(t *testing.T) {

	// Redirect stdin & stdout
	r, w, err := os.Pipe()
	assert.Nil(t, err)
	origStdout := os.Stdout
	os.Stdout = w
	origStdin := os.Stdin
	os.Stdin = r

	// Set default file
	var f *os.File
	filepath := "test/wdefault.yml"
	SetDefaultFile(filepath)
	defer func() {
		err = os.Remove(filepath)
		assert.Nil(t, err)
	}()

	// Test values
	initConf := newTestConf("init")
	fillEmptyConf := newTestConf("fill empty")
	onlyShowConf := newTestConf("only show")
	overWriteConf := newTestConf("overwritten")

	tests := []struct {
		name           string
		inCfg          *TestConf
		input          string
		expectedCfg    *TestConf
		expectedOutput string
		purgeFile      bool
	}{
		{
			name:        "Default file does not exist: create and fill it",
			inCfg:       initConf,
			expectedCfg: initConf,
			purgeFile:   true,
		},
		{
			name:        "Default file exists but is empty: fill it",
			inCfg:       fillEmptyConf,
			expectedCfg: fillEmptyConf,
		},
		{
			name:           "Default file exists with values: only show",
			inCfg:          onlyShowConf,
			expectedCfg:    fillEmptyConf,
			expectedOutput: "CONTENTS OF '" + filepath + "':\n" + String(fillEmptyConf) + "\n",
			input:          "s\n",
		},
		{
			name:        "Default file exists with values: abort",
			inCfg:       overWriteConf,
			expectedCfg: fillEmptyConf,
			input:       "n\n",
		},
		{
			name:        "Default file exists with values: overwrite",
			inCfg:       overWriteConf,
			expectedCfg: overWriteConf,
			input:       "y\n",
		},
		{
			name:           "Default file exists with values: faulty input",
			inCfg:          new(TestConf),
			input:          "123\n",
			expectedOutput: "faulty input",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != "" {
				_, err = w.Write([]byte(tt.input))
				assert.Nil(t, err)
			}

			err = writeToDefaultFile(tt.inCfg)
			assert.Nil(t, err)

			if tt.expectedCfg != nil {
				f, err = os.OpenFile(filepath, os.O_RDWR, 0644)
				cfg := new(TestConf)
				err = decode(cfg, f, filepath)
				assert.Nil(t, err)
				assert.Equal(t, *tt.expectedCfg, *cfg)
			}

			if tt.expectedOutput != "" {
				output := make([]byte, 1024)
				_, err := r.Read(output)
				assert.Nil(t, err)

				//fp := fmt.Sprintf("output=\n%d %s\n--\nexpected: %d %s", len(string(output)), string(output), len(tt.expectedOutput), tt.expectedOutput)
				//origStdout.WriteString(fp)

				assert.True(t, strings.Contains(string(output), tt.expectedOutput))
			} else {
				//empty buffer
				buf := make([]byte, 1024)
				_, err := r.Read(buf)
				assert.Nil(t, err)
			}

			if tt.purgeFile {
				err = os.Truncate(filepath, 0)
				assert.Nil(t, err)
			}
		})
	}

	// Restore
	os.Stdout = origStdout
	os.Stdin = origStdin
}
