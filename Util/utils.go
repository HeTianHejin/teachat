package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
)

/*
   配置文件；
   日志文件；
   版本信息；
   一些常量；
   一些工具函数；
*/

type Configuration struct {
	Address      string
	ReadTimeout  int64
	WriteTimeout int64
	Static       string
	// SysMail_Username string
	// SysMail_Password string
	// SysMail_Host     string
	//SysMail_Port   string
	MaxInviteTeams   int64 // 茶围、茶台最大可邀请团队数
	MaxTeamMembers   int64 // 团队最大成员数
	MaxTeamsCount    int64 // 个人创建的团队数上限
	MaxSurvivalTeams int64 // 个人最大活跃团队数
}

var Config Configuration

// Convenience function for printing to stdout
func PrintStdout(a ...interface{}) {
	fmt.Println(a...)
}

// 初始化配置
func init() {

	loadConfig()

	// 开发模式默认日志配置
	InitLogger(false, LevelDebug) // 默认控制台输出，Debug级别

	file, err := os.OpenFile("teachatWeb.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

// 读取配置文件内容
func loadConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Fatalln("Cannot open config file", err)
	}
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err = decoder.Decode(&Config)
	if err != nil {
		log.Fatalln("Cannot get configuration from file", err)
	}
}

// Version
func Version() string {
	return "0.7"
}

// 检查文件是否已经存在
func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

// Error() 为了显示调试过程中错误信息，而不是终止程序,
// 不记录到日志文件里
// func Error(args ...interface{}) {
// 	log.Println(args...)
// 	//panic(args)
// }

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
func logWithCaller(prefix string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2)        // 跳过两层调用栈
	file = file[strings.LastIndex(file, "/")+1:] // 只保留文件名
	logger.SetPrefix(prefix + " ")
	logger.Printf("%s:%d - %v", file, line, v)
}

// Debug 开发调试信息
func Debug(v ...interface{}) {
	if logLevel <= LevelDebug {
		logWithCaller("DEBUG", v...)
	}
}

// Info 常规信息
func Info(v ...interface{}) {
	if logLevel <= LevelInfo {
		logWithCaller("INFO", v...)
	}
}

// Warning 警告信息
func Warning(v ...interface{}) {
	if logLevel <= LevelWarning {
		logWithCaller("WARNING", v...)
	}
}

// Error 错误信息
func Error(v ...interface{}) {
	if logLevel <= LevelError {
		logWithCaller("ERROR", v...)
	}
}

// Fatal 致命错误并退出
func Fatal(v ...interface{}) {
	logWithCaller("FATAL", v...)
	os.Exit(1)
}
