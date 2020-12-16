package jsonflag

import (
	"os"
	"testing"
)

// See the package documentation on how to run a test.

func TestMain(m *testing.M) {
	// Run example.
	flags()
	// Must explicitly exit because of flag test
	os.Exit(m.Run())
}

// Golden test values.
var golden = Config{
	Flag1:   "cliFlag1",
	Flag2:   "jsonFlag2",
	Flag3:   "defaultFlag3",
	Flag4:   "jsonFlag4",
	Flag5:   5,
	Flag6:   6,
	Flag7:   "FLAG7VALUE",
	Flag7x1: "FLAG7VALUE",
	Flag8:   "Flag8Env",
}

// TestVerifyCorrectFlags
// test with `FLAG7=FLAG7VALUE Flag8=Flag8Env go test --flag1=cliFlag1 --config=test_config.json5`
func TestVerifyCorrectFlags(t *testing.T) {
	if tc.Flag1 != golden.Flag1 {
		mismatchError("Flag1", golden.Flag1, tc.Flag1, t)
	}
	if tc.Flag2 != golden.Flag2 {
		mismatchError("Flag2", golden.Flag2, tc.Flag2, t)
	}
	if tc.Flag3 != golden.Flag3 {
		mismatchError("Flag3", golden.Flag3, tc.Flag3, t)
	}
	if tc.Flag4 != golden.Flag4 {
		mismatchError("Flag4", golden.Flag4, tc.Flag4, t)
	}
	if tc.Flag5 != golden.Flag5 {
		mismatchErrori("Flag5", golden.Flag5, tc.Flag5, t)
	}
	if tc.Flag6 != golden.Flag6 {
		mismatchErrori("Flag6", golden.Flag6, tc.Flag6, t)
	}
	if tc.Flag7 != golden.Flag7 {
		mismatchError("Flag7", golden.Flag7, tc.Flag7, t)
	}
	if tc.Flag7x1 != golden.Flag7x1 {
		mismatchError("Flag7x1", golden.Flag7x1, tc.Flag7x1, t)
	}
	if tc.Flag8 != golden.Flag8 {
		mismatchError("Flag8", golden.Flag8, tc.Flag8, t)
	}

}

func mismatchError(what string, expected string, got string, t *testing.T) {
	t.Error(what+" set incorrectly. Expected", expected, "Got", got)
}

func mismatchErrori(what string, expected int, got int, t *testing.T) {
	t.Error(what+" set incorrectly. Expected", expected, "Got", got)
}
