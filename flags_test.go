package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"testing"

	"github.com/stretchr/testify/assert"
)

/*
NOTES
- uint becomes uint64 due to strconv.IntSize == 64
*/

var (
	b   = "b"
	f64 = "f64"
	i   = "i"
	i64 = "i64"
	str = "str"
	ui  = "ui"
	d   = "d"
)

func testInit() *flag.FlagSet {
	resetConfig()
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	SetFlagSet(flagSet)
	return flagSet
}

func resetConfig() {
	flag_defaults = make(map[string]interface{})
	flags = make(map[string]interface{})
	envs = make(map[string]interface{})
}

func Test_SetFlagDefault(t *testing.T) {
	flagSet := testInit()

	fBool := flagSet.Bool(b, false, "usage")
	f := LookupFlag(b)

	err := SetFlagDefault(b, "true")
	assert.Nil(t, err)
	assert.True(t, *fBool)
	assert.Equal(t, f.DefValue, "true")

	err = SetFlagDefault(ui, "")
	assert.NotNil(t, err)

}

func Test_FlagValueIsString(t *testing.T) {
	flagSet := testInit()

	_ = flagSet.Bool(b, false, "usage")
	_ = flagSet.Float64(f64, 3.14, "usage")
	_ = flagSet.Int(i, 10, "usage")
	_ = flagSet.Int64(i64, 99, "usage")
	_ = flagSet.String(str, "default flag string", "usage")
	_ = flagSet.Uint(ui, 20, "usage")
	_ = flagSet.Duration(d, 5*time.Second, "usage")

	flags := map[string]*flag.Flag{
		b:   LookupFlag(b),
		f64: LookupFlag(f64),
		i:   LookupFlag(i),
		i64: LookupFlag(i64),
		str: LookupFlag(str),
		ui:  LookupFlag(ui),
		d:   LookupFlag(d)}

	// beforeparse ensures Value is of type FlagValue
	flagSet.VisitAll(beforeParse())

	for k, f := range flags {
		if k != str {
			assert.False(t, IsString(f))
		} else {
			assert.True(t, IsString(f))
		}
	}
}

func Test_ParseBoolFLag(t *testing.T) {
	tests := []struct {
		name      string
		defValue  bool
		flagValue string
		expected  bool
		set       bool
		fail      bool
	}{
		{
			name:     "Default flag (False)",
			defValue: false,
			expected: false,
		}, {
			name:     "Default flag (True)",
			defValue: true,
			expected: true,
		},
		{
			name:      "Switch flag: from False to True",
			defValue:  false,
			flagValue: "true",
			expected:  true,
			set:       true,
		},
		{
			name:      "Switch flag: from True to False",
			defValue:  true,
			flagValue: "false",
			expected:  false,
			set:       true,
		},
		{
			name:      "Alternativ true: 1",
			flagValue: "1",
			expected:  true,
			set:       true,
		},
		{
			name:      "Alternativ true: 0",
			flagValue: "0",
			defValue:  true,
			expected:  false,
			set:       true,
		},
		{
			name:      "Fail: wrong datatype",
			flagValue: "42",
			fail:      true,
			set:       true,
		},
	}
	for _, tt := range tests {

		flagSet := testInit()

		fBool := flagSet.Bool(b, tt.defValue, "usage")
		fmt.Println(tt.name, fBool)

		args := make([]string, 0)
		if tt.set {
			a := "-b"
			if tt.flagValue != "" {
				a += fmt.Sprint("=", tt.flagValue)
			}
			args = append(args, a)
		}
		SetFlagSetArgs(args)

		err := ParseFlags()

		if tt.fail {
			assert.Error(t, err)
		} else {
			assert.Nil(t, err)

			bFlag := LookupFlag(b)
			assert.Equal(t, tt.set, ParsedFlag(bFlag))

			assert.Equal(t, tt.expected, *fBool)
		}
	}

}

