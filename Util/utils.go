package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

/*
   存放各个route包文件和data包文件共享的一些方法
*/

type Configuration struct {
	Address          string
	ReadTimeout      int64
	WriteTimeout     int64
	Static           string
	SysMail_Username string
	SysMail_Password string
	SysMail_Host     string
	SysMail_Port     string
	MaxInviteTeams   int // 茶围、茶台最大可邀请团队数
	MaxTeamMembers   int // 团队最大成员数
	MaxTeamCount     int // ����创建的��队数上限
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

// for logging
func Info(args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Println(args...)
}

func Danger(args ...interface{}) {
	logger.SetPrefix("ERROR ")
	logger.Println(args...)
}

func Warning(args ...interface{}) {
	logger.SetPrefix("WARNING ")
	logger.Println(args...)
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
