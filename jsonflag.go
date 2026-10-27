// Package jsonflag provides configuration settings by extending Go's flag
// package with support for JSON/JSON5 config files and environmental variables.
// It does not replace any part of the flag package.
//
// Order of precedence for configuration values:
//  1. Command-line flags (e.g., `--flag1=flag1Value`)
//  2. Environmental variables (e.g., FLAG2=flag2value)
//  3. JSON config values (e.g., `{"flag3": "flag3Value"}`)
//  4. Default values set on flags (e.g., flag.StringVar(&config.Flag4, "flag4Name", "flag4DefaultValue", "flag4Description"))
//     or via struct tags (e.g., `default:"flag4DefaultValue"`)
//
// Flag values are optional in the JSON config file and can be omitted if desired.
// Unrecognized JSON config values (those not in the config struct or unexported)
// are ignored. To prevent a value from being set by flags, include it in the
// config struct and JSON config file; it will still be populated.
//
// jsonflag never expands `$`. Config values and the config path are literal.
// Set a flag from the environment with a per-flag variable such as FLAG1.
//
// # Letter Casing for Flag Names
//
// Flag naming conventions vary by input type:
//   - CLI flag names (not values) must start with lowercase letters (e.g., --flag1).
//   - For environmental variables, flag names are converted to all uppercase,
//     making them case-insensitive (e.g., FLAG1 or flag1 both work). Characters
//     other than ASCII letters, digits, and underscores are replaced with
//     underscores, so dashed and dotted names are settable in a shell:
//     "log-level" is read from LOG_LEVEL and "log.level" from LOG_LEVEL.
//     EnvPrefix is prepended unchanged, so prefix "MYAPP_" and flag
//     "log-level" read MYAPP_LOG_LEVEL.
//   - For JSON names, this package uses Go's json package for decoding. The JSON
//     decoder only accesses exported struct fields and follows this precedence:
//     1. Tags
//     2. Exact case
//     3. Case-insensitive
//   - Flag names in Go code (via the flag package) can be upper or lowercase, but
//     uppercase is recommended to match Go conventions for exported fields.
//
// # Recommended Usage
//
// See the Testing section for a full example.
//  1. Define a config struct with exported fields, optionally using tags:
//     ```go
//     type Config struct {
//     Flag1 string `flag:"flag1" default:"flag1Default" desc:"flag1 description"`
//     Flag2 string // No tags or flags are still okay.
//     }
//     ```
//  2. For non-tag usage, (drop in replacement for Go's `flag` package) use flag functions to set defaults:
//     ```go
//     flag.StringVar(&config.Flag2, "flag2Name", "flag2DefaultValue", "flag2Description")
//     ```
//  3. Add config values to a config.json file (defaults to the current working directory).
//     Use --config=your_config.json to specify a different path.
//  4. Call jsonflag.Parse(&config) to populate the config struct.
//
// # Config Path
//
// Set the config path via command-line:
//
//	--config=your_config.json
//
// Or programmatically (note: CLI flags take precedence over this):
//
//	jsonflag.Path = "assets/config.json"
//
// # Design
//
// This package follows flag.Parse()'s fail-fast design and panics on error.
//
// # Testing
//
// Plain `go test` passes. TestMain fills missing example defaults and leaves
// existing env vars and --flag1/--config flags unchanged:
//   - FLAG6=Flag6EnvValue and JSONFLAG_FLAG10=FLAG10VALUE when unset
//   - --flag1=cliFlag1 and --config=test_config.json5 when absent
//
// These commands still exercise each --config/-config form:
//   - Test 1 (tests --config= form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG6=Flag6EnvValue go test --flag1=cliFlag1 --config=test_config.json5
//   - Test 2 (tests -config= form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG6=Flag6EnvValue go test --flag1=cliFlag1 -config=test_config.json5
//   - Test 3 (tests --config form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG6=Flag6EnvValue go test --flag1=cliFlag1 --config test_config.json5
//   - Test 4 (tests -config form):
//     JSONFLAG_FLAG10=FLAG10VALUE FLAG6=Flag6EnvValue go test --flag1=cliFlag1 -config test_config.json5
package jsonflag

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/DisposaBoy/JsonConfigReader"
)

