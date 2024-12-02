package zlog

import (
	"github.com/rs/zerolog"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(newLogger(STDOUT, LevelDebug))
}
func instance() *Logger { return defaultLogger.Load() }

func NewLogLogger(mode string, level string, options ...Option) {
	var m LogMode
	var le Level
	if m.parse(mode) != nil || le.parse(level) != nil {
		panic("invalid log mode or level")
	}
	defaultLogger.Store(newLogger(m, le, options...))
}

func newLogger(mode LogMode, level Level, options ...Option) *Logger {
	opts := Options{
		Mode:  mode,
		Level: level,
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	l := newZerolog(opts.newWriter(), opts.Level.String())
	return &Logger{
		l: &l,
	}
}

type Logger struct {
	l *zerolog.Logger
}

func newZerolog(writer io.Writer, level string) zerolog.Logger {
	le, err := zerolog.ParseLevel(level)
	if err != nil {
		panic(err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	zerolog.TimeFieldFormat = time.DateTime
	return zerolog.New(writer).Level(le).
		With().Timestamp().Caller().Str("hostname", hostname).
		Logger().Hook(&TraceHook{})
}

func Debug() *zerolog.Event {
	return instance().l.Debug().Str("meth", getFunName(2))
}

func Info() *zerolog.Event {
	return instance().l.Info().Str("meth", getFunName(2))
}

func Warn() *zerolog.Event {
	return instance().l.Warn().Str("meth", getFunName(2))
}

func Error() *zerolog.Event {
	return instance().l.Error().Str("meth", getFunName(2))
}

// 获取第几层函数的名称
// 0-当前层 packageName.getFunName
// 1-上一层 packageName.Info
// 2-再上一层
// 3-再再上一层
func getFunName(l int) string {
	pc, _, _, _ := runtime.Caller(l)
	return runtime.FuncForPC(pc).Name()
}
