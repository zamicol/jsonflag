# jsonflag

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflag)](https://goreportcard.com/report/github.com/zamicol/jsonflag)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflag?status.svg)](https://godoc.org/github.com/zamicol/jsonflag)

jsonflag is a drop in replacement and extension of Go's flag package that
seamlessly adds support for configs (JSON/JSON5), environmental variables,
struct tag defaults, and CLI options.  Struct tag defaults and JSON configs
allows configs to be declarative instead of imperative like the Go flag library.

Values set by a higher precedence overwrite values set by a lower precedence,
**CLI > Env > JSON > JSON config tag default > Flag Default**, making testing
using Env variables, CLI, JSON tags, JSON configs, for CLI flags easy.

**Order of precedence**:
 1. Command line flags         (CLI example: `--flag1=Flag1Value`)
 2. Environmental variables    (Env example: FLAG2=Flag2value)
 3. JSON config values         (JSON example: `{"flag3": "Flag3Value"}`)
 4. Struct tag default         (JSON struct example: config{Flag1 int    `json:"9999"`)
 5. Go flag default            (Go example: `flag.StringVar(&config.Flag5, "Flag5Name", "Flag5DefaultValue", "Flag5Description")`)

To overwrite a value, a CLI parameter may be set `go run main.go
--flag1=flag1CliValue`.  Environmental variables may also be set via CLI, for
example: `FLAG1=Flag1EnvValue go run main.go`, but they are lower priority than
CLI flags.

jsonflag also supports:
- **Default values** in tags
- **Writing a default config** with `WriteDefault` or `--write-config`

## Installation

Go get

```bash
go get github.com/zamicol/jsonflag
```
and import:

```go
import "github.com/zamicol/jsonflag"
```


## Config Path

If not specified, the default path is `config.json5` in the current working
directory. The config path can be specified in two ways:

1. Via command line argument (which takes precedence):
```bash
go run main.go --config=test_config.json
```

2. Programmatically in your Go code, add the line `jsonflag.Path = "test_config.json"`:
```go
type Config struct {
	Flag1 string
}
func main(){
	var config Config
	flag.StringVar(&config.Flag1, "Flag1Name", "Flag1DefaultValue", "Flag1Description")
	jsonflag.Path = "test_config.json"
	jsonflag.Parse(&config)
}
```

(Related reminder: `os.UserConfigDir()` may be idiomatic for some applications.)

## Write a default config

`WriteDefault` writes an indented JSON5 config for a struct:

```go
err := jsonflag.WriteDefault("config.json5", &config)
```

Each exported field is one key. The key is the `json` name, or the Go field
name when that tag is empty. A `default` tag supplies the value; otherwise the
field's current value is written, including values set with `flag` or a struct
literal. `time.Duration` is written as integer nanoseconds. A `desc` tag is
written as a `//` comment above that field. If the file already
exists, `WriteDefault` returns an error that wraps `os.ErrExist`.

`--write-config` writes that file to the config path and exits:

```bash
go run main.go --write-config --config=config.json5
```

## Quick Example
Example `config.json5` file:
```json5
{
"flag1": "jsonFlag1",
"flag2": "jsonFlag2",
"flag3": 3,  // Comments and trailing commas are supported in JSON5 configs and encouraged for readability. 
}
```

Example Go setup:
```go
// Config struct name (tag should) match the json key.  
type Config struct {
	Flag1 string
	Flag2 string
	Flag3 int
}

func main(){
	var config Config
	// `flag` is still from the standard library.
	flag.StringVar(&config.Flag1, "Flag1Name", "Flag1DefaultValue", "Flag1Description")
	flag.StringVar(&config.Flag2, "Flag2Name", "Flag2DefaultValue", "Flag2Description")
	flag.IntVar(&config.Flag3, "Flag3Name", 1, "Flag3Description")

	// Instead of `flag.parse`, use `jsonflag.Parse(&config)` which is the only 
  // line that must be different from using `flag` normally.
	jsonflag.Parse(&config)
}
```

The default values will be overwritten by the values in the JSON config, which
may further be overwritten by environmental variables and CLI parameters.

Example overwriting values using an environmental variable and command line flag:

```
FLAG2=Flag2EnvValue go run main.go --flag1=Flag1CLIValue

```

Which will result in the final values being used:

```
Flag1 = Flag1CLIValue   // From command line flag. 
Flag2 = Flag2EnvValue   // From environmental variable. 
Flag3 = 3               // From JSON file. 
```

## Tag
// TagConfig is for testing "tag design pattern" flags
// Config's values must be exported for package `flag` to be able to set values.
```GO
type TagConfig struct {
	UserName       string        `flag:"username" default:"defaultUserName" desc:"User name"`
	Count          int           `flag:"count" default:"10" desc:"Item count"`
	Enabled        bool          `flag:"enabled" default:"true" desc:"Feature toggle"`
	Timeout        time.Duration `flag:"timeout" default:"5s" desc:"Operation timeout"`
	Threshold      float64       `flag:"threshold" default:"0.75" desc:"Threshold value"`
	MaxRetries     uint          `flag:"max-retries" default:"3" desc:"Max retry attempts"`
	TitleDelimiter string        `default:"_" json:"Title_Delimiter" flag:"title-delimiter" env:"TITLE_DELIMITER" desc:"Delimiter for titles in filenames"`
}
```

An `env` tag names that field's environment variable. The value is used as
written, and `EnvPrefix` is still prepended (`MYAPP_` + `TITLE_DELIMITER` is
`MYAPP_TITLE_DELIMITER`). An `env` tag alone registers the flag.

# Letter Casing For Flag Names
Flag naming conventions vary by input type. 

CLI flag names (not values) must start with lowercase letters (e.g., --flag1).

Environmental variables are converted to all uppercase,
meaning environmental flag names (not values) are case insensitive. Characters
other than ASCII letters, digits, and underscores are replaced with underscores,
which makes dashed and dotted flag names settable in a shell. For example,
`log-level` or `log.level` is read from `LOG_LEVEL`. `EnvPrefix`, when set, is
prepended unchanged (`MYAPP_` + `log-level` is `MYAPP_LOG_LEVEL`).

**Go JSON naming** For JSON names, this package uses the json5 package for
decoding. The JSON decoder only has access to exported fields of structs and
follows its own precedence for JSON decoding:
 1. Tags
 2. Exact case
 3. Case insensitive

The name set by the Go flag package may be upper or lower case, but uppercase is
recommended to align with Go naming conventions for exported fields.

[See also the godocs for more complete documentation and a working
example.](https://godoc.org/github.com/zamicol/jsonflag)