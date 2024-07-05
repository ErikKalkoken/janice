package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/ErikKalkoken/jsonviewer/internal/ui"
)

type logLevelFlag struct {
	value slog.Level
}

func (l logLevelFlag) String() string {
	return l.value.String()
}

func (l *logLevelFlag) Set(value string) error {
	m := map[string]slog.Level{
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}
	v, ok := m[strings.ToUpper(value)]
	if !ok {
		return fmt.Errorf("unknown log level")
	}
	l.value = v
	return nil
}

func main() {
	levelFlag := logLevelFlag{value: slog.LevelWarn}
	flag.Var(&levelFlag, "loglevel", "set log level")
	flag.Parse()
	slog.SetLogLoggerLevel(levelFlag.value)
	u, err := ui.NewUI()
	if err != nil {
		log.Fatalf("Failed to initialize application: %s", err)
	}
	u.ShowAndRun()
}
