// Package jsonflag is a simple example of how to use JSON configs in
// conjunction with Go's flag package, which allows for default values.
//
// Flag values do not need to appear in the config file
// and extra config file values will be ignore.
//
// Environmental variables will be expanded.  See testing for an example where
// "$USER" is expanded to "exampleUser".
//
// Order of precedence for set config values:
//
//  1. Command line flags. (cli Example: `--flag1=flag1Value`)
//  2. JSON values (json Example: `{"flag2": "flag2Value"}`)
//  3. Default flag values (go Example: `flag.StringVar(&config.Flag, "flagName", "flagDefaultValue", "description")`)
//
//
// This package uses Go's json package for decoding.  The json decoder only
// has accesses to exported fields of structs follows it's own
// precedence for json decoding, namely:
//
//  1. Tags
//  2. Exact case
//  3. Case insensitive
//
//
// Recommended Usage
//
//  1. Define a `config` struct with exported fields.
//  2. Use flag's functions to set default config values such as `flag.StringVar(&config.Flag1, "flag1", "defaultFlag1value", "flag1Desc")`
//  3. Put config values in a `config.json` file in the CWD. You can use --config=your_config.json to point somewhere else.
//  4. Call `jsonflag.Parse(&config)`
//
// The config struct is now appropriately populated.
//
//
// Config Path
//
// You can set the config path via the cli,
//
//  --config=your_config.json
//
// You can also set it in your application.  Note that this can be overwritten by the normal precedence.
//
//  path := "assets/config.json"
//  jsonflag.Path = &path
//
//
// Testing
//
// Since this package uses flag, test functions need a cli flag passed to verify
// cli parsing is working.  Test will fail otherwise.
//
//  USER=exampleUser go test --flag1=cliFlag1 --config=test_config.json
//
//
// TODO
//
// I hope to support in the future:
//
//   * Config set by environmental variables (between defaults and json config in the precedence hierarchy).
//   * json5 which permits comments and trailing commas like Go.
//
package jsonflag

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
)

// Path defines default path.
// This will be relative to pwd.
var Path = flag.String("config", "config.json", "Path to json config file.")

// Parse reads config file and parses cli flags into c by calling flag.Parse()
func Parse(c interface{}) {
	// Call Parse() for the first time to get default config path if set.
	flag.Parse()

	parseJSON(*Path, c)
	// Call again to overwrite json values with flags.
	flag.Parse()
}

// parseJSON parses json file configPath into the config struct c.
func parseJSON(configPath string, c interface{}) {
	if configPath == "" {
		return
	}
	file, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		panic(err)
	}
	// Expand env variables in config struct.
	v := reflect.ValueOf(c)
	Expand(v)
}

// Expand expands any environmental variables in config settings recursively.
// For example, on a system where $USER is set to user, $USER will become 'user'
// TODO write tests, specifically for reflect.Slice, Map.
func Expand(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		// Get pointer for reflection
		vv := v.Elem()
		// for nil pointers
		if !vv.IsValid() {
			return
		}
		Expand(vv)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			Expand(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			Expand(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			Expand(v.MapIndex(key))
		}
	case reflect.String:
		str := v.Interface().(string)
		str = os.ExpandEnv(str)
		v.SetString(str) // Value must be exported.
	}
}
