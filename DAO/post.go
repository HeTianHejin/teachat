package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// teaVote 对茶议主张的表态，回复，立场。肯定或者否定？认同或者反感？支持或者反对？
// 一个茶友对一个茶议仅能表态一次一个立场，
// 如果有更多表达需要，可以通过“内涵”内构功能（自己对自己发起新议程）循环立议再表决表达形式
type Post struct {
	Id        int
	Uuid      string
	Body      string
	UserId    int //作者id
	ThreadId  int
	Attitude  bool //表态：肯定（颔首）或者否定（摇头）
	FamilyId  int  //作者发帖时选择的家庭id
	TeamId    int  //作者发帖时选择的成员所属茶团id（team/family）
	IsPrivate bool //指定责任（受益）权属类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	CreatedAt time.Time
	EditAt    *time.Time

	//发布级别（友邻蒙评已通过）：
	// 0 Regular post (by passerby) 路人发布，
	//1 Official post (by team/family admin) 管理方发布(团队/家庭)，
	//2 Post by spaceship crew 飞船机组团队发布，
	Class int

	//不固化保存（不存入数据库）。
	//根据访客身份动态检测决定，仅返回页面渲染用
	ActiveData PublicPData
}

const (
	PostClassNormal             = iota // Regular post (by passerby) 路人发布
	PostClassAdmin                     // Official post (by team/family admin) 管理方发布(团队/家庭)
	PostClassSpaceShipTeam             // Post by spaceship crew 飞船机组团队发布
	PostClassRejectedByNeighbor        // Post rejected by neighbor review 友邻评审已拒绝
)

// 对表态中的attitude进行转换，true为颔首，false为摇头
func (post *Post) Atti() string {
	if post.Attitude {
		return "颔首"
	} else {
		return "摇头"
	}
}

// draftPost 品味（跟帖）草稿
type DraftPost struct {
	Id        int
	Body      string
	UserId    int
	ThreadId  int
	CreatedAt time.Time
	Attitude  bool //肯定or否定=支持or反对，
	IsPrivate bool //类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	TeamId    int  //作者发帖时选择的成员所属茶团id（team/family）
	FamilyId  int  //作者发帖时选择的家庭id

	//发布级别（友邻蒙评已通过）：
	// 0 Regular post (by passerby) 路人发布，
	//1 Official post (by team/family admin) 管理方发布(团队/家庭)，
	//2 Post by spaceship crew 飞船机组团队发布，
	//3 Post rejected by neighbor review 友邻评审已拒绝
	Class int
}

const (
	DraftPostClassNormal             = iota // Regular post (by passerby) 路人发布
	DraftPostClassAdmin                     // Official post (by team/family admin) 管理方发布(团队/家庭)
	DraftPostClassSpaceShipTeam             // Post by spaceship crew 飞船机组团队发布
	DraftPostClassRejectedByNeighbor        // Post rejected by neighbor review 友邻评审已拒绝
)

var DraftPostStatus = map[int]string{
	0: "草稿",
	1: "接纳",
	2: "婉拒",
}

// IsEdited() returns true if the post has been edited
func (post *Post) IsEdited() bool {
	return post.EditAt != nil && !post.EditAt.Equal(post.CreatedAt)
}

