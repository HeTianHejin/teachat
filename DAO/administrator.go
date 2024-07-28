package data

import "time"

type Watchword struct {
	Id              int
	Word            string
	AdministratorId int
	CreatedAt       time.Time
}

type Administrator struct {
	Id            int
	Uuid          string
	UserId        int
	Role          string
	Password      string
	CreatedAt     time.Time
	Valid         bool
	InvalidReason string
	Invalid_at    time.Time
}

var AdminRole = map[string]string{
	"waiter":  "服务员",
	"pilot":   "飞行员",
	"captain": "船长",
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
	rows, err := Db.Query("SELECT id, uuid, user_id, role, password, created_at, valid, invalid_reason, invalid_at FROM administrators")
	if err != nil {
		return
	}
	for rows.Next() {
		administrator := Administrator{}
		if err = rows.Scan(&administrator.Id, &administrator.Uuid, &administrator.UserId, &administrator.Role, &administrator.Password, &administrator.CreatedAt, &administrator.Valid, &administrator.InvalidReason, &administrator.Invalid_at); err != nil {
			return
		}
		administrators = append(administrators, administrator)
	}
	rows.Close()
	return
}

// 添加一个普通管理员（Pilot）
func (administrator *Administrator) Create() (err error) {
	statement := "INSERT INTO administrators (uuid, user_id, role, password, created_at, valid, invalid_reason, invalid_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, uuid, created_at, invalid_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(CreateUuid(), administrator.UserId, administrator.Role, Encrypt(administrator.Password), time.Now(), true, administrator.InvalidReason, time.Now()).Scan(&administrator.Id, &administrator.Uuid, &administrator.CreatedAt, &administrator.Invalid_at)
	return
}

// 根据一个userId获取一个administrator对象
func (user *User) Administrator() (administrator Administrator, err error) {
	administrator = Administrator{}
	err = Db.QueryRow("SELECT id, uuid, user_id, role, password, created_at, valid, invalid_reason, invalid_at FROM administrators WHERE user_id = $1", user.Id).
		Scan(&administrator.Id, &administrator.Uuid, &administrator.UserId, &administrator.Role, &administrator.Password, &administrator.CreatedAt, &administrator.Valid, &administrator.InvalidReason, &administrator.Invalid_at)
	return
}
