package jsonflag

import (
	"flag"
	"fmt"
)

// Create a new config.
var tc Config

// Config's values must be exported
type Config struct {
	Flag1 string // populate everywhere (flag default, json, cli flag)
	Flag2 string // json only
	Flag3 string // flag default only
	Flag4 string // json with flag default (no cli flag)
	Flag5 int    // json only int
	Flag6 int    // json with flag default (no cli flag) int
}

// flags holds all flag definitions for CLI and application set.
// Flag 2 and 5 are missing.  JSON values will still populate.
func flags() {
	flag.StringVar(&tc.Flag1, "flag1", "defaultFlag1", "flag1Desc")
	flag.StringVar(&tc.Flag3, "flag3", "defaultFlag3", "flag3Desc")
	flag.StringVar(&tc.Flag4, "flag4", "defaultFlag4", "flag4Desc")
	flag.IntVar(&tc.Flag6, "flag6", 1, "flag6Desc") // Set default to something other than 6 for testing.

	Parse(&tc)
}

// ExampleConfig prints out values
func Example() {
	fmt.Println(tc)
	// Output: {cliFlag1 jsonFlag2 defaultFlag3 jsonFlag4 5 6}
}
