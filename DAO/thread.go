package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// 茶议 teaThread
// 一个想法（ithink）或者方案(idea)，或者观点，论题...
// 茶议的开放or封闭性是跟随茶台的class，如果茶台是开放式，则茶议是开放式，否则是封闭式，
type Thread struct {
	Id        int
	Uuid      string
	UserId    int    //作者
	Type      int    //哪一种提法？0: "我觉得",1: "出主意"
	Title     string //标题
	Body      string //内容
	CreatedAt time.Time
	EditAt    *time.Time
	Class     int  //1: "开放式",2: "封闭式"
	ProjectId int  //所属茶台号
	IsPrivate bool // 指定责任（受益）权属类型，属于&家庭管理（family）=true，属于$团队管理（team）=false。默认是false
	FamilyId  int  //作者发帖时选择的成员所属家庭id(family_id)
	TeamId    int  //作者发帖时选择的成员身份所属茶团，$事业团队id。换句话说就是选择那个团队负责？（注意个人身份发言是代表“自由职业者”虚拟茶团）
	PostId    int  //是否针对某一个品味？默认=0，普通类型针对茶台（project）发布；如果有>0值，则是该品味（post）的ID，是议中议类型。
	Category  int  //茶议的分类 0:普通，1:嵌套，2:约茶，3:看看，4:脑火，5:建议，6:宝贝，7:手艺

	//仅用于页面渲染，不保存到数据库
	ActiveData PublicPData
}

const (
	ThreadCategoryNormal      = iota //普通
	ThreadCategoryNested             //内涵
	ThreadCategoryAppointment        //约茶
	ThreadCategorySeeSeek            //看看
	ThreadCategoryBrainFire          //脑火
	ThreadCategorySuggestion         //建议
	ThreadCategoryGoods              //宝贝
	ThreadCategoryHandcraft          //手艺
)

const (
	PostIdForThread = 0
)

const (
	ThreadTypeIthink = iota //我觉得
	ThreadTypeIdea          //出主意（解决方案）
)
const (
	ThreadClassPending = iota //待审查
	ThreadClassOpen           //开放式
	ThreadClassClosed         //封闭式
)

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
	Type      int    //哪一种提法？0: "我觉得",1: "出主意",
	Title     string //标题
	Body      string //内容
	Class     int    //1：开放式茶议，2：封闭式茶议，
	Status    int    //0:草稿，1:接纳，2:婉拒
	IsPrivate bool   // 管理类型：&家庭（family）管理=true，$团队（team）管理=false，默认是false
	TeamId    int    //作者发帖时选择的成员身份所属茶团，$事业团队（team）
	FamilyId  int    //作者发帖时选择的成员所属，&家庭(family)
	ProjectId int    //茶台号
	PostId    int    //是否针对某一个品味？默认=0，普通类型针对茶台（project）发布；如果有>0值，则是该品味（post）的ID，是议中议类型。
	CreatedAt time.Time
	Category  int //茶议的分类
}

const (
	DraftThreadTypeIthink = iota //我觉得
	DraftThreadTypeIdea          //出主意（解决方案）
)
const (
	DraftThreadClassPending = iota //待审查
	DraftThreadClassOpen           //开放式茶议
	DraftThreadClassClosed         //封闭式茶议
)
const (
	DraftThreadStatusPending  = iota //草稿
	DraftThreadStatusAccepted        //已接纳
	DraftThreadStatusRejected        //已婉拒
)

// 根据type属性的int值，返回方便阅读的自然语字符
var ThreadType = map[int]string{
	0: "我觉得",
	1: "出主意",
}

func (t *Thread) TypeString() string {
	return ThreadType[t.Type]
}

//	var ThreadStatus = map[int]string{
//		0: "加水",
//		1: "温热",
//		2: "定味",
//		3: "展示",
//		4: "已删除",
//	}
var DraftThreadStatus = map[int]string{
	0: "草稿",
	1: "接纳",
	2: "婉拒",
}

func (dT *DraftThread) StatusString() string {
	return DraftThreadStatus[dT.Status]
}

