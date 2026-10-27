package jsonflag

import (
	"flag"
	"os"
	"testing"
)

// Flags with non-identifier names. Registered in init so flag.Parse inside
// TestMain visits them the same way it visits application flags.
var (
	envLogLevel string
	envDBHost   string
	envBindAddr string
)

func init() {
	flag.StringVar(&envLogLevel, "log-level", "default-log-level", "dashed flag name")
	flag.StringVar(&envDBHost, "db.host", "default-db-host", "dotted flag name")
	flag.StringVar(&envBindAddr, "server.bind-addr", "default-bind-addr", "dotted and dashed flag name")
}

func TestEnvKey(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "log-level", want: "LOG_LEVEL"},
		{name: "log.level", want: "LOG_LEVEL"},
		{name: "Log-Level", want: "LOG_LEVEL"},
		{name: "flag1", want: "FLAG1"},
		{name: "flag_2", want: "FLAG_2"},
		{name: "server.bind-addr", want: "SERVER_BIND_ADDR"},
		{name: "foo--bar", want: "FOO__BAR"},
		{name: "trailing.", want: "TRAILING_"},
		{name: "a b", want: "A_B"},
		{name: "café", want: "CAF_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := envKey(tt.name); got != tt.want {
				t.Fatalf("envKey(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestEnvSetsNormalizedNames(t *testing.T) {
	// Keys are literals so this fails if lookup stops using the shell-safe name.
	tests := []struct {
		flagName string
		prefix   string
		key      string
		legacy   string
		value    string
	}{
		{flagName: "log-level", key: "LOG_LEVEL", legacy: "LOG-LEVEL", value: "debug"},
		{flagName: "db.host", key: "DB_HOST", legacy: "DB.HOST", value: "localhost"},
		{flagName: "server.bind-addr", key: "SERVER_BIND_ADDR", legacy: "SERVER.BIND-ADDR", value: ":8080"},
		{flagName: "log-level", prefix: "MYAPP_", key: "MYAPP_LOG_LEVEL", legacy: "MYAPP_LOG-LEVEL", value: "warn"},
		{flagName: "db.host", prefix: "MYAPP_", key: "MYAPP_DB_HOST", legacy: "MYAPP_DB.HOST", value: "db.internal"},
		{flagName: "server.bind-addr", prefix: "MYAPP_", key: "MYAPP_SERVER_BIND_ADDR", legacy: "MYAPP_SERVER.BIND-ADDR", value: "127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			setEnvPrefix(t, tt.prefix)
			restoreFlag(t, tt.flagName)
			// Only the normalized name is set. The legacy spelling is cleared
			// so a fallback would fail this case.
			mutateEnv(t, map[string]string{tt.key: tt.value}, []string{tt.legacy})

			applyEnvOverride(flag.Lookup(tt.flagName))
			if got := flag.Lookup(tt.flagName).Value.String(); got != tt.value {
				t.Fatalf("flag %q = %q, want %q (from %s)", tt.flagName, got, tt.value, tt.key)
			}
		})
	}
}

// TestEnvPlainNamesStayExact checks flags whose names need no replacement.
func TestEnvPlainNamesStayExact(t *testing.T) {
	t.Run("underscore lookalike is ignored", func(t *testing.T) {
		setEnvPrefix(t, "")
		restoreFlag(t, "flag1")
		mutateEnv(t, map[string]string{"FLAG_1": "nope"}, []string{"FLAG1"})
		if err := flag.Set("flag1", "keep"); err != nil {
			t.Fatal(err)
		}
		applyEnvOverride(flag.Lookup("flag1"))
		if got := flag.Lookup("flag1").Value.String(); got != "keep" {
			t.Fatalf("flag flag1 = %q, want keep", got)
		}
	})

	t.Run("prefix is prepended unchanged", func(t *testing.T) {
		setEnvPrefix(t, "JSONFLAG_")
		restoreFlag(t, "flag1")
		mutateEnv(t, map[string]string{"JSONFLAG_FLAG1": "FLAG1VALUE"}, []string{"FLAG1"})
		applyEnvOverride(flag.Lookup("flag1"))
		if got := flag.Lookup("flag1").Value.String(); got != "FLAG1VALUE" {
			t.Fatalf("flag flag1 = %q, want FLAG1VALUE", got)
		}
	})

	t.Run("prefix is not normalized", func(t *testing.T) {
		setEnvPrefix(t, "my-app_")
		restoreFlag(t, "db.host")
		mutateEnv(t, map[string]string{"my-app_DB_HOST": "prefixed"}, []string{"my-app_DB.HOST", "MY_APP_DB_HOST"})
		applyEnvOverride(flag.Lookup("db.host"))
		if got := flag.Lookup("db.host").Value.String(); got != "prefixed" {
			t.Fatalf("flag db.host = %q, want prefixed", got)
		}
	})
}

