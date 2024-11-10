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
	EditAt    time.Time
	ProjectId int //茶台号
	HitCount  int //点击计数
	Type      int //哪一种提法？0: "我觉得",1: "出主意"
	PostId    int //针对那一个品味？默认0为针对茶台项目
	TeamId    int //作者团队id

	//仅用于页面渲染，不保存到数据库
	PageData PublicPData
}

// struct ThreadGoods 茶议涉及物资
type ThreadGoods struct {
	Id          int
	UserId      int       // 用户ID
	ThreadId    int       // 茶议ID
	GoodsId     int       // 物资ID
	ProjectId   int       // 项目ID
	Type        int       // 物资类型ID 1-装备 2-物品 3-材料
	Number      int       // 数量
	CreatedTime time.Time // 创建时间
	UpdatedTime time.Time // 更新时间
}

// ThreadGoods.Create() 创建1条茶议涉及物资记录
func (tg *ThreadGoods) Create() (err error) {
	statement := "INSERT INTO thread_goods (user_id, thread_id, goods_id, project_id, type, number, created_time, updated_time) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(tg.UserId, tg.ThreadId, tg.GoodsId, tg.ProjectId, tg.Type, tg.Number, time.Now(), time.Now()).Scan(&tg.Id)
	return
}

// ThreadGoods.Update() 更新1条茶议涉及物资记录
func (tg *ThreadGoods) Update() (err error) {
	statement := "UPDATE thread_goods SET user_id = $1, thread_id = $2, goods_id = $3, project_id = $4, type = $5, number = $6, updated_time = $7 WHERE id = $8"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tg.UserId, tg.ThreadId, tg.GoodsId, tg.ProjectId, tg.Type, tg.Number, time.Now(), tg.Id)
	return
}

// ThreadGoods.GetbyThreadId() 获取1条茶议涉及物资记录
func (tg *ThreadGoods) GetbyThreadId() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, goods_id, project_id, type, number, created_time, updated_time FROM thread_goods WHERE thread_id = $1", tg.ThreadId).
		Scan(&tg.Id, &tg.UserId, &tg.ThreadId, &tg.GoodsId, &tg.ProjectId, &tg.Type, &tg.Number, &tg.CreatedTime, &tg.UpdatedTime)
	return
}

// ThreadGoods.GetById() 获取1条茶议涉及物资记录
func (tg *ThreadGoods) GetById() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, goods_id, project_id, type, number, created_time, updated_time FROM thread_goods WHERE id = $1", tg.Id).
		Scan(&tg.Id, &tg.UserId, &tg.ThreadId, &tg.GoodsId, &tg.ProjectId, &tg.Type, &tg.Number, &tg.CreatedTime, &tg.UpdatedTime)
	return
}

// ThreadGoods.Delete() 删除1条茶议涉及物资记录
func (tg *ThreadGoods) Delete() (err error) {
	statement := "DELETE FROM thread_goods WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tg.Id)
	return
}

// ThreadGoods.CountByProjectId() 统计茶台涉及物资总数
func (tg *ThreadGoods) CountByProjectId() (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(id) FROM thread_goods WHERE project_id = $1", tg.ProjectId).Scan(&count)
	return
}

// ThreadGoods.CountByThreadId() 统计茶议涉及物资总数
func (tg *ThreadGoods) CountByThreadId() (count int, err error) {
	err = Db.QueryRow("SELECT COUNT(id) FROM thread_goods WHERE thread_id = $1", tg.ThreadId).Scan(&count)
	return
}

// (thread *Thread) GetThreadGoodsByThreadId() 获取茶议涉及全部物资队列
func (thread *Thread) GetThreadGoodsByThreadId() (threadGoods []ThreadGoods, err error) {
	rows, err := Db.Query("SELECT id, user_id, thread_id, goods_id, project_id, type, number, created_time, updated_time FROM thread_goods WHERE thread_id = $1", thread.Id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tg ThreadGoods
		err = rows.Scan(&tg.Id, &tg.UserId, &tg.ThreadId, &tg.GoodsId, &tg.ProjectId, &tg.Type, &tg.Number, &tg.CreatedTime, &tg.UpdatedTime)
		if err != nil {
			return
		}
		threadGoods = append(threadGoods, tg)
	}
	return
}

// 记录茶议的花费
type ThreadCost struct {
	Id        int
	UserId    int
	ThreadId  int
	Cost      int
	Type      int //0:预算，1:实际
	CreatedAt time.Time
	ProjectId int
}

// ThreadCost.Create() 创建1茶议花费记录
func (tc *ThreadCost) Create() (err error) {
	statement := "INSERT INTO thread_costs (user_id, thread_id, cost, type, created_at, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(tc.UserId, tc.ThreadId, tc.Cost, tc.Type, time.Now(), tc.ProjectId).Scan(&tc.Id)
	return
}

// ThreadCost.UpdateType()
func (tc *ThreadCost) UpdateType() (err error) {
	statement := "UPDATE thread_costs SET type = $1 WHERE id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tc.Type, tc.Id)
	return
}

