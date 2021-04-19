package ekafka

import "github.com/gotomicro/ego/core/elog"

// errorLogger is an elog to kafka-go Logger adapter
type logger struct {
	*elog.Component
}

func (l *logger) Printf(tmpl string, args ...interface{}) {
	l.Debugf(tmpl, args...)
}

// errorLogger is an elog to kafka-go ErrorLogger adapter
type errorLogger struct {
	*elog.Component
}

func (l *errorLogger) Printf(tmpl string, args ...interface{}) {
	l.Errorf(tmpl, args...)
}

func newKafkaLogger(wrappedLogger *elog.Component) *logger {
	return &logger{wrappedLogger}
}

func newKafkaErrorLogger(wrappedLogger *elog.Component) *errorLogger {
	return &errorLogger{wrappedLogger}
}
