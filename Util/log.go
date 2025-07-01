package util

import (
	"log"
	"os"
	"runtime"
)

// 日志级别常量
const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
)

var (
	logger      *log.Logger
	logLevel    int    = LevelDebug // 默认开发模式为Debug级别
	logToFile   bool   = false      // 默认不输出到文件
	logFilePath string = "teachatWeb.log"
)

// InitLogger 初始化日志配置
// writeToFile: 是否写入文件
// level: 日志级别
func InitLogger(writeToFile bool, level int) {
	logToFile = writeToFile
	logLevel = level

	if logToFile {
		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", err)
		}
		logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

// 带调用位置的基础日志方法
func logWithCaller(skip int, prefix string, v ...any) {
	_, file, line, ok := runtime.Caller(skip + 1) // +1 跳过本函数
	if !ok {
		file = "???"
		line = 0
	}
	//	file = filepath.Base(file)

	logger.SetPrefix(prefix + " ")
	logger.Printf("%s:%d - %v", file, line, v)
}

// Debug 开发调试信息
func Debug(v ...any) {
	if logLevel <= LevelDebug {
		logWithCaller(1, "DEBUG", v...)
	}
}

// Info 常规信息
func Info(v ...any) {
	if logLevel <= LevelInfo {
		logWithCaller(1, "INFO", v...)
	}
}

// Warning 警告信息
func Warning(v ...any) {
	if logLevel <= LevelWarning {
		logWithCaller(1, "WARNING", v...)
	}
}

// Error 错误信息
func Error(v ...any) {
	if logLevel <= LevelError {
		logWithCaller(1, "ERROR", v...)
	}
}

// Fatal 致命错误并退出
func Fatal(v ...any) {
	logWithCaller(1, "FATAL", v...)
	os.Exit(1)
}
