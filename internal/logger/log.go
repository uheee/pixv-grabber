package logger

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

const (
	colorBlack = iota + 30
	colorRed
	colorGreen
	colorYellow
	colorBlue
	colorMagenta
	colorCyan
	colorWhite
	colorBold     = 1
	colorDarkGray = 90
)

func InitLog() {
	logLevel := viper.GetString("log.level")
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "[" + time.RFC3339Nano + "]",
		FormatMessage: func(i interface{}) string {
			return colorize(i, colorWhite, false)
		},
		FormatFieldName: func(i interface{}) string {
			return colorize(fmt.Sprintf("%s: ", i), colorYellow, false)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s%s%s",
				colorize("[", colorYellow, false),
				colorize(i, colorCyan, false),
				colorize("]", colorYellow, false),
			)
		},
		FormatErrFieldName: func(i interface{}) string {
			return colorize(fmt.Sprintf("%s: ", i), colorRed, false)
		},
		FormatErrFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s%s%s",
				colorize("[", colorRed, false),
				colorize(i, colorMagenta, false),
				colorize("]", colorRed, false),
			)
		},
	}
	log.Logger = log.Output(writer).Level(getLogLevel(logLevel))
}

var levelDefinitions = []zerolog.Level{
	zerolog.DebugLevel,
	zerolog.InfoLevel,
	zerolog.WarnLevel,
	zerolog.ErrorLevel,
	zerolog.FatalLevel,
	zerolog.PanicLevel,
	zerolog.NoLevel,
	zerolog.Disabled,
}

func getLogLevel(level string) zerolog.Level {
	for _, l := range levelDefinitions {
		if strings.ToLower(level) == l.String() {
			return l
		}
	}
	panic(fmt.Sprintf("error logger level '%s'", level))
}

func colorize(s interface{}, c int, disabled bool) string {
	if disabled {
		return fmt.Sprintf("%s", s)
	}
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", c, s)
}
