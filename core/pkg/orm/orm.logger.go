package orm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"skeyevss/core/pkg/functions"
	"strings"
	"sync/atomic"
	"time"

	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

const (
	// Silent silent log level
	Silent gormLogger.LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
)

// gorm原日志的配置
type FileLogConfig struct {
	gormLogger.Config
}

type StdFileLogger struct {
	FileLogConfig
	infoStr,
	warnStr,
	errStr string

	traceStr,
	traceErrStr,
	traceWarnStr string

	path string
}

type logChanType struct {
	content,
	dir string
}

var (
	logFileChan         = make(chan *logChanType, 100)
	sipLogFile          *os.File
	sipLogFileCreatedAt int64
	logFileDroppedCount int64
)

func init() {
	go logToFile()

	go reportLogDropped()
}

func logToFile() {
	for {
		select {
		case data := <-logFileChan:
			var now = functions.NewTimer().Now()
			if now-sipLogFileCreatedAt >= 24*3600 {
				if sipLogFile != nil {
					_ = sipLogFile.Close()
					sipLogFile = nil
				}
			}

			if sipLogFile == nil {
				var (
					err  error
					file = path.Join(data.dir, fmt.Sprintf("%s.log", functions.NewTimer().Format("ymd")))
				)
				sipLogFile, err = os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("日志文件的打开错误 :", err)
					continue
				}
				sipLogFileCreatedAt = functions.NewTimer().DayInitTimestamp(0)
			}

			if _, err := sipLogFile.WriteString(data.content); err != nil {
				fmt.Println("写入日志文件错误 :", err)
			}
		}
	}
}

func reportLogDropped() {
	var ticker = time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		var dropped = atomic.SwapInt64(&logFileDroppedCount, 0)
		if dropped <= 0 {
			continue
		}

		functions.LogError("orm logger: 写文件队列已满，本周期丢弃日志数量: ", dropped)
	}
}

// 区分有颜色和无颜色