// 获取针对此post的全部threads。
func (post *Post) Threads() (threads []Thread, err error) {
	rows, err := db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 ORDER BY created_at DESC", post.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 根据DraftThread struct生成保存新茶议草稿
func (d *DraftThread) Create() (err error) {
	statement := "INSERT INTO draft_threads (user_id, project_id, title, body, class, created_at, type, post_id, team_id, is_private, family_id, category) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(d.UserId, d.ProjectId, d.Title, d.Body, d.Class, time.Now(), d.Type, d.PostId, d.TeamId, d.IsPrivate, d.FamilyId, d.Category).Scan(&d.Id)
	return
}

// 读取茶议草稿
func (d *DraftThread) Get() (err error) {
	err = db.QueryRow("SELECT id, user_id, project_id, title, body, class, created_at, type, post_id, team_id, is_private, family_id, category FROM draft_threads WHERE id = $1", d.Id).
		Scan(&d.Id, &d.UserId, &d.ProjectId, &d.Title, &d.Body, &d.Class, &d.CreatedAt, &d.Type, &d.PostId, &d.TeamId, &d.IsPrivate, &d.FamilyId, &d.Category)
	return
}

// UpdateClass() 更新茶议草稿级
func (d *DraftThread) UpdateStatus(status int) (err error) {
	_, err = db.Exec("UPDATE draft_threads SET status=$1 WHERE id = $2", status, d.Id)
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
	rows, err := db.Query("SELECT count(*) FROM posts where thread_id = $1", t.Id)
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
	rows, err := db.Query("SELECT count(*) FROM posts where thread_id = $1 and attitude = $2", t.Id, true)
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
	rows, err := db.Query("SELECT count(*) FROM posts where thread_id = $1 and attitude = $2", t.Id, false)
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
	rows, err := db.Query("SELECT count(*) FROM reads where thread_id = $1", thread.Id)
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

// update 追加茶议，补充内容，
func (t *Thread) UpdateBodyAndClass(body string, class int, ctx context.Context) error {
	const query = `
        UPDATE threads 
        SET body = $2, 
            class = $3, 
            edit_at = $4 
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query,
		t.Id,
		body,
		class,
		time.Now().UTC()) // 使用UTC时间

	return err
}

// UpdateClass() 根据Thread.Id更新class
func (t *Thread) UpdateClass() (err error) {
	statement := "UPDATE threads SET class = $1, edit_at = $2 WHERE id = $2"
	stmt, err := db.Prepare(statement)
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
	statement := "INSERT INTO threads (uuid, body, user_id, created_at, class, title, project_id, family_id, type, post_id, team_id, is_private, category) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id, uuid, created_at"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), t.Body, t.UserId, time.Now(), t.Class, t.Title, t.ProjectId, t.FamilyId, t.Type, t.PostId, t.TeamId, t.IsPrivate, t.Category).Scan(&t.Id, &t.Uuid, &t.CreatedAt)
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
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(threadApproved.ProjectId, threadApproved.ThreadId, threadApproved.UserId, time.Now()).Scan(&threadApproved.Id)
	return
}

// thread_approved.GetByThreadId()
func (threadApproved *ThreadApproved) GetByThreadId() (err error) {
	err = db.QueryRow("SELECT id, project_id, thread_id, user_id, created_at FROM thread_approved WHERE thread_id = $1", threadApproved.ThreadId).Scan(&threadApproved.Id, &threadApproved.ProjectId, &threadApproved.ThreadId, &threadApproved.UserId, &threadApproved.CreatedAt)
	return
}

// thread_approved.Delete()
func (threadApproved *ThreadApproved) Delete() (err error) {
	statement := "DELETE FROM thread_approved WHERE project_id = $1 AND thread_id = $2"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(threadApproved.ProjectId, threadApproved.ThreadId)
	return
}

// threadApproved.CountByProjectId() 统计茶台采纳的茶议（方案）数量
func (threadApproved *ThreadApproved) CountByProjectId() (count int) {
	err := db.QueryRow("SELECT COUNT(*) FROM thread_approved WHERE project_id = $1", threadApproved.ProjectId).Scan(&count)
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

// 首页展示的必须是class=1或者2状态,返回thread对象数组，前limit个茶议
// 如果点击数相同，则按创建时间从先到后排序
func HotThreads(limit int, ctx context.Context) (threads []Thread, err error) {
	if limit <= 0 {
		err = fmt.Errorf("limit is %d", limit)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE class IN (1,2) ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// Get a thread by the UUID
func GetThreadByUUID(uuid string) (thread Thread, err error) {
	if uuid == "" {
		err = fmt.Errorf("uuid is empty")
		return
	}
	thread = Thread{}
	err = db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE uuid = $1", uuid).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
	return
}

// Get a objective

// Get a thread by the id
func GetThreadById(id int) (thread Thread, err error) {
	thread = Thread{}
	err = db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE id = $1", id).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
	return
}

// 根据Post.ThreadId获取此品味属于哪一个thread
func (post *Post) Thread() (thread Thread, err error) {
	thread = Thread{}
	err = db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE id = $1", post.ThreadId).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
	return
}

// 获取茶台的普通茶议（category=ThreadCategoryNormal）
func (project *Project) ThreadsNormal(ctx context.Context) ([]Thread, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	const query = `
        SELECT 
            id, uuid, body, user_id, created_at, 
            class, title, edit_at, project_id, 
            family_id, type, post_id, team_id, 
            is_private, category
        FROM threads 
        WHERE post_id = $1 
          AND category = $2 
          AND project_id = $3
        ORDER BY created_at DESC`

	// 使用预编译语句提高性能
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("准备查询语句失败: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, PostIdForThread, ThreadCategoryNormal, project.Id)
	if err != nil {
		return nil, fmt.Errorf("查询执行失败: %w", err)
	}
	defer rows.Close() // 确保在任何情况下都会关闭

	var threads []Thread
	for rows.Next() {
		var thread Thread
		err := rows.Scan(
			&thread.Id, &thread.Uuid, &thread.Body,
			&thread.UserId, &thread.CreatedAt, &thread.Class,
			&thread.Title, &thread.EditAt, &thread.ProjectId,
			&thread.FamilyId, &thread.Type, &thread.PostId,
			&thread.TeamId, &thread.IsPrivate, &thread.Category,
		)
		if err != nil {
			return nil, fmt.Errorf("数据扫描失败: %w", err)
		}
		threads = append(threads, thread)
	}

	// 检查遍历过程中是否出错
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("行遍历出错: %w", err)
	}

	return threads, nil
}

// 获取茶台的“约茶”茶议(category=ThreadCategoryAppointment), return (thread Thread, err error)
func (project *Project) ThreadAppointment(ctx context.Context) (thread Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = db.QueryRowContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategoryAppointment, project.Id).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)

	return
}

// 获取茶台的“看看”茶议(category=ThreadCategorySeeSeek), return ([]Thread,error)
func (project *Project) ThreadsSeeSeek(ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategorySeeSeek, project.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

// 获取茶台的“脑火“茶议(category=ThreadCategoryBrainFire), return ([]Thread, error)
func (project *Project) ThreadsBrainFire(ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategoryBrainFire, project.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

// 获取茶台的“建议”茶议(category=ThreadCategorySuggestion), return ([]Thread, error)
func (project *Project) ThreadsSuggestion(ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategorySuggestion, project.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

// 获取茶台的“物资”茶议(category=ThreadCategoryGoods), return ([]Thread, error)
func (project *Project) ThreadsGoods(ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategoryGoods, project.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

// 获取茶台的“手艺”茶议(category=ThreadCategoryHandcraft), return ([]Thread, error)
func (project *Project) ThreadsHandicraft(ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE post_id = $1 AND category = $2 AND project_id = $3", PostIdForThread, ThreadCategoryHandcraft, project.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		err := rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return threads, nil
}

func (t *Thread) IsEdited() bool {

	return t.EditAt != nil && !t.EditAt.Equal(t.CreatedAt)
}

// 填写入围茶台约茶等6部曲
func CreateRequiredThreads(objective *Objective, project *Project, user_id int, ctx context.Context) error {
	// 使用传入的上下文创建超时控制
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 开始事务（使用传入的上下文）
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("事务启动失败: %w", err)
	}

	// 简化事务处理：确保在函数退出时回滚（如果未提交）
	defer tx.Rollback()

	// 明确定义线程模板类型
	type threadTemplate struct {
		Title    string
		Body     string
		Category int
	}

	templates := []threadTemplate{
		{"约茶", "在此记录具体的茶会时间、地址和出席人员", ThreadCategoryAppointment},
		{"看看", "在此记录“看看”作业情况。", ThreadCategorySeeSeek},
		{"脑火", "在此记录“脑火”作业情况。", ThreadCategoryBrainFire},
		{"建议", "在此记录根据「看看」的结果，提出对应建议。", ThreadCategorySuggestion},
		{"宝贝", "在此记录逐一列出需要准备的物资", ThreadCategoryGoods},
		{"手艺", "在此记录“手艺”作业情况。", ThreadCategoryHandcraft},
	}

	for _, template := range templates {
		thread := Thread{
			UserId:    user_id,
			Type:      ThreadTypeIdea,
			Title:     template.Title,
			Body:      template.Body,
			Class:     project.Class,
			ProjectId: project.Id,
			IsPrivate: project.IsPrivate,
			FamilyId:  FamilyIdUnknown,
			TeamId:    TeamIdVerifier,
			PostId:    PostIdForThread,
			Category:  template.Category,
		}

		if err := thread.CreateInTx(tx); err != nil {
			return fmt.Errorf("创建步骤「%s」失败 (项目:%d 用户:%d): %w",
				template.Title, project.Id, user_id, err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("事务提交失败: %w", err)
	}

	return nil
}

// 事务创建约茶作业6部曲
func (t *Thread) CreateInTx(tx *sql.Tx) error {
	query := `
        INSERT INTO threads (uuid, body, user_id, created_at, class, title, project_id, family_id, type, post_id, team_id, is_private, category)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id, uuid, created_at`

	err := tx.QueryRow(query, Random_UUID(), t.Body, t.UserId, time.Now().UTC(), t.Class, t.Title, t.ProjectId, t.FamilyId, t.Type, t.PostId, t.TeamId, t.IsPrivate, t.Category).
		Scan(&t.Id, &t.Uuid, &t.CreatedAt)
	if err != nil {
		return fmt.Errorf("数据库事务创建Thread失败: %w", err)
	}
	return nil
}

// SearchThreadByTitle(keyword) 根据关键字搜索茶议，返回 []Thread，限制limit条数
func SearchThreadByTitle(keyword string, limit int, ctx context.Context) (threads []Thread, err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := db.QueryContext(ctx, "SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, family_id, type, post_id, team_id, is_private, category FROM threads WHERE title LIKE $1 ORDER BY created_at DESC LIMIT $2", "%"+keyword+"%", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		var thread Thread
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.FamilyId, &thread.Type, &thread.PostId, &thread.TeamId, &thread.IsPrivate, &thread.Category); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}
