package log

import (
	"github.com/sirupsen/logrus"
)

type rlog struct {
	entry     *logrus.Entry
	haveField bool
}

var logger *logrus.Logger

func Build() *logrus.Logger {

	logger = logrus.New()

	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	return logger
}

func newRlog() *rlog {
	return &rlog{
		entry: logrus.NewEntry(logger),
	}
}

func (r *rlog) Info(args ...interface{}) {
	r.entry.Info(args...)
}

func (r *rlog) Debug(args ...interface{}) {
	r.entry.Debug(args...)
}

func (r *rlog) Trace(args ...interface{}) {
	r.entry.Trace(args...)
}

func (r *rlog) Warn(args ...interface{}) {
	r.entry.Warn(args...)
}

func (r *rlog) Error(args ...interface{}) {
	r.entry.Error(args...)
}

func (r *rlog) Fatal(args ...interface{}) {
	r.entry.Fatal(args...)
}

func (r *rlog) Panic(args ...interface{}) {
	r.entry.Fatal(args...)
}

func (r *rlog) Infof(f string, v ...interface{}) {
	r.entry.Infof(f, v...)
}

func (r *rlog) Debugf(f string, v ...interface{}) {
	r.entry.Debugf(f, v...)
}
func (r *rlog) Tracef(f string, v ...interface{}) {
	r.entry.Tracef(f, v...)
}
func (r *rlog) Warnf(f string, v ...interface{}) {
	r.entry.Warnf(f, v...)
}

func (r *rlog) Errorf(f string, v ...interface{}) {
	r.entry.Errorf(f, v...)
}

func (r *rlog) Fatalf(f string, v ...interface{}) {
	r.entry.Fatalf(f, v...)
}

func (r *rlog) Panicf(f string, v ...interface{}) {
	r.entry.Panicf(f, v...)
}

func (r *rlog) WithFields(fields Fields) Logger {
	return &rlog{
		entry:     r.entry.WithFields(logrus.Fields(fields)),
		haveField: true,
	}
}

// WithField return a new logger with fields
func (r *rlog) WithField(key string, val interface{}) Logger {
	return &rlog{
		entry:     r.entry.WithField(key, val),
		haveField: true,
	}
}
