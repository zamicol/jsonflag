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
	Flag1: "cliFlag1",
	Flag2: "jsonFlag2",
	Flag3: "defaultFlag3",
	Flag4: "jsonFlag4",
	Flag5: 5,
	Flag6: 6,
}

// TestVerifyCorrectFlags
// test with `go test --flag1=paramFlag1 --config=test_config.json`
func TestVerifyCorrectFlags(t *testing.T) {
	if tc.Flag1 != golden.Flag1 {
		throwValueMismatchError("Flag1", golden.Flag1, tc.Flag1, t)
	}
	if tc.Flag2 != golden.Flag2 {
		throwValueMismatchError("Flag2", golden.Flag2, tc.Flag2, t)
	}
	if tc.Flag3 != golden.Flag3 {
		throwValueMismatchError("Flag3", golden.Flag3, tc.Flag3, t)
	}
	if tc.Flag4 != golden.Flag4 {
		throwValueMismatchError("Flag4", golden.Flag4, tc.Flag4, t)
	}
	if tc.Flag5 != golden.Flag5 {
		t.Error("Expected", golden.Flag5, "Got", tc.Flag5)
	}
	if tc.Flag6 != golden.Flag6 {
		t.Error("Expected", golden.Flag6, "Got", tc.Flag6)
	}
}

func throwValueMismatchError(what string, expected string, got string, t *testing.T) {
	t.Error(what+" set incorrectly. Expected", expected, "Got", got)
}
