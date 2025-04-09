# jsonflag

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflag)](https://goreportcard.com/report/github.com/zamicol/jsonflag)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflag?status.svg)](https://godoc.org/github.com/zamicol/jsonflag)


jsonflag is an almost drop in replacement for Go's flag package that supports
configs (JSON/JSON5), environmental variables, and CLI options.

Values set by a higher precedence overwrite values set by a lower precedence.
This makes testing using CLI or environmental variables easy. 

Order of precedence for set config values :

 1. Command line flags         (cli Example: `--flag1=Flag1Value`)
 2. Environmental Variables    (env Example: FLAG2=Flag2value)
 3. JSON config values         (json Example: `{"flag3": "Flag3Value"}`)
 4. Default values set on flag (go Example: `flag.StringVar(&config.Flag4,
    "Flag4Name", "Flag4DefaultValue", "Flag4Description")`)

To overwrite a value, say on the config, a CLI parameter may be
set `FLAG1=Flag1EnvValue go run --flag2=flag2CliValue`

## Config Path

If not specified, the default path is `config.json5` in the current working
directory. The config path can be specified in two ways:

1. Via command line argument (which takes precedence):
```bash
go run main.go --config=test_config.json
```

2. Programmatically in your Go code:
```go
jsonflag.Path = "test_config.json"
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
"flag3": 3,  // Comments and trailing commas in JSON5 configs are recommended.  
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

The default values will be overwritten by the values in the JSON config, which may further be overwritten by environmental variables and CLI parameters. For example

Example overwriting values using environmental variable and command line options:
```
FLAG2=Flag2EnvValue go run main.go --flag1=Flag1CLIValue

```

Which will result in the final values being used:

```
Flag1 = Flag1CLIValue  // From command line flag. 
Flag2 = Flag2EnvValue // From environmental variable. 
Flag3 = 3 // From JSON file. 
```


[See also the godocs for more complete documentation and a working example.](https://godoc.org/github.com/zamicol/jsonflag)