func TestEnvLeavesFlagWhenUnset(t *testing.T) {
	setEnvPrefix(t, "")
	restoreFlag(t, "log-level")
	mutateEnv(t, nil, []string{"LOG_LEVEL", "LOG-LEVEL"})
	if err := flag.Set("log-level", "keep-me"); err != nil {
		t.Fatal(err)
	}
	applyEnvOverride(flag.Lookup("log-level"))
	if got := flag.Lookup("log-level").Value.String(); got != "keep-me" {
		t.Fatalf("flag log-level = %q, want keep-me", got)
	}
}

// TestEnvTag checks an env tag.
//   - The tag value is the variable name, as written.
//   - EnvPrefix is still prepended.
//   - An env tag alone registers the flag.
func TestEnvTag(t *testing.T) {
	setEnvPrefix(t, "MYAPP_")
	// cfg: env name differs from the flag name, plus an env-only field.
	var cfg struct {
		TitleDelimiter string `flag:"title-delimiter" env:"TitleSep" default:"_"`
		OnlyEnv        string `env:"ONLY_ENV"`
	}
	registerTagFlags(&cfg)
	if flag.Lookup("onlyenv") == nil {
		t.Fatal("env tag alone did not register onlyenv")
	}
	restoreFlag(t, "title-delimiter")
	restoreFlag(t, "onlyenv")
	// Other spellings are set to different values so only the tag name matches.
	mutateEnv(t, map[string]string{
		"MYAPP_TitleSep":        "colon",
		"MYAPP_TITLESEP":        "normalized-tag",
		"MYAPP_TITLE_DELIMITER": "from-flag-name",
		"MYAPP_ONLY_ENV":        "yes",
		"MYAPP_ONLYENV":         "from-flag-name",
	}, nil)
	applyEnvOverride(flag.Lookup("title-delimiter"))
	applyEnvOverride(flag.Lookup("onlyenv"))
	if got := flag.Lookup("title-delimiter").Value.String(); got != "colon" {
		t.Fatalf("title-delimiter = %q, want colon", got)
	}
	if got := flag.Lookup("onlyenv").Value.String(); got != "yes" {
		t.Fatalf("onlyenv = %q, want yes", got)
	}
}

func setEnvPrefix(t *testing.T, prefix string) {
	t.Helper()
	prev := EnvPrefix
	EnvPrefix = prefix
	t.Cleanup(func() { EnvPrefix = prev })
}

func restoreFlag(t *testing.T, name string) {
	t.Helper()
	f := flag.Lookup(name)
	if f == nil {
		t.Fatalf("flag %q is not defined", name)
	}
	prev := f.Value.String()
	t.Cleanup(func() {
		if err := flag.Set(name, prev); err != nil {
			t.Errorf("restore flag %s: %v", name, err)
		}
	})
}

func mutateEnv(t *testing.T, set map[string]string, unset []string) {
	t.Helper()
	type snap struct {
		key string
		val string
		ok  bool
	}
	seen := map[string]bool{}
	var keys []string
	add := func(k string) {
		if seen[k] {
			return
		}
		seen[k] = true
		keys = append(keys, k)
	}
	for k := range set {
		add(k)
	}
	for _, k := range unset {
		add(k)
	}
	snaps := make([]snap, 0, len(keys))
	for _, k := range keys {
		val, ok := os.LookupEnv(k)
		snaps = append(snaps, snap{k, val, ok})
	}
	t.Cleanup(func() {
		for _, s := range snaps {
			var err error
			if s.ok {
				err = os.Setenv(s.key, s.val)
			} else {
				err = os.Unsetenv(s.key)
			}
			if err != nil {
				t.Errorf("restore env %s: %v", s.key, err)
			}
		}
	})
	for _, k := range unset {
		if err := os.Unsetenv(k); err != nil {
			t.Fatalf("unset %s: %v", k, err)
		}
	}
	for k, v := range set {
		if err := os.Setenv(k, v); err != nil {
			t.Fatalf("set %s: %v", k, err)
		}
	}
}