// Path defines the default config path and is relative to pwd.
var Path string

// EnvPrefix will be prepended to flag names if set. For example, with a prefix
// of "MYAPP_", the flag "flag1" will become "MYAPP_FLAG1" and the flag
// "log-level" will become "MYAPP_LOG_LEVEL". The prefix is not rewritten.
//
// Prefixes are useful for namespacing configurations, ensuring any potential
// name collisions with existing environmental variables are explicitly
// precluded.
var EnvPrefix = ""

func init() {
	// Set default config path to `config.json5`.  Overridable via cli `--config`.
	flag.StringVar(&Path, "config", "config.json5", "Path to json config file.")
}

// Parse reads the config file and parses CLI flags into c with a single flag.Parse() call.
//  1. Parsing --config from command line (if present).
//  2. Registering any missing flags from struct tags.
//  3. Loading and decoding JSON config.
//  4. Applying environment variables.
//  5. Finally parsing CLI flags (highest precedence).
func Parse(c interface{}) {
	// Manually extract --config or -config from os.Args instead of calling
	// flag.Parse twice. By avoiding calling it twice, flag.Parse()'s `--help` works as
	// expected.  Without this, `--help` will not work
	for i := 1; i < len(os.Args); i++ { // Start at 1 to skip program name.
		arg := os.Args[i]
		if arg == "--config" || arg == "-config" {
			if i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "-") {
				Path = os.Args[i+1]
				break
			}
		} else if strings.HasPrefix(arg, "--config=") {
			Path = strings.TrimPrefix(arg, "--config=")
			break
		} else if strings.HasPrefix(arg, "-config=") {
			Path = strings.TrimPrefix(arg, "-config=")
			break
		}
	}

	// flag.*Var writes its default into the field. Register first so JSON, env,
	// and CLI overwrite that default.
	registerTagFlags(c)             // Register struct tags
	parseJSON(Path, c)              // Parse JSON config.
	flag.VisitAll(applyEnvOverride) // Apply environment variables to all defined flags
	flag.Parse()                    // Finally parse CLI flags, override everything else
}

// applyEnvOverride applies environment variables to defined flags. (Example,
// value `$TestEnv` is replaced to `TestValue` if set by an env variable:
// `TESTENV=TestValue`. EnvVars have higher precedence than JSON config but lower than
// CLI flags.
//
// The variable is EnvPrefix plus a shell-safe form of the flag name: uppercased,
// with every character other than an ASCII letter, digit, or underscore replaced
// by '_'. "log-level" uses LOG_LEVEL, or MYAPP_LOG_LEVEL when EnvPrefix is
// "MYAPP_". An unset or empty variable leaves the flag unchanged. flag.Set
// failures panic.
func applyEnvOverride(f *flag.Flag) {
	envName := EnvPrefix + envKey(f.Name)
	value := os.Getenv(envName)
	if value != "" {
		if err := flag.Set(f.Name, value); err != nil {
			panic(fmt.Errorf("jsonflag: failed to set flag --%s from env %s=%q: %w", f.Name, envName, value, err))
		}
	}
}

// envKey converts a flag name into an environment-variable suffix.
// "log-level" and "log.level" both become "LOG_LEVEL".
func envKey(flagName string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, strings.ToUpper(flagName))
}

// parseJSON parses the JSON file at configPath into c.
// The path is used literally; `$` is not expanded.
func parseJSON(configPath string, c interface{}) {
	if configPath == "" {
		return
	}

	var err error
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: cannot get absolute config path: '%s'", err, configPath)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}

	file, err := os.Open(configPath)
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: config '%s' not found", err, configPath)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}
	defer file.Close()

	r := JsonConfigReader.New(file)
	decoder := json.NewDecoder(r)
	err = decoder.Decode(c)
	if err != nil {
		err = fmt.Errorf("%w; jsonflag: unable to decode config", err)
		panic(err) // Fail fast matching flag.Parse() error reporting.
	}
}

