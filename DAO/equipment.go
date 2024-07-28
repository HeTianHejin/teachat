package data

import (
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"log"
	"math"

	"regexp"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// 定义数据库链接常量
const (
	dbhost     = "localhost"
	dbport     = 5432
	dbuser     = "postgres"
	dbpassword = "mima"
	dbname     = "teachatwebdb"
	dbsslmode  = "disable"
	dbTimeZone = "Asia/Shanghai"
	dbdriver   = "postgres"
)

// 使用常量管理文件路径和文件扩展名，增加可维护性
const (
	ImageDir = "./public/image/"
	ImageExt = ".jpeg"
)

var Db *sql.DB

func init() {
	var err error
	//数据库连接
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s TimeZone=%s",
		dbhost, dbport, dbuser, dbpassword, dbname, dbsslmode, dbTimeZone)
	Db, err = sql.Open(dbdriver, psqlInfo)
	if err != nil {
		log.Fatal(err, "星际茶棚数据库迷失～！")
	}
	//测试数据库连接是否成功
	err = Db.Ping()
	if err != nil {
		log.Panic(err, "星际茶棚数据库链接测试失败～！")
	}
	log.Println("星际茶棚开始服务！")

}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func CreateUuid() (uuid string) {
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

// 验证邮箱格式是否正确，正确返回true，错误返回false。
func VerifyEmailFormat(email string) bool {
	pattern := `^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

// 验证team_id_list:"2,19,87..."字符串格式是否正确，正确返回true，错误返回false。
func VerifyTeamIdListFormat(teamIdList string) bool {
	if teamIdList == "" {
		return false
	}
	pattern := `^[0-9]+(,[0-9]+)*$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(teamIdList)
}

// 输入两个统计数（辩论的正方累积得分数，辩论总得分数）（整数），计算前者与后者比值，结果浮点数向上四舍五入取整,
// 返回百分数的分子整数
func ProgressRound(numerator, denominator int) int {

	if denominator == 0 {
		// 分母为0时，视作未有记录，即未进行表决状态，返回100
		return 100
	}
	if numerator == denominator {
		// 分子等于分母时，表示100%正方
		return 100
	}
	ratio := float64(numerator) / float64(denominator) * 100

	// if numerator > denominator {
	// 	// 分子大于分母时，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 0 {
	// 	// 分子小于分母且比例为负数，表示统计数据输入错误，返回一个中间值
	// 	return 50
	// } else if ratio < 1 {
	// 	// 比例小于1时，返回最低限度值1
	// 	return 1
	// }

	// 其他情况，使用math.Floor确保向下取整，然后四舍五入
	return int(math.Floor(ratio + 0.5))
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