// ThreadCost.GetbyThreadId()
func (tc *ThreadCost) GetbyThreadId() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, cost, type, created_at, project_id FROM thread_costs WHERE thread_id = $1", tc.ThreadId).
		Scan(&tc.Id, &tc.UserId, &tc.ThreadId, &tc.Cost, &tc.Type, &tc.CreatedAt, &tc.ProjectId)
	return
}

// ThreadCost.GetById()
func (tc *ThreadCost) GetById() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, cost, type, created_at, project_id FROM thread_costs WHERE id = $1", tc.Id).
		Scan(&tc.Id, &tc.UserId, &tc.ThreadId, &tc.Cost, &tc.Type, &tc.CreatedAt, &tc.ProjectId)
	return
}

// (project *Project) CountThreadCostByProjectId() 统计threadCosts表中全部cost值加起来的总值，这是茶台的总花费用值（克茶叶）
func (project *Project) CountThreadCostByProjectId() (count int, err error) {
	err = Db.QueryRow("SELECT COALESCE(SUM(cost),0) FROM thread_costs WHERE type = 1 AND project_id = $1", project.Id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return
}

// (thread *Thread) Cost() 茶语的花费值（克茶叶）
func (thread *Thread) Cost() (cost int, err error) {
	err = Db.QueryRow("SELECT cost FROM thread_costs WHERE thread_id = $1", thread.Id).Scan(&cost)
	if err != nil {
		return 0, err
	}
	return
}

// 记录茶议的耗时
type ThreadTimeSlot struct {
	Id        int
	UserId    int
	ThreadId  int
	TimeSlot  int
	IsConfirm int //0:未确认，1:已确认
	CreatedAt time.Time
	ProjectId int
}

// ThreadTimeSlot.Create() 创建1茶议费时记录
func (tts *ThreadTimeSlot) Create() (err error) {
	statement := "INSERT INTO thread_time_slots (user_id, thread_id, time_slot, is_confirm, created_at, project_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(tts.UserId, tts.ThreadId, tts.TimeSlot, tts.IsConfirm, time.Now(), tts.ProjectId).Scan(&tts.Id)
	return
}

// ThreadTimeSlot.UpdateIsConfirm()
func (tts *ThreadTimeSlot) UpdateIsConfirm() (err error) {
	statement := "UPDATE thread_time_slots SET is_confirm = $1 WHERE id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(tts.IsConfirm, tts.Id)
	return
}

// ThreadTimeSlot.GetbyThreadId()
func (tts *ThreadTimeSlot) GetbyThreadId() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, time_slot, is_confirm, created_at, project_id FROM thread_time_slots WHERE thread_id = $1", tts.ThreadId).
		Scan(&tts.Id, &tts.UserId, &tts.ThreadId, &tts.TimeSlot, &tts.IsConfirm, &tts.CreatedAt, &tts.ProjectId)
	return
}

// ThreadTimeSlot.GetById()
func (tts *ThreadTimeSlot) GetById() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, time_slot, is_confirm, created_at, project_id FROM thread_time_slots WHERE id = $1", tts.Id).
		Scan(&tts.Id, &tts.UserId, &tts.ThreadId, &tts.TimeSlot, &tts.IsConfirm, &tts.CreatedAt, &tts.ProjectId)
	return
}

// (thread *Thread) TimeSlot() ���语的��时值（分钟）
func (thread *Thread) TimeSlot() (timeSlot int, err error) {
	err = Db.QueryRow("SELECT time_slot FROM thread_time_slots WHERE thread_id = $1", thread.Id).Scan(&timeSlot)
	if err != nil {
		return 0, err
	}
	return
}

