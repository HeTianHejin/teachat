package data

import (
	"time"
)

// 茶议 teaThread
// 议程，主张或者方案，或者观点，论题...
// 茶议的开放性是跟随茶台的class，如果茶台是开放式，则茶议是开放式，否则是封闭式，
type Thread struct {
	Id        int
	Uuid      string
	Body      string //内容
	UserId    int    //作者
	CreatedAt time.Time
	Class     int    //状态0: "加水",1: "品茶",2: "定味",3: "展示",4: "已删除",
	Title     string //标题
	EditAt    *time.Time
	ProjectId int  //茶台号
	FamilyId  int  //作者发帖时选择的成员所属家庭id(family_id)
	Type      int  //哪一种提法？0: "我觉得",1: "出主意"
	PostId    int  //针对那一个品味？默认0为针对茶台项目
	TeamId    int  //作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说茶话？（注意个人身份发言是代表“自由人”茶团）
	IsPrivate bool // 代表类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false

	//仅用于页面渲染，不保存到数据库
	PageData PublicPData
}

// 记录敲杯（阅读）数
type Read struct {
	Id       int
	UserId   int
	ThreadId int
	ReadAt   time.Time
}

// 茶议草稿，未经邻桌蒙评的thread
type DraftThread struct {
	Id        int
	UserId    int    //作者
	ProjectId int    //茶台号
	Title     string //标题
	Body      string //提议？话题？
	Class     int    //分类//0：原始草稿，1:已通过（友邻蒙评），2:（友邻蒙评）已拒绝
	CreatedAt time.Time
	Type      int  //哪一种提法？0: "我觉得",1: "出主意",
	PostId    int  //针对那一个品味？默认为 0 是普通茶议
	TeamId    int  //作者发帖时选择的成员身份所属茶团，$事业团队id或者&family家庭id。换句话说就是代表那个团队或者家庭说话？
	IsPrivate bool // 代表类型，代表&家庭（family）=true，代表$团队（team）=false。默认是true
	FamilyId  int  //作者发帖时选择的成员所属家庭id(family_id)

}

// 根据type属性的int值，返回方便阅读的自然语字符
var TypeStatus = map[int]string{
	0: "我觉得",
	1: "出主意",
}

var ThreadStatus = map[int]string{
	0: "加水",
	1: "温热",
	2: "定味",
	3: "展示",
	4: "已删除",
}
var DraftThreadStatus = map[int]string{
	0: "草稿",
	1: "接纳",
	2: "退回",
}

// 获取针对此post的全部threads。
func (post *Post) Threads() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE post_id = $1 ORDER BY created_at DESC", post.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 根据DraftThread struct生成保存新茶议草稿
func (d *DraftThread) Create() (err error) {
	statement := "INSERT INTO draft_threads (user_id, project_id, title, body, class, created_at, type, post_id, team_id, is_private, family_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(d.UserId, d.ProjectId, d.Title, d.Body, d.Class, time.Now(), d.Type, d.PostId, d.TeamId, d.IsPrivate, d.FamilyId).Scan(&d.Id)
	return
}

// 读取茶议草稿
func (d *DraftThread) Get() (err error) {
	err = Db.QueryRow("SELECT id, user_id, project_id, title, body, class, created_at, type, post_id, team_id, is_private, family_id FROM draft_threads WHERE id = $1", d.Id).
		Scan(&d.Id, &d.UserId, &d.ProjectId, &d.Title, &d.Body, &d.Class, &d.CreatedAt, &d.Type, &d.PostId, &d.TeamId, &d.IsPrivate, &d.FamilyId)
	return
}

// UpdateClass() 更新茶议草稿级
func (d *DraftThread) UpdateClass(class int) (err error) {
	_, err = Db.Exec("UPDATE draft_threads SET class=$1 WHERE id = $2", class, d.Id)
	return
}

