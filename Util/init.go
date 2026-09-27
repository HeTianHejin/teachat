package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

/*
   配置文件；
   日志文件；
   版本信息；
   一些常量；
   一些工具函数；
*/

var AppDir = appBaseDir()

func appBaseDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	wd, _ := os.Getwd()
	return wd
}

func AbsPath(rel string) string {
	if rel == "" {
		return AppDir
	}
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Join(AppDir, rel)
}

// 配置文件结构体
type Configuration struct {
	Address                string
	ReadTimeout            int64
	WriteTimeout           int64
	Static                 string
	ImageDir               string
	UserImageDir           string
	TeamImageDir           string
	ImageExt               string
	TemplatesDir           string
	TemplateExt            string
	ThreadMinWord          int64 //  茶议最小字数限制
	ThreadMaxWord          int64 // 茶议最大字数限制
	PostMinWord            int64 // 品味最小字数限制
	MaxInviteTeams         int64 // 茶围、茶台最大可邀请团队数
	MaxTeamMembers         int64 // 团队最大成员数
	MaxTeamsCount          int64 // 个人创建的团队数上限
	MaxSurvivalTeams       int64 // 个人最大活跃团队数
	PoliteMode             bool  // Debug模式(是否启用“友邻蒙评”审茶)
	DefaultSearchResultNum int64 // 默认搜索结果数

	// SysMail_Username string
	// SysMail_Password string
	// SysMail_Host     string
	//SysMail_Port   string
}

var Config Configuration

// 读取配置文件内容
func LoadConfig() error {
	// 从 exe 所在目录加载 .env（忽略未找到的错误，兼容开发环境/打包环境）
	err := godotenv.Load(filepath.Join(AppDir, ".env"))
	if err != nil {
		return fmt.Errorf("cannot load .env file: %w", err)
	}

	configPath := filepath.Join(AppDir, "config.json")
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("cannot open config.json file: %w", err)
	}
	defer file.Close() // 确保文件关闭

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&Config); err != nil {
		return fmt.Errorf("failed to parse the configuration file: %w", err)
	}

	// 路径标准化处理，支持 exe 直接运行和源码运行
	Config.Static = filepath.Clean(AbsPath(Config.Static))
	Config.ImageDir = filepath.Clean(AbsPath(Config.ImageDir)) + string(filepath.Separator)
	Config.UserImageDir = filepath.Clean(AbsPath(Config.UserImageDir)) + string(filepath.Separator)
	Config.TeamImageDir = filepath.Clean(AbsPath(Config.TeamImageDir)) + string(filepath.Separator)
	Config.TemplatesDir = filepath.Clean(AbsPath(Config.TemplatesDir)) + string(filepath.Separator)

	return nil
}

func (c *Configuration) Validate() error {
	if c.Address == "" {
		return errors.New("server address can't be empty")
	}
	if c.ImageDir == "" {
		return errors.New("the image directory can't be empty")
	}
	if c.UserImageDir == "" {
		return errors.New("the user avatar image directory can't be empty")
	}
	if c.TeamImageDir == "" {
		return errors.New("the team avatar image directory can't be empty")
	}
	if c.TemplatesDir == "" {
		return errors.New("template directory cannot be empty")
	}
	if c.TemplateExt == "" {
		return errors.New("the template file extension cannot be empty")
	}
	if c.MaxInviteTeams == 0 {
		return errors.New("the maximum number of teams that can be invited cannot be empty")
	}
	if c.MaxTeamMembers == 0 {
		return errors.New("the maximum number of team members cannot be empty")
	}
	if c.MaxTeamsCount == 0 {
		return errors.New("the maximum number of teams an individual can create cannot be empty")
	}
	if c.MaxSurvivalTeams == 0 {
		return errors.New("the maximum number of active teams for an individual cannot be empty")
	}
	if c.Static == "" {
		return errors.New("the static file directory can't be empty")
	}
	if c.ThreadMaxWord == 0 {
		return errors.New("the maximum word limit for the tea discussion cannot be empty")
	}
	if c.ThreadMinWord == 0 {
		return errors.New("the minimum word limit for tea discussion cannot be empty")
	}
	if c.ImageExt == "" {
		return errors.New("image file extension can't be empty")
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

// FormatFloat 格式化浮点数，保留指定位数小数
func FormatFloat(num float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, num)
}
