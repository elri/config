package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//Aliasing flag.Errorhandling for clarity and easy of use.
type ErrorHandling flag.ErrorHandling

var osExit = os.Exit //to enable testing

var (
	envs      map[string]interface{}
	envPrefix string

	writedefconf bool
	printconf    bool

	configErrorHandling ErrorHandling
)

const (
	ContinueOnError = ErrorHandling(flag.ContinueOnError) // Return a descriptive error.
	ExitOnError     = ErrorHandling(flag.ExitOnError)     // Call os.Exit(2) or for -h/-help Exit(0).
	PanicOnError    = ErrorHandling(flag.PanicOnError)    // Call panic with a descriptive error.
)

func init() {
	envs = make(map[string]interface{})
	flags = make(map[string]interface{})
	SetFlagSet(flag.CommandLine)
	flagSetArgs = os.Args[1:]
	flag_defaults = make(map[string]interface{})

	flagSet.Usage = Usage
}

/*
Init sets the global error handling property, as well as the error handling property for the flagset.

The error handling for the config package is similar to that of the standard flag package;
there are three modes: Continue, Panic and Exit.

The default mode is Continue.
*/
func Init(errorHandling ErrorHandling) {
	configErrorHandling = errorHandling
	flagSet.Init("elri/config", flag.ErrorHandling(errorHandling))
}

func handleError(err error) {
	switch configErrorHandling {
	case ContinueOnError:
		return
	case ExitOnError:
		osExit(2)
	case PanicOnError:
		panic(err)
	}
}

/*
Set a prefix to use for all environmental variables.

For example to different between what is used in testing and otherwise, the prefix "TEST_" could be used.
The environmental variables TEST_timeout and TEST_angle would then map to the properties 'timeout' and 'angle'.
*/
func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

/*
Set a list of environmental variable names to check when filling out the configuration struct.

The list can consist of variables both containing a set env prefix and not, but the environmental variable that is looked for will be that with the prefix.
That is, if the prefix is set as TEST_ and the list envVarNames is ["timeout", "TEST_angle"], the environmental variables that will be looked for are ["TEST_timeout", "TEST_angle"].

If the environmental variable(s) cannot be find, SetEnvsToParse will return an error containing all the names of the non-existant variables. Note that the error will only be return if
the error handling mode is set to ContinueOnError, else the function will Panic or Exit depending on the mode.
*/
func SetEnvsToParse(envVarNames []string) (err error) {
	for _, e := range envVarNames {
		eFull := e
		if envPrefix != "" {
			if !strings.HasPrefix(eFull, envPrefix) {
				eFull = envPrefix + e
			}
		}
		envVar, ok := os.LookupEnv(eFull)
		if ok {
			e = strings.ToLower(e)
			envs[e] = envVar
		} else {
			newErr := fmt.Errorf("could not find %s", e)
			err = addErr(err, newErr)
			handleError(err)
		}
	}
	return
}

/*
	Parse all the sources (flags, env vars, default config file) and store the result in the value pointer to by cfg.

	If cfg is not a pointer, SetUpConfiguration returns an ErrNotAPointer.

*/
func SetUpConfiguration(cfg interface{}) (err error) {
	return setup(cfg, "")
}

/*
	Parse all the sources (flags, env vars, given config file, default config file) and store the result in the value pointer to by cfg.

	If cfg is not a pointer, SetUpConfigurationWithConfigFile returns an ErrNotAPointer.

	The 'filename' must either be an absolute path to the config file, exist in the current working directory, or in one of the directories given as 'dirs'. If the given file cannot be found, the other sources will still be parsed, but an ErrNoConfigFileToParse will be returned.

*/
func SetUpConfigurationWithConfigFile(cfg interface{}, filename string, dirs ...string) (err error) {
	return setup(cfg, filename)
}

func setup(cfg interface{}, filename string, dirs ...string) (err error) {
	//Check that cfg is pointer
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		err = fmt.Errorf("[setup]: %w ", ErrNotAPointer)
		return
	}

	// DEFAULT CONFIG FILE
	ParseDefaultConfigFile(cfg)

	// DEFAULT FLAGS
	if len(flag_defaults) > 0 {
		parseMapAndSet(cfg, flag_defaults)
	}

	// GIVEN CONFIG FILE
	if filename != "" {
		err = ParseConfigFile(cfg, filename, dirs...)
	}

	// ENVIRONMENTAL VARIABLES
	if len(envs) > 0 {
		rv := reflect.ValueOf(cfg).Elem()
		typ := rv.Type()
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := rv.Field(i)
			name := strings.ToLower(field.Name)
			v := envs[name]
			msg := "type of environmental variable not one that is handled by config"
			env_err := setFieldString(v, name, fieldVal, msg)
			if env_err != nil {
				err = addErr(err, env_err)
			}

		}
	}

	// FLAGS
	if flagSet.Parsed() {
		parseMapAndSet(cfg, flags)
	}

	if writedefconf {
		err = writeToDefaultFile(cfg)
		if err == nil {
			osExit(0)
		}
	}
	if printconf {
		fmt.Println("CONFIGURATION:")
		fmt.Println(String(cfg))
		osExit(0)
	}

	if err != nil {
		handleError(err)
	}

	return
}

