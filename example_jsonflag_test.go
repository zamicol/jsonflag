package jsonflag

import (
	"flag"
	"fmt"
)

// goFlagConfig is for testing "Go flag" design pattern.
//
// Although mixing patterns is supported, idiomatic Go code should use only one
// config design pattern: either the standard library's flag package, the "Go
// flag" design pattern, or the "tag config" design pattern.
//
// The Go flag design pattern provides drop in compatibility with Go's libraries
// while the the "tag config" design pattern avoids stuttering.
//
// Config's values must be exported for package `flag` to be able to set values.
type goFlagConfig struct {
	Flag1 string // Set by flag default, JSON, and CLI - CLI precedence
	Flag2 string // Set by JSON (string) only     - JSON precedence
	Flag3 int    // Set by JSON (int)    only     - JSON precedence
	Flag4 string // Set by flag default only      - Flag default precedence
	Flag5 string // Set by JSON and flag default  - JSON precedence

	// Environment Variables (EnvVars) and environmental variable expansion
	Flag6 string // Set by environmental variable, flag default, and JSON. No CLI, no expansion. - Env precedence
	Flag7 string // Set by environmental value expansion, flag default, and JSON (pre-expansion).  No pure environmental value or CLI. - JSON precedence then unexpanded.
	Flag8 string // Test expanding the flag default value ($FlagDefault8) with a value from an environmental variable. No JSON or CLI.  Expansion variable set in ENV and pre-expansion value set in JSON config.

	//Flag6 int    `json:"9999"` // Tests JSON tag. JSON with flag default int and JSON tag - (Three places are set: flag default, json tag, and JSON) - JSON precedence.
}

// fc is the global for Go Flag examples.
var fc goFlagConfig

// exampleInitFlags is an example of using jsonflag as a drop-in replacement for Go's flag.  Do not run in testing since it is not exported.
// how an application can initialize flag definitions using the "Go flag" design pattern.
func exampleInitGoFlagConfig() {
	flag.StringVar(&fc.Flag1, "flag1", "flagDefault1", "flag1Desc") // Basic CLI flag
	// Flag 2 and 3 are missing here in order to test values that populate only from the JSON config.
	flag.StringVar(&fc.Flag4, "flag4", "flagDefault4", "flag4Desc")
	flag.StringVar(&fc.Flag5, "flag5", "flagDefault5", "flag5Desc")

	//Env Vars
	flag.StringVar(&fc.Flag6, "flag6", "flagDefault6", "Flag6's value comes from environmental variable.")
	flag.StringVar(&fc.Flag7, "flag7", "flagDefault7", "Flag7 tests environmental expansion for JSON config.")
	flag.StringVar(&fc.Flag8, "flag8", "$FLAG8ENVEXPANSION", "Flag8 tests environmental expansion for default flag value.")

	//flag.StringVar(&fc.Flag8, "flag8", "$FLAG8", "Flag8's value comes from expanding the default flag value ($FLAG) with a variable to an environmental variable.")

	//flag.IntVar(&fc.Flag6, "flag6", 1111, "flag6Desc") // Default value to '1111' for testing.  (JSON config set to '6')
	Parse(&fc)
}

// Example prints out values
// go test -run Example_goFlagConfig --config=test_config.json5
//
//	FLAG6=Flag6EnvValue FLAG7ENVEXPANSION=Flag7EnvExpansionValue  FLAG8ENVEXPANSION=Flag8EnvExpansionValue go test -run Example_goFlagConfig --flag1=cliFlag1 --config=test_config.json5
//	Deprecated: JSONFLAG_FLAG10=FLAG10EnvValue
func Example_goFlagConfig() {
	//exampleInitGoFlagConfig() is called by TestMain since it uses the flag
	//package cannot be called inside of an example since flags must be
	//registered before by the binary.  This breaks the typical, idiomatic go
	//testing pattern, but this is the only way that this is possible since Go
	//flag's package does not export the variables that would be needed for the
	//idiomatic pattern.

	fmt.Println(fc)
	// Output: {cliFlag1 jsonFlag2 3 flagDefault4 jsonFlag5 Flag6EnvValue Flag7EnvExpansionValue Flag8EnvExpansionValue}
}

var envPreConfig EnvPrefixConfig

// EnvPrefixConfig is for testing environmental prefix which is useful for
// namespacing an application's configuration options on a system wide level.
type EnvPrefixConfig struct {
	Flag10 string // Test EnvPrefix.  Value is set by CLI only and is prefixed by an environmental prefix.
}

// ExampleEnvPrefix is an example using Environmental variable prefix.
//
// Prefixes are useful for namespacing configurations, ensuring any potential
// name collisions with existing environmental variables are explicitly
// precluded.
//
// Test with:
// JSONFLAG_FLAG10=FLAG10VALUE go test -run ExampleEnvPrefix --config=test_config.json5
func ExampleEnvPrefix() {
	fmt.Println("ExampleEnvPrefix")
	EnvPrefix = "JSONFLAG_"
	flag.StringVar(&envPreConfig.Flag10, "flag10", "defaultValue", "Flag10 tests prefixing the EnvPrefix to env vars")
	Parse(&envPreConfig)
	fmt.Println(envPreConfig)
	// // Output: {FLAG10VALUE}
}
