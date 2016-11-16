package jsonflags

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flags()
	os.Exit(m.Run())
}

// Config's values must be exported
type Config struct {
	Flag1 string // populate everywhere (flag default, json, cli flag)
	Flag2 string // empty everywhere (no default, json, or cli flag)
	Flag3 string // json only
	Flag4 string // flag default only
	Flag5 string // json with flag default (no cli flag)
	// Flag6 json only string
	Flag7 int // json only int
	Flag8 int // json with flag default (no cli flag) int
}

// Golden settings
var golden = Config{
	Flag1: "paramFlag1",
	Flag3: "jsonFlag3",
	Flag4: "defaultFlag4",
	Flag5: "jsonFlag5",
}

var tc Config

// flags holds all flag definitions
func flags() {
	flag.StringVar(&tc.Flag1, "flag1", "defaultFlag1", "flag1Desc")
	flag.StringVar(&tc.Flag4, "flag4", "defaultFlag4", "flag4Desc")
	flag.StringVar(&tc.Flag5, "flag5", "defaultFlag5", "flag5Desc")
	flag.IntVar(&tc.Flag7, "flag7", 1, "flag7Desc")
	Parse(&tc)
}

// TestVerifyCorrectFlags
// test with `go test --flag1=paramFlag1`
func TestVerifyCorrectFlags(t *testing.T) {
	if tc.Flag1 != golden.Flag1 {
		throwValuesMismatchError("Flag1", golden.Flag1, tc.Flag1, t)
	}
	if tc.Flag3 != golden.Flag3 {
		throwValuesMismatchError("Flag3", golden.Flag3, tc.Flag3, t)
	}
	if tc.Flag4 != golden.Flag4 {
		throwValuesMismatchError("Flag4", golden.Flag4, tc.Flag4, t)
	}
}

func throwValuesMismatchError(what string, expected string, got string, t *testing.T) {
	t.Error(what+" set incorrectly. Expected", expected, "Got", got)
}

// ExampleConfig prints out values
func ExampleConfig() {
	fmt.Println(tc)
	// Output: {paramFlag1  jsonFlag3 defaultFlag4 defaultFlag5 7 0}
}
