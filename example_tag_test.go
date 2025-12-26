package jsonflag

import "time"

var tc TagConfig

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

// // Golden test values for tag-based test.
// var tagGolden = TagConfig{
// 	UserName:   "cliName",        // CLI override
// 	Count:      20,               // JSON value
// 	Enabled:    false,            // Env override
// 	Timeout:    10 * time.Second, // JSON value
// 	Threshold:  0.75,             // Default value
// 	MaxRetries: 5,                // Env override
// }
