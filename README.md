# jsonflags

[![Go Report Card](https://goreportcard.com/badge/github.com/zamicol/jsonflags)](https://goreportcard.com/report/github.com/zamicol/jsonflags)
[![GoDoc](https://godoc.org/github.com/zamicol/jsonflags?status.svg)](https://godoc.org/github.com/zamicol/jsonflags)


Use JSON configs in conjunction with Go's flag package.


Example Go code:
```
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
  jsonflags.Parse(&config)
}
```

Example json file:
```
{
    "flag1": "jsonFlag1",
    "flag2": "jsonFlag2",
    "flag3": 3,
}
```
