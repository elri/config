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

var (
	envs      map[string]interface{}
	envPrefix string

	writedefconf bool
	printconf    bool

	configErrorHandling flag.ErrorHandling
)

var osExit = os.Exit //to enable testing
var ErrNotAPointer = errors.New("cfg should be pointer")

func init() {
	envs = make(map[string]interface{})
	flags = make(map[string]interface{})
	SetFlagSet(flag.CommandLine)
	flagSetArgs = os.Args[1:]
	flag_defaults = make(map[string]interface{})

	flagSet.Usage = Usage
}

// Init sets the error handling property.
// The error handling for the config package is the same as that
// for the std flag package
func Init(errorHandling flag.ErrorHandling) {
	configErrorHandling = errorHandling
	flagSet.Init("config", errorHandling)
}

func handleError(err error) {
	switch configErrorHandling {
	case flag.ContinueOnError:
		return
	case flag.ExitOnError:
		osExit(2)
	case flag.PanicOnError:
		panic(err)
	}
}

func SetEnvPrefix(prefix string) {
	envPrefix = prefix
}

func SetEnvsToParse(envVarNames []string) (err error) {
	for _, e := range envVarNames {
		eFull := e
		if envPrefix != "" {
			eFull = envPrefix + e
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

func SetUpConfiguration(cfg interface{}) (err error) {
	return setup(cfg, "")
}

func SetUpConfigurationWithConfigFile(cfg interface{}, filename string, dirs ...string) (err error) {
	return setup(cfg, filename)
}

func setup(cfg interface{}, filename string, dirs ...string) (err error) {
	//Check that cfg is pointer
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return fmt.Errorf("invalid argument: "+ErrNotAPointer.Error()+"but is %s", reflect.ValueOf(cfg).Kind())
	}

	// DEFAULT CONFIG FILE
	ParseDefaultConfigFile(cfg)

	// DEFAULT FLAGS
	if len(flag_defaults) > 0 {
		parseMapAndSet(cfg, flag_defaults)
	}

	// GIVEN CONFIG FILE
	if filename != "" {
		cf_err := ParseConfigFile(cfg, filename, dirs...)
		if cf_err != nil {
			err = addErr(err, cf_err)
		}
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

// Creates a string given a ptr to a struct
// E.g.
// type Person struct {
//	Name string
// 	Age int
// }
//
// mio := &Person{Name: "Mio", Age: 9}
//
// String(mio):
// name: Mio
// age: 9
func String(c interface{}) string {
	return createString(c, true)
}

// Same as String, except ignores zero values e.g. empty strings and zeroes
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
