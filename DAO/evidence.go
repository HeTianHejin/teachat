package data

import "time"

// 凭据，依据，指音视频等视觉证据，证明手工艺作业符合描述的资料,
// 最好能反映作业劳动成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id             int
	Uuid           string
	Description    string // 描述记录
	RecorderUserId int    // 记录人id
	Note           string //备注,特别说明
	Category       int    //分类：1、图片，2、视频，3、音频，4、其他
	Link           string // 储存链接（地址）
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

// 标记属于那一个手工艺，
type EvidenceHandicraft struct {
	Id           int
	Uuid         string
	HandicraftId int
	EvidenceId   int
	CreatedAt    time.Time
}

// EvidenceHandicraft.Create()
func (e_h *EvidenceHandicraft) Create() (err error) {
	statement := "INSERT INTO evidence_handicrafts (uuid, handicraft_id, evidence_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), e_h.HandicraftId, e_h.EvidenceId, time.Now()).Scan(&e_h.Id, &e_h.Uuid)
	if err != nil {
		return
	}
	return
}

// EvidenceHandicraft.Get()
func (e_h *EvidenceHandicraft) Get() (err error) {
	statement := "SELECT id, uuid, handicraft_id, evidence_id, created_at FROM evidence_handicrafts WHERE id=$1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(e_h.Id).Scan(&e_h.Id, &e_h.Uuid, &e_h.HandicraftId, &e_h.EvidenceId, &e_h.CreatedAt)
	if err != nil {
		return
	}
	return
}
func (e_h *EvidenceHandicraft) GetByUuid() (err error) {
	statement := "SELECT id, uuid, handicraft_id, evidence_id, created_at FROM evidence_handicrafts WHERE uuid=$1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(e_h.Uuid).Scan(&e_h.Id, &e_h.Uuid, &e_h.HandicraftId, &e_h.EvidenceId, &e_h.CreatedAt)
	if err != nil {
		return
	}
	return
}

// 凭据，依据，指音视频等视觉证据，
type SeeSeekEvidence struct {
	Id         int
	Uuid       string
	SeeSeekId  int // 标记属于那一个“看看”，
	EvidenceId int
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

// SeeSeekEvidence.Create()
func (s_s_e *SeeSeekEvidence) Create() (err error) {
	statement := "INSERT INTO see_seek_evidences (uuid, see_seek_id, evidence_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id, uuid"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), s_s_e.SeeSeekId, s_s_e.EvidenceId, time.Now()).Scan(&s_s_e.Id, &s_s_e.Uuid)
	if err != nil {
		return
	}
	return
}

// SeeSeekEvidence.Get()
func (s_s_e *SeeSeekEvidence) Get() (err error) {
	statement := "SELECT id, uuid, see_seek_id, evidence_id, created_at FROM see_seek_evidences WHERE id=$1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s_s_e.Id).Scan(&s_s_e.Id, &s_s_e.Uuid, &s_s_e.SeeSeekId, &s_s_e.EvidenceId, &s_s_e.CreatedAt)
	if err != nil {
		return
	}
	return
}
func (s_s_e *SeeSeekEvidence) GetByUuid() (err error) {
	statement := "SELECT id, uuid, see_seek_id, evidence_id, created_at FROM see_seek_evidences WHERE uuid=$1"
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(s_s_e.Uuid).Scan(&s_s_e.Id, &s_s_e.Uuid, &s_s_e.SeeSeekId, &s_s_e.EvidenceId, &s_s_e.CreatedAt)
	if err != nil {
		return
	}
	return
}
