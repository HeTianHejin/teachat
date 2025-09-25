package data

import "time"

// 凭据，依据，指音视频等视觉证据，证明手工艺作业符合描述的资料,
// 最好能反映作业劳动成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id          int
	Uuid        string
	Description string           // 描述记录
	VerifierId  int              // 记录人id
	Note        string           //备注,特别说明
	Category    EvidenceCategory //分类：1、图片，2、视频，3、音频，4、其他
	Link        string           // 储存链接（地址）
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
type EvidenceCategory int

const (
	Unknown EvidenceCategory = iota //0、初始化默认值
	IMAGE                           //1、图片，
	VIDEO                           //2、视频，
	AUDIO                           //3、音频，
	OTHER                           //4、其他
)

// 凭据，依据，指音视频等视觉证据，
type SeeSeekEvidence struct {
	Id         int
	Uuid       string
	SeeSeekId  int // 缺省值 0，标记属于那一个“看看”，see-seek.go -> SeeSeek{}
	EvidenceId int // 缺省值 0，指向 Evidence{}
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
