// Package jsonflags is a simple example of how to use JSON configs in
// conjunction with Go's flag package.
//
// This package uses Go's json package for decoding.  The json decoder only
// has accesses the exported fields of struct types and follows it's own
// precedence for decoding, namely tags, exact case, and non-case sensitive.
//
// Flag values do not need to appear in the config file
// and extra config file values will be ignore.
//
// Order of Precedence for defined values
//    1. Command-line flags. (cli Example: `--flag1=flag1Value`)
//    2. JSON values (json Example: `{"flag2": "flag2Value"}`)
//    3. Default flag values (go Example: `flag.StringVar(&config.Flag3, "flag3", "flag3Value", "flag3Desc")`)
//
// Recommended Usage
//    1. Define a config struct with fields.
//    2. Use flag's functions to set default config values such as `flag.StringVar(&config.Flag1, "flag1", "defaultFlag1value", "flag1Desc")`
//    3. Put configs in a `config.json` file. (or use --config=foobar to point somewhere else)
//    4. Call `jsonflags.Parse(&config)`
// The config struct is now appropriately populated.
//
// Testing
//
// Since this package uses flag, test functions need a cli flag passed to verify
// cli parsing is working.
//    go test --flag1=paramFlag1 --config=test_config.json
//
package jsonflags

import (
	"encoding/json"
	"flag"
	"os"
)

// flag for config path
var config = flag.String("config", "config.json", "Path to json config file.")

// Parse reads config file and parses cli flags into c by calling flag.Parse()
func Parse(c interface{}) {
	flag.Parse() // Parse first time to get config path.
	parseJSON(*config, c)
	flag.Parse() // Call again to overwrite json values with flags.
}

func parseJSON(configPath string, c interface{}) {
	if *config == "" {
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
}
