// Package config parses environment-driven configuration. A .env file is
// loaded when present, caarlos0/env maps variables into a string-only raw
// struct, Parse converts it into the typed Config, and AssertRequiredEnv fails
// fast on values that cannot be defaulted.
package config
