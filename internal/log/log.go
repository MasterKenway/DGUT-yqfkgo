package log

import (
	"fmt"
	"os"
	"strings"
	"time"
)

import (
	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger() {
	wr := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC822Z}
	wr.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	Logger = zerolog.New(wr).With().Timestamp().Logger()
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Panic() *zerolog.Event {
	return Logger.Panic()
}
