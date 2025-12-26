package jsonflag

import (
	"os"
	"testing"
)

// See the package documentation on how to run a test.
func TestMain(m *testing.M) {
	exampleInitGoFlagConfig() // Must be called in TestMain for CLI to see flags.
	// Example_goFlagConfig() // Run "Go flag design pattern" example.
	// tags()           // Run "tag config design pattern" example.
	//envVarPrefix()   // Run example using environmental variable prefix, useful for "namespacing" config settings to your specific application.
	os.Exit(m.Run()) // Must explicitly exit because of flag test
}

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
