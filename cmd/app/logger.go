package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func logger(debug bool) error {
	logLevel := zap.InfoLevel
	if debug {
		logLevel = zap.DebugLevel
	}

	cfg := zap.Config{
		Encoding:    "console",
		Level:       zap.NewAtomicLevelAt(logLevel),
		OutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,
			TimeKey:     "time",
			EncodeTime:  zapcore.ISO8601TimeEncoder,
		},
	}

	log, err := cfg.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(log)

	return nil
}
