package jsonflag

import (
	"flag"
	"fmt"
)

// Create two new configs.  Typical usage uses only one config.
var tc Config
var tc2 Config2

// Config's values must be exported
type Config struct {
	Flag1 string // Set by flag default, JSON, and CLI - CLI precedence
	Flag2 string // Set by JSON only - JSON precedence
	Flag3 string // Set by flag default only - default precedence
	Flag4 string // Set by JSON and flag default - JSON precedence
	Flag5 int    // Set by JSON only int - JSON precedence
	Flag6 int    `json:"flagsix"` // JSON with flag default int - JSON precedence.  Also tests tags.
	Flag7 string // Set by environmental variable and JSON.  No default, or CLI. - Env precedence
	Flag8 string // Set by environmental value expansion from value in JSON config.
	Flag9 string // Test expanding the default flag value ($FLAG7) with a variable to an environmental variable."
}

type Config2 struct {
	Flag10 string // Test EnvPrefix.  Value is set by CLI only.
}

// flags holds all flag definitions for CLI and application set.
// Flag 2 and 5 are missing to test JSON values which will still populate.
func flags() {
	flag.StringVar(&tc.Flag1, "flag1", "defaultFlag1", "flag1Desc")
	flag.StringVar(&tc.Flag3, "flag3", "defaultFlag3", "flag3Desc")
	flag.StringVar(&tc.Flag4, "flag4", "defaultFlag4", "flag4Desc")
	flag.IntVar(&tc.Flag6, "flag6", 1, "flag6Desc") // Set default value to something other than 6 for testing.
	flag.StringVar(&tc.Flag7, "flag7", "defaultFlag7", "Flag7's value comes from environmental variable.")
	flag.StringVar(&tc.Flag8, "flag8", "defaultFlag8", "Flag8 tests environmental expansion.")
	flag.StringVar(&tc.Flag9, "flag9", "$FLAG7", "Flag9's value comes from expanding the default flag value ($FLAG7) with a variable to an environmental variable.")
	Parse(&tc)

	EnvPrefix = "JSONFLAG_"
	flag.StringVar(&tc2.Flag10, "flag10", "", "Flag10 tests prefixing the EnvPrefix to env vars")
	Parse(&tc2)
}

// Example prints out values
func Example() {
	fmt.Println(tc)
	fmt.Println(tc2)
	// Output: {cliFlag1 jsonFlag2 defaultFlag3 jsonFlag4 5 6 FLAG7VALUE Flag8Env FLAG7VALUE}
	// {FLAG10VALUE}
}
