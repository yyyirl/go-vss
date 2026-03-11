package server

import (
	"fmt"

	"github.com/ghettovoice/gosip/log"
	"github.com/zeromicro/go-zero/core/logx"

	"skeyevss/core/pkg/functions"
)

var _ log.Logger = (*Logger)(nil)

type Logger struct {
}

func NewLogger() *Logger {
	return new(Logger)
}

func (l *Logger) Print(args ...interface{}) {
	// logx.Info(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Printf(format string, args ...interface{}) {
	// logx.Info(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Trace(args ...interface{}) {
	// logx.Info(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	// logx.Info(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Debug(args ...interface{}) {
	// logx.Info(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	// logx.Info(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Info(args ...interface{}) {
	// logx.Info(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	// logx.Info(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Warn(args ...interface{}) {
	// logx.Alert(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	// logx.Alert(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Error(args ...interface{}) {
	logx.Error(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	logx.Error(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Fatal(args ...interface{}) {
	logx.Error(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	logx.Error(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) Panic(args ...interface{}) {
	// logx.Info(fmt.Sprint(append(args, " \n"+functions.Caller(2))...))
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	// logx.Info(fmt.Sprintf("[SIP Handler Log]"+format+" \n"+functions.Caller(2), args...))
}

func (l *Logger) WithPrefix(_ string) log.Logger {
	return NewLogger()
}

func (l *Logger) Prefix() string {
	return ""
}

func (l *Logger) WithFields(_ map[string]interface{}) log.Logger {
	return NewLogger()
}

func (l *Logger) Fields() log.Fields {
	return nil
}

func (l *Logger) SetLevel(_ uint32) {

}
