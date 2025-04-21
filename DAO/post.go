package data

import "time"

// teaVote 对茶议主张的表态，回复，立场
// 喜欢或者讨厌？认同或者反感？支持或者反对？
// 回复需要风控人吗？慎独原则适用于表态吗？
type Post struct {
	Id        int
	Uuid      string
	Body      string
	UserId    int
	ThreadId  int
	CreatedAt time.Time
	EditAt    *time.Time
	Attitude  bool //表态：肯定（颔首）或者否定（摇头）
	FamilyId  int  //作者发帖时选择的家庭id
	TeamId    int  //作者发帖时选择的成员所属茶团id（team/family）
	IsPrivate bool //类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	Class     int  //发布级别（友邻蒙评已通过）：0、普通发布，1、管理团队（家庭）发布，2、飞行机组团队发布，3、

	//仅页面渲染用
	PageData PublicPData
}

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
	Attitude  bool
	Class     int  //发布级别：0、普通发布，1、管理团队（家庭）发布，2、飞行机组团队发布，3、监管部门发布，00:（友邻蒙评）已拒绝
	TeamId    int  //作者发帖时选择的成员所属茶团id（team/family）
	IsPrivate bool //类型，代表&家庭（family）=true，代表$团队（team）=false。默认是false
	FamilyId  int  //作者发帖时选择的家庭id
}

var DraftPostStatus = map[int]string{
	0: "草稿",
	1: "接纳",
	2: "婉拒",
}

// IsEdited() returns true if the post has been edited
// 检测edit_at是否晚于created_at 5秒以上
func (post *Post) IsEdited() bool {
	return post.EditAt != nil && post.EditAt.After(post.CreatedAt.Add(time.Second*5))
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
func (post *Post) UpdateBody(body string) (err error) {
	statement := "UPDATE posts SET body = $2, edit_at = $3 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, body, time.Now())
	return
}

// (post *Post) Create() 按照post的struct创建一个post
func (post *Post) Create() (err error) {
	statement := "INSERT INTO posts (uuid, body, user_id, thread_id, created_at, attitude, family_id, team_id, is_private, class) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, uuid"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), post.Body, post.UserId, post.ThreadId, time.Now(), post.Attitude, post.FamilyId, post.TeamId, post.IsPrivate, post.Class).Scan(&post.Id, &post.Uuid)
	return
}

// user create a new post to a thread
// 用户初次发表跟帖（对主贴主张进行表态）
// (post *Post) Get()
func (post *Post) Get() (err error) {
	statement := "SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE id = $1"
	stmt, err := Db.Prepare(statement)
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
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(post.Uuid).Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.FamilyId, &post.TeamId, &post.IsPrivate, &post.Class)
	return
}

// get posts to a thread
// 获取某个thread的全部posts,按照  DESC 排序
func (t *Thread) Posts() (posts []Post, err error) {
	posts = []Post{}
	rows, err := Db.Query("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, family_id, team_id, is_private, class FROM posts WHERE thread_id = $1", t.Id)
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
	err := Db.QueryRow("SELECT count(*) FROM threads WHERE post_id = $1", post.Id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// Create() 创建一个新的品味（DraftPost）草稿
func (user *User) CreateDraftPost(thread_id, family_id, team_id int, attitude, is_private bool, body string) (post DraftPost, err error) {
	statement := "INSERT INTO draft_posts (user_id, thread_id, body, created_at, attitude, class, team_id, is_private, family_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(user.Id, thread_id, body, time.Now(), attitude, 0, team_id, is_private, family_id).Scan(&post.Id)
	return
}

// Get() 读取一个DraftPost品味（跟帖=DraftPost）草稿
func (draft_post *DraftPost) Get() (err error) {
	err = Db.QueryRow("SELECT id, user_id, thread_id, body, created_at, attitude, class, team_id, is_private, family_id FROM draft_posts WHERE id = $1", draft_post.Id).
		Scan(&draft_post.Id, &draft_post.UserId, &draft_post.ThreadId, &draft_post.Body, &draft_post.CreatedAt, &draft_post.Attitude, &draft_post.Class, &draft_post.TeamId, &draft_post.IsPrivate, &draft_post.FamilyId)
	return
}

// UpdateClass() 更新一个品味（DraftPost）草稿
func (post *DraftPost) UpdateClass(class int) (err error) {
	statement := "UPDATE draft_posts SET class = $2 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, class)
	return
}