func NewStdFileLogger(config FileLogConfig, path string) *StdFileLogger {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if config.Colorful {
		infoStr = Green + "%s\n" + Reset + Green + "[info] " + Reset
		warnStr = BlueBold + "%s\n" + Reset + Magenta + "[warn] " + Reset
		errStr = Magenta + "%s\n" + Reset + Red + "[error] " + Reset
		traceStr = Green + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
		traceWarnStr = Green + "%s " + Yellow + "%s\n" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = RedBold + "%s " + MagentaBold + "%s\n" + Reset + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s"
	}

	return &StdFileLogger{
		FileLogConfig: config,
		path:          path,
		// Loggers:      loggers,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

func (logger *StdFileLogger) printf(msg string, data ...interface{}) {
	// if len(data) > 0 {
	// 	fmt.Printf("\n data[0]: %+v \n", data[0])
	// 	if v, ok := data[0].(string); ok {
	// 		data[0] = v + ", " + functions.CallerFile(7)
	// 	}
	// }

	var content = fmt.Sprintf(msg, data...)

	if logger.path == "" {
		log.Printf(content)
		return
	}

	var (
		now       = time.Now()
		formatted = now.Format("2006-01-02 15:04:05")
	)

	// 替换掉彩色打印符号
	content = strings.ReplaceAll(content, Reset, "")
	content = strings.ReplaceAll(content, Red, "")
	content = strings.ReplaceAll(content, Green, "")
	content = strings.ReplaceAll(content, Yellow, "")
	content = strings.ReplaceAll(content, Blue, "")
	content = strings.ReplaceAll(content, Magenta, "")
	content = strings.ReplaceAll(content, Cyan, "")
	content = strings.ReplaceAll(content, White, "")
	content = strings.ReplaceAll(content, BlueBold, "")
	content = strings.ReplaceAll(content, MagentaBold, "")
	content = strings.ReplaceAll(content, RedBold, "")
	content = strings.ReplaceAll(content, YellowBold, "")

	content = formatted + " " + content
	logger.LogToFile(logger.path, content+"\n")
}

func (logger *StdFileLogger) LogMode(lv gormLogger.LogLevel) gormLogger.Interface {
	logger.LogLevel = lv
	return logger
}

func (logger *StdFileLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Info {
		var (
			fileCaller = utils.FileWithLineNum()
			funcCaller = ctx.Value(callerFileCtxName)
		)
		if v, ok := funcCaller.(string); ok && v != "" {
			fileCaller = v + "\n		    " + fileCaller
		}

		logger.xPrintf(logger.infoStr+msg, append([]interface{}{fileCaller}, data...)...)
	}
}

// warn
func (logger *StdFileLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Warn {
		var (
			fileCaller = utils.FileWithLineNum()
			funcCaller = ctx.Value(callerFileCtxName)
		)
		if v, ok := funcCaller.(string); ok && v != "" {
			fileCaller = v + "\n		    " + fileCaller
		}

		logger.xPrintf(logger.warnStr+msg, append([]interface{}{fileCaller}, data...)...)
	}
}

// error
func (logger *StdFileLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if logger.LogLevel >= Error {
		var (
			fileCaller = utils.FileWithLineNum()
			funcCaller = ctx.Value(callerFileCtxName)
		)
		if v, ok := funcCaller.(string); ok && v != "" {
			fileCaller = v + "\n		    " + fileCaller
		}

		logger.xPrintf(logger.errStr+msg, append([]interface{}{fileCaller}, data...)...)
	}
}

func (logger *StdFileLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if logger.LogLevel <= Silent {
		return
	}

	var (
		elapsed    = time.Since(begin)
		fileCaller = utils.FileWithLineNum()
		funcCaller = ctx.Value(callerFileCtxName)
	)

	if v, ok := funcCaller.(string); ok && v != "" {
		fileCaller = v + "\n		    " + fileCaller
	}

	switch {
	case err != nil && logger.LogLevel >= Error && (!errors.Is(err, gormLogger.ErrRecordNotFound) || !logger.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			logger.xPrintf(logger.traceErrStr+"\n", fileCaller, "\n		    "+err.Error(), float64(elapsed.Nanoseconds())/1e6, "-", strings.ReplaceAll(sql, "%", "%%"))
		} else {
			logger.xPrintf(logger.traceErrStr+"\n", fileCaller, "\n		    "+err.Error(), float64(elapsed.Nanoseconds())/1e6, rows, strings.ReplaceAll(sql, "%", "%%"))
		}

	case elapsed > logger.SlowThreshold && logger.SlowThreshold != 0 && logger.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", logger.SlowThreshold)
		if rows == -1 {
			logger.xPrintf(logger.traceWarnStr, fileCaller, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", strings.ReplaceAll(sql, "%", "%%"))
		} else {
			logger.xPrintf(logger.traceWarnStr, fileCaller, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, strings.ReplaceAll(sql, "%", "%%"))
		}

	case logger.LogLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			logger.xPrintf(logger.traceStr, fileCaller, float64(elapsed.Nanoseconds())/1e6, "-", strings.ReplaceAll(sql, "%", "%%"))
		} else {
			logger.xPrintf(logger.traceStr, fileCaller, float64(elapsed.Nanoseconds())/1e6, rows, strings.ReplaceAll(sql, "%", "%%"))
		}
	}
}

func (logger *StdFileLogger) PrintStackTrace(_ error) string {
	var buf = bytes.NewBuffer(nil)
	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		_, _ = fmt.Fprintf(buf, "%d: %s:%d (0x%x)\n", i, file, line, pc)
	}

	return buf.String()
}

func (logger *StdFileLogger) LogToFile(dir, content string) {
	select {
	case logFileChan <- &logChanType{
		content: content,
		dir:     dir,
	}:
		return

	default:
		atomic.AddInt64(&logFileDroppedCount, 1)
		return
	}
}

func (logger *StdFileLogger) xPrintf(msg string, data ...interface{}) {
	logger.printf(msg, data...)
}
