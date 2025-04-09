// Package jsonflag is for configuration settings.  It extends Go's flag
// package, designed for cli flags, with json config files and environmental
// variables.  It does not replace any part of flag.
//
// Order of precedence for set config values:
//
//  1. Command line flags. (cli Example: `--flag1=flag1Value`)
//  2. Environmental Variables (env Example: FLAG2=flag2value)
//  3. JSON config values. (json Example: `{"flag3": "flag3Value"}`)
//  4. Default values set on flag. (go Example: `flag.StringVar(&config.Flag4, "flag4Name", "flag4DefaultValue", "flag4Description")`)
//
// Flag values do not need to appear in the json config file and can be left
// blank if desired. If not set in config struct or exported, extra json config
// file values will be ignored. If a value should not be set by flags, add the
// value in the config struct and json config file.  It will still be set.
//
// Environmental variables can be set using the flag name.  The flag's name will
// be converted to all upper case.  If set, "EnvPrefix" will be prefixed when
// looking up environment variables.
//
// Environmental variables in json config will be expanded.  See testing for an
// example where "$Flag8" is expanded.
//
// This package uses Go's json package for decoding.  The json decoder only
// has accesses to exported fields of structs and follows its own
// precedence for json decoding:
//
//  1. Tags
//  2. Exact case
//  3. Case insensitive
//
// CLI names must start with lower case.
//
// # Recommended Usage
//
// See testing for an example.
//
//  1. Define a `config` struct with exported fields.
//  2. Use flag's functions to set default config values. `flag.StringVar(&config.Flag1, "flag1Name", "flag1DefaultValue", "flag1Description")`
//  3. Put config values in a `config.json` file. The config file path defaults to the cwd.  You can use `--config=your_config.json` to point somewhere else.
//  4. Call `jsonflag.Parse(&config)`
//
// The config struct is now appropriately populated.
//
// # Config Path
//
// You can set the config path via the cli,
//
//	--config=your_config.json
//
// You can also set it in your application.  Note that this can be overwritten
// by the normal precedence via a cli flag as previously mentioned.
//
//	jsonflag.Path = "assets/config.json"
//
// # Design
//
// This package follows flag.Parse() fail fast design and panics on error.
//
// # Testing
//
// Since this package uses flag, test functions need a cli flag passed to verify
// cli parsing is working.  Test will otherwise fail.
//
// Test 1 (tests --config= form):
//
//	JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 --config=test_config.json5
//
// Test 2 (tests -config= form):
//
// JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 -config=test_config.json5
//
// Test 3 (tests --config form):
//
// JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 --config test_config.json5
//
// Test 4 (tests -config form):
//
// JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 -config test_config.json5
package jsonflag

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/DisposaBoy/JsonConfigReader"
)

// Path defines the default config path and is relative to pwd.
var Path string

// EnvPrefix will be prepended to flag names if set. For example, with a prefix
// of "MYAPP_", the flag "flag1" will become "MYAPP_FLAG1".
var EnvPrefix = ""

func init() {
	flag.StringVar(&Path, "config", "config.json5", "Path to json config file.")
}

// Parse reads the config file and parses CLI flags into c with a single flag.Parse() call.
func Parse(c interface{}) {
	// Manually extract --config or -config from os.Args instead of calling
	// flag.Parse twice. By avoiding calling flag.Parse() twice help works as
	// expected.
	for i := 1; i < len(os.Args); i++ { // Start at 1 to skip program name.
		arg := os.Args[i]
		if arg == "--config" || arg == "-config" {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				Path = os.Args[i+1]
				break
			}
		} else if strings.HasPrefix(arg, "--config=") {
			Path = strings.TrimPrefix(arg, "--config=")
			break
		} else if strings.HasPrefix(arg, "-config=") {
			Path = strings.TrimPrefix(arg, "-config=")
			break
		}
	}

	// Parse the JSON config first, using the determined Path.
	parseJSON(Path, c)

	// Set environmental variables on all flags.
	flag.VisitAll(env)

	// Single call to parse CLI flags, overriding JSON/env values as needed.
	flag.Parse()
}

// env sets environmental values on all flags based on flag name.
func env(f *flag.Flag) {
	v := os.Getenv(EnvPrefix + strings.ToUpper(f.Name))
	if v != "" {
		flag.Set(f.Name, v)
	}
}

// parseJSON parses the JSON file at configPath into the config struct c.
func parseJSON(configPath string, c interface{}) {
	if configPath == "" {
		return
	}

	var err error
	configPath, err = filepath.Abs(os.ExpandEnv(configPath))
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: cannot get absolute config path: '%s'", err, configPath)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}

	file, err := os.Open(configPath)
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: config '%s' not found", err, configPath)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}
	defer file.Close()

	r := JsonConfigReader.New(file)
	decoder := json.NewDecoder(r)
	err = decoder.Decode(c)
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: unable to decode config", err)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}

	// Expand env variables in the config struct.
	v := reflect.ValueOf(c)
	expand(v)
}

// Expand recursively expands from interface{} any structs, slices, pointers,
// and maps looking for variables with the  underlying type of string.  If the
// underlying type is string, it will attempt to expand any environmental
// variable.
//
// For an environmental variable expansion example, on a system where $USER is
// set to user, $USER will become 'user'
func expand(v reflect.Value) {
	switch v.Kind() {
	case reflect.Ptr:
		vv := v.Elem()     // Get value pointer is pointing to.
		if !vv.IsValid() { // For nil pointers
			return
		}
		expand(vv)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			expand(v.Field(i))
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			expand(v.Index(i))
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			expand(v.MapIndex(key))
		}
	case reflect.String:
		str := v.String()
		str = os.ExpandEnv(str)
		v.SetString(str)
	}
}
