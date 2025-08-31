/*
fully featured, public domain go logging library.
made by elisiei, published under cc0 1.0 (public domain)
with an additional ip waiver.
*/
package zlog

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type F map[string]any

var (
	ansiReset  = "\u001b[0m"
	ansiBlack  = "\u001b[30m"
	levelColor = map[Level]string{
		LevelDebug: "\u001b[34m",   // dark magenta
		LevelInfo:  "\u001b[32m",   // dark green
		LevelWarn:  "\u001b[33m",   // dark yellow
		LevelError: "\u001b[31m",   // dark red
		LevelFatal: "\u001b[35;1m", // brighter/deeper red
	}
	levelColorBg = map[Level]string{
		LevelDebug: "\u001b[44m",   // white background
		LevelInfo:  "\u001b[42m",   // green background
		LevelWarn:  "\u001b[43m",   // yellow background
		LevelError: "\u001b[41m",   // red background
		LevelFatal: "\u001b[45;1m", // deeper red background
	}
)

type Logger struct {
	mu         sync.Mutex
	out        io.Writer
	level      Level
	timeStamp  bool
	timeFormat string
	json       bool
	colors     bool
	caller     bool
	fields     F
}

func New() *Logger {
	return &Logger{
		out:        os.Stderr,
		level:      LevelInfo,
		timeStamp:  true,
		timeFormat: time.RFC3339,
		colors:     isTerminal(os.Stderr),
		caller:     true,
		fields:     make(F),
	}
}

var std = New()

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = w
	if file, ok := w.(*os.File); ok {
		l.colors = isTerminal(file)
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}
func (l *Logger) EnableTimestamps(on bool) {
	l.mu.Lock()
	l.timeStamp = on
	l.mu.Unlock()
}
func (l *Logger) SetTimeFormat(tf string) {
	l.mu.Lock()
	l.timeFormat = tf
	l.mu.Unlock()
}
func (l *Logger) SetJSON(on bool) {
	l.mu.Lock()
	l.json = on
	l.mu.Unlock()
}
func (l *Logger) EnableColors(on bool) {
	l.mu.Lock()
	l.colors = on
	l.mu.Unlock()
}
func (l *Logger) ShowCaller(on bool) {
	l.mu.Lock()
	l.caller = on
	l.mu.Unlock()
}

func (l *Logger) WithFields(f F) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	newFields := make(F, len(l.fields)+len(f))
	maps.Copy(newFields, l.fields)
	maps.Copy(newFields, f)
	return &Logger{
		out: l.out, level: l.level,
		timeStamp: l.timeStamp, timeFormat: l.timeFormat,
		json: l.json, colors: l.colors, caller: l.caller,
		fields: newFields,
	}
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func (l *Logger) log(level Level, msg string, extra F) {
	l.mu.Lock()
	out, jsonMode, colors, timeStamp, timeFormat, callerOn := l.out, l.json, l.colors, l.timeStamp, l.timeFormat, l.caller
	baseFields := make(F, len(l.fields))
	maps.Copy(baseFields, l.fields)
	l.mu.Unlock()

	if level < l.level {
		return
	}
	maps.Copy(baseFields, extra)

	var callerStr string
	if callerOn {
		if _, file, line, ok := runtime.Caller(3); ok {
			callerStr = fmt.Sprintf("%s:%d", shortFile(file), line)
		}
	}

	if jsonMode {
		entry := make(map[string]any, 4+len(baseFields))
		entry["level"], entry["msg"] = level.String(), msg
		if timeStamp {
			entry["time"] = time.Now().Format(timeFormat)
		}
		if callerStr != "" {
			entry["caller"] = callerStr
		}
		maps.Copy(entry, baseFields)
		b, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(out, "json marshal error: %v\n", err)
			return
		}
		fmt.Fprintln(out, string(b))
		if level == LevelFatal {
			os.Exit(1)
		}
		return
	}

	var b strings.Builder
	if colors {
		if c, ok := levelColor[level]; ok {
			b.WriteString(c)
		}
	}
	if timeStamp {
		b.WriteString(time.Now().Format(timeFormat))
		b.WriteString(" ")
	}
	if colors {
		if c, ok := levelColorBg[level]; ok {
			b.WriteString(c)
			b.WriteString(ansiBlack)
		}
	}

	b.WriteString(" ")
	b.WriteString(level.String())
	b.WriteString(" ")

	if colors {
		if c, ok := levelColor[level]; ok {
			b.WriteString(ansiReset)
			b.WriteString(c)
		}
	}
	b.WriteString(" ")
	b.WriteString(msg)

	if len(baseFields) > 0 {
		b.WriteString(" ")
		keys := make([]string, 0, len(baseFields))
		for k := range baseFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for i, k := range keys {
			if i > 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(&b, "%s=%v", k, baseFields[k])
		}
	}
	if callerStr != "" {
		fmt.Fprintf(&b, " (%s)", callerStr)
	}
	if colors {
		b.WriteString(ansiReset)
	}

	fmt.Fprintln(out, b.String())
	if level == LevelFatal {
		os.Exit(1)
	}
}

