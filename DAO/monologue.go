package data

import "time"

// 独白，个人自由茶语戏剧表演
// 例：李白诗，举杯邀明月，对影成三人。《石头记》林黛玉：……莫若随花飞到天尽头，天尽头，何处有香丘……
type Monologue struct {
	Id        int
	Uuid      string
	Title     string
	Content   string
	UserId    int    // 作者ID
	Note      string //备注
	Category  int    // 种类，类别
	CreatedAt time.Time
}

// Create() (m *Monologue) 创建1独白记录在monologues表中
func (m *Monologue) Create() (monologue Monologue, err error) {
	statement := "INSERT INTO monologues (uuid, title, content, user_id, note, category, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, uuid, title, content, user_id, note, category, created_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), m.Title, m.Content, m.UserId, m.Note, m.Category, time.Now()).Scan(&monologue.Id, &monologue.Uuid, &monologue.Title, &monologue.Content, &monologue.UserId, &monologue.Note, &monologue.Category, &monologue.CreatedAt)
	return
}

// Get() (m *Monologue) 读取1独白记录
func (m *Monologue) Get() (err error) {
	err = Db.QueryRow("SELECT id, uuid, title, content, user_id, note, category, created_at FROM monologues WHERE id = $1", m.Id).
		Scan(&m.Id, &m.Uuid, &m.Title, &m.Content, &m.UserId, &m.Note, &m.Category, &m.CreatedAt)
	return
}

// GetMonologuesbyUserId() 读取某个用户的全部独白记录
func GetMonologuesbyUserID(userId int) (monologues []Monologue, err error) {
	rows, err := Db.Query("SELECT id, uuid, title, content, user_id, note, category, created_at FROM monologues WHERE user_id = $1 ORDER BY created_at DESC", userId)
	if err != nil {
		return
	}
	for rows.Next() {
		monologue := Monologue{}
		if err = rows.Scan(&monologue.Id, &monologue.Uuid, &monologue.Title, &monologue.Content, &monologue.UserId, &monologue.Note, &monologue.Category, &monologue.CreatedAt); err != nil {
			return
		}
		monologues = append(monologues, monologue)
	}
	rows.Close()
	return
}

// UpdateMonologueNoteAndCategory() 更新独白的备注和类别
func (m *Monologue) UpdateMonologueNoteAndCategory() (err error) {
	statement := "UPDATE monologues SET note = $1, category = $2 WHERE id = $3"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Note, m.Category, m.Id)
	return
}