// (project *Project) CountThreadTimeSlotByProjectId() 统计threadTimeSlots表中time_slot值加起来的总值，这是茶台总耗时（分钟）
func (project *Project) CountThreadTimeSlotByProjectId() (count int, err error) {
	err = Db.QueryRow("SELECT COALESCE(SUM(time_slot),0) FROM thread_time_slots WHERE is_confirm = 1 AND project_id = $1", project.Id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return
}

// 记录敲杯（评论）数
// 记录敲杯（阅读）数
type Read struct {
	Id       int
	UserId   int
	ThreadId int
	ReadAt   time.Time
}

// 茶议草稿，未经邻桌盲评的thread
type DraftThread struct {
	Id        int
	UserId    int    //作者
	ProjectId int    //茶台号
	Title     string //标题
	Body      string //提议？话题？
	Class     int    //分类//0：原始草稿，1:已通过（友邻盲评），2:（友邻盲评）已拒绝
	CreatedAt time.Time
	Type      int //哪一种提法？0: "我觉得",1: "出主意",
	PostId    int //针对那一个品味？默认为 0 是普通茶议
	TeamId    int //作者团队id
	Cost      int //预计花费价值 （克茶叶）
	TimeSlot  int //预计消耗时间（分钟）
}

// 根据type属性的int值，返回方便阅读的自然语字符
var TypeStatus = map[int]string{
	0: "我觉得",
	1: "出主意",
}

var ThreadStatus = map[int]string{
	0: "加水",
	1: "品茶",
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
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE post_id = $1 ORDER BY created_at DESC", post.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 根据DraftThread struct生成保存新茶议草稿
func (d *DraftThread) Create() (err error) {
	statement := "INSERT INTO draft_threads (user_id, project_id, title, body, class, created_at, type, post_id, team_id, cost, time_slot) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(d.UserId, d.ProjectId, d.Title, d.Body, d.Class, time.Now(), d.Type, d.PostId, d.TeamId, d.Cost, d.TimeSlot).Scan(&d.Id)
	return
}

// 读取茶议草稿
func (d *DraftThread) GetById() (err error) {
	err = Db.QueryRow("SELECT id, user_id, project_id, title, body, class, created_at, type, post_id, team_id, cost, time_slot FROM draft_threads WHERE id = $1", d.Id).
		Scan(&d.Id, &d.UserId, &d.ProjectId, &d.Title, &d.Body, &d.Class, &d.CreatedAt, &d.Type, &d.PostId, &d.TeamId, &d.Cost, &d.TimeSlot)
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
// 追加内容之后class=0，需要邻桌盲评，内容是否符合茶棚礼仪公约
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
	statement := "UPDATE threads SET class = $1 WHERE id = $2"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Class, t.Id)
	return
}

// AddHitCount 更新茶议的访问量，运行一次就是hit_count加1
func (t *Thread) AddHitCount() (err error) {
	statement := "UPDATE threads SET hit_count = hit_count + 1 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Id)
	return
}

// Create a new thread
// 保存新的茶议
func (t *Thread) Create() (err error) {
	statement := "INSERT INTO threads (uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(Random_UUID(), t.Body, t.UserId, time.Now(), t.Class, t.Title, time.Now(), t.ProjectId, t.HitCount, t.Type, t.PostId, t.TeamId)
	if err != nil {
		return
	}
	return
}

// 批准，采纳，赞成某个主张/方案/观点
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

// thread_approved.GetByProjectIdAndThreadId()
func (threadApproved *ThreadApproved) GetByProjectIdAndThreadId() (err error) {
	err = Db.QueryRow("SELECT id, project_id, thread_id, user_id, created_at FROM thread_approved WHERE project_id = $1 AND thread_id = $2", threadApproved.ProjectId, threadApproved.ThreadId).Scan(&threadApproved.Id, &threadApproved.ProjectId, &threadApproved.ThreadId, &threadApproved.UserId, &threadApproved.CreatedAt)
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

// 获取一些threads当其等级=0时，这是某个会员新发布的thread，为了稳妥起见，需要随机双盲评估确认内容符合茶棚公约，才能公诸于所有会员，
// 这是AWS CodeWhisperer 协助写的
func ThreadsVisibleToPilot() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE class = 0 ORDER BY created_at DESC")
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 获取一些threads当其等级=1或者2，这是团体成员可表态的threads，
func ThreadsVisibleToTeam(limit int) (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE class = 1 OR class = 2 ORDER BY created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 首页展示的必须是class=1或者2状态,返回thread对象数组，按照点击数thread.hit_count从高到低排序的前limit个茶议
// 如果点击数相同，则按创建时间从先到后排序
func ThreadsIndex(limit int) (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE class = 1 OR class = 2 ORDER BY hit_count DESC, created_at DESC LIMIT $1", limit)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId); err != nil {
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
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE uuid = $1", uuid).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId)
	return
}

// Get a objective

// Get a thread by the id
func GetThreadById(id int) (thread Thread, err error) {
	thread = Thread{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE id = $1", id).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId)
	return
}

// 根据Post.ThreadId获取此品味属于哪一个thread
func (post *Post) Thread() (thread Thread, err error) {
	thread = Thread{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, created_at, class, title, edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE id = $1", post.ThreadId).
		Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId)
	return
}

// 获取茶台的全部茶议
func (project *Project) Threads() (threads []Thread, err error) {
	rows, err := Db.Query("SELECT id, uuid, body, user_id, created_at, class, title ,edit_at, project_id, hit_count, type, post_id, team_id FROM threads WHERE post_id = 0 AND project_id = $1 order by edit_at ASC", project.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		thread := Thread{}
		if err = rows.Scan(&thread.Id, &thread.Uuid, &thread.Body, &thread.UserId, &thread.CreatedAt, &thread.Class, &thread.Title, &thread.EditAt, &thread.ProjectId, &thread.HitCount, &thread.Type, &thread.PostId, &thread.TeamId); err != nil {
			return
		}
		threads = append(threads, thread)
	}
	rows.Close()
	return
}

// 通过比较thread的EditAt和CreatedAt时间，如果前者等于后者，则是没有编辑过，如果前者晚于后者，说明曾经编辑过（补充内容）true，返回 bool
func (t *Thread) IsEdited() bool {
	return t.EditAt.After(t.CreatedAt)
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
