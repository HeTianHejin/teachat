package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// 团队服务项目
// 与技能区别：团队愿意为外部目标解决什么问题？
// 一个团队服务项目可以关连多个团队技能
// 确定身份、当前状态、所属团队
// 服务项目是团队的“可承诺能力”，技能是团队的“能力依据”，茶围目标是“需求上下文”，茶订单是“经确认的履约合同”。
// 生命周期：草稿 -> 待审核 -> 已上架 <-> 暂停接单 -> 已下架 -> 已废止；
// 只有"已上架"状态可以承接新茶订单，且进入"已上架"时必须锁定一个新版本，
// 茶订单通过 tea_orders.service_version_id 引用该版本，历史订单不受后续修改影响。
type TeamServiceOffering struct {
	Id               int
	Uuid             string
	TeamId           int
	Name             string
	Summary          string //服务项目简介
	Description      string //服务项目详细描述
	TargetProblem    string //服务项目解决的目标问题
	Deliverables     string //交付物清单
	Requirements     string //需求方需配合的前置条件
	EstimatedMinutes int    //预计耗时，单位分钟
	PriceMilligrams  int64  //价格，单位毫克（星茶），1克=1000毫克
	Status           TeamServiceOfferingStatus
	Availability     TeamServiceOfferingAvailability
	CurrentVersionId int //当前在架版本ID，未上架为0
	ApprovedBy       int //审核人（见证者团队成员）用户ID，未审核为0
	ApprovedAt       *time.Time
	RecorderUserId   int //登记人用户ID
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	PublishedAt      *time.Time
	UnpublishedAt    *time.Time
	RetiredAt        *time.Time
	DeletedAt        *time.Time //软删除时间戳，NULL表示未删除
}

// 服务项目状态
// 0 为"未知"零值哨兵，避免未初始化的结构体被误判为"草稿"而误上架。
type TeamServiceOfferingStatus int

const (
	UnknownTeamServiceOfferingStatus     TeamServiceOfferingStatus = iota //未知，零值哨兵
	DraftTeamServiceOfferingStatus                                        //草稿
	PendingTeamServiceOfferingStatus                                      //待审核
	PublishedTeamServiceOfferingStatus                                    //已上架
	PausedTeamServiceOfferingStatus                                       //暂停接单
	RejectedTeamServiceOfferingStatus                                     //已婉拒，审核不通过，可修改后重新提交
	UnpublishedTeamServiceOfferingStatus                                  //已下架
	RetiredTeamServiceOfferingStatus                                      //已废止
)

// 状态名称映射
var TeamServiceOfferingStatusNameMap = map[TeamServiceOfferingStatus]string{
	UnknownTeamServiceOfferingStatus:     "未知",
	DraftTeamServiceOfferingStatus:       "草稿",
	PendingTeamServiceOfferingStatus:     "待审核",
	PublishedTeamServiceOfferingStatus:   "已上架",
	PausedTeamServiceOfferingStatus:      "暂停接单",
	RejectedTeamServiceOfferingStatus:    "已婉拒",
	UnpublishedTeamServiceOfferingStatus: "已下架",
	RetiredTeamServiceOfferingStatus:     "已废止",
}

// StatusString 返回状态的中文描述
func (s TeamServiceOfferingStatus) StatusString() string {
	if name, ok := TeamServiceOfferingStatusNameMap[s]; ok {
		return name
	}
	return "未知状态"
}