func Test_ParseFlags(t *testing.T) {
	flagSet := testInit()

	_ = flagSet.Bool(b, false, "usage")
	_ = flagSet.Float64(f64, 3.14, "usage")
	_ = flagSet.Int(i, 10, "usage")
	_ = flagSet.Int64(i64, 99, "usage")
	_ = flagSet.String(str, "default flag string", "usage")
	_ = flagSet.Uint(ui, 20, "usage")
	_ = flagSet.Duration(d, 5*time.Second, "usage")

	flags := map[string]*flag.Flag{
		b:   LookupFlag(b),
		f64: LookupFlag(f64),
		i:   LookupFlag(i),
		i64: LookupFlag(i64),
		str: LookupFlag(str),
		ui:  LookupFlag(ui),
		d:   LookupFlag(d)}

	args := []string{
		"-f64", "3.1415",
		"-i", "342",
		"-str", "hello",
		"-b",
	}
	SetFlagSetArgs(args)

	//parse
	err := ParseFlags()
	assert.Nil(t, err)

	for _, arg := range args {
		argSplit := strings.Split(arg, "-")
		if len(argSplit) > 1 {
			k := argSplit[1]
			if f := flags[k]; f != nil {
				assert.True(t, ParsedFlag(f))
				assert.NotEqual(t, f.Value.String(), f.DefValue)
			}
		}
	}
}

func Test_ensureFlagValue(t *testing.T) {
	flagSet := testInit()

	var f *flag.Flag
	var changed bool

	// Change
	_ = flagSet.Bool(b, false, "usage")
	f = LookupFlag(b)

	changed = ensureFlagValue(f)
	assert.True(t, changed)

	// No change (+ ensure that flagvalue isn't reset)
	_ = flagSet.Uint(ui, 20, "usage")
	f = LookupFlag(ui)
	copy := FlagValue{Value: f.Value, parsed: true}
	f.Value = &FlagValue{Value: f.Value, parsed: true}

	changed = ensureFlagValue(f)
	assert.False(t, changed)
	fv := f.Value.(*FlagValue)
	assert.Equal(t, copy, *fv)
}

func Test_GetFlagValue(t *testing.T) {
	flagSet := testInit()

	_ = flagSet.String(str, "hello", "usage")
	f := flagSet.Lookup(str)

	//Not
	fv := GetFlagValue(f)
	assert.Nil(t, fv)

	//FlagValue
	f.Value = &FlagValue{Value: f.Value}

	fv = GetFlagValue(f)
	assert.NotNil(t, fv)

	//FlagValueBool
	fvb := new(FlagValueBool)
	fvb.Value = fv.Value
	f.Value = fvb

	fv = GetFlagValue(f)
	assert.NotNil(t, fv)
}

func Test_beforeParse(t *testing.T) {
	flagSet := testInit()

	//reset flag_defaults
	flag_defaults = make(map[string]interface{})

	fBool := flagSet.Bool(b, false, "usage")
	fFloat64 := flagSet.Float64(f64, 3.14, "usage")
	fInt := flagSet.Int(i, 10, "usage")
	fInt64 := flagSet.Int64(i64, 99, "usage")
	fStr := flagSet.String(str, "default flag string", "usage")
	fUint := flagSet.Uint(ui, 20, "usage")
	fDuration := flagSet.Duration(d, 5*time.Second, "usage")

	flagSet.VisitAll(beforeParse())

	flagValues := map[string]interface{}{
		b:   *fBool,
		f64: *fFloat64,
		i:   *fInt,
		i64: *fInt64,
		str: *fStr,
		ui:  *fUint,
		d:   *fDuration,
	}

	flagPtrs := map[string]*flag.Flag{
		b:   LookupFlag(b),
		f64: LookupFlag(f64),
		i:   LookupFlag(i),
		i64: LookupFlag(i64),
		str: LookupFlag(str),
		ui:  LookupFlag(ui),
		d:   LookupFlag(d)}

	// check that flag_defaults got populated correctly
	// and that all flags have FlagValue type in place of flag.Value
	assert.True(t, len(flag_defaults) >= len(flagValues))
	for k, v := range flagValues {
		assert.EqualValues(t, v, flag_defaults[k])
		notFlagValue := ensureFlagValue(flagPtrs[k])
		assert.False(t, notFlagValue)
	}
}

