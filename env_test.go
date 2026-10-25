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

			env(flag.Lookup(tt.flagName))
			if got := flag.Lookup(tt.flagName).Value.String(); got != tt.value {
				t.Fatalf("flag %q = %q, want %q (from %s)", tt.flagName, got, tt.value, tt.key)
			}
		})
	}
}

func TestEnvLegacyFallback(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		prefix   string
		set      map[string]string
		unset    []string
		want     string
	}{
		{
			name:     "dashed",
			flagName: "log-level",
			set:      map[string]string{"LOG-LEVEL": "from-legacy"},
			unset:    []string{"LOG_LEVEL"},
			want:     "from-legacy",
		},
		{
			name:     "dotted",
			flagName: "db.host",
			set:      map[string]string{"DB.HOST": "legacy-host"},
			unset:    []string{"DB_HOST"},
			want:     "legacy-host",
		},
		{
			name:     "dashed with prefix",
			flagName: "log-level",
			prefix:   "MYAPP_",
			set:      map[string]string{"MYAPP_LOG-LEVEL": "prefixed-legacy"},
			unset:    []string{"MYAPP_LOG_LEVEL"},
			want:     "prefixed-legacy",
		},
		{
			name:     "dotted with prefix",
			flagName: "db.host",
			prefix:   "MYAPP_",
			set:      map[string]string{"MYAPP_DB.HOST": "prefixed-dot-legacy"},
			unset:    []string{"MYAPP_DB_HOST"},
			want:     "prefixed-dot-legacy",
		},
		{
			name:     "normalized wins over legacy",
			flagName: "log-level",
			set:      map[string]string{"LOG_LEVEL": "from-normalized", "LOG-LEVEL": "from-legacy"},
			want:     "from-normalized",
		},
		{
			name:     "empty normalized falls back to legacy",
			flagName: "log-level",
			set:      map[string]string{"LOG_LEVEL": "", "LOG-LEVEL": "from-legacy"},
			want:     "from-legacy",
		},
		{
			name:     "mixed separators with prefix",
			flagName: "server.bind-addr",
			prefix:   "MYAPP_",
			set:      map[string]string{"MYAPP_SERVER.BIND-ADDR": "legacy-bind"},
			unset:    []string{"MYAPP_SERVER_BIND_ADDR"},
			want:     "legacy-bind",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnvPrefix(t, tt.prefix)
			restoreFlag(t, tt.flagName)
			mutateEnv(t, tt.set, tt.unset)

			if got := envValue(tt.flagName); got != tt.want {
				t.Fatalf("envValue(%q) = %q, want %q", tt.flagName, got, tt.want)
			}
			env(flag.Lookup(tt.flagName))
			if got := flag.Lookup(tt.flagName).Value.String(); got != tt.want {
				t.Fatalf("flag %q = %q, want %q", tt.flagName, got, tt.want)
			}
		})
	}
}

func TestEnvPlainNamesStayExact(t *testing.T) {
	t.Run("underscore lookalike is ignored", func(t *testing.T) {
		setEnvPrefix(t, "")
		mutateEnv(t, map[string]string{"FLAG_1": "nope"}, []string{"FLAG1"})
		if got := envValue("flag1"); got != "" {
			t.Fatalf("envValue(flag1) = %q, want empty", got)
		}
	})

	t.Run("prefix is prepended unchanged", func(t *testing.T) {
		setEnvPrefix(t, "JSONFLAG_")
		mutateEnv(t, map[string]string{"JSONFLAG_FLAG10": "FLAG10VALUE"}, nil)
		if got := envValue("flag10"); got != "FLAG10VALUE" {
			t.Fatalf("envValue(flag10) = %q, want FLAG10VALUE", got)
		}
	})

	t.Run("prefix is not normalized", func(t *testing.T) {
		setEnvPrefix(t, "my-app_")
		restoreFlag(t, "db.host")
		mutateEnv(t, map[string]string{"my-app_DB_HOST": "prefixed"}, []string{"my-app_DB.HOST", "MY_APP_DB_HOST"})
		env(flag.Lookup("db.host"))
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
	env(flag.Lookup("log-level"))
	if got := flag.Lookup("log-level").Value.String(); got != "keep-me" {
		t.Fatalf("flag log-level = %q, want keep-me", got)
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
