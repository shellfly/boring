package log

import (
	"io"
	"log"
	"os"
	"strings"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError

	DefaultLevel = LevelInfo
)

type logger struct {
	l     *log.Logger
	level LogLevel
}

var std = New(os.Stderr, "", log.LstdFlags|log.Llongfile)

func New(out io.Writer, prefix string, flag int) *logger {
	l := log.New(out, prefix, flag)
	return &logger{
		l:     l,
		level: DefaultLevel,
	}
}

func SetLevel(level string) {
	var l LogLevel
	switch strings.ToLower(level) {
	case "debug":
		l = LevelDebug
	case "info":
		l = LevelInfo
	case "warn":
		l = LevelWarn
	case "error":
		l = LevelError
	default:
		l = LevelInfo
	}
	std.level = l
}

// Debug calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Debug(v ...interface{}) {
	if std.level > LevelDebug {
		return
	}
	std.l.Print(v...)
}

// Debugf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	if std.level > LevelDebug {
		return
	}
	std.l.Printf(format, v...)
}

// Debugln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Debugln(v ...interface{}) {
	if std.level > LevelDebug {
		return
	}
	std.l.Println(v...)
}

// Info calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Info(v ...interface{}) {
	if std.level > LevelInfo {
		return
	}
	std.l.Print(v...)
}

// Infof calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	if std.level > LevelInfo {
		return
	}
	std.l.Printf(format, v...)
}

// Infoln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Infoln(v ...interface{}) {
	if std.level > LevelInfo {
		return
	}
	std.l.Println(v...)
}

// Warn calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Warn(v ...interface{}) {
	if std.level > LevelWarn {
		return
	}
	std.l.Print(v...)
}

// Warnf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	if std.level > LevelWarn {
		return
	}
	std.l.Printf(format, v...)
}

// Warnln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Warnln(v ...interface{}) {
	if std.level > LevelWarn {
		return
	}
	std.l.Println(v...)
}

// Error calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Error(v ...interface{}) {
	if std.level > LevelError {
		return
	}
	std.l.Print(v...)
}

// Errorf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	if std.level > LevelError {
		return
	}
	std.l.Printf(format, v...)
}

// Errorln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Errorln(v ...interface{}) {
	if std.level > LevelError {
		return
	}
	std.l.Println(v...)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func Panic(v ...interface{}) {
	std.l.Panic(v...)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	std.l.Panicf(format, v...)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	std.l.Panicln(v...)
}
