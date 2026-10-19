package util

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

// 日志级别常量
// 开发环境
// InitLogger(true, LevelDebug) // 看到所有日志
// 生产环境
// InitLogger(true, LevelWarning) // 只看到警告和错误
// 或
// InitLogger(true, LevelError) // 只看到错误
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
)

var (
	loggerStdout *log.Logger
	infoLogger   *log.Logger
	errorLogger  *log.Logger
	logLevel     int = LevelDebug
)

// InitLogger 初始化日志配置
// writeToFile: 是否写入文件
// level: 日志级别
func InitLogger(writeToFile bool, level int) {
	logLevel = level
	// 设置日志输出到标准输出
	loggerStdout = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	// 设置日志输出到文件
	infoLogger = nil
	errorLogger = nil
	if writeToFile {
		infoFile, err := os.OpenFile("info.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("fatal! cannot open or create info.log file:", err)
		}
		errorFile, err := os.OpenFile("error.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			_ = infoFile.Close()
			log.Fatalln("fatal! cannot open or create error.log file:", err)
		}
		flags := log.Ldate | log.Ltime | log.Lshortfile
		infoLogger = log.New(infoFile, "", flags)
		errorLogger = log.New(errorFile, "", flags)
	}
}

// logWithCaller 根据调用栈信息记录日志
func logWithCaller(target *log.Logger, skip int, prefix string, format string, v ...any) {
	if loggerStdout == nil {
		InitLogger(false, LevelDebug)
	}

	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		file = "???"
		line = 0
	}

	msg := fmt.Sprintf(format, v...)
	entry := fmt.Sprintf("%s %s:%d - %s", prefix, file, line, msg)
	loggerStdout.Output(2, entry)
	if target != nil {
		target.Output(2, entry)
	}
}

// 支持格式化的日志函数
func Debug(format string, v ...any) {
	if logLevel <= LevelDebug {
		logWithCaller(loggerStdout, 1, "DEBUG", format, v...)
	}
}

func Info(format string, v ...any) {
	if logLevel <= LevelInfo {
		logWithCaller(infoLogger, 1, "INFO", format, v...)
	}
}

func Warning(format string, v ...any) {
	if logLevel <= LevelWarning {
		logWithCaller(infoLogger, 1, "WARNING", format, v...)
	}
}

func Error(format string, v ...any) {
	if logLevel <= LevelError {
		logWithCaller(errorLogger, 1, "ERROR", format, v...)
	}
}

func Panic(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	logWithCaller(errorLogger, 1, "PANIC", "%s", msg)
	panic(msg)
}
