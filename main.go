package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2/app"
	"github.com/ErikKalkoken/janice/internal/ui"
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
	versionFlag := flag.Bool("v", false, "show current version")
	flag.Usage = myUsage
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Llongfile)
	slog.SetLogLoggerLevel(levelFlag.value)
	a := app.NewWithID("io.github.erikkalkoken.janice")
	if *versionFlag {
		fmt.Printf("Current version is: %s", a.Metadata().Version)
		return
	}
	u, err := ui.NewUI(a)
	if err != nil {
		log.Fatalf("Failed to initialize application: %s", err)
	}
	source := flag.Arg(0)
	u.ShowAndRun(source)
}

// myUsage writes a custom usage message to configured output stream.
func myUsage() {
	s := "Usage: janice [options] [<inputfile>]\n\n" +
		"A desktop app for viewing large JSON files.\n" +
		"For more information please see: https://github.com/ErikKalkoken/janice\n\n" +
		"Options:\n"
	fmt.Fprint(flag.CommandLine.Output(), s)
	flag.PrintDefaults()
}
