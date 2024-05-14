package log

import (
	"fmt"

	"go.uber.org/atomic"
)

// _defaultLevel is package default logging level.
var _defaultLevel = atomic.NewUint32(uint32(InfoLevel))

func SetLevel(level Level) {
	_defaultLevel.Store(uint32(level))
}

func Debugf(format string, args ...any) {
	logf(DebugLevel, format, args...)
}

func Infof(format string, args ...any) {
	logf(InfoLevel, format, args...)
}

func Warnf(format string, args ...any) {
	logf(WarnLevel, format, args...)
}

func Errorf(format string, args ...any) {
	logf(ErrorLevel, format, args...)
}

func Fatalf(format string, args ...any) {
	logf(ErrorLevel, format, args...)
}

func logf(level Level, format string, args ...any) {
	if uint32(level) > _defaultLevel.Load() {
		return
	}
	fmt.Println("[", level.String(), "]", fmt.Sprintf(format, args...))
}
