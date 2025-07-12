package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

/*
   配置文件；
   日志文件；
   版本信息；
   一些常量；
   一些工具函数；
*/

// 初始化日志
func init() {
	// 开发模式默认日志配置
	InitLogger(false, LevelDebug) // 默认控制台输出，Debug级别
}

// 配置文件结构体
type Configuration struct {
	Address          string
	ReadTimeout      int64
	WriteTimeout     int64
	Static           string
	ImageDir         string
	ImageExt         string
	TemplatesDir     string
	TemplateExt      string
	ThreadMinWord    int64 //  茶议最小字数限制
	ThreadMaxWord    int64 // 茶议最大字数限制
	PostMinWord      int64 // 品味最小字数限制
	MaxInviteTeams   int64 // 茶围、茶台最大可邀请团队数
	MaxTeamMembers   int64 // 团队最大成员数
	MaxTeamsCount    int64 // 个人创建的团队数上限
	MaxSurvivalTeams int64 // 个人最大活跃团队数
	PoliteMode       bool  // Debug模式

	// SysMail_Username string
	// SysMail_Password string
	// SysMail_Host     string
	//SysMail_Port   string
}

var Config Configuration

// 读取配置文件内容
func LoadConfig() error {
	configPath := getConfigPath()

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close() // 确保文件关闭

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&Config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 路径标准化处理
	Config.ImageDir = filepath.Clean(Config.ImageDir) + string(filepath.Separator)
	Config.TemplatesDir = filepath.Clean(Config.TemplatesDir) + string(filepath.Separator)

	return nil
}

// 使用绝对路径获取配置文件
func getConfigPath() string {
	// 优先使用环境变量指定的配置路径
	if path := os.Getenv("APP_CONFIG"); path != "" {
		return path
	}

	// 其次尝试当前目录下的 config.json
	if _, err := os.Stat("config.json"); err == nil {
		return "config.json"
	}

	// 最后尝试用户主目录
	if home, err := os.UserHomeDir(); err == nil {
		defaultPath := filepath.Join(home, ".teachat", "config.json")
		if _, err := os.Stat(defaultPath); err == nil {
			return defaultPath
		}
	}

	return "config.json" // 默认使用当前目录下的 config.json
}
func (c *Configuration) Validate() error {
	if c.Address == "" {
		return errors.New("服务器地址不能为空")
	}
	if c.ImageDir == "" {
		return errors.New("图片目录不能为空")
	}
	if c.TemplatesDir == "" {
		return errors.New("模板目录不能为空")
	}
	if c.TemplateExt == "" {
		return errors.New("模板扩展名不能为空")
	}
	if c.MaxInviteTeams == 0 {
		return errors.New("最大可邀请团队数不能为空")
	}
	if c.MaxTeamMembers == 0 {
		return errors.New("团队最大成员数不能为空")
	}
	if c.MaxTeamsCount == 0 {
		return errors.New("个人创建的团队数上限不能为空")
	}
	if c.MaxSurvivalTeams == 0 {
		return errors.New("个人最大活跃团队数不能为空")
	}
	if c.Static == "" {
		return errors.New("静态文件目录不能为空")
	}
	if c.ThreadMaxWord == 0 {
		return errors.New("茶议最大字数限制不能为空")
	}
	if c.ThreadMinWord == 0 {
		return errors.New("茶议最小字数限制不能为空")
	}
	if c.ImageExt == "" {
		return errors.New("图片扩展名不能为空")
	}
	return nil
}

// Version
func Version() string {
	return "0.7"
}

// Convenience function for printing to stdout
func PrintStdout(a ...any) {
	fmt.Println(a...)
}

// 检查文件是否已经存在
func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
