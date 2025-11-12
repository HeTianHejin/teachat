package data

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	util "teachat/Util"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
   涉及数据库存取操作的定义和一些方法
*/

var db *sql.DB //数据库实例

func init() {
	var err error

	// 获取项目根目录（无论从哪里调用）
	_, filename, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(filename)) // 根据实际层级调整
	// 加载 .env 文件
	err = godotenv.Load(filepath.Join(rootDir, ".env"))
	if err != nil {
		util.PrintStdout("Error loading .env file")
	}
	// 从环境变量获取数据库配置
	dbdriver := os.Getenv("DB_DRIVER")
	dbhost := os.Getenv("DB_HOST")
	dbport, _ := strconv.Atoi(os.Getenv("DB_PORT")) // 字符串转整数
	dbuser := os.Getenv("DB_USER")
	dbpassword := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	dbsslmode := os.Getenv("DB_SSLMODE")
	dbTimeZone := os.Getenv("DB_TIMEZONE")
	//数据库连接
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbhost, dbport, dbuser, dbpassword, dbname, dbsslmode, dbTimeZone)
	db, err = sql.Open(dbdriver, psqlInfo)
	if err != nil {
		util.Fatal("星际迷失->茶棚数据库打开时：", err)
	}
	//测试数据库连接是否成功
	if err = db.Ping(); err != nil {
		util.Fatal("ping teachat database failure - 测试链接茶话会数据库失败", err)
	}

	//ok
	util.PrintStdout("Open tea chat database success, 星际茶棚数据库打开成功")

}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func Random_UUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		log.Println("Cannot generate UUID", err)
	}

	// 0x40 is reserved variant from RFC 4122
	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}

// hash plaintext with SHA-1
func Encrypt(plaintext string) (cryptext string) {
	cryptext = fmt.Sprintf("%x", sha1.Sum([]byte(plaintext)))
	return
}

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
	stmt, err := db.Prepare(statement)
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
	return db.Begin()
}
