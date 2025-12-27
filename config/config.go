package config

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/charmbracelet/log"
)

type DB struct {
	Url string `env:"URL" envDefault:"postgres://postgres:postgres@localhost/postgres?sslmode=disable"`
}

type Log struct {
	Level slog.Level `env:"LEVEL" envDefault:"info"`
	Type  string     `env:"TYPE" envDefault:"pretty"`
}

func (l Log) New() *slog.Logger {
	var h slog.Handler
	switch l.Type {
	case "pretty":
		h = log.NewWithOptions(os.Stderr, log.Options{
			Level: log.Level(l.Level),
		})
	case "json":
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: l.Level})
	default:
		h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l.Level})
	}

	return slog.New(h)
}

type Telemetry struct {
	Enabled        bool   `env:"ENABLED" envDefault:"false"`
	OTELEndpoint   string `env:"OTEL_ENDPOINT" envDefault:"localhost:4317"`
	ServiceName    string `env:"SERVICE_NAME" envDefault:"goovern"`
	ServiceVersion string `env:"SERVICE_VERSION" envDefault:"0.1.0"`
}

type GoovernD struct {
	DB        DB        `envPrefix:"DB_"`
	Log       Log       `envPrefix:"LOG_"`
	Telemetry Telemetry `envPrefix:"TELEMETRY_"`
}

func Load() (GoovernD, error) {
	return env.ParseAsWithOptions[GoovernD](env.Options{Prefix: "GOO_"})
}