// registerTagFlags registers flags based on struct tags if not already defined.
func registerTagFlags(c interface{}) {
	v := reflect.ValueOf(c)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)

		// Skip unexported fields
		if !fv.CanSet() {
			continue
		}

		// Get tag values
		flagName := field.Tag.Get("flag")
		defaultVal := field.Tag.Get("default")
		desc := field.Tag.Get("desc")

		// Default flag name to field name (lowercase) if not specified
		if flagName == "" {
			flagName = strings.ToLower(field.Name)
		}

		// Skip if flag already defined (preserves manual flag definitions)
		if flag.Lookup(flagName) != nil {
			continue
		}

		// Only register flags for fields with explicit tags
		hasTag := field.Tag.Get("flag") != "" ||
			field.Tag.Get("default") != "" ||
			field.Tag.Get("desc") != ""
		if !hasTag {
			continue
		}

		// Register flags based on field type
		switch field.Type.Kind() {
		case reflect.String:
			flag.StringVar(fv.Addr().Interface().(*string), flagName, defaultVal, desc)
		case reflect.Int:
			var def int
			if defaultVal != "" {
				var err error
				def, err = strconv.Atoi(defaultVal)
				if err != nil {
					panic(fmt.Errorf("jsonflag: invalid default int value for %s: %v", flagName, err))
				}
			}
			flag.IntVar(fv.Addr().Interface().(*int), flagName, def, desc)
		case reflect.Int64:
			if field.Type == reflect.TypeOf(time.Duration(0)) {
				var def time.Duration
				if defaultVal != "" {
					var err error
					def, err = time.ParseDuration(defaultVal)
					if err != nil {
						panic(fmt.Errorf("jsonflag: invalid default duration value for %s: %v", flagName, err))
					}
				}
				flag.DurationVar(fv.Addr().Interface().(*time.Duration), flagName, def, desc)
			} else {
				var def int64
				if defaultVal != "" {
					var err error
					def, err = strconv.ParseInt(defaultVal, 10, 64)
					if err != nil {
						panic(fmt.Errorf("jsonflag: invalid default int64 value for %s: %v", flagName, err))
					}
				}
				flag.Int64Var(fv.Addr().Interface().(*int64), flagName, def, desc)
			}
		case reflect.Uint:
			var def uint
			if defaultVal != "" {
				var err error
				def64, err := strconv.ParseUint(defaultVal, 10, 32)
				if err != nil {
					panic(fmt.Errorf("jsonflag: invalid default uint value for %s: %v", flagName, err))
				}
				def = uint(def64)
			}
			flag.UintVar(fv.Addr().Interface().(*uint), flagName, def, desc)
		case reflect.Uint64:
			var def uint64
			if defaultVal != "" {
				var err error
				def, err = strconv.ParseUint(defaultVal, 10, 64)
				if err != nil {
					panic(fmt.Errorf("jsonflag: invalid default uint64 value for %s: %v", flagName, err))
				}
			}
			flag.Uint64Var(fv.Addr().Interface().(*uint64), flagName, def, desc)
		case reflect.Float64:
			var def float64
			if defaultVal != "" {
				var err error
				def, err = strconv.ParseFloat(defaultVal, 64)
				if err != nil {
					panic(fmt.Errorf("jsonflag: invalid default float64 value for %s: %v", flagName, err))
				}
			}
			flag.Float64Var(fv.Addr().Interface().(*float64), flagName, def, desc)
		case reflect.Bool:
			var def bool
			if defaultVal != "" {
				var err error
				def, err = strconv.ParseBool(defaultVal)
				if err != nil {
					panic(fmt.Errorf("jsonflag: invalid default bool value for %s: %v", flagName, err))
				}
			}
			flag.BoolVar(fv.Addr().Interface().(*bool), flagName, def, desc)
		}
	}
}