func Test_afterParse(t *testing.T) {
	flagSet := testInit()

	fBool := flagSet.Bool(b, false, "usage")
	fFloat64 := flagSet.Float64(f64, 3.14, "usage")
	fInt := flagSet.Int(i, 10, "usage")
	fInt64 := flagSet.Int64(i64, 99, "usage")
	fStr := flagSet.String(str, "default flag string", "usage")
	fUint := flagSet.Uint(ui, 20, "usage")
	fDuration := flagSet.Duration(d, 5*time.Second, "usage")

	args := []string{
		"-f64", "3.1415",
		"-i", "342",
		"-str", "hello",
		"-d", "7m",
	}

	flagSet.VisitAll(func(f *flag.Flag) {
		f.Value = &FlagValue{Value: f.Value}
	})

	err := flagSet.Parse(args)
	assert.Nil(t, err)

	flagSet.VisitAll(afterParse())

	flagValues := map[string]interface{}{
		b:   *fBool,
		f64: *fFloat64,
		i:   *fInt,
		i64: *fInt64,
		str: *fStr,
		ui:  *fUint,
		d:   *fDuration,
	}

	// check that flags got populated correctly
	// and that all flags in it has been parsed
	assert.Equal(t, len(args)/2, len(flags))
	for k, v := range flagValues {
		if ok := flags[k]; ok != nil {
			assert.EqualValues(t, v, flags[k])
			f := LookupFlag(k)
			assert.True(t, ParsedFlag(f))
		}
	}
}

func Test_Config_SetDefaults(t *testing.T) {
	flagSet := testInit()

	fBool := flagSet.Bool(b, false, "usage")
	fFloat64 := flagSet.Float64(f64, 3.14, "usage")
	fInt := flagSet.Int(i, 10, "usage")
	fInt64 := flagSet.Int64(i64, 99, "usage")
	fStr := flagSet.String(str, "default flag string", "usage")
	fUint := flagSet.Uint(ui, 20, "usage")
	fDuration := flagSet.Duration(d, 5*time.Second, "usage")

	flags := []struct {
		f     *flag.Flag
		value string
	}{
		{LookupFlag(b), "true"},
		{LookupFlag(f64), "3.0"},
		{LookupFlag(i), "20"},
		{LookupFlag(i64), "33"},
		{LookupFlag(str), "new default str"},
		{LookupFlag(ui), "20"},
		{LookupFlag(d), "7m"}}

	for _, u := range flags {
		SetFlagDefault(u.f.Name, u.value)
	}

	args := []string{
		"-f64", "3.1415",
		"-i", "342",
		"-str", "hello",
	}

	flagSet.VisitAll(beforeParse())
	err := flagSet.Parse(args)
	assert.Nil(t, err)

	assert.Equal(t, true, *fBool)
	assert.Equal(t, 3.1415, *fFloat64) //parsed
	assert.Equal(t, 342, *fInt)        //parsed
	assert.Equal(t, int64(33), *fInt64)
	assert.Equal(t, "hello", *fStr) //parsed
	assert.Equal(t, uint(0x14), *fUint)
	assert.Equal(t, 7*time.Minute, *fDuration)

	flag_defaults := GetDefaultFlags()

	assert.Equal(t, true, flag_defaults[b])
	assert.Equal(t, 3.0, flag_defaults[f64])
	assert.Equal(t, 20, flag_defaults[i])
	assert.Equal(t, int64(33), flag_defaults[i64])
	assert.Equal(t, "new default str", flag_defaults[str])
	assert.EqualValues(t, 0x14, flag_defaults[ui])
	assert.Equal(t, 7*time.Minute, flag_defaults[d])
}

func Test_DefaultFlagPrint(t *testing.T) {
	flagSet := testInit()

	flagSet.String("pim", "just a random pim flag", "pimflag")
	flagSet.Bool("dreams", false, "none")
	flagSet.Int("age", 35, "flag age")
	flagSet.Float64("pi", 3.141592, "magic")
	flagSet.String("name", "piglet name flag", "pigflag")

	args_print := []string{"-write-def-conf", "-print-conf"}
	SetFlagSetArgs(args_print)

	var err error
	err = ParseFlags()
	assert.Nil(t, err)

	// set default config file
	err = SetDefaultFile("example/default_conf.yml")
	assert.Nil(t, err)

	// intercept exit func
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	var got int
	myExit := func(code int) {
		got = code
	}

	osExit = myExit

	// parse conf
	conf := new(TestConfig)
	SetUpConfiguration(conf)
	if exp := 0; got != exp {
		t.Errorf("Expected exit code: %d, got: %d", exp, got)
	}

}
