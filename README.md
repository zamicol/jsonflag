# jsonflag

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflag)](https://goreportcard.com/report/github.com/zamicol/jsonflag)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflag?status.svg)](https://godoc.org/github.com/zamicol/jsonflag)


Use JSON configs in conjunction with Go's flag package.

[See the godocs for documentation and a working example.](https://godoc.org/github.com/zamicol/jsonflag)


Example:
```go
type ExampleConfig struct {
	Flag1 string
	Flag2 string
	Flag3 int
}

func main(){
  var config ExampleConfig
  flag.StringVar(&config.Flag1, "flag1", "defaultFlag1", "flag1Desc")
  flag.StringVar(&config.Flag2, "flag2", "defaultFlag2", "flag2Desc")
  flag.IntVar(&config.Flag3, "flag3", 1, "flag3Desc")

  jsonflag.Parse(&config)
}
```

Example `config.json` file:
```json
{
    "flag1": "jsonFlag1",
    "flag2": "jsonFlag2",
    "flag3": 3,
}
```
