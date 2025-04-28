package data

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	util "teachat/Util"
	"time"

	_ "github.com/lib/pq"
)

/*
   涉及数据库存取操作的定义和一些方法
*/

// 定义数据库链接常量
const (
	dbdriver   = "postgres"
	dbhost     = "localhost"
	dbport     = 5432
	dbuser     = "postgres"
	dbpassword = "teachat"
	dbname     = "teachat"
	dbsslmode  = "disable"
	dbTimeZone = "Asia/Shanghai"
)

var Db *sql.DB // postgres 数据库实例

// 使用常量管理文件路径和文件扩展名，增加可维护性
const (
	ImageDir = "./public/image/"
	ImageExt = ".jpeg"
)

func init() {
	var err error
	//数据库连接
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbhost, dbport, dbuser, dbpassword, dbname, dbsslmode, dbTimeZone)
	Db, err = sql.Open(dbdriver, psqlInfo)
	if err != nil {
		util.Fatal("星际迷失->茶棚数据库打开时",err)
	}
	//测试数据库连接是否成功
	err = Db.Ping()
	if err != nil {
		util.Fatal("ping teachat database failure - 测试链接茶话会数据库失败~~~",err)
	}

	//ok
	util.PrintStdout("Well done, 星际茶棚开始服务")

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
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(thread_id, user_id, time.Now()).Scan(&read.Id, &read.ThreadId, &read.UserId, &read.ReadAt)
	return
}

// 生成头像默认（占位）图片，jpeg格式，64*64
func DefaultAvatar(uuid string) {
	// 生成64*64新画布
	bg := image.NewRGBA(image.Rect(0, 0, 64, 64))

	newFilePath := ImageDir + uuid + ImageExt
	newFile, err := os.Create(newFilePath)
	if err != nil {

		return
	}
	// 确保文件在函数执行完毕后关闭
	defer newFile.Close()

	// 先写内存缓存
	buff := bufio.NewWriter(newFile)
	//转换格式
	jpeg.Encode(buff, bg, nil)

	// 写入硬盘
	buff.Flush()
}
