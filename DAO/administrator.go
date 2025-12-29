package dao

import "time"

type Watchword struct {
	Id              int
	Word            string
	AdministratorId int
	CreatedAt       time.Time
}

type Administrator struct {
	Id        int
	Uuid      string
	UserId    int
	Role      string
	Password  string
	CreatedAt time.Time
	Valid     bool
}

var AdminRole = map[string]string{
	"cabin crew": "客舱机组人员",
	"pilot":      "飞行员",
	"captain":    "船长",
}

var Status = map[bool]string{
	true:  "正常",
	false: "停机",
}

// format the CreatedAt date to display nicely on the screen
func (administrator *Administrator) CreatedAtDate() string {
	return administrator.CreatedAt.Format("2006-01-02 15:04:05")
}

// 根据管理员的Valid属性返回其状态文字
func (administrator *Administrator) Status() string {
	return Status[administrator.Valid]
}

// 获取全部管理员对象
func GetAdministrators() (administrators []Administrator, err error) {

	return
}

// 添加一个普通管理员（carbin crew）
func (administrator *Administrator) Create() (err error) {
	statement := "INSERT INTO administrators (uuid, user_id, role, password, created_at, valid) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), administrator.UserId, administrator.Role, Encrypt(administrator.Password), time.Now(), true).Scan(&administrator.Id, &administrator.Uuid)
	return
}
