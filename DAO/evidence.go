package dao

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// 凭据，依据，指音视频等视觉证据，
// 证明活动作业符合描述的资料,
// 最好能反映作业劳动成就。或者人力消耗、工具的折旧情况。
type Evidence struct {
	Id             int
	Uuid           string
	Description    string           // 描述记录
	RecorderUserId int              // 记录人id
	Note           string           //备注,特别说明
	Category       EvidenceCategory //分类：1、图片，2、视频，3、音频，4、其他
	Path           string           // 储存路径
	OriginalURL    string           // 原始URL
	FileName       string           // 文件名
	MimeType       string           // MIME类型
	FileSize       int64            // 文件大小，单位：字节
	FileHash       string           // 文件哈希值，用于防止重复上传
	Width          int              // 图片/视频宽度，单位：像素
	Height         int              // 图片/视频高度，单位：像素
	Duration       int              // 视频/音频时长，单位：秒
	Visibility     int              //可见性：0、公开，1、私有，仅当事家庭/团队可见
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time //软删除

}

const (
	// 可见性常量
	VisibilityPublic  = 0 // 公开
	VisibilityPrivate = 1 // 私有

	// 文件大小限制
	MaxFileSize = 10 * 1024 * 1024 // 10MB
)

type EvidenceCategory int

const (
	Unknown EvidenceCategory = iota //0、初始化默认值
	IMAGE                           //1、图片，
	VIDEO                           //2、视频，
	AUDIO                           //3、音频，
	OTHER                           //4、其他
)

// String 返回分类的字符串表示
func (ec EvidenceCategory) String() string {
	switch ec {
	case IMAGE:
		return "图片"
	case VIDEO:
		return "视频"
	case AUDIO:
		return "音频"
	case OTHER:
		return "其他"
	default:
		return "未知"
	}
}

// IsValid 检查分类是否有效
func (ec EvidenceCategory) IsValid() bool {
	return ec >= IMAGE && ec <= OTHER
}

// "手工艺"的凭据
type HandicraftEvidence struct {
	Id           int
	Uuid         string
	HandicraftId int    // 手工艺Id
	EvidenceId   int    // 指向 Evidence{}
	Note         string //备注,特别说明
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time //软删除
}

// Create 创建手工艺凭据关联
func (he *HandicraftEvidence) Create() (err error) {
	statement := `INSERT INTO handicraft_evidences 
		(uuid, handicraft_id, evidence_id, created_at, note) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), he.HandicraftId, he.EvidenceId, time.Now(), he.Note).Scan(&he.Id, &he.Uuid)
	return err
}

// GetByHandicraftId 根据手工艺ID获取凭据列表
func GetHandicraftEvidencesByHandicraftId(handicraftId int) ([]HandicraftEvidence, error) {
	rows, err := DB.Query("SELECT id, uuid, handicraft_id, evidence_id, created_at, updated_at, deleted_at, note FROM handicraft_evidences WHERE handicraft_id = $1 AND deleted_at IS NULL", handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []HandicraftEvidence
	for rows.Next() {
		var he HandicraftEvidence
		err := rows.Scan(&he.Id, &he.Uuid, &he.HandicraftId, &he.EvidenceId, &he.CreatedAt, &he.UpdatedAt, &he.DeletedAt, &he.Note)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, he)
	}
	return evidences, nil
}

// Delete 软删除手工艺凭据关联
func (he *HandicraftEvidence) Delete() error {
	statement := `UPDATE handicraft_evidences SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(he.Id, now)
	if err == nil {
		he.DeletedAt = &now
	}
	return err
}

// Validate 验证凭据数据
func (e *Evidence) Validate() error {
	if e.RecorderUserId <= 0 {
		return fmt.Errorf("记录人ID不能为空")
	}
	if !e.Category.IsValid() {
		return fmt.Errorf("无效的凭据类型")
	}
	if e.FileSize < 0 {
		return fmt.Errorf("文件大小不能为负数")
	}
	if e.FileSize > MaxFileSize {
		return fmt.Errorf("文件大小超过限制")
	}
	return nil
}

// Create 创建新凭据
func (e *Evidence) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err = e.Validate(); err != nil {
		return err
	}

	statement := `INSERT INTO evidences 
		(uuid, description, recorder_user_id, note, category, path, original_url, filename, 
		 mime_type, file_size, file_hash, width, height, duration, created_at, visibility) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), e.Description, e.RecorderUserId, e.Note,
		e.Category, e.Path, e.OriginalURL, e.FileName, e.MimeType, e.FileSize, e.FileHash,
		e.Width, e.Height, e.Duration, time.Now(), e.Visibility).Scan(&e.Id, &e.Uuid)
	return err
}

// GetByIdOrUUID 根据ID或UUID获取凭据
func (e *Evidence) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if e.Id <= 0 && e.Uuid == "" {
		return errors.New("invalid Evidence ID or UUID")
	}
	statement := `SELECT id, uuid, description, recorder_user_id, note, category, path, 
		original_url, filename, mime_type, file_size, file_hash, width, height, duration, 
		created_at, updated_at, deleted_at, visibility
		FROM evidences WHERE (id=$1 OR uuid=$2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, e.Id, e.Uuid).Scan(&e.Id, &e.Uuid, &e.Description,
		&e.RecorderUserId, &e.Note, &e.Category, &e.Path, &e.OriginalURL, &e.FileName,
		&e.MimeType, &e.FileSize, &e.FileHash, &e.Width, &e.Height, &e.Duration,
		&e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.Visibility)
	return err
}

