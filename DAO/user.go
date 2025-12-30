package dao

import (
	"context"
	"database/sql"
	"fmt"
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

// 系统预设用户ID常量
const (
	UserId_None              = 0 //
	UserId_Captain_Spaceship = 1 // 太空船长
	UserId_Verifier          = 2 // 见证集团第一团队CEO id
)

// 系统预设“游客”用户
var UserUnknown = User{
	Id:   UserId_None,
	Uuid: "x",
	Name: "游客",
}

// 系统预设用户UUID常量
const (
	UserUUIDCaptain  = "396d7fac-2f29-44a7-7f77-63cbaf423438"
	UserUUIDVerifier = "070a7e98-d5ab-4506-4e59-093a053bc32b"
	UserUUIDUnknown  = "x"
)

var UserRole = map[string]string{
	"captain": "太空船长",

	// 茶博士 ——古时专指陆羽。陆羽著《茶经》，唐德宗李适曾当面称陆羽为“茶博士”。
	// 茶博士 -teaOffice，是古代中华传统文化对茶馆工作人员的昵称，如：富家宴会，犹有专供茶事之人，谓之茶博士。——唐代《西湖志馀》
	// 现在多指精通茶艺的师傅，尤其是四川的长嘴壶茶艺，茶博士个个都是身怀绝技的“高手”。
	"teaoffice":    "茶博士",
	"traveller":    "太空旅客",
	"hijacker":     "劫机者",
	"zebra":        "莽撞者",
	"troublemaker": "捣乱者",
	"UFO":          "外星人",
}

const (
	// 用户角色
	User_Role_Captain      = "captain"      //太空船长
	User_Role_TeaOffice    = "teaoffice"    //茶博士
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
	rows, err := DB.QueryContext(ctx, "SELECT * FROM users WHERE name LIKE $1 Limit $2", "%"+keyword+"%", limit)
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
// var ObjectName = map[int]string{
// 	1:  "objective",
// 	2:  "project",
// 	3:  "thread",
// 	4:  "post",
// 	5:  "team",
// 	6:  "goods",
// 	7:  "magic",
// 	8:  "skill",
// 	9:  "family",
// 	10: "user",
// 	//...
// }

// Create a new user, save user info into the database
func (user *User) Create() (err error) {
	// Postgres does not automatically return the last insert id, because it would be wrong to assume
	// you're always using a sequence.You need to use the RETURNING keyword in your insert to get this
	// information from postgres.

	statement := "INSERT INTO users (uuid, name, email, password, created_at, biography, role, gender, avatar) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, uuid"
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRowContext(ctx, "SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE email = $1", email).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据给出的邮箱地址，通过快速计数法QueryRow检查user是否已经存在注册记录
// 返回: exist - 是否存在, err - 错误信息(包含原始SQL错误和上下文)
func UserExistByEmail(email string) (exist bool, err error) {
	const query = "SELECT count(*) FROM users WHERE email = $1"
	var count int

	// 执行查询并捕获错误
	err = DB.QueryRow(query, email).Scan(&count)
	if err != nil {
		// 包装原始错误，添加更多上下文信息
		return false, fmt.Errorf("查询邮箱存在性失败: %v, 查询: %q, 参数: %s", err, query, email)
	}

	return count > 0, nil
}

// Get a single user given the UUID or id
func GetUserByID(uuid string) (user User, err error) {
	if uuid == "" {
		return user, fmt.Errorf("uuid is empty")
	}
	// 先以uuid查询，如果不存在，再以id查询
	user = User{}
	err = DB.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE uuid = $1", uuid).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			err = DB.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", uuid).
				Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
		} else {
			return user, fmt.Errorf("查询用户失败:参数: %s, %v", uuid, err)
		}
	}
	return
}

// Get the user who created the objective
// 获取创建茶话会的用户，茶围作者（撰写人），目标主理人
func (o *Objective) Admin() (user User, err error) {
	user = User{}
	DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", o.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// get the user who created this project
// 获取创建该茶台（项目）的用户（撰写人），即项目主理人
func (project *Project) Master() (user User, err error) {
	user = User{}
	DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", project.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)

	return
}

// Get the user who created the thread
func (t *Thread) Author() (user User, err error) {
	user = User{}
	DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", t.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get the user who wrote the post
func (post *Post) Author() (user User, err error) {
	user = User{}
	DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", post.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据团队创建人FounderId获取User信息
// AWS CodeWhisperer assist in writing
func (team *Team) Founder() (user User, err error) {
	user = User{}
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team.FounderId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// UserCount（）获取注册用户数
func UserCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM users")
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err = DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err = DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow("SELECT id, uuid, name, email, password, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", id).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// Get ToUser by invitation's invite_email
// 根据邀请函的邮箱获取受邀请人资料
func (invitation *Invitation) ToUser() (user User, err error) {
	user = User{}
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE email = $1", invitation.InviteEmail).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据团队成员的UserId获取User信息
// AWS CodeWhisperer assist in writing
func (team_member *TeamMember) User() (user User, err error) {
	user = User{}
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team_member.UserId).
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
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", team.FounderId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 获取撰写邀请函的茶团时任CEO（原撰写人），可能不是现任CEO
func (invitation *Invitation) Author() (user User, err error) {
	user = User{}
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", invitation.AuthorUserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}

// 根据invite_email查询一个User收到的全部邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationsCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM invitations WHERE invite_email = $1", user.Email)
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

// 据invite_email查询一个User收到的未查看class=0请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationUnviewedCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 0", user.Email)
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
	rows, err := DB.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 1", user.Email)
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
	rows, err := DB.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 2", user.Email)
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

// 据invite_email查询一个User收到的已拒绝邀请class=3邀请函数量
// AWS CodeWhisperer assist in writing
func (user *User) InvitationRejectedCount() (count int) {
	rows, err := DB.Query("SELECT count(*) FROM invitations WHERE invite_email = $1 AND status = 3", user.Email)
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
