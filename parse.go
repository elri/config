package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var defaultFile = ""

/*
Set default file. fpath must be absolute path. If the file cannot be opened, the function will return an error. Note that the error will only be return if
the error handling mode is set to ContinueOnError, else the function will Panic or Exit depending on the mode.
*/
func SetDefaultFile(fpath string) (err error) {
	defaultFile = fpath

	var f *os.File
	f, err = os.Open(fpath)
	if err != nil {
		err = fmt.Errorf("failed to set default file '%s': %s", fpath, err.Error())
		handleError(err)
		//fmt.Println("DEBUG DefaultFile set successfully", defaultFile)
	}
	defer f.Close()
	return
}

func GetDefaultFile() string {
	return defaultFile
}

var (
	ErrNoDefaultConfig            = errors.New("no default config file to parse")
	ErrFailedToParseDefaultConfig = fmt.Errorf("failed to parse default config (%s)", defaultFile)
	ErrNotAPointer                = errors.New("argument to must be a pointer")
	ErrInvalidConfigFile          = errors.New("unsupported or invalid file")
	ErrInvalidFormat              = errors.New("invalid format of file")
	ErrNoConfigFileToParse        = errors.New("no file given to parse")
	ErrNoFileFound                = syscall.Errno(2) // "could not find file"
)

/*
Compounds errors. Used rather than errors.Wrap since there's no hierarchy in the errors;
errors can stack up one another without one being dependant on one another.
*/
func addErr(prev error, add error) error {
	if prev == nil {
		return errors.New(add.Error())
	}
	return errors.New(prev.Error() + ", " + add.Error())
}

/*
Parse the default config fiĺe into the value pointed to by cfg. Returns error regardless of error handling mode.

If cfg is not a pointer, ParseDefaultConfigFile returns an ErrNotAPointer.
*/
func ParseDefaultConfigFile(cfg interface{}) (err error) {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr { //TODO: to a struct, map or list?
		err = fmt.Errorf("[ParseDefaultConfigFile]: %w ", ErrNotAPointer)
		return
	}

	if defaultFile == "" {
		err = ErrNoDefaultConfig
		return
	}

	var f *os.File
	f, err = os.Open(defaultFile)
	if err != nil {
		return
	}
	defer f.Close()

	filename := f.Name()
	derr := decode(cfg, f, filename)
	if derr != nil {
		err = addErr(err, derr)
	}
	return
}

/*
Parse the given config fiĺe into the value pointed to by cfg. Returns error regardless of error handling scheme.

If cfg is not a pointer, ParseConfigFile returns an ErrNotAPointer.

The 'filename' must either be an absolute path to the config file, exist in the current working directory, or in one of the directories given as 'dirs'. If the given file cannot be found, ParseConfig file returns an ErrNoConfigFileToParse.
*/
func ParseConfigFile(cfg interface{}, filename string, dirs ...string) (err error) {
	if reflect.TypeOf(cfg).Kind() != reflect.Ptr {
		err = fmt.Errorf("[ParseConfigFile]: %w ", ErrNotAPointer)
		return
	}

	if filename == "" {
		err = ErrNoConfigFileToParse
		return
	}

	// Parse default file first -- it's ok if it fails
	ParseDefaultConfigFile(cfg)

	// If not found as is, check through relevant directories
	found := true
	f, ferr := os.Open(filename)
	if ferr != nil {
		found = false
		if len(dirs) > 0 {
			var ftmp *os.File
			for _, dir := range dirs {
				fpath := filepath.Join(dir, filename)
				ftmp, ferr = os.Open(fpath)
				if ferr == nil {
					f.Close()
					f = ftmp
					filename = fpath
					found = true
					break
				}
			}
		}
		if !found {
			err = addErr(err, ErrNoFileFound)
		}
	}
	defer f.Close()

	derr := decode(cfg, f, filename)
	if derr != nil {
		err = addErr(err, derr)
	}

	return
}

