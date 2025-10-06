package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type LogFormat string

const (
	LogFormatText    LogFormat = "text"
	LogFormatJSON    LogFormat = "json"
	LogFormatConsole LogFormat = "console"
)

func (f *LogFormat) UnmarshalText(text []byte) error {
	switch format := string(text); format {
	case string(LogFormatText), string(LogFormatJSON), string(LogFormatConsole):
		*f = LogFormat(format)
		return nil
	default:
		return fmt.Errorf("invalid log format: %s", format)
	}
}

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
	return nil
}

type Config struct {
	Logging                LoggingConfig `mapstructure:"logging"`
	CheckIntervalInSeconds int           `default:"86400" split_words:"true"`
	Domains                []string      `mapstructure:"domains"`
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
	_ = v.BindEnv("checkIntervalInSeconds")

	// Default values
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("checkIntervalInSeconds", 86400)

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
	default:
		slog.SetLogLoggerLevel(config.Logging.SlogLevel)
	}

	for _, domain := range config.Domains {
		_, err := url.Parse(domain)
		if err != nil {
			return config, fmt.Errorf("invalid domain format: %s", domain)
		}
	}

	return config, nil
}