// format the CreatedAt date to display nicely on the screen
func (post *Post) CreatedAtDate() string {
	return post.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// format the EditAt date to display nicely on the screen
func (post *Post) EditAtDate() string {
	return post.EditAt.Format("2006-01-02 15:04:05")
}

// update a post
// 用户补充（追加）其表态内容
func (post *Post) UpdateBody() (err error) {
	statement := "UPDATE posts SET body = $2, edit_at = $3 WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, post.Body, time.Now())
	return
}

// (post *Post) Create() 按照post的struct创建一个post
func (post *Post) Create() (err error) {
	statement := "INSERT INTO posts (uuid, body, user_id, thread_id, created_at, attitude, family_id, team_id, is_private, class) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), post.Body, post.UserId, post.ThreadId, time.Now(), post.Attitude, post.FamilyId, post.TeamId, post.IsPrivate, post.Class).Scan(&post.Id, &post.Uuid)
	return
}

// (post *Post) Get()
func (post *Post) Get() (err error) {
	statement := "SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(post.Id).Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.FamilyId, &post.TeamId, &post.IsPrivate, &post.Class)
	return
}

// (post *Post) GetByUuid() gets a post by the UUID
func (post *Post) GetByUuid() (err error) {
	statement := "SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE uuid = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(post.Uuid).Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.FamilyId, &post.TeamId, &post.IsPrivate, &post.Class)
	return
}

// get posts to a thread
// 获取某个thread的全部普通posts,class = 0,
func (t *Thread) Posts() (posts []Post, err error) {
	posts = []Post{}
	rows, err := db.Query("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE class = $1 AND thread_id = $2 ORDER BY created_at DESC", PostClassNormal, t.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.FamilyId, &post.TeamId, &post.IsPrivate, &post.Class); err != nil {
			return
		}
		posts = append(posts, post)
	}
	rows.Close()
	return
}

// get post to a thread,class = 1,
// post by admin team/family
func (t *Thread) PostsAdmin() (posts []Post, err error) {
	posts = []Post{}
	rows, err := db.Query("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE class = $1 AND thread_id = $2 ORDER BY created_at DESC", PostClassAdmin, t.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.FamilyId, &post.TeamId, &post.IsPrivate, &post.Class); err != nil {
			return
		}
		posts = append(posts, post)
	}
	rows.Close()
	return
}

// NumReplies() returns the number of threads where Thread.PostId = Post.Id
func (post *Post) NumReplies() (count int) {
	err := db.QueryRow("SELECT count(*) FROM threads WHERE post_id = $1", post.Id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// (draft_post *DraftPost) Create() 创建一个新的品味（DraftPost）草稿
func (draft_post *DraftPost) Create() (err error) {
	statement := "INSERT INTO draft_posts (user_id, thread_id, body, created_at, attitude, class, team_id, is_private, family_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(draft_post.UserId, draft_post.ThreadId, draft_post.Body, time.Now(), draft_post.Attitude, draft_post.Class, draft_post.TeamId, draft_post.IsPrivate, draft_post.FamilyId).Scan(&draft_post.Id)
	return
}

// Get() 读取一个DraftPost品味（跟帖=DraftPost）草稿
func (draft_post *DraftPost) Get() (err error) {
	err = db.QueryRow("SELECT id, user_id, thread_id, body, created_at, attitude, class, team_id, is_private, family_id FROM draft_posts WHERE id = $1", draft_post.Id).
		Scan(&draft_post.Id, &draft_post.UserId, &draft_post.ThreadId, &draft_post.Body, &draft_post.CreatedAt, &draft_post.Attitude, &draft_post.Class, &draft_post.TeamId, &draft_post.IsPrivate, &draft_post.FamilyId)
	return
}

// UpdateClass() 更新一个品味（DraftPost）草稿
func (post *DraftPost) UpdateClass(class int) (err error) {
	statement := "UPDATE draft_posts SET class = $2 WHERE id = $1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, class)
	return
}

// HasUserPostedInThread 检查用户是否在指定话题下发表过回复
// 结构体必备参数: userID - 用户ID, threadID - 话题ID
// 返回值: bool - 是否发表过, error - 错误信息
func (p *Post) HasUserPostedInThread(ctx context.Context) (bool, error) {
	// 5秒查询超时则取消
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM posts 
            WHERE thread_id = $1 
            AND user_id = $2
            LIMIT 1
        )`

	var exists bool
	err := db.QueryRowContext(ctx, query, p.ThreadId, p.UserId).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 没有记录属于正常情况，返回false
			return false, nil
		}
		return false, fmt.Errorf("查询用户回复记录失败: %w", err)
	}

	return exists, nil
}
