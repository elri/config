package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	DEFAULT_TEST_CONFIG = "test/test.toml"
)

func Test_String(t *testing.T) {
	type Person struct {
		Name    string
		Surname string
		Age     int
		Adult   bool
	}

	mio := &Person{Name: "Mio", Age: 9}
	mioStr := String(mio)
	expc := "name: Mio\nsurname: \nage: 9\nadult: false\n"
	assert.Equal(t, expc, mioStr)

	mio = &Person{Name: "Mio", Age: 9, Adult: false}
	mioStr = StringIgnoreZeroValues(mio)
	expc = "name: Mio\nage: 9\nadult: false\n"
	assert.Equal(t, expc, mioStr)
}

func Test_handleError(t *testing.T) {
	/*	errStr := "handleError testing"
		err := errors.New(errStr)

		// CONTINUEONERROR
		Init(flag.ContinueOnError)
		handleError(err)
		//nothing happens

		// EXITONERROR
		Init(flag.ExitOnError)

		// intercept exit func
		oldOsExit := osExit
		defer func() { osExit = oldOsExit }()

		var got int
		myExit := func(code int) {
			got = code
		}

		osExit = myExit

		// parse conf
		handleError(err)
		if exp := 2; got != exp {
			t.Errorf("Expected exit code: %d, got: %d", exp, got)
		}
	*/
}

func Test_SetEnvsToParse(t *testing.T) {
	resetConfig()
	SetEnvPrefix("CONFTEST_")

	envs := make(map[string]string, 0)
	envs["env1"] = "text"
	envs["env2"] = "3.141595"
	envs["env3"] = "true"
	envs["env4"] = "79"

	envStrs := []string{"env1", "env2", "env3", "env4"}

	err := SetEnvsToParse(envStrs)
	assert.NotNil(t, err)
	for _, e := range envStrs {
		exp := fmt.Sprintf("could not find %s", e)
		assert.Contains(t, err.Error(), exp)
	}

	for e, v := range envs {
		envVarName := "CONFTEST_" + e
		os.Setenv(envVarName, v)
	}
	err = SetEnvsToParse(envStrs)
	assert.Nil(t, err)

	for _, env := range envStrs {
		err = os.Unsetenv("CONFTEST_" + env)
		assert.Nil(t, err)
	}
}

func Test_ConfigDefault(t *testing.T) {
	var err error

	// don't care if default file doesn't exist
	err = SetDefaultFile("doesntexist.yml")
	assert.NotNil(t, err)

	err = SetUpConfiguration(new(TestConfig))
	assert.Nil(t, err)

	// set default config file
	err = SetDefaultFile(DEFAULT_TEST_CONFIG)
	assert.Nil(t, err)

	// get config
	conf := new(TestConfig)
	err = SetUpConfiguration(conf)
	assert.Nil(t, err)

	// check config is correct
	resultConf := fullTestConfigToml()

	assert.Equal(t, resultConf, conf)
}

// NOTE: Not sure how to test env vars of type int64 or uint, is it even possible?
func Test_ConfigEnv(t *testing.T) {
	var err error

	resetConfig()

	// env setup
	SetEnvPrefix("CONFTEST_")

	envs := make(map[string]string, 0)
	envs["Pim"] = "env_pim"
	envs["Pi"] = "3.141595"
	envs["Dreams"] = "true"
	envs["Age"] = "79"

	envStrs := make([]string, 0)
	for e, v := range envs {
		envStrs = append(envStrs, e)
		envVarName := "CONFTEST_" + e
		os.Setenv(envVarName, v)
	}

	err = SetEnvsToParse(envStrs)
	assert.Nil(t, err)

	// set default config file
	err = SetDefaultFile(DEFAULT_TEST_CONFIG)
	assert.Nil(t, err)

	// get config
	conf := new(TestConfig)
	err = SetUpConfigurationWithConfigFile(conf, "test/test.yml")
	assert.Nil(t, err)

	// check config is correct
	fullYml := fullTestConfigYml()
	expected := &TestConfig{
		//Pi
		//Dreams
		Perfection: fullYml.Perfection,
		DOB:        fullYml.DOB,
		Piglet:     fullYml.Piglet,
	}
	expected.Pim = envs["Pim"]
	//Age
	expected.Cats = fullYml.Cats
	expected.Pi, err = strconv.ParseFloat(envs["Pi"], 64)
	assert.Nil(t, err)
	ageStr := envs["Age"]
	expected.Age, err = strconv.Atoi(ageStr)
	assert.Nil(t, err)
	expected.Dreams, err = strconv.ParseBool(envs["Dreams"])
	assert.Nil(t, err)

	assert.Equal(t, expected, conf)

	for _, env := range envStrs {
		err = os.Unsetenv("CONFTEST_" + env)
		assert.Nil(t, err)
	}

}

