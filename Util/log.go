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
	logger   *log.Logger
	logLevel int = LevelDebug
)

// InitLogger 初始化日志配置
// writeToFile: 是否写入文件
// level: 日志级别
func InitLogger(writeToFile bool, level int) {
	logLevel = level
	var writer = os.Stdout
	if writeToFile {
		file, err := os.OpenFile("teachatWeb.log",
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", err)
		}
		writer = file
	}
	logger = log.New(writer, "", log.Ldate|log.Ltime|log.Lshortfile)
}

// logWithCaller 根据调用栈信息记录日志
func logWithCaller(skip int, prefix string, format string, v ...any) {
	if logger == nil {
		InitLogger(false, LevelDebug)
	}

	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		file = "???"
		line = 0
	}

	msg := fmt.Sprintf(format, v...)
	logger.SetPrefix(prefix + " ")
	logger.Output(2, fmt.Sprintf("%s:%d - %s", file, line, msg))
}

// 支持格式化的日志函数
func Debug(format string, v ...any) {
	if logLevel <= LevelDebug {
		logWithCaller(1, "DEBUG", format, v...)
	}
}

func Info(format string, v ...any) {
	if logLevel <= LevelInfo {
		logWithCaller(1, "INFO", format, v...)
	}
}

func Warning(format string, v ...any) {
	if logLevel <= LevelWarning {
		logWithCaller(1, "WARNING", format, v...)
	}
}

func Error(format string, v ...any) {
	if logLevel <= LevelError {
		logWithCaller(1, "ERROR", format, v...)
	}
}

func Panic(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	logWithCaller(1, "PANIC", "%s", msg)
	panic(msg)
}
