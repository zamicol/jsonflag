package jsonflag

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/zamicol/json5"
)

// writeSample is the WriteDefault fixture.
//   - default and desc tags, a current value, and a duration
//   - json:"-" and unexported fields are omitted
type writeSample struct {
	Name    string        `json:"name" default:"ada" desc:"Display name"`
	Manual  string        `json:"manual"`
	Timeout time.Duration `json:"timeout" default:"5s" desc:"Operation timeout"`
	Hidden  string        `json:"-" default:"secret"`
	skip    string
}

// wantWriteConfig is the JSON5 document for writeSample.
const wantWriteConfig = `{
    // Display name
    "name": "ada",
    "manual": "hand",
    // Operation timeout
    "timeout": 5000000000,
}
`

// TestWriteDefault round-trips a default config through json5.NewDecoder.
func TestWriteDefault(t *testing.T) {
	cfg := writeSample{Name: "no", Manual: "hand", Timeout: time.Second, Hidden: "x", skip: "y"}
	path := filepath.Join(t.TempDir(), "cfg.json5")
	if err := WriteDefault(path, &cfg); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantWriteConfig {
		t.Fatalf("config =\n%s\nwant =\n%s", got, wantWriteConfig)
	}
	var decoded writeSample
	if err := json5.NewDecoder(bytes.NewReader(got)).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	want := writeSample{Name: "ada", Manual: "hand", Timeout: 5 * time.Second}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("decoded = %+v\nwant %+v", decoded, want)
	}
}

// TestWriteDefaultExisting refuses to overwrite a file.
func TestWriteDefaultExisting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "exists.json5")
	if err := os.WriteFile(path, []byte("keep\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := WriteDefault(path, &writeSample{}); !errors.Is(err, os.ErrExist) {
		t.Fatalf("error = %v, want os.ErrExist", err)
	}
}

// TestParseWriteConfig checks that --write-config writes Path and exits 0.
func TestParseWriteConfig(t *testing.T) {
	args, path, exit, write := os.Args, Path, osExit, writeConfig
	t.Cleanup(func() { os.Args, Path, osExit, writeConfig = args, path, exit, write })

	out := filepath.Join(t.TempDir(), "out.json5")
	os.Args = []string{"jsonflag-test", "--write-config", "--config=" + out}
	code := -1
	osExit = func(c int) { code = c }

	Parse(&writeSample{Manual: "hand"})
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != wantWriteConfig {
		t.Fatalf("config =\n%s\nwant =\n%s", got, wantWriteConfig)
	}
}
