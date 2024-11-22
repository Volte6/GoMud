package util

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"
)

type ColorHandler struct {
	slog.Handler
	l                    *log.Logger
	minimumMessageLength int
}

func (h *ColorHandler) Handle(ctx context.Context, r slog.Record) error {

	var level string

	msgLen := len(r.Message)
	attrCt := r.NumAttrs()
	if msgLen == 0 && attrCt == 0 {
		h.l.Println(``)
		return nil
	}

	switch r.Level {
	case slog.LevelDebug:
		level = fmt.Sprintf("\033[95m%s:\033[0m", r.Level.String()) // magenta
	case slog.LevelInfo:
		level = fmt.Sprintf("\033[32m%s: \033[0m", r.Level.String()) // green
	case slog.LevelWarn:
		level = fmt.Sprintf("\033[33m%s: \033[0m", r.Level.String()) // yellow
	case slog.LevelError:
		level = fmt.Sprintf("\033[31m%s:\033[0m", r.Level.String()) // red
	}

	finalOut := strings.Builder{}

	if attrCt > 0 {
		r.Attrs(func(a slog.Attr) bool {

			strVal := a.Value.String()

			fgcolor := 39
			bgcolor := 49

			if a.Key == `err` || a.Key == `error` {
				fgcolor = 37 // red
				bgcolor = 41 // red
			} else {

				switch a.Value.Kind() {
				case slog.KindString:
					fgcolor = 33 // yellow
					if strings.ContainsAny(strVal, "\r\n") {
						strVal = strings.ReplaceAll(strVal, "\n", `\n`)
						strVal = strings.ReplaceAll(strVal, "\r", `\r`)
					}
					strVal = fmt.Sprintf(`"%s"`, strVal)
				case slog.KindBool:
					fgcolor = 92 // bright green
				case slog.KindInt64:
					fgcolor = 31 // red
				case slog.KindUint64:
					fgcolor = 31 // red
				case slog.KindFloat64:
					fgcolor = 31 // red
				case slog.KindDuration:
					fgcolor = 35 // magenta
				case slog.KindTime:
					fgcolor = 35 // magenta
				default:
					fgcolor = 37 // white
				}

			}

			finalOut.WriteString(fmt.Sprintf("%s=\033[%d;%dm%s\033[0m ", a.Key, fgcolor, bgcolor, strVal))
			strlen := len(a.Key) + len(strVal)

			if padding := 24 - strlen; padding > 0 {
				finalOut.WriteString(strings.Repeat(` `, padding))
			}

			return true
		})
	}

	timeStr := "\033[90m" + r.Time.Format("[15:04:05]") + "\033[0m" // bright black

	if msgLen > 24 {
		r.Message = r.Message[:24]
		msgLen = 24
	}
	if msgLen > h.minimumMessageLength {
		h.minimumMessageLength = len(r.Message)
	}

	if padding := h.minimumMessageLength - msgLen; padding > 0 {
		r.Message += strings.Repeat(` `, padding)
	}

	msg := fmt.Sprintf("\033[36m%s\033[39;49m", r.Message) // cyan

	h.l.Println(timeStr, level, msg, finalOut.String())

	return nil
}

func GetColorLogHandler(out io.Writer, logLvl slog.Level) *ColorHandler {

	opt := &slog.HandlerOptions{

		Level: logLvl,
		//AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("time", time.Now().Format("15:04:05"))
			}

			if a.Key == slog.MessageKey {
				if strings.HasPrefix(a.Value.String(), "INFO ") {
					return slog.String("msg", a.Value.String()[5:])
				}
			}

			if a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}

	h := &ColorHandler{
		Handler: slog.NewTextHandler(out, opt),
		l:       log.New(out, "", 0),
	}

	return h
}

func GetLogHandler(output *os.File, logLvl slog.Level) slog.Handler {
	return slog.NewTextHandler(
		output,
		&slog.HandlerOptions{

			Level: logLvl,
			//AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.String("time", time.Now().Format("15:04:05"))
				}

				if a.Key == slog.MessageKey {
					if strings.HasPrefix(a.Value.String(), "INFO ") {
						return slog.String("msg", a.Value.String()[5:])
					}
				}

				if a.Key == slog.LevelKey {
					return slog.Attr{}
				}
				return a
			},
		})
}
