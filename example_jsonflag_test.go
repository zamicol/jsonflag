package jsonflag

import (
	"flag"
	"fmt"
)

// Create new configs.
//
// Although mixing patterns is supported, idiomatic Go code should use only one
// config design pattern: either the standard library's flag package, the "Go
// flag" design pattern, or the "tag config" design pattern.
//
// The Go flag design pattern provides drop in compatibility with Go's libraries
// while the the "tag config" design pattern avoids stuttering.
var fc FlagConfig

// FlagConfig is for testing "Go flag design pattern" flags
// Config's values must be exported for package `flag` to be able to set values.
type FlagConfig struct {
	Flag1 string // Set by flag default, JSON, and CLI - CLI precedence
	Flag2 string // Set by JSON only     - JSON precedence
	Flag3 int    // Set by JSON only int - JSON precedence
	Flag4 string // Set by flag default only - default precedence
	Flag5 string // Set by JSON and flag default - JSON precedence
	Flag6 int    `json:"flagsix"` // Tests JSON tag. JSON with flag default int - JSON precedence.
	Flag7 string // Set by environmental variable and JSON.  No default, or CLI. - Env precedence
	Flag8 string // Set by environmental value expansion from value in JSON config.
	Flag9 string // Test expanding the default flag value ($FLAG7) with a variable to an environmental variable."
}

// flags holds all flag definitions for CLI and application set.
// Flag 2 and 3 are missing to test JSON values which will still populate.
func flags() {
	// Example of using jsonflag as a drop-in replacement for Go's flag.
	flag.StringVar(&fc.Flag1, "flag1", "defaultFlag1", "flag1Desc")
	// Flag 2 and 3 are missing to test JSON values which will still populate.
	flag.StringVar(&fc.Flag4, "flag4", "defaultFlag4", "flag4Desc")
	flag.StringVar(&fc.Flag5, "flag5", "defaultFlag5", "flag5Desc")
	flag.IntVar(&fc.Flag6, "flag6", 1, "flag6Desc") // Set default value to something other than 6 for testing.
	flag.StringVar(&fc.Flag7, "flag7", "defaultFlag7", "Flag7's value comes from environmental variable.")
	flag.StringVar(&fc.Flag8, "flag8", "defaultFlag8", "Flag8 tests environmental expansion.")
	flag.StringVar(&fc.Flag9, "flag9", "$FLAG7", "Flag9's value comes from expanding the default flag value ($FLAG7) with a variable to an environmental variable.")
	Parse(&fc)
}

// Example prints out values
func Example() {
	fmt.Println(fc)
	// Output: {cliFlag1 jsonFlag2 defaultFlag3 jsonFlag4 5 6 FLAG7VALUE Flag8Env FLAG7VALUE}
}
