

[![build](https://github.com/elri/config/actions/workflows/go.yml/badge.svg)](https://github.com/elri/config/actions/workflows/go.yml)
[![coverage](https://coveralls.io/repos/github/elri/config/badge.svg?branch=main)](https://coveralls.io/github/elri/config?branch=main)
[![goreport](https://goreportcard.com/badge/github.com/elri/config)](https://goreportcard.com/report/github.com/elri/config)

# Light-weight configuration parsing
Package config can be used to parse flags, environmental variables and configuration files, and store them in a given struct.

The priority of the sources is the following:
1. flags
2. env. variables
3. given config file
4. flag defaults 
5. default config file

For example, if values from the following sources were loaded:
```
Defaults : {
		"user": "default",
		"secret": "",
		"endpoint": "https:localhost"
        "port": 99,
}

Config : {
		"user": "root"
		"secret": "confsecret"
}

Env : {
		"secret": "somesecretkey"
        "port": 88
}

Flags : {
        "port": 77 
}
```
 The resulting config will have the following values:
```  
	{
		"user": "root",
		"secret": "somesecretkey",
		"endpoint": "https:localhost"
        "port": 77,
	}
```

## Supported file types

The config files may be of the following types:
- toml
- yml
- json

## Keep in mind
- There is no case sensitivty, i.e. "pim", "Pim" and "PIM" are all considered the same
- The names of the environmental variables must match that of the struct. It is possible to set a prefix, so that i.e. if "MYVAR_" is set as a prefix, "MYVAR_PIM" will map to the property "pim"/"Pim"/"PIM". 
- For flags to map to the config automatically they must have the same name


## Currently not supported
- Anonymous structs for yaml files. Example:
```
type Inner struct {
	hello string
}

type Outer struct {
	Inner
	goodbye string
}

```
This is due to how the standard yml packages (currently) parses structs.  

- Setting more complex structures via flags. 

Example, in the case of this struct

```
type MyConfig struct {
	Debug   bool   `yaml:"debug" toml:"debug"`
	Log     string `yaml:"log" toml:"log"`

	Local struct {
		Host string `yaml:"host" toml:"host"`
		Port int    `yaml:"port" toml:"port"`
	} `yaml:"local" toml:"local"`

	Remote struct {
		Host string `yaml:"host" toml:"host"`
		Port int    `yaml:"port" toml:"port"`
	} `yaml:"remote" toml:"remote"`
}
```

for the flags `-host` / `-port`, the first corresponding property will be set, which in this case is the `Local` struct's properties.