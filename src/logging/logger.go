package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	NONE  Level = iota // 0
	ERROR              // 1
	INFO               // 2
	TRAF               // 3
	DEBUG              // 4
	TRACE              // 5
)

type Logger struct {
	mu    sync.Mutex
	out   io.Writer
	level Level
}

func New(level Level) *Logger {
	return &Logger{
		out:   os.Stdout,
		level: level,
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

func (l *Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) Log(level Level, tag string, format string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if level > l.level {
		return
	}
	ts := time.Now().Format("15:04:05")
	levelStr := [...]string{"NONE", "ERROR", "INFO", "TRAF", "DEBUG", "TRACE"}[level]
	msg := fmt.Sprintf(format, args...)
	//fmt.Fprintf(l.out, "[%s] %-5s %s: %s\n", ts, levelStr, tag, msg) // with generatingFunc tag
	fmt.Fprintf(l.out, "[%s] %-5s: %s\n", ts, levelStr, msg) // without generatingFunc tag
}
