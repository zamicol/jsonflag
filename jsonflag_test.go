package jsonflag

import (
	"os"
	"testing"
	"time"
)

// See the package documentation on how to run a test.
func TestMain(m *testing.M) {
	exampleInitGoFlagConfig() // Must be called in TestMain for CLI to see flags.
	os.Exit(m.Run())          // Must explicitly exit because of flag test
}

// TagConfig is for testing "tag design pattern" flags
// Config's values must be exported for package `flag` to be able to set values.
type TagConfig struct {
	UserName   string        `flag:"username" default:"defaultUserName" desc:"User name"`
	Count      int           `flag:"count" default:"10" desc:"Item count"`
	Enabled    bool          `flag:"enabled" default:"true" desc:"Feature toggle"`
	Timeout    time.Duration `flag:"timeout" default:"5s" desc:"Operation timeout"`
	Threshold  float64       `flag:"threshold" default:"0.75" desc:"Threshold value"`
	MaxRetries uint          `flag:"max-retries" default:"3" desc:"Max retry attempts"`
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

func TestTagConfig(t *testing.T) {
	os.Setenv("ENABLED", "false")
	os.Setenv("MAX_RETRIES", "5")
	defer os.Unsetenv("ENABLED")
	defer os.Unsetenv("MAX_RETRIES")

	var tc TagConfig

	// Use a test-specific config file
	origPath := Path
	Path = "test_tags.json5" // should contain: {"count": 20, "timeout": "10s"}
	defer func() { Path = origPath }()

	Parse(&tc)

	// Now assert values:
	if tc.UserName != "cliName" { // assuming run with --username=cliName
		t.Errorf("expected cliName, got %s", tc.UserName)
	}
	if tc.Count != 20 {
		t.Errorf("expected 20, got %d", tc.Count)
	}
	if tc.Enabled != false {
		t.Errorf("expected false, got %v", tc.Enabled)
	}
	if tc.Timeout != 10*time.Second {
		t.Errorf("expected 10s, got %v", tc.Timeout)
	}
	if tc.Threshold != 0.75 {
		t.Errorf("expected 0.75, got %f", tc.Threshold)
	}
	if tc.MaxRetries != 5 {
		t.Errorf("expected 5, got %d", tc.MaxRetries)
	}
}