func parseMapAndSet(cfg interface{}, m map[string]interface{}) {
	var err error
	rv := reflect.ValueOf(cfg).Elem()
	typ := rv.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := rv.Field(i)

		setFieldAux := func(name string, fieldVal reflect.Value) {
			v := m[name]
			if v != nil {
				msg := fmt.Sprintf("type mismatch between flag and corresponding field (%s)", name)
				err = setField(v, fieldVal, msg)
				if err != nil {
					log.Println(err.Error())
				}
			}
		}

		if fieldVal.Kind() == reflect.Struct {
			innerTyp := fieldVal.Type()
			for j := 0; j < fieldVal.NumField(); j++ {
				innerFieldVal := fieldVal.Field(j)
				name := strings.ToLower(innerTyp.Field(j).Name)
				setFieldAux(name, innerFieldVal)
			}
		} else {
			name := strings.ToLower(field.Name)
			setFieldAux(name, fieldVal)
		}
	}
}

func setField(toInsert interface{}, fieldVal reflect.Value, defaultMsg string) (err error) {
	if toInsert != nil {
		var converted interface{}
		switch c := toInsert.(type) {
		case string:
			converted = c
		case bool:
			converted = c
		case float64:
			converted = c
		case int:
			converted = c
		case int64:
			converted = c
		case uint:
			converted = c
		case uint64:
			converted = c
		case time.Duration:
			converted = c
		//TODO more types?
		default:
			err = errors.New("SWITCH DEFAULT " + defaultMsg)
		}

		newVal := reflect.ValueOf(converted)
		if newVal.Kind() != fieldVal.Kind() {
			err = errors.New("WRONG KIND " + defaultMsg)
		}

		if err == nil {
			fieldVal.Set(newVal)
		}

	}
	return
}

func setFieldString(toInsert interface{}, fieldName string, fieldVal reflect.Value, defaultMsg string) (err error) {
	if toInsert != nil {
		toInsertVal := reflect.ValueOf(toInsert)
		var converted interface{}
		toInsertValStr := toInsertVal.String()
		k := fieldVal.Kind()
		switch k {
		case reflect.String:
			inK := toInsertVal.Kind()
			if inK == reflect.String {
				fieldVal.Set(toInsertVal)
			} else {
				err = errors.New(defaultMsg)
			}
			return
		case reflect.Bool:
			converted, err = strconv.ParseBool(toInsertValStr)
		case reflect.Float64:
			converted, err = strconv.ParseFloat(toInsertValStr, 64)
		case reflect.Int:
			converted, err = strconv.Atoi(toInsertValStr)
		case reflect.Int64:
			converted, err = strconv.ParseInt(toInsertValStr, 2, 64)
		case reflect.Uint, reflect.Uint64:
			converted, err = strconv.ParseUint(toInsertValStr, 10, 64)
		default:
			err = errors.New(defaultMsg)
		}

		if err == nil {
			newVal := reflect.ValueOf(converted)
			fieldVal.Set(newVal)
		} else {
			errStr := fmt.Sprintf("env var '%s' trying to set field '%s' with type %s to '%s' (ignored)", envPrefix+fieldName, fieldName, k, toInsertValStr)
			err = errors.New(errStr)
			handleError(err)
			fmt.Println("WARNING: " + errStr)
		}
	}
	return
}

/*
Creates a string given a ptr to a struct.

Example:
 type Person struct {
	Name string
 	Age int
 }

 func printMio() {
	 mio := &Person{Name: "Mio", Age: 9}
	 fmt.Println(String(mio))
 }

output:
 name: Mio
 age: 9
*/
func String(c interface{}) string {
	return createString(c, true)
}

// Same as String(), except ignores zero values e.g. empty strings and zeroes
func StringIgnoreZeroValues(c interface{}) string {
	return createString(c, false)
}

func createString(c interface{}, printZeroValues bool) string {

	doPrint := func(fieldVal reflect.Value) bool {
		return !fieldVal.IsZero() || fieldVal.Kind() == reflect.Bool || printZeroValues
	}

	isTime := func(fieldVal reflect.Value) bool {
		now := time.Now()
		return reflect.Indirect(fieldVal).Type() == reflect.ValueOf(now).Type()
	}

	ret := ""
	rv := reflect.ValueOf(c).Elem()
	typ := rv.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := rv.Field(i)
		switch fieldVal.Kind() {
		case reflect.Struct:
			// Check if time.Time type
			if isTime(fieldVal) {
				ret += fmt.Sprint(strings.ToLower(field.Name), ": ", reflect.Indirect(fieldVal), "\n")
			} else {
				ret += fmt.Sprintf("%s: \n", strings.ToLower(field.Name))
				jTyp := fieldVal.Type()
				for j := 0; j < fieldVal.NumField(); j++ {
					jField := fieldVal.Field(j)
					name := strings.ToLower(jTyp.Field(j).Name)
					if doPrint(jField) {
						ret += fmt.Sprint("    ", name, ": ", jField, "\n")
					}
				}
			}
		default:
			if doPrint(fieldVal) {
				ret += fmt.Sprint(strings.ToLower(field.Name), ": ", fieldVal, "\n")
			}
		}
	}

	return ret
}