// Update 更新凭据
func (e *Evidence) Update() error {
	if err := e.Validate(); err != nil {
		return err
	}
	statement := `UPDATE evidences SET description = $2, note = $3, visibility = $4, updated_at = $5  
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.Id, e.Description, e.Note, e.Visibility, time.Now())
	return err
}

// SoftDelete 软删除凭据
func (e *Evidence) SoftDelete() error {
	statement := `UPDATE evidences SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(e.Id, now)
	if err == nil {
		e.DeletedAt = &now
	}
	return err
}

// IsDeleted 检查是否已删除
func (e *Evidence) IsDeleted() bool {
	return e.DeletedAt != nil
}

// CreatedDateTime 格式化创建时间
func (e *Evidence) CreatedDateTime() string {
	return e.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// CategoryString 获取分类字符串
func (e *Evidence) CategoryString() string {
	return e.Category.String()
}

// VisibilityString 获取可见性字符串
func (e *Evidence) VisibilityString() string {
	if e.Visibility == VisibilityPublic {
		return "公开"
	}
	return "私有"
}

// GetEvidenceByUUID 根据UUID获取凭据
func GetEvidenceByUUID(uuid string, ctx context.Context) (Evidence, error) {
	var e Evidence
	e.Uuid = uuid
	err := e.GetByIdOrUUID(ctx)
	return e, err
}

// GetEvidencesByUser 获取用户的凭据列表
func GetEvidencesByUser(userId int, limit int, ctx context.Context) ([]Evidence, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, description, category, filename, file_size, created_at, visibility
		FROM evidences WHERE recorder_user_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC LIMIT $2`
	rows, err := DB.QueryContext(ctx, statement, userId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []Evidence
	for rows.Next() {
		var e Evidence
		err := rows.Scan(&e.Id, &e.Uuid, &e.Description, &e.Category, &e.FileName,
			&e.FileSize, &e.CreatedAt, &e.Visibility)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, e)
	}
	return evidences, nil
}

// GetEvidencesByCategory 根据分类获取凭据列表
func GetEvidencesByCategory(category EvidenceCategory, limit int, ctx context.Context) ([]Evidence, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, description, recorder_user_id, filename, file_size, created_at, visibility
		FROM evidences WHERE category = $1 AND deleted_at IS NULL AND visibility = $2
		ORDER BY created_at DESC LIMIT $3`
	rows, err := DB.QueryContext(ctx, statement, category, VisibilityPublic, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []Evidence
	for rows.Next() {
		var e Evidence
		err := rows.Scan(&e.Id, &e.Uuid, &e.Description, &e.RecorderUserId, &e.FileName,
			&e.FileSize, &e.CreatedAt, &e.Visibility)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, e)
	}
	return evidences, nil
}

// “看看”的凭据,指音视频等视觉证据，
type SeeSeekLookEvidence struct {
	Id         int
	Uuid       string
	SeeSeekId  int    // 缺省值 0，标记属于那一个“看看”，see-seek.go -> SeeSeek{}
	EvidenceId int    // 缺省值 0，指向 Evidence{}
	Note       string //备注,特别说明
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time //软删除
}

// Create 创建“看看”凭据关联
func (ssle *SeeSeekLookEvidence) Create() (err error) {
	statement := `INSERT INTO see_seek_look_evidences 
		(uuid, see_seek_id, evidence_id, created_at, note) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(Random_UUID(), ssle.SeeSeekId, ssle.EvidenceId, time.Now(), ssle.Note).Scan(&ssle.Id, &ssle.Uuid)
	return err
}

// GetBySeeSeekId 根据“看看”ID获取凭据列表
func GetSeeSeekEvidencesBySeeSeekId(seeSeekId int) ([]SeeSeekLookEvidence, error) {
	rows, err := DB.Query("SELECT id, uuid, see_seek_id, evidence_id, created_at, updated_at, deleted_at, note FROM see_seek_look_evidences WHERE see_seek_id = $1 AND deleted_at IS NULL", seeSeekId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var evidences []SeeSeekLookEvidence
	for rows.Next() {
		var ssle SeeSeekLookEvidence
		err := rows.Scan(&ssle.Id, &ssle.Uuid, &ssle.SeeSeekId, &ssle.EvidenceId, &ssle.CreatedAt, &ssle.UpdatedAt, &ssle.DeletedAt, &ssle.Note)
		if err != nil {
			return nil, err
		}
		evidences = append(evidences, ssle)
	}
	return evidences, nil
}

// Delete 软删除“看看”凭据关联
func (ssle *SeeSeekLookEvidence) Delete() error {
	statement := `UPDATE see_seek_look_evidences SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(ssle.Id, now)
	if err == nil {
		ssle.DeletedAt = &now
	}
	return err
}