// format the CreatedAt date to display nicely on the screen
func (t *Thread) CreatedAtDate() string {
	return t.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// format the EditAt date to display nicely on the screen
func (t *Thread) EditAtDate() string {
	return t.EditAt.Format("2006-01-02 15:04:05")
}

// get the number of posts in a thread
func (t *Thread) NumReplies() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM posts where thread_id = $1", t.Id)
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

// 统计某个thread的全部posts属性Attitude=true的数量
func (t *Thread) NumSupport() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM posts where thread_id = $1 and attitude = $2", t.Id, true)
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

// 统计某个thread的全部posts属性Attitude=false的数量
func (t *Thread) NumOppose() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM posts where thread_id = $1 and attitude = $2", t.Id, false)
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

// get the number of reads in a thread
func (thread *Thread) NumReads() (count int) {
	rows, err := Db.Query("SELECT count(*) FROM reads where thread_id = $1", thread.Id)
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

// IsAuthor returns true if the user who created the thread is the same as the user passed in
func (t *Thread) IsAuthor(u User) bool {
	return t.UserId == u.Id
}

// update 追加茶议，补充主张内容，不能修改标题，
// 追加内容之后class=0，需要邻桌蒙评，内容是否符合茶棚礼仪公约
func (t *Thread) UpdateTopicAndClass(body string, class int) (err error) {
	statement := "UPDATE threads SET topic = $2, class = $3, edit_at = $4 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Id, body, class, time.Now())
	return
}

// UpdateClass() 根据Thread.Id更新class
func (t *Thread) UpdateClass() (err error) {
	statement := "UPDATE threads SET class = $1, edit_at = $2 WHERE id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Class, time.Now(), t.Id)
	return
}

// AddHitCount 更新茶议的访问量，运行一次就是hit_count加1

// Create a new thread
// 保存新的茶议
func (t *Thread) Create() (err error) {
	statement := "INSERT INTO threads (uuid, body, user_id, created_at, class, title, project_id, family_id, type, post_id, team_id, is_private) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), t.Body, t.UserId, time.Now(), t.Class, t.Title, t.ProjectId, t.FamilyId, t.Type, t.PostId, t.TeamId, t.IsPrivate).Scan(&t.Id, &t.Uuid)
	if err != nil {
		return
	}
	return
}

// 批准，采纳，赞成某个茶议的主张/方案/观点
type ThreadApproved struct {
	Id        int
	ProjectId int       //项目茶台id
	ThreadId  int       //茶议id
	UserId    int       //采纳(批准)者id
	CreatedAt time.Time //采纳时间
}

// thread_approved.Create()
func (threadApproved *ThreadApproved) Create() (err error) {
	statement := "INSERT INTO thread_approved (project_id, thread_id, user_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(threadApproved.ProjectId, threadApproved.ThreadId, threadApproved.UserId, time.Now()).Scan(&threadApproved.Id)
	return
}

// thread_approved.GetByThreadId()
func (threadApproved *ThreadApproved) GetByThreadId() (err error) {
	err = Db.QueryRow("SELECT id, project_id, thread_id, user_id, created_at FROM thread_approved WHERE thread_id = $1", threadApproved.ThreadId).Scan(&threadApproved.Id, &threadApproved.ProjectId, &threadApproved.ThreadId, &threadApproved.UserId, &threadApproved.CreatedAt)
	return
}

// thread_approved.Delete()
func (threadApproved *ThreadApproved) Delete() (err error) {
	statement := "DELETE FROM thread_approved WHERE project_id = $1 AND thread_id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(threadApproved.ProjectId, threadApproved.ThreadId)
	return
}

// threadApproved.CountByProjectId() 统计茶台采纳的茶议（方案）数量
func (threadApproved *ThreadApproved) CountByProjectId() (count int) {
	err := Db.QueryRow("SELECT COUNT(*) FROM thread_approved WHERE project_id = $1", threadApproved.ProjectId).Scan(&count)
	if err != nil {
		return
	}
	return
}

// thread.IsApproved() 主张方案（主意）是否已被台主采纳
func (thread *Thread) IsApproved() bool {
	threadApproved := ThreadApproved{ThreadId: thread.Id}
	err := threadApproved.GetByThreadId()
	return err == nil
}

// 获取一些threads当其等级=0时，这是某个会员新发布的thread，为了稳妥起见，需要随机双蒙评估确认内容符合茶棚公约，才能公诸于所有会员，
// 这是AWS CodeWhisperer 协助写的
func ThreadsVisibleToPilot() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE class = 0 ORDER BY created_at DESC")
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 获取一些threads当其等级=1或者2，这是团体成员可表态的threads，
func ThreadsVisibleToTeam(limit int) (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE class = 1 OR class = 2 ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 首页展示的必须是class=1或者2状态,返回thread对象数组，前limit个茶议
// 如果点击数相同，则按创建时间从先到后排序
func HotThreads(limit int) (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE class IN (1,2) LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// Get a thread by the UUID
func ThreadByUUID(uuid string) (thread Thread, err error) {
	thread = Thread{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE uuid = $1", uuid).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate)
	return
}

// Get a objective

// Get a thread by the id
func GetThreadById(id int) (thread Thread, err error) {
	thread = Thread{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE id = $1", id).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate)
	return
}

// 根据Post.ThreadId获取此品味属于哪一个thread
func (post *Post) Thread() (thread Thread, err error) {
	thread = Thread{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE id = $1", post.ThreadId).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate)
	return
}

// 获取茶台的全部茶议
func (project *Project) Threads() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title ,edit_at, project_id, family_id, type, post_id, team_id, is_private FROM threads WHERE post_id = 0 AND project_id = $1 order by edit_at ASC", project.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

func (t *Thread) IsEdited() bool {

	return t.EditAt != nil && !t.EditAt.Equal(t.CreatedAt)
}

// 获取thread的状态string
func (t *Thread) Status() string {
	return ThreadStatus[t.Class]
}

// 获取draftThread的状态string
func (d *DraftThread) Status() string {
	return DraftThreadStatus[d.Class]
}

// 获取thread的type的状态string
func (t *Thread) TypeStatus() string {
	return TypeStatus[t.Type]
}
