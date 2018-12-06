package jsonflag

import (
	"flag"
	"fmt"
)

// Create a new config.
var tc Config

// Config's values must be exported
type Config struct {
	Flag1 string // Set by flag default, json, and CLI - CLI precedence
	Flag2 string // Set by json only - json  precedence
	Flag3 string // Set by flag default only - default precedence
	Flag4 string // Set by json and flag default - json precedence
	Flag5 int    // json only int - json precedence
	Flag6 int    // json with flag default int - json precedence
	Flag7 string // Test environmental variable retrieval.  No json, default, or CLI. - Env precedence
	Flag8 string // Test environmental value expansion from value in json config.
}

// flags holds all flag definitions for CLI and application set.
// Flag 2 and 5 are missing to test jsons values which will still populate.
func flags() {
	flag.StringVar(&tc.Flag1, "flag1", "defaultFlag1", "flag1Desc")
	flag.StringVar(&tc.Flag3, "flag3", "defaultFlag3", "flag3Desc")
	flag.StringVar(&tc.Flag4, "flag4", "defaultFlag4", "flag4Desc")
	flag.IntVar(&tc.Flag6, "flag6", 1, "flag6Desc") // Set default value to something other than 6 for testing.
	flag.StringVar(&tc.Flag7, "FLAG7", "defaultFlag7", "Flag7's value comes from environmental variable.")
	flag.StringVar(&tc.Flag8, "Flag8", "defaultFlag8", "Flag8 test environmental expansion.")

	Parse(&tc)
}

// Example prints out values
func Example() {
	fmt.Println(tc)
	// Output: {cliFlag1 jsonFlag2 defaultFlag3 jsonFlag4 5 6 FLAG7VALUE Flag8Env}
}
