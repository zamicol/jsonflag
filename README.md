# jsonflag

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflag)](https://goreportcard.com/report/github.com/zamicol/jsonflag)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflag?status.svg)](https://godoc.org/github.com/zamicol/jsonflag)


jsonflag is a drop in replacement and extension of Go's flag package that
seamlessly adds support for configs (JSON/JSON5), environmental variables,
struct tag defaults, and CLI options.  This allows configs to be more declarative and less imparitive than the Go flag library.

Values set by a higher precedence overwrite values set by a lower precedence,
**CLI > Env > JSON > JSON config tag default > Flag Default**, making testing
using Env variables, CLI, JSON tags, JSON configs, for CLI flags easy.

**Order of precedence**:
 1. Command line flags         (CLI example: `--flag1=Flag1Value`)
 2. Environmental variables    (Env example: FLAG2=Flag2value)
 3. JSON config values         (JSON example: `{"flag3": "Flag3Value"}`)
 4. JSON config tag default    (JSON Config struct example: `Flag6 int    `json:"9999"``)
 5. Default values set using go flag initialization (Go example: `flag.StringVar(&config.Flag4,
    "Flag4Name", "Flag4DefaultValue", "Flag4Description")`)

To overwrite a value, a CLI parameter may be set `go run main.go
--flag1=flag1CliValue`.  

Environmental variables may also be set via CLI, for example:
`FLAG1=Flag1EnvValue go run main.go`, but they are lower priority than CLI
flags.

jsonflag also supports:
- **Default values** in tags
- **Environmental variable expansion**  where if FLAG1=EnvFlagValue, and anywhere the value is $FLAG1, $FLAG1 will be expanded to EnvFlagValue. It does not matter where the variable is set, and `$` character is the signifier for a variable. `$


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


## Installation

Go get

```bash
go get github.com/zamicol/jsonflag
```
and import:

```go
import "github.com/zamicol/jsonflag"
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

	// Instead of `flag.parse`, use `jsonflag.Parse(&config)` which is the only line that must be different from using `flag` normally.  
	
	jsonflag.Parse(&config)
}
```

The default values will be overwritten by the values in the JSON config, which may further be overwritten by environmental variables and CLI parameters.

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
type TagConfig struct {
	UserName   string        `flag:"username" default:"defaultUserName" desc:"User name"`
	Count      int           `flag:"count" default:"10" desc:"Item count"`
	Enabled    bool          `flag:"enabled" default:"true" desc:"Feature toggle"`
	Timeout    time.Duration `flag:"timeout" default:"5s" desc:"Operation timeout"`
	Threshold  float64       `flag:"threshold" default:"0.75" desc:"Threshold value"`
	MaxRetries uint          `flag:"max-retries" default:"3" desc:"Max retry attempts"`
}











# Letter Casing For Flag Names
Flag naming conventions vary by input type. 

CLI flag names (not values) must start with lowercase letters (e.g., --flag1).

For environmental variables, the flag’s name is converted to all uppercase, meaning environmental flag names (not values) are case insensitive.

For JSON names, this package uses Go’s json package for decoding. The JSON decoder only has access to exported fields of structs and follows its own precedence for JSON decoding:
 1. Tags
 2. Exact case
 3. Case insensitive

The name set by the Go flag package may be upper or lower case, but uppercase is recommended to align with Go naming conventions for exported fields.

[See also the godocs for more complete documentation and a working example.](https://godoc.org/github.com/zamicol/jsonflag)