func decode(cfg interface{}, f *os.File, filename string) (err error) {
	switch {
	case strings.Contains(filename, "toml"):
		_, err = toml.DecodeFile(filename, cfg)
	case strings.Contains(filename, "yml"), strings.Contains(filename, "yaml"):
		decoder := yaml.NewDecoder(f)
		err = decoder.Decode(cfg)
	case strings.Contains(filename, "json"):
		var content []byte
		content, err = ioutil.ReadFile(filename)
		if err == nil {
			err = json.Unmarshal(content, cfg)
		} else {
			return
		}
	default:
		err = errors.New(ErrInvalidConfigFile.Error() + " of type " + filename)
	}

	if err != nil && !strings.Contains(err.Error(), ErrInvalidConfigFile.Error()) {
		err = ErrInvalidFormat
	}
	return
}

func writeToDefaultFile(cfg interface{}) (err error) {
	if defaultFile == "" {
		fmt.Println("WARNING! Trying to write to default file but no default file path set")
		osExit(1)
	} else {
		var write bool
		var fs os.FileInfo
		fs, err = os.Stat(defaultFile)

		if err == nil && fs.Size() != 0 { // if file DOES exist, warn overwriting -> show, overwrite, abort
			fmt.Printf("'%s' exists, would you like to overwrite it? \n", defaultFile)
			fmt.Print("Options: Yes/Overwrite [y], Show content [s], No/Abort [n]: ")
			reader := bufio.NewReader(os.Stdin)

			var choice string
			choice, err = reader.ReadString('\n')
			choice = strings.Split(choice, "\n")[0]
			switch choice {
			case "y", "yes", "overwrite":
				write = true
			case "s", "show":
				var prevDef []byte
				prevDef, err = os.ReadFile(defaultFile)
				if err == nil {
					fmt.Printf("\nCONTENTS OF '%s':\n%s\n", defaultFile, string(prevDef))
				}
			case "n", "no", "abort":
				fmt.Println("Aborting.")
			default:
				fmt.Println("faulty input")
			}
		} else if err == nil { // file exists but is empty
			write = true
		} else if os.IsNotExist(err) { // if file doesn't exist, create it and write
			_, err = os.Create(defaultFile)
			if err == nil {
				fmt.Printf("Created %s", defaultFile)
				write = true
			}
		}

		var buf *bytes.Buffer
		if write {
			buf, err = encode(cfg, defaultFile)
			fmt.Println("Wrote to", defaultFile)
		}

		if err == nil {
			if printconf {
				fmt.Println("CONFIGURATION:")
				if buf != nil {
					fmt.Println(buf)
				} else {
					fmt.Println(String(cfg))
				}
			}
		} else {
			fmt.Print(err.Error(), "\n")
			osExit(1)
		}
	}
	return
}

func encode(cfg interface{}, filename string) (buf *bytes.Buffer, err error) {
	buf = new(bytes.Buffer)
	var bytes []byte

	switch {
	case strings.Contains(filename, "toml"):
		encoder := toml.NewEncoder(buf)
		err = encoder.Encode(cfg)
		if err == nil {
			bytes = buf.Bytes()
		}
	case strings.Contains(filename, "yml"), strings.Contains(filename, "yaml"):
		encoder := yaml.NewEncoder(buf)
		err = encoder.Encode(cfg)
		if err == nil {
			bytes = buf.Bytes()
		}
	case strings.Contains(filename, "json"):
		bytes, err = json.Marshal(cfg)
	default:
		err = ErrInvalidConfigFile
		//err = errors.New("can't handle " + filename)
	}

	if err == nil {
		var f *os.File
		f, err = os.OpenFile(filename, os.O_WRONLY, 0644)
		if err == nil {
			f.Write(bytes)
		}
		defer f.Close()
	}

	if err != nil {
		err = errors.WithStack(err)
	}

	return
}
