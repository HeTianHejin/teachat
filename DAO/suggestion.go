package dao

import (
	"context"
	"database/sql"
	"time"
)

// 建议,
// 根据检查或者踏勘，讨论所得，拟采取最终处置方案：处理or搁置
type Suggestion struct {
	Id        int
	Uuid      string
	UserId    int //记录者，见证人
	ProjectId int //项目id

	Resolution bool //表态：处理（颔首）或者搁置（摇头）=不处理
	Body       string

	Category int //分类：0、公开，1、保密，仅当事家庭/团队可见内容
	Status   int //状态：0、未开始，1、已提交，

	CreatedAt time.Time
	UpdatedAt *time.Time
}
type SuggestionCategory int

const (
	SuggestionCategoryPublic  SuggestionCategory = iota // 公开
	SuggestionCategoryPrivate                           // 保密
)

type SuggestionStatus int

const (
	SuggestionStatusDraft     SuggestionStatus = iota // 未开始/草稿
	SuggestionStatusSubmitted                         // 已提交
	// SuggestionStatusApproved, StatusRejected ... 未来状态
)

// Suggestion.Create() 创建一个Suggestion记录
func (s *Suggestion) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO suggestions 
		(uuid, user_id, project_id, resolution, body, category, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id, uuid`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), s.UserId, s.ProjectId, s.Resolution,
		s.Body, s.Category, s.Status).Scan(&s.Id, &s.Uuid)
	return err
}

// Suggestion.Update() 更新Suggestion记录
func (s *Suggestion) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE suggestions SET resolution = $2, body = $3, category = $4, 
		status = $5, updated_at = $6 WHERE id = $1`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, s.Id, s.Resolution, s.Body, s.Category,
		s.Status, time.Now())
	return err
}

// Suggestion.GetByIdOrUUID() 根据ID或UUID获取Suggestion记录
func (s *Suggestion) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, user_id, project_id, resolution, body, category, status,
		created_at, updated_at
		FROM suggestions WHERE id=$1 OR uuid=$2`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, s.Id, s.Uuid).Scan(&s.Id, &s.Uuid, &s.UserId, &s.ProjectId,
		&s.Resolution, &s.Body, &s.Category, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	return err
}

// GetSuggestionByProjectId 根据project_id查找Suggestion记录
func GetSuggestionByProjectId(projectId int, ctx context.Context) (Suggestion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var s Suggestion
	statement := `SELECT id, uuid, user_id, project_id, resolution, body, category, status,
		created_at, updated_at
		FROM suggestions WHERE project_id=$1 ORDER BY created_at DESC LIMIT 1`
	stmt, err := db.PrepareContext(ctx, statement)
	if err != nil {
		return s, err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, projectId).Scan(&s.Id, &s.Uuid, &s.UserId, &s.ProjectId,
		&s.Resolution, &s.Body, &s.Category, &s.Status, &s.CreatedAt, &s.UpdatedAt)

	if err == sql.ErrNoRows {
		s.Id = 0
		return s, err
	} else if err != nil {
		return s, err
	}

	return s, nil
}

// Suggestion.StatusString() 返回状态的中文描述
func (s *Suggestion) StatusString() string {
	switch s.Status {
	case int(SuggestionStatusDraft):
		return "草稿"
	case int(SuggestionStatusSubmitted):
		return "已提交"
	default:
		return "未知状态"
	}
}

// Suggestion.CategoryString() 返回分类的中文描述
func (s *Suggestion) CategoryString() string {
	switch s.Category {
	case int(SuggestionCategoryPublic):
		return "公开"
	case int(SuggestionCategoryPrivate):
		return "保密"
	default:
		return "未知分类"
	}
}

// Suggestion.ResolutionString() 返回处置方案的中文描述
func (s *Suggestion) ResolutionString() string {
	if s.Resolution {
		return "处理"
	} else {
		return "搁置"
	}
}

// Suggestion.CreatedDateTime() 返回格式化的创建时间
func (s *Suggestion) CreatedDateTime() string {
	return s.CreatedAt.Format(FMT_DATE_TIME_CN)
}
