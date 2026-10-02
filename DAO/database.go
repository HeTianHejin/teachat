package dao

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	util "teachat/Util"
	"time"
	"uuid"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
   涉及数据库存取操作的定义和一些方法
*/

//var db *sql.DB //数据库实例

// DB 数据库实例
var DB *sql.DB

func InitDB() error {
	// 统一从 exe 所在目录读取 .env，避免双击运行时找不到配置
	envPath := filepath.Join(util.AppDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		util.PrintStdout("error load .env file")
		util.Error("load .env file: %v", err)
	}

	// 生产阶段从环境变量获取数据库配置
	dbdriver := os.Getenv("DB_DRIVER")
	dbhost := os.Getenv("DB_HOST")
	dbport, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	dbuser := os.Getenv("DB_USER")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dbsslmode := os.Getenv("DB_SSLMODE")
	dbTimeZone := os.Getenv("DB_TIMEZONE")

	//数据库连接
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbhost, dbport, dbuser, dbpassword, dbname, dbsslmode, dbTimeZone)
	db, err := sql.Open(dbdriver, psqlInfo)
	if err != nil {
		return fmt.Errorf("cannot open database teachat: %w", err)
	}
	// 配置连接池
	db.SetMaxOpenConns(25)                 // 最多同时打开 25 个连接（并发上限）
	db.SetMaxIdleConns(25)                 // 建议和 Open 一致，避免连接抖动
	db.SetConnMaxLifetime(2 * time.Minute) // 每个连接最多活 2 分钟，到期后主动回收重建
	// 可选：DB.SetConnMaxIdleTime(time.Minute) // 空闲连接多久后回收

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping database teachat: %w", err)
	}

	DB = db
	return nil
}

func Random_UUID() (uuidString string) {
	uuidString = uuid.NewV4().String()
	return
}

// hash plaintext with SHA-1
// func Encrypt(plaintext string) (cryptext string) {
// 	cryptext = fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
// 	return
// }

// 时间处理格式化
const (
	FMT_DATE_TIME        = "2006-01-02 15:04:05"
	FMT_DATE             = "2006-01-02"
	FMT_TIME             = "15:04:05"
	FMT_DATE_FULLTIME_CN = "2006年01月02日 15时04分05秒"
	FMT_DATE_TIME_CN     = "2006年01月02日 15时04分"
	FMT_DATE_CN          = "2006年01月02日"
	FMT_TIME_CN          = "15时04分05秒"
)

// 字符串时间转时间类型
func TimeParse(timeStr, layout string) (time.Time, error) {
	return time.Parse(layout, timeStr)
}

// return yyyyMMdd
func GetDay(time time.Time) int {
	ret, _ := strconv.Atoi(time.Format("20060102"))
	return ret
}

// create a new read
func SaveReadedUserId(thread_id int, user_id int) (read Read, err error) {
	statement := "INSERT INTO reads (thread_id, user_id, read_at) VALUES ($1, $2, $3) RETURNING id, thread_id, user_id, read_at"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(thread_id, user_id, time.Now()).Scan(&read.Id, &read.ThreadId, &read.UserId, &read.ReadAt)
	return
}

// contains 检查切片中是否包含特定元素
func contains(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// BeginTx 开始一个数据库事务
func BeginTx() (*sql.Tx, error) {
	return DB.Begin()
}
