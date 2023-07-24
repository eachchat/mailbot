package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// LogConf the logger configuration
type LogConf struct {
	Level string `yaml:"level"`
}

func (l *LogConf) New() *zerolog.Logger {
	switch l.Level {
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "WARN":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
	multi := zerolog.MultiLevelWriter(os.Stdout)
	logg := zerolog.New(multi).With().Timestamp().Logger()

	return &logg
}
