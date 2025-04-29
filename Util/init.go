package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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
