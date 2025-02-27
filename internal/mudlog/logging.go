package mudlog

import (
	"log/slog"
)

var (
	logLevel           = new(slog.LevelVar)
	textLoggerFailover *slog.TextHandler
)

type TeeLogger interface {
	Println(level string, v ...any)
}

func SetLogLevel(lvl string) {

	if len(lvl) > 0 {
		if lvl[0:1] == `M` {
			logLevel.Set(slog.LevelInfo)
			return
		} else if lvl[0:1] == `L` {
			logLevel.Set(slog.LevelWarn)
			return
		}
	}

	logLevel.Set(slog.LevelDebug)

}
