package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	util "teachat/Util"
	"time"
)

// 已激活账号用户
type User struct {
	Id        int
	Uuid      string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	Biography string
	Role      string
	Gender    int // 0: "女",1: "男",
	Avatar    string
	UpdatedAt *time.Time

	//Footprint 浏览页面足迹，不保存到数据库，
	//用于临时记录点击‘登录’按钮时页面，以便登船成功后返回同一页面，
	Footprint string
	Query     string //查询参数
}

const (
	UserId_None             = 0
	UserId_SpaceshipCaptain = 1
	UserId_Verifier         = 67 // 见证团队代表人id
)
const sessionDuration = 7 * 24 * time.Hour

var UserRole = map[string]string{
	"traveller":    "太空旅客",
	"hijacker":     "劫机者",
	"zebra":        "莽撞者",
	"troublemaker": "捣乱者",
	"UFO":          "外星人",
}

const (
	// 用户角色
	User_Role_Traveller    = "traveller"    //太空普通旅客
	User_Role_Hijacker     = "hijacker"     //劫机者
	User_Role_Zebra        = "zebra"        //莽撞者
	User_Role_Troublemaker = "troublemaker" //捣乱者
	User_Role_UFO          = "UFO"          //外星人
	// 用户性别
	User_Gender_Female = 0 // 女
	User_Gender_Male   = 1 // 男
)

// SearchUserByNameKeyword() 根据给出的关键词（keyword）,从users.name模糊查询用户，WHERE column LIKE 'keyword%',返回[]User,err
// limit int 表示查询结果数量，5秒超时取消
func SearchUserByNameKeyword(keyword string, limit int, ctx context.Context) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := Db.QueryContext(ctx, "SELECT * FROM users WHERE name LIKE $1 Limit $2", "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// 未激活账号用户
type UserUnactivated struct {
	Id        int
	Uuid      string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	Biography string
	Role      string
	Gender    int
	Avatar    string
}

// 朋友
type Friend struct {
	Id                int
	Uuid              string
	UserId            int
	FriendUserId      int    //朋友的用户号
	Nickname          string //绰号，备注名
	Note              string //备注事项
	RelationshipLevel int    //熟悉程度，0：刚见面的，1:见过几面，3:了解一些背景，4：比较熟，5，非常熟识，6：无所不谈，7:志同道合，8:
	IsRival           bool   //是否竞争对手
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// fans 粉丝
type Fan struct {
	Id                int
	Uuid              string
	UserId            int
	FanUserId         int
	Nickname          string //绰号，备注名
	Note              string //备注事项
	RelationshipLevel int    //熟悉程度，0：刚见面的，1:见过几面，3:了解一些背景，4：比较熟，5，非常熟识，6：无所不谈，7:志同道合，8:
	IsBlackSlice      bool   //是否黑名单
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// follow 关注的
type Follow struct {
	Id                int
	Uuid              string
	UserId            int
	FollowUserId      int
	Nickname          string //绰号，备注名
	Note              string //备注事项
	RelationshipLevel int    //熟悉程度，0：刚见面的，1:见过几面，3:了解一些背景，4：比较熟，5，非常熟识，6：无所不谈，7:志同道合，8:
	IsDisdain         bool   //是否鄙视，蔑视
	CreatedAt         time.Time
	UpdatedAt         *time.Time
}

// 会话
type Session struct {
	Id        int
	Uuid      string
	Email     string
	UserId    int
	CreatedAt time.Time
	Gender    int
}

// 用户的星标本（收藏夹），收藏的茶议=3或者茶话会=1/茶台=2/茶团=5，甚至是品味post=4
// 宝贝=6，魔法=7，宝物=8，
type UserStar struct {
	Id        int
	Uuid      string
	UserId    int
	Type      int
	ObjectId  int
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// 根据UserStar.Type int反射object对象名称string,
// 茶议=3或者茶话会=1,茶台=2,茶团=5，品味post=4， 好东西=6，魔法=7，宝物=8，
// 未知=9=？
var ObjectName = map[int]string{
	1: "objective",
	2: "project",
	3: "thread",
	4: "post",
	5: "team",
	6: "goody",
	7: "magic",
	8: "treasure",
	9: "?",
}

// Create a new session for an existing user
func (user *User) CreateSession() (session Session, err error) {
	statement := "INSERT INTO sessions (uuid, email, user_id, created_at, gender) VALUES ($1, $2, $3, $4 ,$5) RETURNING id, uuid, email, user_id, created_at, gender"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	// use QueryRow to return a row and scan the returned id into the Session struct
	err = stmt.QueryRow(Random_UUID(), user.Email, user.Id, time.Now(), user.Gender).Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.Gender)
	return
}

// Get the session for an existing user
func (user *User) Session() (session Session, err error) {
	session = Session{}
	err = Db.QueryRow("SELECT id, uuid, email, user_id, created_at, gender FROM sessions WHERE user_id = $1", user.Id).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.Gender)
	return
}

// 删除用户的session
func (user *User) DeleteSession() (err error) {
	statement := /* sql */ "DELETE FROM sessions WHERE user_id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Id)
	return
}

// Check if session is valid in the database
func (session *Session) Check() (bool, error) {
	err := Db.QueryRow(
		"SELECT id, uuid, email, user_id, created_at, gender FROM sessions WHERE uuid = $1",
		session.Uuid,
	).Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.Gender)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 会话不存在不算错误
		}
		return false, fmt.Errorf("database query failed: %w", err)
	}

	if session.Id == 0 {
		return false, nil
	}

	expiryTime := session.CreatedAt.Add(sessionDuration)
	if time.Now().Before(expiryTime) {
		return true, nil
	}

	// 会话过期
	if err := session.Delete(); err != nil {
		util.Debug("failed to delete expired session: %v", err)
	}
	return false, nil
}

