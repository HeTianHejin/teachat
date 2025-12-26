package dao

import (
	"context"
	"database/sql"
	"time"
)

// 限时思考？脑火，指在有效（限）时间情景下，智力资源运作消耗糖的表观记录
// 烧脑活动，根据“看看”所得线索作出快刀斩乱麻的判断
type BrainFire struct {
	Id        int
	Uuid      string
	ProjectId int
	// 日期时间
	StartTime     time.Time // 开始时间，默认为当前时间
	EndTime       time.Time // 结束时间，默认为开始时间+1小时
	EnvironmentId int       // 环境id

	Title     string
	Inference string //快速推理
	Diagnose  string //诊断、验证
	Judgement string //断言、最终解决方案

	PayerUserId   int //付茶叶代表人Id
	PayerTeamId   int //付茶叶团队Id
	PayerFamilyId int //付茶叶家庭Id

	PayeeUserId   int //收茶叶代表人Id
	PayeeTeamId   int //收茶叶团队Id
	PayeeFamilyId int //收茶叶家庭Id

	VerifierUserId   int
	VerifierFamilyId int
	VerifierTeamId   int
	CreatedAt        time.Time
	UpdatedAt        *time.Time

	//0、未点火
	//1、已点火
	//2、燃烧中
	//3、已熄灭
	Status BrainFireStatus

	//1、文艺类
	//2、理工类
	BrainFireClass int

	//1、公开
	//2、私密（专利？）
	BrainFireType int
}

const (
	BrainFireClassLiterature  = 1 //1、文艺类
	BrainFireClassEngineering = 2 //2、理工类
)
const (
	BrainFireTypePublic  = 1 //1、公开
	BrainFireTypePrivate = 2 //2、私密（专利？）
)

type BrainFireStatus int

const (
	BrainFireStatusUnlit        BrainFireStatus = iota // 未点火
	BrainFireStatusLit                                 // 已点火
	BrainFireStatusBurning                             // 燃烧中
	BrainFireStatusExtinguished                        // 已熄灭
)

// BrainFire.Create() 创建一个BrainFire记录
func (bf *BrainFire) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO brain_fires 
		(uuid, project_id, start_time, end_time, environment_id, title, inference, diagnose, judgement,
		 payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id,
		 verifier_user_id, verifier_family_id, verifier_team_id, status, brain_fire_class, brain_fire_type) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) 
		RETURNING id, uuid`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, Random_UUID(), bf.ProjectId, bf.StartTime, bf.EndTime, bf.EnvironmentId,
		bf.Title, bf.Inference, bf.Diagnose, bf.Judgement, bf.PayerUserId, bf.PayerTeamId,
		bf.PayerFamilyId, bf.PayeeUserId, bf.PayeeTeamId, bf.PayeeFamilyId, bf.VerifierUserId,
		bf.VerifierFamilyId, bf.VerifierTeamId, bf.Status, bf.BrainFireClass, bf.BrainFireType).Scan(&bf.Id, &bf.Uuid)
	return err
}

// BrainFire.Update() 更新BrainFire记录
func (bf *BrainFire) Update(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE brain_fires SET title = $2, inference = $3, diagnose = $4, judgement = $5, 
		status = $6, start_time = $7, end_time = $8, updated_at = $9 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, bf.Id, bf.Title, bf.Inference, bf.Diagnose, bf.Judgement,
		bf.Status, bf.StartTime, bf.EndTime, time.Now())
	return err
}

// BrainFire.GetByIdOrUUID() 根据ID或UUID获取BrainFire记录
func (bf *BrainFire) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, project_id, start_time, end_time, environment_id, title, inference, diagnose, judgement,
		payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id,
		verifier_user_id, verifier_family_id, verifier_team_id, status, brain_fire_class, brain_fire_type,
		created_at, updated_at
		FROM brain_fires WHERE id=$1 OR uuid=$2`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, bf.Id, bf.Uuid).Scan(&bf.Id, &bf.Uuid, &bf.ProjectId, &bf.StartTime, &bf.EndTime,
		&bf.EnvironmentId, &bf.Title, &bf.Inference, &bf.Diagnose, &bf.Judgement, &bf.PayerUserId, &bf.PayerTeamId,
		&bf.PayerFamilyId, &bf.PayeeUserId, &bf.PayeeTeamId, &bf.PayeeFamilyId, &bf.VerifierUserId,
		&bf.VerifierFamilyId, &bf.VerifierTeamId, &bf.Status, &bf.BrainFireClass, &bf.BrainFireType,
		&bf.CreatedAt, &bf.UpdatedAt)
	return err
}

// GetBrainFireByProjectId 根据project_id查找BrainFire记录
func GetBrainFireByProjectId(projectId int, ctx context.Context) (BrainFire, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var bf BrainFire
	statement := `SELECT id, uuid, project_id, start_time, end_time, environment_id, title, inference, diagnose, judgement,
		payer_user_id, payer_team_id, payer_family_id, payee_user_id, payee_team_id, payee_family_id,
		verifier_user_id, verifier_family_id, verifier_team_id, status, brain_fire_class, brain_fire_type,
		created_at, updated_at
		FROM brain_fires WHERE project_id = $1 ORDER BY created_at DESC LIMIT 1`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return bf, err
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, projectId).Scan(&bf.Id, &bf.Uuid, &bf.ProjectId, &bf.StartTime, &bf.EndTime,
		&bf.EnvironmentId, &bf.Title, &bf.Inference, &bf.Diagnose, &bf.Judgement, &bf.PayerUserId, &bf.PayerTeamId,
		&bf.PayerFamilyId, &bf.PayeeUserId, &bf.PayeeTeamId, &bf.PayeeFamilyId, &bf.VerifierUserId,
		&bf.VerifierFamilyId, &bf.VerifierTeamId, &bf.Status, &bf.BrainFireClass, &bf.BrainFireType,
		&bf.CreatedAt, &bf.UpdatedAt)

	if err == sql.ErrNoRows {
		bf.Id = 0
		return bf, err
	} else if err != nil {
		return bf, err
	}

	return bf, nil
}

// BrainFire.StatusString() 返回状态的中文描述
func (bf *BrainFire) StatusString() string {
	switch bf.Status {
	case BrainFireStatusUnlit:
		return "未点火"
	case BrainFireStatusLit:
		return "已点火"
	case BrainFireStatusBurning:
		return "燃烧中"
	case BrainFireStatusExtinguished:
		return "已熄灭"
	default:
		return "未知状态"
	}
}

// BrainFire.CreatedDateTime() 返回格式化的创建时间
func (bf *BrainFire) CreatedDateTime() string {
	return bf.CreatedAt.Format(FMT_DATE_TIME_CN)
}
