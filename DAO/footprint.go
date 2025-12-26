package dao

import "time"

// 茶团加盟申请书查看足迹记录
type Footprint struct {
	Id        int
	UserId    int
	TeamId    int
	TeamName  string
	TeamType  int
	Content   string
	ContentId int
	CreatedAt time.Time
}

// Footprint.Create()
func (footprint *Footprint) Create() (err error) {
	statement := `INSERT INTO footprints (user_id, team_id, team_name, team_type, content, content_id, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		footprint.UserId,
		footprint.TeamId,
		footprint.TeamName,
		footprint.TeamType,
		footprint.Content,
		footprint.ContentId,
		time.Now())
	return
}

// Footprint.GetByUserIdAndTeamId()
func (footprint *Footprint) GetByUserIdAndTeamId() (err error) {
	statement := `SELECT * FROM footprints WHERE user_id = $1 AND team_id = $2`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(footprint.UserId, footprint.TeamId).Scan(
		&footprint.Id,
		&footprint.UserId,
		&footprint.TeamId,
		&footprint.TeamName,
		&footprint.TeamType,
		&footprint.Content,
		&footprint.ContentId,
		&footprint.CreatedAt)
	return
}