func Test_ConfigEnvFaultyVals(t *testing.T) {
	var err error

	// Redirect stdin & stdout
	var r, w *os.File
	r, w, err = os.Pipe()
	assert.Nil(t, err)
	origStdout := os.Stdout
	os.Stdout = w
	origStdin := os.Stdin
	os.Stdin = r
	defer func() {
		// Restore
		os.Stdout = origStdout
		os.Stdin = origStdin
	}()

	resetConfig()

	// env setup
	SetEnvPrefix("CONFTEST_")

	envs := make(map[string]string, 0)
	envs["Pim"] = "3.141595"
	envs["Pi"] = "env_pim"
	envs["Dreams"] = "79"
	envs["Age"] = "true"

	envStrs := make([]string, 0)
	for e, v := range envs {
		envStrs = append(envStrs, e)
		envVarName := "CONFTEST_" + e
		os.Setenv(envVarName, v)
	}

	err = SetEnvsToParse(envStrs)
	assert.Nil(t, err)

	// set default config file
	err = SetDefaultFile(DEFAULT_TEST_CONFIG)
	assert.Nil(t, err)

	// get config
	conf := new(TestConfig)
	err = SetUpConfigurationWithConfigFile(conf, "test/test.yml")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ignored")

	output := make([]byte, 1024)
	_, err = r.Read(output)
	assert.Nil(t, err)
	assert.Contains(t, string(output), "WARNING")

}

func Test_ConfigFlags(t *testing.T) {
	var err error

	flagSet := testInit()

	type TestConfigOnlyFlags struct {
		B   bool
		F64 float64
		I   int
		I64 int64
		Str string
		Ui  uint64
		D   time.Duration
	}

	fBool := flagSet.Bool(b, false, "usage")
	fFloat64 := flagSet.Float64(f64, 3.14, "usage")
	fInt := flagSet.Int(i, 10, "usage")
	fInt64 := flagSet.Int64(i64, 99, "usage")
	fStr := flagSet.String(str, "default flag string", "usage")
	fUint := flagSet.Uint64(ui, 20, "usage")
	fDuration := flagSet.Duration(d, 5*time.Second, "usage")

	setFlags := map[string]string{
		b:   "true",
		f64: "6.28",
		i:   "1",
		i64: "66",
		str: "set flag",
		ui:  "3",
		d:   "8m",
	}

	args := []string{
		"-" + b, setFlags[b],
		"-" + f64, setFlags[f64],
		"-" + i, setFlags[i],
		"-" + i64, setFlags[i64],
		"-" + str, setFlags[str],
		"-" + ui, setFlags[ui],
		"-" + d, setFlags[d],
	}
	SetFlagSetArgs(args)

	err = ParseFlags()
	assert.Nil(t, err)

	conf := new(TestConfigOnlyFlags)
	err = SetUpConfiguration(conf)
	assert.Nil(t, err)

	resultConf := &TestConfigOnlyFlags{
		B:   *fBool,
		F64: *fFloat64,
		I:   *fInt,
		I64: *fInt64,
		Str: *fStr,
		Ui:  *fUint,
		D:   *fDuration,
	}

	assert.EqualValues(t, resultConf, conf)
}

