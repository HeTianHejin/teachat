package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
)

/*
   存放各个route包文件和data包文件共享的一些方法
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

var logger *log.Logger

// Convenience function for printing to stdout
func PrintStdout(a ...interface{}) {
	fmt.Println(a...)
}

// 初始化配置
func init() {
	loadConfig()
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

// 尝试记录错误信息发生的位置(文件，行)
func LogError(err error) (error_info string) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	error_info = fmt.Sprintf("Error occurred in %s:%d - %v", file, line, err)
	//log.Printf("%s", error_info) // 打印日志
	return error_info // 返回错误信息
}

// for logging
func Info(args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Println(args...)
}

func Warning(args ...interface{}) {
	logger.SetPrefix("WARNING ")
	logger.Println(args...)
}

func Danger(args ...interface{}) {
	logger.SetPrefix("ERROR ")
	logger.Println(args...)
}

// PanicTea() 为了显示调试过程中错误信息，而不是终止程序,
// 不记录到日志文件里
func PanicTea(args ...interface{}) {
	log.Println(args...)
	//panic(args)
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
