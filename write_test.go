package jsonflag

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/zamicol/json5"
)

// fromFlagDefault is a flag.*Var default copied into WriteDefault fixtures.
var fromFlagDefault string

func init() {
	flag.StringVar(&fromFlagDefault, "from-flag-default", "flagDefault", "from flag")
}

// writeSample is the WriteDefault fixture.
//   - default tags override the current value
//   - manual and flag values are kept when there is no default tag
//   - json:"-" and unexported fields are omitted
type writeSample struct {
	UserName   string        `json:"username" default:"defaultUserName" desc:"User name"`
	Count      int           `json:"count,omitempty" default:"10" desc:"Item count"`
	Enabled    bool          `json:"enabled" default:"true" desc:"Feature toggle"`
	Timeout    time.Duration `json:"timeout" default:"5s" desc:"Operation timeout"`
	Threshold  float64       `json:"threshold" default:"0.75" desc:"Threshold value"`
	MaxRetries uint          `json:"max-retries" default:"3" desc:"Max retry attempts"`
	Big        int64         `json:"big" default:"64"`
	Huge       uint64        `json:"huge" default:"128"`
	Manual     string        `json:"manual" desc:"Set from the struct"`
	FromFlag   string        `json:"fromFlag"`
	Elapsed    time.Duration `json:"elapsed"`
	Fallback   string        `json:",omitempty"`
	Note       string        `json:"note" desc:"line one\nline two"`
	Hidden     string        `json:"-" default:"secret" desc:"hidden"`
	skip       string        `default:"nope" desc:"unexported"`
}

// wantDefaultConfig is the JSON5 document WriteDefault should write for writeSample.
const wantDefaultConfig = `{
    // User name
    "username": "defaultUserName",
    // Item count
    "count": 10,
    // Feature toggle
    "enabled": true,
    // Operation timeout
    "timeout": 5000000000,
    // Threshold value
    "threshold": 0.75,
    // Max retry attempts
    "max-retries": 3,
    "big": 64,
    "huge": 128,
    // Set from the struct
    "manual": "hand",
    "fromFlag": "flagDefault",
    "elapsed": 1500000000,
    "Fallback": "fb",
    // line one
    // line two
    "note": "kept",
}
`