// 允许的状态流转表：当前状态 -> 可达状态集合。已废止为终态。
var teamServiceOfferingTransitions = map[TeamServiceOfferingStatus][]TeamServiceOfferingStatus{
	DraftTeamServiceOfferingStatus:       {PendingTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	PendingTeamServiceOfferingStatus:     {PublishedTeamServiceOfferingStatus, RejectedTeamServiceOfferingStatus, DraftTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	RejectedTeamServiceOfferingStatus:    {DraftTeamServiceOfferingStatus, PendingTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	PublishedTeamServiceOfferingStatus:   {PausedTeamServiceOfferingStatus, UnpublishedTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	PausedTeamServiceOfferingStatus:      {PublishedTeamServiceOfferingStatus, UnpublishedTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	UnpublishedTeamServiceOfferingStatus: {DraftTeamServiceOfferingStatus, PendingTeamServiceOfferingStatus, RetiredTeamServiceOfferingStatus},
	RetiredTeamServiceOfferingStatus:     {},
}

// CanTransitionTo 判断状态是否允许流转到目标状态
func (s TeamServiceOfferingStatus) CanTransitionTo(target TeamServiceOfferingStatus) bool {
	for _, allowed := range teamServiceOfferingTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// 服务项目接单可用性
// 注意："产能已满"是派生状态，由进行中的茶订单数量实时计算，不落库，避免状态过期。
type TeamServiceOfferingAvailability int

const (
	UnknownTeamServiceOfferingAvailability TeamServiceOfferingAvailability = iota //未知，零值哨兵
	AvailableTeamServiceOffering                                                  //当前可接
	TemporarilyOffTeamServiceOffering                                             //暂时不可接
)

// 接单可用性名称映射
var TeamServiceOfferingAvailabilityNameMap = map[TeamServiceOfferingAvailability]string{
	UnknownTeamServiceOfferingAvailability: "未知",
	AvailableTeamServiceOffering:           "当前可接",
	TemporarilyOffTeamServiceOffering:      "暂时不可接",
}

// AvailabilityString 返回接单可用性的中文描述
func (a TeamServiceOfferingAvailability) AvailabilityString() string {
	if name, ok := TeamServiceOfferingAvailabilityNameMap[a]; ok {
		return name
	}
	return "未知状态"
}

// 服务项目与团队技能的关联（能力依据）
// 关联团队技能（skill_teams），同时冗余记录 team_id，
// 用于约束所关联技能必须属于服务项目的所属团队。
type TeamServiceOfferingSkill struct {
	Id                int
	Uuid              string
	ServiceOfferingId int
	SkillId           int  //技能ID，对应 skills.id
	TeamId            int  //所属团队ID，必须与服务项目所属团队一致
	RequiredLevel     int  //要求的团队技能等级，1-9
	IsPrimary         bool //是否为主要能力依据
	CreatedAt         time.Time
}

// 服务项目生命周期事件记录表
// 记录项目发布者
// 记录下架原因
// 记录服务，暂停次数
// 记录修改服务能力的核心成员
// 记录当前服务是否经历过审核？
// 建议上架、下架、暂停、废止、恢复都写事件记录
type ServiceOfferingEvent struct {
	Id                int
	Uuid              string
	ServiceOfferingId int
	VersionId         int //关联版本ID，0表示本次流转不涉及版本
	FromStatus        TeamServiceOfferingStatus
	ToStatus          TeamServiceOfferingStatus
	Action            string //动作，见 ServiceOfferingAction* 常量
	Reason            string //原因，默认'-'
	EvidenceId        int    //凭据材料ID，0表示无
	OperatorUserId    int    //操作人用户ID
	CreatedAt         time.Time
}

// 服务项目生命周期动作
const (
	ServiceOfferingActionSubmit    = "提交审核"
	ServiceOfferingActionApprove   = "审核通过"
	ServiceOfferingActionReject    = "审核婉拒"
	ServiceOfferingActionPause     = "暂停接单"
	ServiceOfferingActionResume    = "恢复接单"
	ServiceOfferingActionUnpublish = "下架"
	ServiceOfferingActionRetire    = "废止"
)

// 服务版本
// 上架服务项目是锁定一个版本，后续修改服务时创建新版本
// 服务 v1 上架, 订单 A 使用 v1
// 团队修改服务
// 服务 v2 创建并上架
// 订单 A 仍然显示 v1
// 新订单使用 v2
type TeamServiceOfferingVersion struct {
	Id                int
	Uuid              string
	ServiceOfferingId int    //所属服务项目ID
	VersionNo         int    //版本号，从1开始递增
	Name              string //服务项目名称，冗余自 TeamServiceOffering.name
	Summary           string //服务项目简介，冗余自 TeamServiceOffering.summary
	Description       string //服务项目详细描述，冗余自 TeamServiceOffering.description
	TargetProblem     string //服务项目解决的目标问题，冗余自 TeamServiceOffering.target_problem
	Deliverables      string //服务项目交付物清单，冗余自 TeamServiceOffering.deliverables
	Requirements      string //服务项目需求方需配合的前置条件，冗余自 TeamServiceOffering.requirements
	EstimatedMinutes  int    //服务项目预计耗时，单位分钟，冗余自 TeamServiceOffering.estimated_minutes
	PriceMilligrams   int64  //服务项目价格，单位毫克（星茶），1克=1000毫克，冗余自 TeamServiceOffering.price_milligrams
	IsCurrent         bool   //是否为当前在架版本
	RecorderUserId    int    //登记人id
	CreatedAt         time.Time
}

// ============================================
// 服务项目查询列与扫描辅助
// ============================================

// serviceOfferingColumns 服务项目查询列，保证各查询的 Scan 顺序一致
const serviceOfferingColumns = `id, uuid, team_id, name, summary, description, target_problem, deliverables, 
	requirements, estimated_minutes, price_milligrams, status, availability, 
	COALESCE(current_version_id, 0), COALESCE(approved_by, 0), approved_at, recorder_user_id, 
	created_at, updated_at, published_at, unpublished_at, retired_at, deleted_at`

// scanTeamServiceOffering 按 serviceOfferingColumns 的顺序扫描一行
func scanTeamServiceOffering(row interface {
	Scan(dest ...interface{}) error
}, tso *TeamServiceOffering) error {
	return row.Scan(
		&tso.Id, &tso.Uuid, &tso.TeamId, &tso.Name, &tso.Summary, &tso.Description, &tso.TargetProblem,
		&tso.Deliverables, &tso.Requirements, &tso.EstimatedMinutes, &tso.PriceMilligrams, &tso.Status,
		&tso.Availability, &tso.CurrentVersionId, &tso.ApprovedBy, &tso.ApprovedAt, &tso.RecorderUserId,
		&tso.CreatedAt, &tso.UpdatedAt, &tso.PublishedAt, &tso.UnpublishedAt, &tso.RetiredAt, &tso.DeletedAt)
}

// ============================================
// 团队服务项目 CRUD
// ============================================

// Validate 校验服务项目字段
func (tso *TeamServiceOffering) Validate() error {
	if tso.TeamId <= 0 {
		return errors.New("服务项目必须属于某个团队")
	}
	if strings.TrimSpace(tso.Name) == "" {
		return errors.New("服务项目名称不能为空")
	}
	if tso.PriceMilligrams < 0 {
		return errors.New("服务项目价格不能为负数")
	}
	if tso.EstimatedMinutes < 0 {
		return errors.New("预计耗时不能为负数")
	}
	if tso.RecorderUserId <= 0 {
		return errors.New("服务项目须记录登记人")
	}
	return nil
}

// Create 创建服务项目，默认为草稿且当前可接
func (tso *TeamServiceOffering) Create(ctx context.Context) error {
	if err := tso.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 未显式指定时使用安全默认值，避免零值被误判
	if tso.Status == UnknownTeamServiceOfferingStatus {
		tso.Status = DraftTeamServiceOfferingStatus
	}
	if tso.Availability == UnknownTeamServiceOfferingAvailability {
		tso.Availability = AvailableTeamServiceOffering
	}

	statement := `INSERT INTO team_service_offerings 
		(team_id, name, summary, description, target_problem, deliverables, requirements, 
		 estimated_minutes, price_milligrams, status, availability, recorder_user_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		RETURNING id, uuid, created_at`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRowContext(ctx, tso.TeamId, tso.Name, tso.Summary, tso.Description, tso.TargetProblem,
		tso.Deliverables, tso.Requirements, tso.EstimatedMinutes, tso.PriceMilligrams,
		tso.Status, tso.Availability, tso.RecorderUserId, time.Now()).Scan(&tso.Id, &tso.Uuid, &tso.CreatedAt)
}

// GetByIdOrUUID 根据ID或UUID获取服务项目（不含已软删除）
func (tso *TeamServiceOffering) GetByIdOrUUID(ctx context.Context) error {
	if tso.Id <= 0 && strings.TrimSpace(tso.Uuid) == "" {
		return errors.New("无效的服务项目ID或UUID")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT ` + serviceOfferingColumns + ` 
		FROM team_service_offerings 
		WHERE (id = $1 OR uuid = $2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = scanTeamServiceOffering(stmt.QueryRowContext(ctx, tso.Id, tso.Uuid), tso)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("服务项目不存在：id=%d uuid=%s", tso.Id, tso.Uuid)
	}
	return err
}

// Update 修改服务内容
// 已上架（版本已锁定）的条目必须先暂停接单或下架才能修改，以保证已锁定版本不可变。
func (tso *TeamServiceOffering) Update(ctx context.Context) error {
	if !tso.IsEditable() {
		return fmt.Errorf("当前状态（%s）不允许修改服务内容，请先暂停接单或下架", tso.Status.StatusString())
	}
	if err := tso.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE team_service_offerings SET name = $2, summary = $3, description = $4, 
		target_problem = $5, deliverables = $6, requirements = $7, estimated_minutes = $8, 
		price_milligrams = $9, availability = $10, updated_at = $11 
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	_, err = stmt.ExecContext(ctx, tso.Id, tso.Name, tso.Summary, tso.Description, tso.TargetProblem,
		tso.Deliverables, tso.Requirements, tso.EstimatedMinutes, tso.PriceMilligrams,
		tso.Availability, now)
	if err == nil {
		tso.UpdatedAt = &now
	}
	return err
}

// IsEditable 判断当前状态是否允许修改服务内容
func (tso *TeamServiceOffering) IsEditable() bool {
	switch tso.Status {
	case DraftTeamServiceOfferingStatus, PendingTeamServiceOfferingStatus,
		RejectedTeamServiceOfferingStatus, PausedTeamServiceOfferingStatus,
		UnpublishedTeamServiceOfferingStatus:
		return true
	default:
		return false
	}
}

// IsOrderable 判断当前是否可承接新茶订单：已上架且当前可接
// 说明："产能已满"由进行中的茶订单数量派生，调用方可另行校验。
func (tso *TeamServiceOffering) IsOrderable() bool {
	return tso.Status == PublishedTeamServiceOfferingStatus &&
		tso.Availability == AvailableTeamServiceOffering
}

// SoftDelete 软删除服务项目，仅草稿或已废止的条目可删除
func (tso *TeamServiceOffering) SoftDelete(ctx context.Context) error {
	if tso.Status != DraftTeamServiceOfferingStatus && tso.Status != RetiredTeamServiceOfferingStatus {
		return fmt.Errorf("仅草稿或已废止的服务项目可以删除，当前状态：%s", tso.Status.StatusString())
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `UPDATE team_service_offerings SET deleted_at = $2, updated_at = $2 
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	_, err = stmt.ExecContext(ctx, tso.Id, now)
	if err == nil {
		tso.DeletedAt = &now
		tso.UpdatedAt = &now
	}
	return err
}

// ============================================
// 团队服务项目生命周期流转
// ============================================

// transition 执行不涉及版本锁定的状态流转：在事务中更新状态并写入生命周期事件
func (tso *TeamServiceOffering) transition(ctx context.Context, to TeamServiceOfferingStatus,
	action, reason string, operatorUserId int) error {

	if tso.Id <= 0 {
		return errors.New("无效的服务项目ID")
	}
	if !tso.Status.CanTransitionTo(to) {
		return fmt.Errorf("服务项目状态不允许从 %s 流转到 %s",
			tso.Status.StatusString(), to.StatusString())
	}
	if strings.TrimSpace(reason) == "" {
		reason = "-"
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	from := tso.Status
	now := time.Now()

	setClause := "status = $2, updated_at = $3"
	switch to {
	case UnpublishedTeamServiceOfferingStatus:
		setClause += ", unpublished_at = $3"
	case RetiredTeamServiceOfferingStatus:
		setClause += ", retired_at = $3"
	}
	statement := fmt.Sprintf(`UPDATE team_service_offerings SET %s 
		WHERE id = $1 AND status = $4 AND deleted_at IS NULL`, setClause)
	result, err := tx.ExecContext(ctx, statement, tso.Id, to, now, from)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("服务项目状态已被其他操作变更，请重新加载后再试")
	}

	event := &ServiceOfferingEvent{
		ServiceOfferingId: tso.Id,
		FromStatus:        from,
		ToStatus:          to,
		Action:            action,
		Reason:            reason,
		OperatorUserId:    operatorUserId,
	}
	if err = event.createWithTx(ctx, tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true

	tso.Status = to
	tso.UpdatedAt = &now
	switch to {
	case UnpublishedTeamServiceOfferingStatus:
		tso.UnpublishedAt = &now
	case RetiredTeamServiceOfferingStatus:
		tso.RetiredAt = &now
	}
	return nil
}

// insertVersionSnapshot 在当前事务中把服务项目的内容锁定为一个新版本快照
// version_no 由数据库取当前最大值加一，唯一索引保证同一项目内版本号不重复。
func insertVersionSnapshot(ctx context.Context, tx *sql.Tx, tso *TeamServiceOffering,
	recorderUserId int, now time.Time) (*TeamServiceOfferingVersion, error) {

	// 先释放旧的当前版本，满足"每个服务项目最多一个当前版本"的唯一约束
	if _, err := tx.ExecContext(ctx, `UPDATE team_service_offering_versions SET is_current = FALSE 
		WHERE service_offering_id = $1 AND is_current`, tso.Id); err != nil {
		return nil, err
	}

	version := &TeamServiceOfferingVersion{
		ServiceOfferingId: tso.Id,
		Name:              tso.Name,
		Summary:           tso.Summary,
		Description:       tso.Description,
		TargetProblem:     tso.TargetProblem,
		Deliverables:      tso.Deliverables,
		Requirements:      tso.Requirements,
		EstimatedMinutes:  tso.EstimatedMinutes,
		PriceMilligrams:   tso.PriceMilligrams,
		IsCurrent:         true,
		RecorderUserId:    recorderUserId,
		CreatedAt:         now,
	}
	statement := `INSERT INTO team_service_offering_versions 
		(service_offering_id, version_no, name, summary, description, target_problem, deliverables, 
		 requirements, estimated_minutes, price_milligrams, is_current, recorder_user_id, created_at) 
		VALUES ($1, (SELECT COALESCE(MAX(version_no), 0) + 1 FROM team_service_offering_versions 
			WHERE service_offering_id = $1), $2, $3, $4, $5, $6, $7, $8, $9, TRUE, $10, $11) 
		RETURNING id, uuid, version_no`
	err := tx.QueryRowContext(ctx, statement, version.ServiceOfferingId, version.Name, version.Summary,
		version.Description, version.TargetProblem, version.Deliverables, version.Requirements,
		version.EstimatedMinutes, version.PriceMilligrams, version.RecorderUserId,
		version.CreatedAt).Scan(&version.Id, &version.Uuid, &version.VersionNo)
	if err != nil {
		return nil, err
	}
	return version, nil
}

// publish 流转到"已上架"：锁定一个新版本并写入事件，全部在一个事务内完成
// reviewerUserId 大于 0 表示本次为审核通过，同时记录审核人与审核时间。
func (tso *TeamServiceOffering) publish(ctx context.Context, action string,
	reviewerUserId, operatorUserId int) error {

	if tso.Id <= 0 {
		return errors.New("无效的服务项目ID")
	}
	if !tso.Status.CanTransitionTo(PublishedTeamServiceOfferingStatus) {
		return fmt.Errorf("服务项目状态不允许从 %s 流转到 %s",
			tso.Status.StatusString(), PublishedTeamServiceOfferingStatus.StatusString())
	}
	if err := tso.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	from := tso.Status
	now := time.Now()

	version, err := insertVersionSnapshot(ctx, tx, tso, operatorUserId, now)
	if err != nil {
		return err
	}

	statement := `UPDATE team_service_offerings SET status = $2, current_version_id = $3, 
		published_at = $4, updated_at = $4, 
		approved_by = CASE WHEN $5 > 0 THEN $5 ELSE approved_by END, 
		approved_at = CASE WHEN $5 > 0 THEN $4 ELSE approved_at END 
		WHERE id = $1 AND status = $6 AND deleted_at IS NULL`
	result, err := tx.ExecContext(ctx, statement, tso.Id, PublishedTeamServiceOfferingStatus,
		version.Id, now, reviewerUserId, from)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("服务项目状态已被其他操作变更，请重新加载后再试")
	}

	event := &ServiceOfferingEvent{
		ServiceOfferingId: tso.Id,
		VersionId:         version.Id,
		FromStatus:        from,
		ToStatus:          PublishedTeamServiceOfferingStatus,
		Action:            action,
		Reason:            "-",
		OperatorUserId:    operatorUserId,
	}
	if err = event.createWithTx(ctx, tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true

	tso.Status = PublishedTeamServiceOfferingStatus
	tso.CurrentVersionId = version.Id
	tso.PublishedAt = &now
	tso.UpdatedAt = &now
	if reviewerUserId > 0 {
		tso.ApprovedBy = reviewerUserId
		tso.ApprovedAt = &now
	}
	return nil
}

// Submit 提交审核：草稿/已婉拒/已下架 -> 待审核
func (tso *TeamServiceOffering) Submit(ctx context.Context, operatorUserId int) error {
	return tso.transition(ctx, PendingTeamServiceOfferingStatus,
		ServiceOfferingActionSubmit, "-", operatorUserId)
}

// Approve 审核通过并上架：待审核 -> 已上架，同时锁定新版本并记录审核人
func (tso *TeamServiceOffering) Approve(ctx context.Context, reviewerUserId, operatorUserId int) error {
	if reviewerUserId <= 0 {
		return errors.New("审核通过必须记录审核人（见证者团队成员）")
	}
	return tso.publish(ctx, ServiceOfferingActionApprove, reviewerUserId, operatorUserId)
}

// Reject 审核婉拒：待审核 -> 已婉拒，须填写原因
func (tso *TeamServiceOffering) Reject(ctx context.Context, operatorUserId int, reason string) error {
	return tso.transition(ctx, RejectedTeamServiceOfferingStatus,
		ServiceOfferingActionReject, reason, operatorUserId)
}

// Pause 暂停接单：已上架 -> 暂停接单，须填写原因
func (tso *TeamServiceOffering) Pause(ctx context.Context, operatorUserId int, reason string) error {
	return tso.transition(ctx, PausedTeamServiceOfferingStatus,
		ServiceOfferingActionPause, reason, operatorUserId)
}

// Resume 恢复接单：暂停接单 -> 已上架，恢复时重新锁定版本（暂停期间内容可能已修改）
func (tso *TeamServiceOffering) Resume(ctx context.Context, operatorUserId int) error {
	return tso.publish(ctx, ServiceOfferingActionResume, 0, operatorUserId)
}

// Unpublish 下架：已上架/暂停接单 -> 已下架
func (tso *TeamServiceOffering) Unpublish(ctx context.Context, operatorUserId int, reason string) error {
	return tso.transition(ctx, UnpublishedTeamServiceOfferingStatus,
		ServiceOfferingActionUnpublish, reason, operatorUserId)
}

// Retire 废止：任何非终态 -> 已废止（终态）
func (tso *TeamServiceOffering) Retire(ctx context.Context, operatorUserId int, reason string) error {
	return tso.transition(ctx, RetiredTeamServiceOfferingStatus,
		ServiceOfferingActionRetire, reason, operatorUserId)
}

// ============================================
// 团队服务项目查询与展示
// ============================================

// GetTeamServiceOfferingsByTeamId 获取团队的服务项目列表
// onlyPublished 为 true 时仅返回已上架的条目（供团队详情页展示）。
func GetTeamServiceOfferingsByTeamId(teamId int, onlyPublished bool, ctx context.Context) ([]*TeamServiceOffering, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT ` + serviceOfferingColumns + ` 
		FROM team_service_offerings WHERE team_id = $1 AND deleted_at IS NULL`
	if onlyPublished {
		statement += fmt.Sprintf(" AND status = %d", PublishedTeamServiceOfferingStatus)
	}
	statement += " ORDER BY published_at DESC NULLS LAST, created_at DESC"

	rows, err := DB.QueryContext(ctx, statement, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	offerings := make([]*TeamServiceOffering, 0)
	for rows.Next() {
		tso := &TeamServiceOffering{}
		if err = scanTeamServiceOffering(rows, tso); err != nil {
			return nil, err
		}
		offerings = append(offerings, tso)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return offerings, nil
}

// GetPublishedTeamServiceOfferingsByTeamId 获取团队已上架的服务项目列表
func GetPublishedTeamServiceOfferingsByTeamId(teamId int, ctx context.Context) ([]*TeamServiceOffering, error) {
	return GetTeamServiceOfferingsByTeamId(teamId, true, ctx)
}

// CountTeamServiceOfferingsByTeamIdAndStatus 统计团队指定状态的服务项目数量
func CountTeamServiceOfferingsByTeamIdAndStatus(teamId int, status TeamServiceOfferingStatus, ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT COUNT(*) FROM team_service_offerings 
		WHERE team_id = $1 AND status = $2 AND deleted_at IS NULL`
	var count int
	err := DB.QueryRowContext(ctx, statement, teamId, status).Scan(&count)
	return count, err
}

// PriceGrams 返回服务价格，以克为单位（1克=1000毫克）
func (tso *TeamServiceOffering) PriceGrams() float64 {
	return float64(tso.PriceMilligrams) / 1000.0
}

// PriceDisplay 返回服务价格的展示文本
func (tso *TeamServiceOffering) PriceDisplay() string {
	return fmt.Sprintf("%.3f克", tso.PriceGrams())
}

// EstimatedDurationDisplay 返回预计耗时的展示文本
func (tso *TeamServiceOffering) EstimatedDurationDisplay() string {
	if tso.EstimatedMinutes <= 0 {
		return "-"
	}
	if tso.EstimatedMinutes < 60 {
		return fmt.Sprintf("%d分钟", tso.EstimatedMinutes)
	}
	hours := tso.EstimatedMinutes / 60
	minutes := tso.EstimatedMinutes % 60
	if minutes == 0 {
		return fmt.Sprintf("%d小时", hours)
	}
	return fmt.Sprintf("%d小时%d分钟", hours, minutes)
}

// CreatedDateTime 格式化服务项目的创建时间
func (tso *TeamServiceOffering) CreatedDateTime() string {
	return tso.CreatedAt.Format(FMT_DATE_TIME_CN)
}

// ============================================
// 服务项目版本 TeamServiceOfferingVersion
// ============================================

// GetTeamServiceOfferingVersionById 按版本ID读取版本快照
// 茶订单展示使用：按订单锁定的 service_version_id 读取，保证历史订单显示旧版本。
func GetTeamServiceOfferingVersionById(versionId int, ctx context.Context) (*TeamServiceOfferingVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	version := &TeamServiceOfferingVersion{}
	statement := `SELECT id, uuid, service_offering_id, version_no, name, summary, description, 
		target_problem, deliverables, requirements, estimated_minutes, price_milligrams, 
		is_current, recorder_user_id, created_at 
		FROM team_service_offering_versions WHERE id = $1`
	err := DB.QueryRowContext(ctx, statement, versionId).Scan(
		&version.Id, &version.Uuid, &version.ServiceOfferingId, &version.VersionNo, &version.Name,
		&version.Summary, &version.Description, &version.TargetProblem, &version.Deliverables,
		&version.Requirements, &version.EstimatedMinutes, &version.PriceMilligrams,
		&version.IsCurrent, &version.RecorderUserId, &version.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("服务项目版本不存在：%d", versionId)
	}
	if err != nil {
		return nil, err
	}
	return version, nil
}

// GetTeamServiceOfferingVersionsByOfferingId 获取服务项目的全部版本快照（新版本在前）
func GetTeamServiceOfferingVersionsByOfferingId(offeringId int, ctx context.Context) ([]*TeamServiceOfferingVersion, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, service_offering_id, version_no, name, summary, description, 
		target_problem, deliverables, requirements, estimated_minutes, price_milligrams, 
		is_current, recorder_user_id, created_at 
		FROM team_service_offering_versions WHERE service_offering_id = $1 
		ORDER BY version_no DESC`
	rows, err := DB.QueryContext(ctx, statement, offeringId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make([]*TeamServiceOfferingVersion, 0)
	for rows.Next() {
		version := &TeamServiceOfferingVersion{}
		if err = rows.Scan(&version.Id, &version.Uuid, &version.ServiceOfferingId, &version.VersionNo,
			&version.Name, &version.Summary, &version.Description, &version.TargetProblem,
			&version.Deliverables, &version.Requirements, &version.EstimatedMinutes,
			&version.PriceMilligrams, &version.IsCurrent, &version.RecorderUserId,
			&version.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return versions, nil
}

// ============================================
// 服务项目生命周期事件 ServiceOfferingEvent
// ============================================

// createWithTx 在给定事务中写入生命周期事件
func (e *ServiceOfferingEvent) createWithTx(ctx context.Context, tx *sql.Tx) error {
	if e.ServiceOfferingId <= 0 {
		return errors.New("事件必须属于某个服务项目")
	}
	if e.OperatorUserId <= 0 {
		return errors.New("事件必须记录操作人")
	}
	if strings.TrimSpace(e.Reason) == "" {
		e.Reason = "-"
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}

	// version_id 可为空，用指针避免写入0值时触发外键失败
	var versionId interface{}
	if e.VersionId > 0 {
		versionId = e.VersionId
	}

	statement := `INSERT INTO team_service_offering_events 
		(service_offering_id, version_id, from_status, to_status, action, reason, evidence_id, 
		 operator_user_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		RETURNING id, uuid`
	stmt, err := tx.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return stmt.QueryRowContext(ctx, e.ServiceOfferingId, versionId, e.FromStatus, e.ToStatus,
		e.Action, e.Reason, e.EvidenceId, e.OperatorUserId, e.CreatedAt).Scan(&e.Id, &e.Uuid)
}

// Create 创建一条生命周期事件（独立事务）
func (e *ServiceOfferingEvent) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err = e.createWithTx(ctx, tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// GetServiceOfferingEventsByOfferingId 获取服务项目的生命周期事件列表（按时间倒序）
func GetServiceOfferingEventsByOfferingId(offeringId int, ctx context.Context) ([]*ServiceOfferingEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, service_offering_id, version_id, from_status, to_status, action, 
		reason, evidence_id, operator_user_id, created_at 
		FROM team_service_offering_events WHERE service_offering_id = $1 
		ORDER BY created_at DESC, id DESC`
	rows, err := DB.QueryContext(ctx, statement, offeringId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versionId sql.NullInt64
	var fromStatus sql.NullInt64
	events := make([]*ServiceOfferingEvent, 0)
	for rows.Next() {
		event := &ServiceOfferingEvent{}
		if err = rows.Scan(&event.Id, &event.Uuid, &event.ServiceOfferingId, &versionId, &fromStatus,
			&event.ToStatus, &event.Action, &event.Reason, &event.EvidenceId,
			&event.OperatorUserId, &event.CreatedAt); err != nil {
			return nil, err
		}
		if versionId.Valid {
			event.VersionId = int(versionId.Int64)
		}
		if fromStatus.Valid {
			event.FromStatus = TeamServiceOfferingStatus(fromStatus.Int64)
		}
		events = append(events, event)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// ============================================
// 服务项目技能关联 TeamServiceOfferingSkill
// ============================================

// Create 创建服务项目与团队技能的关联
// 通过 EXISTS 子句校验：服务项目与技能必须属于同一团队，且技能须已登记为该团队的技能。
func (tsos *TeamServiceOfferingSkill) Create(ctx context.Context) error {
	if tsos.ServiceOfferingId <= 0 {
		return errors.New("技能关联必须属于某个服务项目")
	}
	if tsos.SkillId <= 0 || tsos.TeamId <= 0 {
		return errors.New("技能关联必须同时指定技能ID与团队ID")
	}
	if tsos.RequiredLevel < 1 || tsos.RequiredLevel > 9 {
		return errors.New("要求的团队技能等级必须在1-9之间")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `INSERT INTO team_service_offering_skills 
		(service_offering_id, skill_id, team_id, required_level, is_primary, created_at) 
		SELECT $1, $2, $3, $4, $5, $6 
		WHERE EXISTS (SELECT 1 FROM team_service_offerings 
			WHERE id = $1 AND team_id = $3 AND deleted_at IS NULL) 
		  AND EXISTS (SELECT 1 FROM skill_teams 
			WHERE skill_id = $2 AND team_id = $3 AND deleted_at IS NULL) 
		RETURNING id, uuid`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, tsos.ServiceOfferingId, tsos.SkillId, tsos.TeamId,
		tsos.RequiredLevel, tsos.IsPrimary, time.Now()).Scan(&tsos.Id, &tsos.Uuid)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("技能关联失败：服务项目与技能必须属于同一团队，且技能须已登记为该团队的团队技能")
	}
	return err
}

// Delete 删除服务项目与团队技能的关联
func (tsos *TeamServiceOfferingSkill) Delete(ctx context.Context) error {
	if tsos.ServiceOfferingId <= 0 || tsos.SkillId <= 0 {
		return errors.New("删除技能关联必须指定服务项目ID与技能ID")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `DELETE FROM team_service_offering_skills 
		WHERE service_offering_id = $1 AND skill_id = $2`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, tsos.ServiceOfferingId, tsos.SkillId)
	return err
}

// GetTeamServiceOfferingSkillsByOfferingId 获取服务项目的技能关联列表（主要能力在前）
func GetTeamServiceOfferingSkillsByOfferingId(offeringId int, ctx context.Context) ([]*TeamServiceOfferingSkill, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	statement := `SELECT id, uuid, service_offering_id, skill_id, team_id, required_level, 
		is_primary, created_at 
		FROM team_service_offering_skills WHERE service_offering_id = $1 
		ORDER BY is_primary DESC, required_level DESC, id`
	rows, err := DB.QueryContext(ctx, statement, offeringId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make([]*TeamServiceOfferingSkill, 0)
	for rows.Next() {
		skill := &TeamServiceOfferingSkill{}
		if err = rows.Scan(&skill.Id, &skill.Uuid, &skill.ServiceOfferingId, &skill.SkillId,
			&skill.TeamId, &skill.RequiredLevel, &skill.IsPrimary, &skill.CreatedAt); err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return skills, nil
}