func shortFile(path string) string {
	parts := strings.Split(path, "/")
	n := len(parts)
	if n >= 2 {
		return strings.Join(parts[n-2:], "/")
	}
	return path
}

func (l *Logger) Debug(msg string) { l.log(LevelDebug, msg, nil) }
func (l *Logger) Info(msg string)  { l.log(LevelInfo, msg, nil) }
func (l *Logger) Warn(msg string)  { l.log(LevelWarn, msg, nil) }
func (l *Logger) Error(msg string) { l.log(LevelError, msg, nil) }
func (l *Logger) Fatal(msg string) { l.log(LevelFatal, msg, nil) }

func (l *Logger) Debugf(f string, a ...any) { l.log(LevelDebug, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Infof(f string, a ...any)  { l.log(LevelInfo, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Warnf(f string, a ...any)  { l.log(LevelWarn, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Errorf(f string, a ...any) { l.log(LevelError, fmt.Sprintf(f, a...), nil) }
func (l *Logger) Fatalf(f string, a ...any) { l.log(LevelFatal, fmt.Sprintf(f, a...), nil) }

func (l *Logger) Debugw(msg string, f F) { l.log(LevelDebug, msg, f) }
func (l *Logger) Infow(msg string, f F)  { l.log(LevelInfo, msg, f) }
func (l *Logger) Warnw(msg string, f F)  { l.log(LevelWarn, msg, f) }
func (l *Logger) Errorw(msg string, f F) { l.log(LevelError, msg, f) }
func (l *Logger) Fatalw(msg string, f F) { l.log(LevelFatal, msg, f) }

func SetOutput(w io.Writer)    { std.SetOutput(w) }
func SetLevel(l Level)         { std.SetLevel(l) }
func EnableTimestamps(on bool) { std.EnableTimestamps(on) }
func SetTimeFormat(tf string)  { std.SetTimeFormat(tf) }
func SetJSON(on bool)          { std.SetJSON(on) }
func EnableColors(on bool)     { std.EnableColors(on) }
func ShowCaller(on bool)       { std.ShowCaller(on) }
func WithFields(f F) *Logger   { return std.WithFields(f) }

func Debug(msg string) { std.Debug(msg) }
func Info(msg string)  { std.Info(msg) }
func Warn(msg string)  { std.Warn(msg) }
func Error(msg string) { std.Error(msg) }
func Fatal(msg string) { std.Fatal(msg) }

func Debugf(f string, a ...any) { std.Debugf(f, a...) }
func Infof(f string, a ...any)  { std.Infof(f, a...) }
func Warnf(f string, a ...any)  { std.Warnf(f, a...) }
func Errorf(f string, a ...any) { std.Errorf(f, a...) }
func Fatalf(f string, a ...any) { std.Fatalf(f, a...) }

func Debugw(msg string, f F) { std.Debugw(msg, f) }
func Infow(msg string, f F)  { std.Infow(msg, f) }
func Warnw(msg string, f F)  { std.Warnw(msg, f) }
func Errorw(msg string, f F) { std.Errorw(msg, f) }
func Fatalw(msg string, f F) { std.Fatalw(msg, f) }

func ParseLevel(s string) (Level, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error", "err":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("unknown level: %s", s)
	}
}