// TestWriteDefault checks the JSON5 written by WriteDefault.
//   - default tags, struct values, and flag defaults
//   - desc comments and time.Duration nanoseconds
//   - round trip through json5.NewDecoder
//   - flag.CommandLine is unchanged
func TestWriteDefault(t *testing.T) {
	prev := fromFlagDefault
	fromFlagDefault = "flagDefault"
	t.Cleanup(func() { fromFlagDefault = prev })

	cfg := writeSample{
		UserName:   "ignore-me",
		Count:      99,
		Enabled:    false,
		Timeout:    time.Second,
		Threshold:  0.1,
		MaxRetries: 9,
		Big:        1,
		Huge:       1,
		Manual:     "hand",
		FromFlag:   fromFlagDefault,
		Elapsed:    1500 * time.Millisecond,
		Fallback:   "fb",
		Note:       "kept",
		Hidden:     "visible",
		skip:       "nope",
	}
	before := commandLineSnapshot()

	// `$` in the name stays literal.
	dir := t.TempDir()
	path := filepath.Join(dir, "$config.json5")
	if err := WriteDefault(path, &cfg); err != nil {
		t.Fatal(err)
	}
	if got := commandLineSnapshot(); !reflect.DeepEqual(got, before) {
		t.Fatalf("flags changed:\n got %v\nwant %v", got, before)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantDefaultConfig {
		t.Fatalf("config =\n%s\nwant =\n%s", got, wantDefaultConfig)
	}
	if bytes.Contains(got, []byte("secret")) || bytes.Contains(got, []byte("unexported")) || bytes.Contains(got, []byte("nope")) {
		t.Fatalf("omitted fields were written:\n%s", got)
	}

	var decoded writeSample
	if err := json5.NewDecoder(bytes.NewReader(got)).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	want := writeSample{
		UserName:   "defaultUserName",
		Count:      10,
		Enabled:    true,
		Timeout:    5 * time.Second,
		Threshold:  0.75,
		MaxRetries: 3,
		Big:        64,
		Huge:       128,
		Manual:     "hand",
		FromFlag:   "flagDefault",
		Elapsed:    1500 * time.Millisecond,
		Fallback:   "fb",
		Note:       "kept",
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("decoded = %+v\nwant %+v", decoded, want)
	}

	// A struct value writes the same document as a pointer.
	valuePath := filepath.Join(dir, "value.json5")
	if err := WriteDefault(valuePath, cfg); err != nil {
		t.Fatal(err)
	}
	valueGot, err := os.ReadFile(valuePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(valueGot) != wantDefaultConfig {
		t.Fatalf("struct value config =\n%s\nwant =\n%s", valueGot, wantDefaultConfig)
	}
	if got := commandLineSnapshot(); !reflect.DeepEqual(got, before) {
		t.Fatalf("flags changed after struct value write:\n got %v\nwant %v", got, before)
	}
}

// TestIndentMultiline checks that a multi-line JSON value stays indented.
func TestIndentMultiline(t *testing.T) {
	got := indentMultiline([]byte("{\n\"k\": 1\n}"))
	want := "{\n    \"k\": 1\n    }"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := indentMultiline([]byte("1")); got != "1" {
		t.Fatalf("single line = %q", got)
	}
}

// TestWriteDefaultErrors checks rejected WriteDefault inputs.
//   - empty path, nil, and non-struct configs
//   - an invalid default tag does not create a file
//   - an existing file (os.ErrExist) is left unchanged
func TestWriteDefaultErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json5")

	t.Run("empty path", func(t *testing.T) {
		if err := WriteDefault("", &writeSample{}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("nil", func(t *testing.T) {
		if err := WriteDefault(path, nil); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("nil pointer", func(t *testing.T) {
		var cfg *writeSample
		if err := WriteDefault(path, cfg); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("not a struct", func(t *testing.T) {
		if err := WriteDefault(path, 5); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("invalid default", func(t *testing.T) {
		cases := []interface{}{
			&struct {
				Count int `default:"nope"`
			}{},
			&struct {
				Timeout time.Duration `default:"nope"`
			}{},
			&struct {
				Nested struct{ N int } `default:"x"`
			}{},
		}
		for i, cfg := range cases {
			badPath := filepath.Join(dir, "bad"+string(rune('0'+i))+".json5")
			err := WriteDefault(badPath, cfg)
			if err == nil {
				t.Fatalf("case %d: expected error", i)
			}
			if _, statErr := os.Stat(badPath); !os.IsNotExist(statErr) {
				t.Fatalf("case %d: invalid default created a file: %v", i, statErr)
			}
		}
	})
	t.Run("existing file", func(t *testing.T) {
		existing := filepath.Join(dir, "exists.json5")
		const keep = "keep\n"
		if err := os.WriteFile(existing, []byte(keep), 0644); err != nil {
			t.Fatal(err)
		}
		err := WriteDefault(existing, &writeSample{})
		if !errors.Is(err, os.ErrExist) {
			t.Fatalf("error = %v, want os.ErrExist", err)
		}
		got, readErr := os.ReadFile(existing)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(got) != keep {
			t.Fatalf("file = %q, want %q", got, keep)
		}
	})
}

// writeFlagConfig is the --write-config fixture used by Parse.
type writeFlagConfig struct {
	Name string        `flag:"write-config-name" json:"name" default:"ada" desc:"Display name"`
	Wait time.Duration `flag:"write-config-wait" json:"wait" default:"2s" desc:"Wait duration"`
}

// wantFlagConfig is the file Parse should write for writeFlagConfig.
const wantFlagConfig = `{
    // Display name
    "name": "ada",
    // Wait duration
    "wait": 2000000000,
}
`

// TestParseWriteConfig checks --write-config.
//   - handled before a missing config file is opened
//   - each accepted spelling exits 0 after writing Path
func TestParseWriteConfig(t *testing.T) {
	restoreCLI(t)
	dir := t.TempDir()
	forms := [][]string{
		{"--write-config", "--config"},
		{"--config", "-write-config"},
		{"--write-config=true", "--config"},
		{"-config", "-write-config=true"},
	}
	for i, form := range forms {
		path := filepath.Join(dir, "out"+string(rune('0'+i))+".json5")
		args := []string{"jsonflag-test"}
		for _, a := range form {
			if a == "--config" || a == "-config" {
				args = append(args, a+"="+path)
				continue
			}
			args = append(args, a)
		}
		os.Args = args
		code := -1
		osExit = func(c int) { code = c }

		var cfg writeFlagConfig
		Parse(&cfg)
		if code != 0 {
			t.Fatalf("args %q exit = %d, want 0", args, code)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("args %q: %v", args, err)
		}
		if string(got) != wantFlagConfig {
			t.Fatalf("args %q config =\n%s\nwant =\n%s", args, got, wantFlagConfig)
		}
		var decoded writeFlagConfig
		if err := json5.NewDecoder(bytes.NewReader(got)).Decode(&decoded); err != nil {
			t.Fatal(err)
		}
		want := writeFlagConfig{Name: "ada", Wait: 2 * time.Second}
		if !reflect.DeepEqual(decoded, want) {
			t.Fatalf("decoded = %+v, want %+v", decoded, want)
		}
	}
}

// TestParseWriteConfigExistingPanics checks that Parse panics when WriteDefault fails.
func TestParseWriteConfigExistingPanics(t *testing.T) {
	restoreCLI(t)
	path := filepath.Join(t.TempDir(), "exists.json5")
	const keep = "keep\n"
	if err := os.WriteFile(path, []byte(keep), 0644); err != nil {
		t.Fatal(err)
	}
	os.Args = []string{"jsonflag-test", "--write-config", "--config=" + path}
	called := false
	osExit = func(int) { called = true }

	defer func() {
		rec := recover()
		if rec == nil {
			t.Fatal("Parse did not panic")
		}
		err, ok := rec.(error)
		if !ok || !errors.Is(err, os.ErrExist) {
			t.Fatalf("panic = %v, want os.ErrExist", rec)
		}
		if called {
			t.Fatal("osExit was called")
		}
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(got) != keep {
			t.Fatalf("file = %q, want %q", got, keep)
		}
	}()

	var cfg writeFlagConfig
	Parse(&cfg)
}

// restoreCLI puts os.Args, Path, osExit, and writeConfig back after the test.
func restoreCLI(t *testing.T) {
	t.Helper()
	args := append([]string(nil), os.Args...)
	path := Path
	exit := osExit
	write := writeConfig
	t.Cleanup(func() {
		os.Args = args
		Path = path
		osExit = exit
		writeConfig = write
	})
}

// commandLineSnapshot copies flag names and values.
func commandLineSnapshot() map[string]string {
	got := map[string]string{}
	flag.VisitAll(func(f *flag.Flag) {
		got[f.Name] = f.Value.String()
	})
	return got
}
