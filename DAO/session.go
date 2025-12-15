package dao

import (
	"database/sql"
	"errors"
	"fmt"
	util "teachat/Util"
	"time"
)

const sessionDuration = 7 * 24 * time.Hour

// 会话
type Session struct {
	Id        int
	Uuid      string
	Email     string
	UserId    int
	CreatedAt time.Time
	Gender    int
}

// Create a new session for an existing user
func (user *User) CreateSession() (session Session, err error) {
	statement := "INSERT INTO sessions (uuid, email, user_id, created_at, gender) VALUES ($1, $2, $3, $4 ,$5) RETURNING id, uuid, email, user_id, created_at, gender"
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow("SELECT id, uuid, email, user_id, created_at, gender FROM sessions WHERE user_id = $1", user.Id).
		Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.Gender)
	return
}

// 删除用户的session
func (user *User) DeleteSession() (err error) {
	statement := /* sql */ "DELETE FROM sessions WHERE user_id = $1"
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Id)
	return
}

// Check if session is valid in the database
func (session *Session) Check() (bool, error) {
	err := DB.QueryRow(
		"SELECT id, uuid, email, user_id, created_at, gender FROM sessions WHERE uuid = $1",
		session.Uuid,
	).Scan(&session.Id, &session.Uuid, &session.Email, &session.UserId, &session.CreatedAt, &session.Gender)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	err = DB.QueryRow("SELECT id, word FROM watchwords WHERE word = $1 ", watchword).Scan(&watchword_db.Id, &watchword_db.Word)
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
	stmt, err := DB.Prepare(statement)
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
	err = DB.QueryRow("SELECT id, uuid, name, email, created_at, biography, role, gender, avatar, updated_at FROM users WHERE id = $1", session.UserId).
		Scan(&user.Id, &user.Uuid, &user.Name, &user.Email, &user.CreatedAt, &user.Biography, &user.Role, &user.Gender, &user.Avatar, &user.UpdatedAt)
	return
}
