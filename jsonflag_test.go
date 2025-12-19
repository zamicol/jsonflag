package jsonflag

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

var envPreConfig EnvPrefixConfig

var goldenEnvPrefix = EnvPrefixConfig{
	Flag10: "FLAG10VALUE",
}

// EnvPrefixConfig is for testing environmental prefix which is useful for
// namespacing you application's configuration options on a system wide level.
type EnvPrefixConfig struct {
	Flag10 string // Test EnvPrefix.  Value is set by CLI only and is prefixed by an environmental prefix.
}

// ExampleEnvPrefix is an example using Environmental variable prefix.
//
// Prefixes are useful for namespacing configurations, ensuring any potential
// name collisions with existing environmental variables are explicitly
// precluded.
func ExampleEnvPrefix() {
	EnvPrefix = "JSONFLAG_"
	flag.StringVar(&envPreConfig.Flag10, "flag10", "defaultValue", "Flag10 tests prefixing the EnvPrefix to env vars")
	Parse(&envPreConfig)
	fmt.Println(envPreConfig)
	// // Output: {FLAG10VALUE}
}

// See the package documentation on how to run a test.
func TestMain(m *testing.M) {
	flags() // Run "Go flag design pattern" example.
	// tags()           // Run "tag config design pattern" example.
	//envVarPrefix()   // Run example using environmental variable prefix, useful for "namespacing" config settings to your specific application.
	os.Exit(m.Run()) // Must explicitly exit because of flag test
}

// Golden test values.
// See struct FlagConfig for expected behavior.
var golden = FlagConfig{
	Flag1: "cliFlag1",
	Flag2: "jsonFlag2",
	Flag3: 3,
	Flag4: "defaultFlag4",
	Flag5: "jsonFlag5",
	Flag6: 6,
	Flag7: "FLAG7VALUE",
	Flag8: "Flag8Env",
	Flag9: "FLAG7VALUE",
}

// TestVerifyCorrectFlags
// test with:
// JSONFLAG_FLAG10=FLAG10VALUE FLAG7=FLAG7VALUE Flag8=Flag8Env go test -run TestVerifyCorrectFlags --flag1=cliFlag1 --config=test_config.json5
func TestVerifyCorrectFlags(t *testing.T) {
	if fc.Flag1 != golden.Flag1 {
		mismatchError("Flag1", golden.Flag1, fc.Flag1, t)
	}
	if fc.Flag2 != golden.Flag2 {
		mismatchError("Flag2", golden.Flag2, fc.Flag2, t)
	}
	if fc.Flag3 != golden.Flag3 {
		mismatchError("Flag3", fmt.Sprintf("%v", golden.Flag3), fmt.Sprintf("%v", fc.Flag3), t)
	}
	if fc.Flag4 != golden.Flag4 {
		mismatchError("Flag4", golden.Flag4, fc.Flag4, t)
	}
	if fc.Flag5 != golden.Flag5 {
		mismatchError("Flag5", golden.Flag5, fc.Flag5, t)
	}
	if fc.Flag6 != golden.Flag6 {
		mismatchError("Flag6", fmt.Sprintf("%v", golden.Flag6), fmt.Sprintf("%v", fc.Flag6), t)
	}
	if fc.Flag7 != golden.Flag7 {
		mismatchError("Flag7", golden.Flag7, fc.Flag7, t)
	}
	if fc.Flag8 != golden.Flag8 {
		mismatchError("Flag8", golden.Flag8, fc.Flag8, t)
	}
	if fc.Flag9 != golden.Flag9 {
		mismatchError("Flag9", golden.Flag9, fc.Flag9, t)
	}

	// // Environmental Prefix Example
	// if envPreConfig.Flag10 != goldenEnvPrefix.Flag10 {
	// 	mismatchError("Flag10", goldenEnvPrefix.Flag10, envPreConfig.Flag10, t)
	// }
}

// // Golden test values for tag-based test.
// var tagGolden = TagConfig{
// 	UserName:   "cliName",        // CLI override
// 	Count:      20,               // JSON value
// 	Enabled:    false,            // Env override
// 	Timeout:    10 * time.Second, // JSON value
// 	Threshold:  0.75,             // Default value
// 	MaxRetries: 5,                // Env override
// }

// // TestVerifyTagFlags
// // test with:
// // ENABLED=false MAX_RETRIES=5 go test -run TestVerifyTagFlags --username=cliName --config=test_tags.json5
// func TestVerifyTagFlags(t *testing.T) {
// 	var tagConfig TagConfig

// 	// Temporarily override Path to use tag-specific config (tests may be ran out of order)
// 	origPath := Path
// 	Path = "test_tags.json5"
// 	defer func() { Path = origPath }()

// 	// Parse flags
// 	Parse(&tagConfig)

// 	fmt.Println(tagConfig)

// 	func tags() {
// 		Parse(&tc) // TagConfig
// 		fmt.Println(tc)
// 		// Output: {defaultUserName 10 true 5s 0.75 3}
// 	}

// 	// Verify values
// 	if tagConfig.UserName != tagGolden.UserName {
// 		mismatchError("UserName", tagGolden.UserName, tagConfig.UserName, t)
// 	}
// 	if tagConfig.Count != tagGolden.Count {
// 		mismatchError("Count", fmt.Sprintf("%v", tagGolden.Count), fmt.Sprintf("%v", tagConfig.Count), t)
// 	}
// 	if tagConfig.Enabled != tagGolden.Enabled {
// 		mismatchError("Enabled", fmt.Sprintf("%v", tagGolden.Enabled), fmt.Sprintf("%v", tagConfig.Enabled), t)
// 	}
// 	if tagConfig.Timeout != tagGolden.Timeout {
// 		mismatchError("Timeout", fmt.Sprintf("%v", tagGolden.Timeout), fmt.Sprintf("%v", tagConfig.Timeout), t)
// 	}
// 	if tagConfig.Threshold != tagGolden.Threshold {
// 		mismatchError("Threshold", fmt.Sprintf("%v", tagGolden.Threshold), fmt.Sprintf("%v", tagConfig.Threshold), t)
// 	}
// 	if tagConfig.MaxRetries != tagGolden.MaxRetries {
// 		mismatchError("MaxRetries", fmt.Sprintf("%v", tagGolden.MaxRetries), fmt.Sprintf("%v", tagConfig.MaxRetries), t)
// 	}
// }

func mismatchError(what string, expected string, got string, t *testing.T) {
	t.Error(what+" set incorrectly. Expected", expected, "Got", got)
}
