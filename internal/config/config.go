package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"golang.org/x/net/idna"
)

type LogFormat string

const (
	LogFormatText    LogFormat = "text"
	LogFormatJSON    LogFormat = "json"
	LogFormatConsole LogFormat = "console"
)

type LoggingConfig struct {
	Level     string     `mapstructure:"level"`
	Format    LogFormat  `mapstructure:"format"`
	SlogLevel slog.Level `mapstructure:"-"`
}

func (c *LoggingConfig) Parse() error {
	switch strings.ToLower(c.Level) {
	case "debug":
		c.SlogLevel = slog.LevelDebug
	case "info":
		c.SlogLevel = slog.LevelInfo
	case "warn":
		c.SlogLevel = slog.LevelWarn
	case "error":
		c.SlogLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	switch c.Format {
	case LogFormatText, LogFormatJSON, LogFormatConsole:
	default:
		return fmt.Errorf("invalid log format: %s", c.Format)
	}

	return nil
}

type Config struct {
	Logging                LoggingConfig `mapstructure:"logging"`
	CheckIntervalInSeconds int           `mapstructure:"check_interval_in_seconds"`
	Domains                []string      `mapstructure:"domains"`
}

func validateDomain(domain string) error {
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	if _, err := idna.Lookup.ToASCII(domain); err != nil {
		return fmt.Errorf("invalid domain format: %s: %w", domain, err)
	}
	return nil
}

func Get() (Config, error) {
	var config Config

	godotenv.Load()

	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	_ = v.BindEnv("logging.level")
	_ = v.BindEnv("logging.format")
	_ = v.BindEnv("check_interval_in_seconds")

	// Default values
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("check_interval_in_seconds", 86400)

	// Optionally load config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return config, fmt.Errorf("error reading config file: %w", err)
		}
	}

	if err := v.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	if err := config.Logging.Parse(); err != nil {
		return config, fmt.Errorf("invalid logging configuration: %w", err)
	}

	// Set logger
	switch config.Logging.Format {
	case LogFormatText:
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: config.Logging.SlogLevel,
		})))
	case LogFormatJSON:
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: config.Logging.SlogLevel,
		})))
	case LogFormatConsole:
		// Only works as long as slog.SetDefault has not been called before
		slog.SetLogLoggerLevel(config.Logging.SlogLevel)
	}

	if config.CheckIntervalInSeconds <= 0 {
		return config, fmt.Errorf("invalid check interval: %d (must be greater than 0)", config.CheckIntervalInSeconds)
	}

	for _, domain := range config.Domains {
		if err := validateDomain(domain); err != nil {
			return config, err
		}
	}

	return config, nil
}
