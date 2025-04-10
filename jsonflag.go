// Package jsonflag provides configuration settings by extending Go's flag
// package with support for JSON/JSON5 config files and environmental variables.
// It does not replace any part of the flag package.
//
// Order of precedence for configuration values:
//  1. Command-line flags (e.g., `--flag1=flag1Value`)
//  2. Environmental variables (e.g., FLAG2=flag2value)
//  3. JSON config values (e.g., `{"flag3": "flag3Value"}`)
//  4. Default values set on flags (e.g., flag.StringVar(&config.Flag4, "flag4Name", "flag4DefaultValue", "flag4Description"))
//
// Flag values are optional in the JSON config file and can be omitted if desired.
// Unrecognized JSON config values (those not in the config struct or unexported)
// are ignored. To prevent a value from being set by flags, include it in the
// config struct and JSON config file; it will still be populated.
//
// Environmental variables in JSON configs are expanded. See the Testing section
// for an example where "$Flag8" is replaced with its environmental value.
//
// # Letter Casing for Flag Names
//
// Flag naming conventions vary by input type:
//   - CLI flag names (not values) must start with lowercase letters (e.g., --flag1).
//   - For environmental variables, flag names are converted to all uppercase,
//     making them case-insensitive (e.g., FLAG1 or flag1 both work).
//   - For JSON names, this package uses Go's json package for decoding. The JSON
//     decoder only accesses exported struct fields and follows this precedence:
//     1. Tags
//     2. Exact case
//     3. Case-insensitive
//   - Flag names in Go code (via the flag package) can be upper or lowercase, but
//     uppercase is recommended to match Go conventions for exported fields.
//
// # Recommended Usage
//
// See the Testing section for a full example.
//  1. Define a config struct with exported fields.
//  2. Use flag functions to set defaults (e.g., flag.StringVar(&config.Flag1, "flag1Name", "flag1DefaultValue", "flag1Description")).
//  3. Add config values to a config.json file (defaults to the current working directory).
//     Use --config=your_config.json to specify a different path.
//  4. Call jsonflag.Parse(&config) to populate the config struct.
//
// # Config Path
//
// Set the config path via command-line:
//
//	--config=your_config.json
//
// Or programmatically (note: CLI flags take precedence over this):
//
//	jsonflag.Path = "assets/config.json"
//
// # Design
//
// This package follows flag.Parse()'s fail-fast design and panics on error.
//
// # Testing
//
// Since jsonflag builds on flag, tests must include CLI flags to verify parsing.
// Without flags, tests may fail. Example test commands:
//   - Test 1 (tests --config= form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 --config=test_config.json5
//   - Test 2 (tests -config= form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 -config=test_config.json5
//   - Test 3 (tests --config form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 --config test_config.json5
//   - Test 4 (tests -config form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 -config test_config.json5
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
