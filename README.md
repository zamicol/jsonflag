# jsonflag

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflag)](https://goreportcard.com/report/github.com/zamicol/jsonflag)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflag?status.svg)](https://godoc.org/github.com/zamicol/jsonflag)


Use JSON configs and environmental variables in conjunction with Go's flag package.

[See the godocs for documentation and a working example.](https://godoc.org/github.com/zamicol/jsonflag)


## Example
Example `config.json` file:
```json
{
"flag1": "jsonFlag1",
"flag2": "jsonFlag2",
"flag3": 3,  // Trailing commas and comments in json config are recommended.  
}
```

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
	flag.StringVar(&config.Flag1, "flag1name", "defaultFlag1", "flag1Desc")
	flag.StringVar(&config.Flag2, "flag2name", "defaultFlag2", "flag2Desc")
	flag.IntVar(&config.Flag3, "flag3name", 1, "flag3Desc")

	// Instead of `flag.parse`, use `jsonflag.Parse(&config)`
	// This is the only line that must be different from using `flag` normally.  
	jsonflag.Parse(&config)
}
```


Environmental variables and cli flags give further flexibility.  See the documentation for order of precedence.  

	`FLAG1=flag1EnvValue go run --flag2=flag2CliValue`



	# TODO
Add support for go tags, possibly for json and description
