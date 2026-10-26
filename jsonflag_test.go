package jsonflag

import (
	"os"
	"strings"
	"testing"
	"time"
)

// See the package documentation on how to run a test.
//
// Bare `go test` omits the example env vars and CLI flags. Fill those in
// before flag.Parse, and only when the caller did not set them:
//   - FLAG6=Flag6EnvValue and JSONFLAG_FLAG10=FLAG10VALUE when unset
//   - --flag1=cliFlag1 and --config=test_config.json5 when absent
func TestMain(m *testing.M) {
	setExampleDefaults()
	exampleInitGoFlagConfig() // Must be called in TestMain for CLI to see flags.
	os.Exit(m.Run())          // Must explicitly exit because of flag test
}

// setExampleDefaults supplies Example_goFlagConfig and ExampleEnvPrefix inputs
// when they are missing. Existing env vars and CLI flags are left unchanged.
func setExampleDefaults() {
	setEnvIfUnset("FLAG6", "Flag6EnvValue")
	setEnvIfUnset("JSONFLAG_FLAG10", "FLAG10VALUE")
	if !flagOnArgs("config") {
		// Equals form, after test flags, so flag.Parse accepts it.
		os.Args = append(os.Args, "--config=test_config.json5")
	}
	if !flagOnArgs("flag1") {
		os.Args = append(os.Args, "--flag1=cliFlag1")
	}
}

// setEnvIfUnset sets key to val only when the variable is unset.
func setEnvIfUnset(key, val string) {
	if _, ok := os.LookupEnv(key); ok {
		return
	}
	_ = os.Setenv(key, val)
}

// flagOnArgs reports whether name is already on os.Args.
// Forms: --name=v, -name=v, --name v, -name v.
func flagOnArgs(name string) bool {
	long, short := "--"+name, "-"+name
	longEq, shortEq := long+"=", short+"="
	for _, arg := range os.Args[1:] {
		switch {
		case arg == long || arg == short:
			return true
		case strings.HasPrefix(arg, longEq) || strings.HasPrefix(arg, shortEq):
			return true
		}
	}
	return false
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

// TestTagConfig checks tag-flag precedence.
//
//   - JSON: username, count, timeout
//   - env: enabled, max-retries
//   - default: threshold
//
// --config overrides Path, so the same keys are in test_config.json5.
// --username is not passed: TestMain parses flags before tag flags exist.
func TestTagConfig(t *testing.T) {
	// Host env outranks JSON. USERNAME is commonly set.
	os.Unsetenv("USERNAME")
	os.Unsetenv("COUNT")
	os.Unsetenv("TIMEOUT")
	os.Unsetenv("THRESHOLD")
	os.Setenv("ENABLED", "false")
	os.Setenv("MAX_RETRIES", "5")
	defer os.Unsetenv("ENABLED")
	defer os.Unsetenv("MAX_RETRIES")

	var tc TagConfig

	// Use a test-specific config file. --config overrides Path inside Parse.
	origPath := Path
	Path = "test_tags.json5" // username, count, timeout
	defer func() { Path = origPath }()

	Parse(&tc)

	// Now assert values:
	if tc.UserName != "cliName" { // JSON value in the test configs
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