// 检查登录口令是否正确
func CheckWatchword(watchword string) (valid bool, err error) {
	watchword_db := Watchword{}
	err = Db.QueryRow("SELECT id, word FROM watchwords WHERE word = $1 ", watchword).Scan(&watchword_db.Id, &watchword_db.Word)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			valid = false
			return valid, err
		} else {
			valid = false
			return
		}

	}
	if watchword_db.Word == watchword {
		valid = true
		return
	} else {
		valid = false
		return
	}

}

// Delete session from database
func (session *Session) Delete() (err error) {
	statement := "delete from sessions where uuid = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(session.Uuid)
	return
}

// Get the user from the session
func (session *Session) User() (user User, err error) {
	//user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", session.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// get the number of threads  by a user
func (user *User) NumThreads() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM threads where user_id = $1", user.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// Create a new user, save user info into the database
func (user *User) Create() (err error) {
	// Postgres does not automatically return the last insert id, because it would be wrong to assume
	// you're always using a sequence.You need to use the RETURNING keyword in your insert to get this
	// information from postgres.

	statement := "INSERT INTO users (uuid, name, email, password, created_at, biography, role, gender, avatar) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	// use QueryRow to return a row and scan the returned id into the User struct
	err = stmt.QueryRow(Random_UUID(), user.Name, user.Email, Encrypt(user.Password), time.Now(), user.Biography, user.Role, user.Gender, user.Avatar).Scan(&user.Id, &user.Uuid)
	return
}

// UpdateUserNameAndBiography user information in the database
func UpdateUserNameAndBiography(user_id int, user_name string, user_biography string) (err error) {
	statement := "UPDATE users SET name = $2, biography = $3, updated_at = $4 where id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user_id, user_name, user_biography, time.Now())
	return
}

// 修改数据库中用户身份（角色）
func (user *User) UpdateRole() (err error) {
	statement := "UPDATE users SET role = $2, updated_at = $3 where id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Id, user.Role, time.Now())
	return
}

// Update user avatar in the database
func (user *User) UpdateAvatar() (err error) {
	statement := "UPDATE users SET avatar = $2, updated_at = $3 where id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Id, user.Avatar, time.Now())
	return
}

// Get a single user given the email，limit - 限制查询结果数量,5秒超时就取消
func GetUserByEmail(email string, ctx context.Context) (user User, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	user = User{}
	err = Db.QueryRowContext(ctx, "SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE email = $1", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据给出的邮箱地址，通过快速计数法QueryRow检查user是否已经存在注册记录
// 返回: exist - 是否存在, err - 错误信息(包含原始SQL错误和上下文)
func UserExistByEmail(email string) (exist bool, err error) {
	const query = "SELECT count(*) FROM users WHERE email = $1"
	var count int

	// 执行查询并捕获错误
	err = Db.QueryRow(query, email).Scan(&count)
	if err != nil {
		// 包装原始错误，添加更多上下文信息
		return false, fmt.Errorf("查询邮箱存在性失败: %v, 查询: %q, 参数: %s", err, query, email)
	}

	return count > 0, nil
}

// Get a single user given the ID
func UserById(id int) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get a single user given the UUID
func GetUserByUUID(uuid string) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE uuid = $1", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据管理员的UserId查询其user对象
func (administrator *Administrator) User() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", administrator.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)

	return
}

// Get the user who created the objective
// 获取创建茶话会的用户，即围主代表
func (o *Objective) User() (user User, err error) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", o.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get the user who created the thread
func (t *Thread) User() (user User, err error) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", t.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get the user who wrote the post
func (post *Post) User() (user User, err error) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", post.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// get the user who created this project
// 获取创建该茶台（项目）的用户，即台主代表
func (project *Project) User() (user User, err error) {
	user = User{}
	Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", project.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)

	return
}

// 根据团队创建人FounderId获取User信息
// AWS CodeWhisperer assist in writing
func (team *Team) Founder() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team.FounderId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// UserCount（）获取注册用户数
func UserCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM users")
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// SessionCount() 获取session数
func SessionCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM sessions")
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// Get2GenderRandomSUserIdExceptId() 获取两个性别（一男一女），已登陆的随机userId，排除指定的user_id
func Get2GenderRandomSUserIdExceptId(id int) (user_ids []int, err error) {
	user_ids = []int{}

	// 先随机获取一个在线女士id
	statement := "SELECT user_id FROM sessions WHERE id != $1 AND gender = 0 ORDER BY RANDOM()"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var lady_id int
	if err = stmt.QueryRow(id).Scan(&lady_id); err != nil {
		return
	}
	user_ids = append(user_ids, lady_id)

	// 再随机获取一个在线男士id
	statement = "SELECT user_id FROM sessions WHERE id != $1 AND gender = 1 ORDER BY RANDOM()"
	stmt, err = Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var gentleman_id int
	if err = stmt.QueryRow(id).Scan(&gentleman_id); err != nil {
		return
	}
	user_ids = append(user_ids, gentleman_id)

	return user_ids, nil
}

// Get2GenderRandomUserIdExceptId() 获取两个性别（一男一女），随机userId，排除指定的user_id
func Get2GenderRandomUserIdExceptId(id int) (user_ids []int, err error) {
	user_ids = []int{}

	// 先随机获取一个女士id
	statement := "SELECT id FROM users WHERE id != $1 AND gender = 0 ORDER BY RANDOM()"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var lady_id int
	if err = stmt.QueryRow(id).Scan(&lady_id); err != nil {
		return
	}
	user_ids = append(user_ids, lady_id)

	// 再随机获取一个男士id
	statement = "SELECT user_id FROM users WHERE id != $1 AND gender = 1 ORDER BY RANDOM()"
	stmt, err = Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	var gentleman_id int
	if err = stmt.QueryRow(id).Scan(&gentleman_id); err != nil {
		return
	}
	user_ids = append(user_ids, gentleman_id)

	return user_ids, nil
}

// Get2RandomSUserIdExceptId() 获取两个已登陆随机userId，排除指定的user_id
func Get2RandomSUserIdExceptId(id int) (user_ids []int, err error) {
	user_ids = []int{}
	statement := "SELECT user_id FROM sessions WHERE id != $1 ORDER BY RANDOM() LIMIT 2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(id)
	if err != nil {
		return
	}
	for rows.Next() {
		var user_id int
		if err = rows.Scan(&user_id); err != nil {
			return
		}
		user_ids = append(user_ids, user_id)
	}
	return
}

// Get2RandomUserExceptId() 随机俩位用户,排除指定的id
func Get2RandomUserExceptId(id int) (users []User, err error) {
	users = []User{}
	statement := "SELECT user_id FROM users WHERE id != $1 ORDER BY RANDOM() LIMIT 2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(id)
	if err != nil {
		return
	}
	for rows.Next() {
		var user_id int
		if err = rows.Scan(&user_id); err != nil {
			return
		}
		var user_online User
		user_online, err = GetUser(user_id)
		if err != nil {
			return
		}
		users = append(users, user_online)
	}
	return

}

// Get2RandomUserId() 获取两个随机userId
func Get2RandomUserId() (user_ids []int, err error) {
	user_ids = []int{}
	statement := "SELECT id FROM users ORDER BY RANDOM() LIMIT 2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		return
	}
	for rows.Next() {
		var user_id int
		if err = rows.Scan(&user_id); err != nil {
			return
		}
		user_ids = append(user_ids, user_id)
	}
	return
}

// GetUser() Get a single user given the id
func GetUser(id int) (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get ToUser by invitation's invite_email
// 根据邀请函的邮箱获取受邀请人资料
func (invitation *Invitation) ToUser() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE email = $1", invitation.InviteEmail).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据团队成员的UserId获取User信息
// AWS CodeWhisperer assist in writing
func (team_member *TeamMember) User() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team_member.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 获取邀请函的茶团创建人
func (invitation *Invitation) TeamFounder() (user User, err error) {
	user = User{}
	team, err := GetTeam(invitation.TeamId)
	if err != nil {
		return
	}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team.FounderId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 获取撰写邀请函的茶团时任CEO（原撰写人），可能不是现任CEO
func (invitation *Invitation) AuthorCEO() (user User, err error) {
	user = User{}
	err = Db.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", invitation.AuthorUserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据invite_email查询一个User收到的全部邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationsCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM invitations WHERE invite_email = $1", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// ��据invite_email查询一个User收到的未查看class=0请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationUnviewedCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 0", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 根据invite_email查询一个User收到的已查看class=1邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationViewedCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 1", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// 根据invite_email查询一个User收到的已接受邀请class=2邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationAcceptedCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 2", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// ��据invite_email查询一个User收到的已拒绝邀请class=3邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationRejectedCount() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 3", user.Email)
	if err != nil {
		return
	}
	for rows.Next() {
		if err = rows.Scan(&count); err != nil {
			return
		}
	}
	rows.Close()
	return
}

// isUserInAnyTeam() 检查用户是否在特定一团队中
func isUserInAnyTeam(user_id int, team_ids []int) (bool, error) {
	for _, team_id := range team_ids {
		members, err := GetAllMemberUserIdsByTeamId(team_id)
		if err != nil {
			return false, fmt.Errorf("获取团队 #%d 全部成员失败: %v", team_id, err)
		}
		if contains(members, user_id) {
			return true, nil
		}
	}
	return false, nil
}

// isUserInAnyFamily() 检查用户是否在特定一家庭中
func isUserInAnyFamily(user_id int, family_ids []int) (bool, error) {
	for _, family_id := range family_ids {
		members, err := GetAllMembersUserIdsByFamilyId(family_id)
		if err != nil {
			return false, fmt.Errorf("获取家庭%d成员失败: %v", family_id, err)
		}
		if contains(members, user_id) {
			return true, nil
		}
	}
	return false, nil
}
