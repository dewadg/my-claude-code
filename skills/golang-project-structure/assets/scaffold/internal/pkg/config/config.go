package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/app/internal/pkg/logging"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const envDevelopment = "development"

// Config is the validated, typed snapshot every consumer receives via fx.
type Config struct {
	Env       string
	Port      int
	LogLevel  slog.Level
	LogFormat string

	DatabaseURL       string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBPingTimeout     time.Duration

	ShutdownTimeout time.Duration

	// Per-signal switches are already ANDed with OTelEnabled, so consumers
	// read exactly one boolean.
	OTelEnabled        bool
	OTelEndpoint       string
	OTelLogsEnabled    bool
	OTelTracesEnabled  bool
	OTelMetricsEnabled bool
}

// IsDevelopment reports whether ENV=development.
func (c Config) IsDevelopment() bool {
	return c.Env == envDevelopment
}

// raw is what caarlos0/env populates. Every field is a string so Parse owns
// conversion and can choose per field between warn-and-default and fail.
type raw struct {
	Env       string `env:"ENV"        envDefault:""`
	Port      string `env:"PORT"       envDefault:"3000"`
	LogLevel  string `env:"LOG_LEVEL"  envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"text"`

	DatabaseURL       string `env:"DATABASE_URL"         envDefault:""`
	DBMaxOpenConns    string `env:"DB_MAX_OPEN_CONNS"    envDefault:""`
	DBMaxIdleConns    string `env:"DB_MAX_IDLE_CONNS"    envDefault:""`
	DBConnMaxLifetime string `env:"DB_CONN_MAX_LIFETIME" envDefault:""`
	DBPingTimeout     string `env:"DB_PING_TIMEOUT"      envDefault:""`

	ShutdownTimeout string `env:"SHUTDOWN_TIMEOUT" envDefault:""`

	OTelEnabled        string `env:"OTEL_ENABLED"                envDefault:""`
	OTelEndpoint       string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	OTelLogsEnabled    string `env:"OTEL_LOGS_ENABLED"           envDefault:""`
	OTelTracesEnabled  string `env:"OTEL_TRACES_ENABLED"         envDefault:""`
	OTelMetricsEnabled string `env:"OTEL_METRICS_ENABLED"        envDefault:""`
}

// LoadDotenv loads ./.env without overriding variables already set in the
// process. A missing file is normal; a malformed one is logged and skipped.
func LoadDotenv() error {
	err := godotenv.Load()
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	slog.With(logging.KeySource, "config.LoadDotenv").Warn("dotenv load failed", logging.KeyError, err)
	return nil
}

// Parse reads the environment into a Config. Values that can be safely
// defaulted warn and fall back; values that cannot return an error.
func Parse() (Config, error) {
	var r raw
	if err := env.Parse(&r); err != nil {
		return Config{}, err
	}

	port, err := strconv.Atoi(strings.TrimSpace(r.Port))
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("invalid PORT %q: must be an integer in 1..65535", r.Port)
	}

	otelEnabled := parseBool(r.OTelEnabled, false)

	return Config{
		Env:       strings.ToLower(strings.TrimSpace(r.Env)),
		Port:      port,
		LogLevel:  parseLogLevel(r.LogLevel),
		LogFormat: parseLogFormat(r.LogFormat),

		DatabaseURL:       strings.TrimSpace(r.DatabaseURL),
		DBMaxOpenConns:    parseInt(r.DBMaxOpenConns, 25, "DB_MAX_OPEN_CONNS"),
		DBMaxIdleConns:    parseInt(r.DBMaxIdleConns, 5, "DB_MAX_IDLE_CONNS"),
		DBConnMaxLifetime: parseDuration(r.DBConnMaxLifetime, 30*time.Minute, "DB_CONN_MAX_LIFETIME"),
		DBPingTimeout:     parseDuration(r.DBPingTimeout, 5*time.Second, "DB_PING_TIMEOUT"),

		ShutdownTimeout: parseDuration(r.ShutdownTimeout, 15*time.Second, "SHUTDOWN_TIMEOUT"),

		OTelEnabled:        otelEnabled,
		OTelEndpoint:       strings.TrimSpace(r.OTelEndpoint),
		OTelLogsEnabled:    otelEnabled && parseBool(r.OTelLogsEnabled, true),
		OTelTracesEnabled:  otelEnabled && parseBool(r.OTelTracesEnabled, true),
		OTelMetricsEnabled: otelEnabled && parseBool(r.OTelMetricsEnabled, true),
	}, nil
}

// AssertRequiredEnv rejects configurations that must not boot.
func AssertRequiredEnv(cfg Config) error {
	if cfg.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.OTelEnabled && cfg.OTelEndpoint == "" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required when OTEL_ENABLED=true")
	}
	return nil
}

func parseBool(raw string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return def
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseInt(raw string, def int, name string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		slog.With(logging.KeySource, "config.parseInt").Warn("invalid integer env, using default", "env", name, "raw", raw, "default", def)
		return def
	}
	return n
}

func parseDuration(raw string, def time.Duration, name string) time.Duration {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		slog.With(logging.KeySource, "config.parseDuration").Warn("invalid duration env, using default", "env", name, "raw", raw, "default", def.String())
		return def
	}
	return d
}

func parseLogLevel(raw string) slog.Level {
	var l slog.Level
	if err := l.UnmarshalText([]byte(strings.TrimSpace(raw))); err != nil {
		slog.With(logging.KeySource, "config.parseLogLevel").Warn("invalid LOG_LEVEL, using info", "raw", raw)
		return slog.LevelInfo
	}
	return l
}

func parseLogFormat(raw string) string {
	switch f := strings.ToLower(strings.TrimSpace(raw)); f {
	case "text", "json":
		return f
	default:
		slog.With(logging.KeySource, "config.parseLogFormat").Warn("invalid LOG_FORMAT, using text", "raw", raw)
		return "text"
	}
}
