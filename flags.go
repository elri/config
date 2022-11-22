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
	flagSet       *flag.FlagSet
	flagSetArgs   []string
	flags         map[string]interface{}
	flag_defaults map[string]interface{}
)

var (
	writeConfFlagName = "write-def-conf"
	printConfFlagName = "print-conf"
)

func SetFlagSet(f *flag.FlagSet) {
	flagSet = f

	_ = flagSet.Bool(writeConfFlagName, false, "writes default configuration to default file. if default file already exists, options of overwrite, show and abort are given. ")
	_ = flagSet.Bool(printConfFlagName, false, " prints configuration for current run. if combined with write-def-conf the print format is that of default file.")
}

func SetFlagSetArgs(args []string) {
	flagSetArgs = args
}

func GetDefaultFlags() map[string]interface{} {
	return flag_defaults
}

func LookupFlag(name string) *flag.Flag {
	return flagSet.Lookup(name)
}

//based on flag package's PrintDefaults()
func Usage() {
	fmt.Fprintf(flagSet.Output(), "Usage of %s:\n", os.Args[0])

	fmt.Fprint(flagSet.Output(), "[!] Use the flag '-write-def-conf' to write default values to the default config file. The default file is created if it doesn't exist. \n    If the default file exists and isn't empty, options to overwrite, show content and abort are given.", "\n")
	fmt.Fprint(flagSet.Output(), "[!] Use the flag '-print-conf' to just print the current configuration to stdout. If -print-conf is combined with -write-def-conf the print format is that of default file.", "\n")

	if defaultFile != "" {
		fmt.Fprintf(flagSet.Output(), "[!] Default config file is '%s'.\n", defaultFile)
	} else {
		fmt.Fprint(flagSet.Output(), "[!] No default config file is set.\n")
	}
	flagSet.VisitAll(func(f *flag.Flag) {
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s", f.Name) // Two spaces before -; see next two comments.
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if b.Len() <= 4 { // space, space, '-', 'x'.
			b.WriteString("\t")
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			b.WriteString("\n    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))

		if f.Name != writeConfFlagName && f.Name != printConfFlagName {
			if !reflect.ValueOf(f.DefValue).IsZero() {
				if IsString(f) {
					// put quotes on the value
					fmt.Fprintf(&b, " (default %q)", f.DefValue)
				} else {
					fmt.Fprintf(&b, " (default %v)", f.DefValue)

				}
			}
			fmt.Fprint(flagSet.Output(), b.String(), "\n")
		}

	})

}

// Value type insert

type FlagValue struct {
	Value  flag.Value
	parsed bool
}

func (fv *FlagValue) defSet(val string) error {
	return fv.Value.Set(val)
}

func (fv *FlagValue) Set(val string) error {
	fv.parsed = true
	return fv.Value.Set(val)
}

func (fv *FlagValue) String() string {
	var ret string
	if fv.Value != nil {
		ret = fv.Value.String()
	}
	return ret
}

func IsString(f *flag.Flag) bool {
	ensureFlagValue(f)
	fv := GetFlagValue(f)
	val := reflect.Indirect(reflect.ValueOf(fv.Value))
	return val.Kind() == reflect.String
}

func GetFlagValue(f *flag.Flag) *FlagValue {
	if f.Value != nil {
		fv, ok := f.Value.(*FlagValue)
		if ok {
			return fv
		}
		fvb, ok := f.Value.(*FlagValueBool)
		if ok {
			fv := &FlagValue{Value: fvb.Value, parsed: fvb.parsed}
			return fv
		}
	}
	return nil

}

func ensureFlagValue(f *flag.Flag) (changed bool) {
	if f.Value == nil {
		err := errors.New("flag Value is nil")
		handleError(err)
		return
	}
	val := reflect.Indirect(reflect.ValueOf(f.Value))
	typ := reflect.TypeOf(FlagValue{})
	typ2 := reflect.TypeOf(FlagValueBool{})
	convAlreadyDone := val.CanConvert(typ) || val.CanConvert(typ2)
	if !convAlreadyDone {
		val := reflect.ValueOf(f.Value)
		kind := reflect.Indirect(val).Kind()
		if kind == reflect.Bool {
			fvb := new(FlagValueBool)
			fvb.Value = f.Value
			f.Value = fvb
		} else {
			f.Value = &FlagValue{Value: f.Value}
		}
		changed = true
	}
	return
}

// Special case for bool to enable using bool flags like switches
type FlagValueBool struct {
	FlagValue
}

func (fvb *FlagValueBool) IsBoolFlag() bool { return true }

// Flag Parsing

func ParseFlags() error {
	flagSet.VisitAll(beforeParse())
	err := flagSet.Parse(flagSetArgs)
	if err != nil && (strings.Contains(err.Error(), writeConfFlagName) || strings.Contains(err.Error(), printConfFlagName)) {
		err = nil
	}

	if err == nil {
		flagSet.VisitAll(afterParse())
	}
	return err
}

func ParsedFlag(f *flag.Flag) bool {
	ensureFlagValue(f)
	fv := GetFlagValue(f)
	if fv != nil {
		return fv.parsed
	}
	return false
}

func addToFlagDefaults(f *flag.Flag, defVal string) {
	addFlagValueToMap(flag_defaults, f, defVal)
}

func SetFlagDefault(fName, def string) error {
	f := LookupFlag(fName)
	if f == nil {
		return fmt.Errorf("flag '%s' not found?", fName)
	}
	addToFlagDefaults(f, def)
	ensureFlagValue(f)
	f.DefValue = def
	fv := GetFlagValue(f)
	if fv != nil {
		return fv.defSet(def)
	}
	errMsg := fmt.Sprint("not FlagValue type: ", reflect.TypeOf(f.Value))
	return errors.New(errMsg)
}

func beforeParse() func(*flag.Flag) {
	return func(f *flag.Flag) {
		addToFlagDefaults(f, f.DefValue)
	}
}

func afterParse() func(*flag.Flag) {
	return func(f *flag.Flag) {
		if !flagSet.Parsed() {
			err := errors.New("flagSet not parsed")
			handleError(err)
			//			panic(errors.New("flagSet not parsed")) //TODO
		}
		if ParsedFlag(f) {
			if f.Name == printConfFlagName && f.Value.String() == "true" {
				printconf = true
			} else if f.Name == writeConfFlagName && f.Value.String() == "true" {
				writedefconf = true
			} else {
				addFlagValueToMap(flags, f, f.Value.String())
			}
		}
	}
}

func addFlagValueToMap(m map[string]interface{}, f *flag.Flag, value string) {
	var err error
	name := f.Name

	//Get reflect.Kind of the data that's stored in the Flag
	ensureFlagValue(f)
	fv := GetFlagValue(f)
	if fv == nil { //TODO if debug
		log.Println("not FlagValue type: ", reflect.TypeOf(f.Value))
		return
	}
	val := reflect.ValueOf(fv.Value)
	kind := reflect.Indirect(val).Kind()

	switch kind {
	case reflect.String:
		m[name] = value
	case reflect.Bool:
		m[name], err = strconv.ParseBool(value)
	case reflect.Int:
		m[f.Name], err = strconv.Atoi(value)
	case reflect.Int64:
		m[f.Name], err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			var derr error
			m[f.Name], derr = time.ParseDuration(value)
			if derr != nil {
				err = errors.Wrap(err, derr.Error())
			} else {
				err = derr
			}
		}
	case reflect.Float64:
		m[name], err = strconv.ParseFloat(value, 64)
	case reflect.Uint, reflect.Uint64:
		m[name], err = strconv.ParseUint(value, 10, 64)
	}

	if err != nil { //probably won't reach here, flag.Parse() will protest before this
		fmt.Println("panicking in addFlagValueToMap", f.Name, value)
		handleError(err)
	}

}
