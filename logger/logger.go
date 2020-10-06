package logger

import (
	fmt "fmt"

	log "github.com/sirupsen/logrus"
)

// Logger is a common logger
type Logger interface {
	Tracef(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
}

// SimpleLogger ...
type SimpleLogger struct {
	Pfx string
}

// NewSimpleLogger ...
func NewSimpleLogger(pfx string) *SimpleLogger {
	return &SimpleLogger{pfx}
}

func (l *SimpleLogger) msg(f string, a ...interface{}) string {
	msg := fmt.Sprintf(f, a...)
	return fmt.Sprintf("%s: %s", l.Pfx, msg)
}

// Tracef ...
func (l *SimpleLogger) Tracef(f string, a ...interface{}) {
	log.Trace(l.msg(f, a...))
}

// Debugf ...
func (l *SimpleLogger) Debugf(f string, a ...interface{}) {
	log.Debug(l.msg(f, a...))
}

// Infof ...
func (l *SimpleLogger) Infof(f string, a ...interface{}) {
	log.Info(l.msg(f, a...))
}

// Warnf ...
func (l *SimpleLogger) Warnf(f string, a ...interface{}) {
	log.Warn(l.msg(f, a...))
}

// Errorf ...
func (l *SimpleLogger) Errorf(f string, a ...interface{}) {
	log.Error(l.msg(f, a...))
}
