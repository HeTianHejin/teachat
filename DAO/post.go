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
	EditAt    time.Time
	Attitude  bool
	Score     int
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
	Class     int //0：原始草稿，1:已通过（友邻盲评），2:（友邻盲评）已拒绝
}

var DraftPostStatus = map[int]string{
	0: "草稿",
	1: "接纳",
	2: "婉拒",
}

// IsEdited() returns true if the post has been edited
// 检测edit_at是否晚于created_at一秒以上
func (post *Post) IsEdited() bool {
	return post.CreatedAt.Sub(post.EditAt) >= time.Second
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

// user create a new post to a thread
// 用户初次发表跟帖（对主贴主张进行表态）
func (user *User) CreatePost(thread_id int, attitude bool, body string) (post Post, err error) {
	statement := "INSERT INTO posts (uuid, body, user_id, thread_id, created_at, edit_at, attitude) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, score"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	// use QueryRow to return a row and scan the returned id into the post struct
	err = stmt.QueryRow(Random_UUID(), body, user.Id, thread_id, time.Now(), time.Now(), attitude).Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.Score)
	return
}

// GetPostByUuid() gets a post by the UUID
func GetPostByUuid(uuid string) (post Post, err error) {
	post = Post{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, score FROM posts WHERE uuid = $1", uuid).
		Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.Score)
	return
}

// GetPostbyId() gets a post by the id
func GetPostbyId(id int) (post Post, err error) {
	post = Post{}
	err = Db.QueryRow("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, score FROM posts WHERE id = $1", id).
		Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.Score)
	return
}

// update a post.score
func (post *Post) UpdateScore(score int) (err error) {
	statement := "UPDATE posts SET score = $2 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, score)
	return
}

// get posts to a thread
// 获取某个thread的全部posts,按照 .score DESC 排序
func (t *Thread) Posts() (posts []Post, err error) {
	posts = []Post{}
	rows, err := Db.Query("SELECT id, uuid, body, user_id, thread_id, created_at, edit_at, attitude, score FROM posts WHERE thread_id = $1 ORDER BY score DESC", t.Id)
	if err != nil {
		return
	}
	for rows.Next() {
		post := Post{}
		if err = rows.Scan(&post.Id, &post.Uuid, &post.Body, &post.UserId, &post.ThreadId, &post.CreatedAt, &post.EditAt, &post.Attitude, &post.Score); err != nil {
			return
		}
		posts = append(posts, post)
	}
	rows.Close()
	return
}

// 统计某个thread的全部posts的.score总值，返回int
func (t *Thread) PostsScore() (score int, err error) {
	err = Db.QueryRow("SELECT sum(score) FROM posts WHERE thread_id = $1", t.Id).Scan(&score)
	if err != nil {
		return
	}
	return
}

// 统计某个thread的posts.attitude=true的.score总值，返回int
func (t *Thread) PostsScoreSupport() (score int, err error) {
	//first, check the posts table to see if there are any posts with attitude=true for this thread
	//if there are none, return 0
	//if there are some, sum the score for those posts
	//return the sum
	//check if there are any posts with attitude=true for this thread
	var count int
	err = Db.QueryRow("SELECT count(*) FROM posts WHERE thread_id = $1 AND attitude = true", t.Id).Scan(&count)
	//if there are none, return 0
	if err != nil {
		return 0, err
	}
	err = Db.QueryRow("SELECT sum(score) FROM posts WHERE thread_id = $1 AND attitude = true", t.Id).Scan(&score)
	if err != nil {
		return
	}
	return
}

// 统计某个thread的posts.attitude=false的.score总值，返回int
func (t *Thread) PostsScoreOppose() (score int, err error) {
	//first, check the posts table to see if there are any posts with attitude=false for this thread
	//if there are none, return 0
	//if there are some, sum the score for those posts
	//return the sum
	//check if there are any posts with attitude=false for this thread
	var count int
	err = Db.QueryRow("SELECT count(*) FROM posts WHERE thread_id = $1 AND attitude = false", t.Id).Scan(&count)
	//if there are none, return 0
	if err != nil {
		return 0, err
	}
	err = Db.QueryRow("SELECT sum(score) FROM posts WHERE thread_id = $1 AND attitude = false", t.Id).Scan(&score)
	if err != nil {
		return
	}
	return
}

// get posts to a thread, with pagination

// NumReplies() returns the number of threads where Thread.PostId = Post.Id
func (post *Post) NumReplies() (count int) {
	err := Db.QueryRow("SELECT count(*) FROM threads WHERE post_id = $1", post.Id).Scan(&count)
	if err != nil {
		return
	}
	return
}

// Create() 创建一个新的DraftPost草稿
func (user *User) CreateDraftPost(thread_id int, attitude bool, body string) (post DraftPost, err error) {
	statement := "INSERT INTO draft_posts (user_id, thread_id, body, created_at, attitude, class) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, user_id, thread_id, body, created_at, attitude, class"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(user.Id, thread_id, body, time.Now(), attitude, 0).Scan(&post.Id, &post.UserId, &post.ThreadId, &post.Body, &post.CreatedAt, &post.Attitude, &post.Class)
	return
}

// Get() 读取一个DraftPost品味（跟帖）稿
func GetDraftPost(id int) (post DraftPost, err error) {
	post = DraftPost{}
	err = Db.QueryRow("SELECT id, user_id, thread_id, body, created_at, attitude, class FROM draft_posts WHERE id = $1", id).
		Scan(&post.Id, &post.UserId, &post.ThreadId, &post.Body, &post.CreatedAt, &post.Attitude, &post.Class)
	return
}

// GetDraftPostbyUserId() 读取��个用户的全部DraftPost品��（����）稿
func GetDraftPostbyUserId(user_id int) (posts []DraftPost, err error) {
	rows, err := Db.Query("SELECT id, user_id, thread_id, body, created_at, attitude, class FROM draft_posts WHERE user_id = $1", user_id)
	if err != nil {
		return
	}
	for rows.Next() {
		post := DraftPost{}
		if err = rows.Scan(&post.Id, &post.UserId, &post.ThreadId, &post.Body, &post.CreatedAt, &post.Attitude, &post.Class); err != nil {
			return
		}
		posts = append(posts, post)
	}
	rows.Close()
	return
}

// UpdateDraftPost() 更新一个DraftPost品��（����）稿
func (post *DraftPost) UpdateDraftPost(class int) (err error) {
	statement := "UPDATE draft_posts SET class = $2 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(post.Id, class)
	return
}