func Test_ConfigAll(t *testing.T) {
	var err error

	// flags and flagset setup
	flagSet := testInit()

	flagSet.String("pim", "just a random pim flag", "pimflag")
	flagSet.Bool("dreams", false, "none")
	flagSet.Int("age", 45, "flag age")
	flagSet.Float64("pi", 3.141592, "magic")
	flagSet.String("name", "piglet name flag", "pigflag")

	args := []string{
		"-pim", "pimflag",
		"-name", "Penny",
	}
	SetFlagSetArgs(args)

	// parse
	err = ParseFlags()
	assert.Nil(t, err)
	fmt.Println("FLAGS: ", flags, "\n DEFAULTS:", flag_defaults)

	// env setup
	SetEnvPrefix("CONFTEST_")

	envs := make(map[string]string, 0)
	envs["Pim"] = "env_pim"
	envs["Pi"] = "3.141595"

	envStrs := make([]string, 0)
	for e, v := range envs {
		envStrs = append(envStrs, e)
		envVarName := "CONFTEST_" + e
		os.Setenv(envVarName, v)
	}

	err = SetEnvsToParse(envStrs)
	assert.Nil(t, err)

	// set default config file
	err = SetDefaultFile(DEFAULT_TEST_CONFIG)
	assert.Nil(t, err)

	// get config
	conf := new(TestConfig)
	err = SetUpConfigurationWithConfigFile(conf, "test/test_partial.yml")
	assert.Nil(t, err)

	// check config is correct
	fullYml := fullTestConfigYml()
	fullToml := fullTestConfigToml()
	expected := &TestConfig{
		//Pi
		Dreams:     fullToml.Dreams,     //from default config
		Perfection: fullToml.Perfection, //from default config
		DOB:        fullYml.DOB,         //from given config
		Piglet:     fullToml.Piglet,     //from default config
	}
	expected.Pim = "pimflag"                            //from flag
	expected.Age = fullYml.Age                          //from given config
	expected.Cats = fullYml.Cats                        //from given config
	expected.Pi, _ = strconv.ParseFloat(envs["Pi"], 64) //from env
	expected.Piglet.Name = "Penny"

	assert.Equal(t, expected, conf)

	for _, env := range envStrs {
		os.Unsetenv("CONFTEST_" + env)
	}
}

func Test_ConfigFail(t *testing.T) {
	var err error
	var cfg interface{}

	tests := []struct {
		name           string
		cfg            interface{}
		defaultFile    string
		configFile     string
		expectedErrors []error
	}{
		{
			name:           "nonexistant given config",
			configFile:     "none.yml",
			expectedErrors: []error{ErrNoFileFound},
		},
		{
			name:           "no default file, given config file doesn't exist and is of unvalid type",
			configFile:     "test.fake",
			expectedErrors: []error{ErrNoDefaultConfig, ErrInvalidConfigFile, ErrNoFileFound},
		},
		{
			name:           "cfg is not a pointer",
			cfg:            TestConfig{},
			expectedErrors: []error{ErrNotAPointer},
		},
		{
			name:           "type errors in file",
			configFile:     "test/faulty.yml",
			expectedErrors: []error{ErrInvalidFormat},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultFile(tt.defaultFile)

			if tt.cfg == nil {
				cfg = new(TestConfig)
			} else {
				cfg = tt.cfg
			}
			err = SetUpConfigurationWithConfigFile(cfg, tt.configFile)
			assert.NotNil(t, err)
			for _, expectedErr := range tt.expectedErrors {
				assert.Contains(t, err.Error(), expectedErr.Error())
			}
		})
	}

}

func Test_setField(t *testing.T) {
	type MyStruct struct {
		Name   string
		Age    int
		Height float64
		Online bool
	}

	ms := &MyStruct{Name: "Elsa"}

	wrong := []interface{}{
		43,      // - str
		43.0,    // - int
		"hello", // - float64
		"false", // - bool
	}
	rv := reflect.ValueOf(ms).Elem()
	typ := rv.Type()
	for i := 0; i < typ.NumField(); i++ {
		fieldVal := rv.Field(i)

		err := setField(wrong[i], fieldVal, "arbitrary error msg")
		assert.NotNil(t, err)
	}
}

func Test_setFieldString(t *testing.T) {
	type MyStruct struct {
		Name   string
		Age    int
		Height float64
		Online bool
	}

	ms := &MyStruct{Name: "Elsa"}

	wrong := []interface{}{
		43,      // - str
		43.0,    // - int
		"hello", // - float64
		"no",    // - bool
	}
	rv := reflect.ValueOf(ms).Elem()
	typ := rv.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := rv.Field(i)
		name := strings.ToLower(field.Name)

		err := setFieldString(wrong[i], name, fieldVal, "arbitrary error msg")
		assert.NotNil(t, err)
	}
}
