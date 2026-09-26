package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

/*
在一个茶围objective目标（愿望、需求）里，某个项目project(主意/应对方案）被选中"入围"之后，代表围主管理团队是明确需要启动线下作业服务以解决问题。
--如果没有“入围”动作，意味无需线下作业行动，没有TeaOrder启动，继续线上讨论即可；
--如果启动tea_order实例并且获得合规批准，这是一个类似"大观园的诗社活动"，有见证方verify team（主持&裁判），需求方payer team（出题），解题方payee team（作答），一个总解题的过程通常可以拆分为多个部分，一个部分的实操环节对应一个handicraft（手工艺），
--为了慎独，另外引入可选监护方（care team）与解题方共同承担责任风险；
--启动tea_order之后，系统会生成一个tea_order实体，记录该解题服务的具体内容；
--一个tea_order可能包含多个handicraft，每个handicraft对应一个具体的解题任务(见handicraft.go文件)；
--考虑作业环境影响甚大，手工艺添加主客场标识，以解题方为主视角，如果是在解题方指定地点作业，则为主场，由需求方选定的地方，则是客场（例如上门或者突发事件地点等)；
(作为存档记录，tea_order涉及的茶围objective与项目project信息及参与方实体将固定（已经完成的order不可删除）以实现历史还原)
。。。
*/
type TeaOrder struct {
	Id   int
	Uuid string

	ObjectiveId    int         // 茶围目标ID
	ProjectId      int         // 项目ID
	UserId         int         // 茶围管理团队成员，选择入围操作者
	Status         string      // tea_order状态：pending/active/pause/completed/cancelled
	VerifyTeamId   int         // 见证方团队ID
	PayerTeamId    int         // 需求方（出题方）团队ID
	PayeeTeamId    int         // 解题方团队ID
	CareTeamId     int         // 监护方团队ID
	ServiceMode    ServiceMode // 服务模式
	DefaultPlaceId int         //默认地点id
	// 审批人（见证者）填写，必填
	// 审批人角色是类似大观园海棠诗社活动中的李纨社长角色，批准主题、主持活动及裁判"违规"情形，将阻止贾宝玉作西厢记类那种"男女礼教脱轨诗"或者禁止薛蟠那种酒色情诗；
	// 又或者是老师组织的多团队协作任务活动里的老师角色，不过在这茶会里不负责技术方面的审核，所以说"见证"记录事件发生的真实性、合规性。
	// 见证人也是活动进程主持人，类似教堂神父主持婚礼活动，发现不道德的欺瞒情况，例如新郎或者新娘竟然是重婚者之类不符合道德规范的活动将取消或者宣布无效。
	TeaTopic                string        // 茶会主题，默认值'-'。审批时即使不批准也应当根据茶围出题内容提炼，例如：热水器维修，宠物狗口腔护理,等。
	IsApproved              bool          // 是否批准，审批人填写（必填）,默认false
	ApproverUserId          sql.NullInt64 // 审批人ID，也是关联订单负责人，必须是见证者团队成员,如果Valid为false代表待审批
	ApprovalRejectionReason string        // 审批意见，如果拒绝，填写原因,默认值:'-'
	ApprovedAt              *time.Time    // 审批时间

	FinalScore sql.NullInt64 // 根据另外的打分表计算后得到的最终的解题评分，NULL代表未评分，0代表得0分。
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	DeletedAt  *time.Time //软删除时间（未完成的tea_order可以被取消删除，已完成的tea_order不可删除）

	ServiceOfferingId int // 服务项目ID
	ServiceVersionId  int // 服务项目版本ID
}

// TeaOrderMember 是订单创建时生成的上场名单快照。
// 成员后来离团、改名或技能等级变化，不影响这份历史记录。
type TeaOrderMember struct {
	Id                   int
	Uuid                 string
	TeaOrderId           int
	RequirementId        int
	TeamId               int
	TeamMemberId         int
	UserId               int
	SkillId              int
	ServiceRole          string
	TeamRoleSnapshot     int
	SkillLevelSnapshot   int
	ResponsibilityWeight int
	ParticipationStatus  string
	JoinedAt             time.Time
	LeftAt               *time.Time
}

const (
	TeaOrderStatusPending   = "pending"   // 待处理
	TeaOrderStatusActive    = "active"    // 活动中
	TeaOrderStatusPause     = "pause"     // 暂停
	TeaOrderStatusCompleted = "completed" // 完成
	TeaOrderStatusCancelled = "cancelled" // 已取消
)

// 见证日志
type WitnessLog struct {
	Id         int
	Uuid       string
	TeaOrderId int
	Action     string // "审批"/"暂停"/"恢复"/"终止"/"罚没"/"退款"
	Reason     string
	WitnessId  int // 见证人用户ID
	EvidenceId int // 证据材料 ->Evidence{}
	WitnessAt  time.Time
}

const (
	WitnessActionApprove = "审批"
	WitnessActionPause   = "暂停"
	WitnessActionResume  = "恢复"
	WitnessActionCancel  = "终止"
	WitnessActionForfeit = "罚没" // 见证人对违规恶意/不道德行为的处罚，罚没星茶转入系统特殊团队"公共治理团队"
	WitnessActionRefund  = "退回" // 见证人对无恶意但超出预设讨论范围的约茶的处理，退款星茶原路退回双方团队
)

type ServiceMode string

const (
	ServiceModeCustomerDelivers ServiceMode = "customer_delivers" // 需求方送到解题方场所
	ServiceModeProviderVisits   ServiceMode = "provider_visits"   // 解题方上门或者到达需求方指定场所
	ServiceModeThirdPartyVenue  ServiceMode = "third_party_venue" // 第三方（例如道路商场公园等公共场所）
	ServiceModeMobile           ServiceMode = "mobile"
)

func (m ServiceMode) IsValid() bool {
	switch m {
	case ServiceModeCustomerDelivers, ServiceModeProviderVisits, ServiceModeThirdPartyVenue, ServiceModeMobile:
		return true
	default:
		return false
	}
}

// WitnessLog.Create() 创建见证日志
func (w *WitnessLog) Create(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `INSERT INTO witness_logs (tea_order_id, action, reason, witness_id, evidence_id, witness_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, w.TeaOrderId, w.Action, w.Reason, w.WitnessId, w.EvidenceId, w.WitnessAt)
	return err
}

// WitnessLog.GetByTeaOrderId() 获取见证日志列表
func (w *WitnessLog) GetByTeaOrderId(ctx context.Context) ([]*WitnessLog, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, tea_order_id, action, reason, witness_id, evidence_id, witness_at FROM witness_logs WHERE tea_order_id = $1 ORDER BY witness_at DESC`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, w.TeaOrderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	witnessLogs := make([]*WitnessLog, 0)
	for rows.Next() {
		witnessLog := &WitnessLog{}
		err := rows.Scan(&witnessLog.Id, &witnessLog.Uuid, &witnessLog.TeaOrderId, &witnessLog.Action, &witnessLog.Reason, &witnessLog.WitnessId, &witnessLog.EvidenceId, &witnessLog.WitnessAt)
		if err != nil {
			return nil, err
		}
		witnessLogs = append(witnessLogs, witnessLog)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return witnessLogs, nil
}

// 根据状态获取茶订单列表
func GetTeaOrdersByStatus(ctx context.Context, status string, page int, pageSize int) ([]*TeaOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, status, pageSize, page*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teaOrders := make([]*TeaOrder, 0)
	for rows.Next() {
		teaOrder := &TeaOrder{}
		err := rows.Scan(&teaOrder.Id, &teaOrder.Uuid, &teaOrder.ObjectiveId, &teaOrder.ProjectId, &teaOrder.UserId, &teaOrder.Status, &teaOrder.VerifyTeamId, &teaOrder.PayerTeamId, &teaOrder.PayeeTeamId, &teaOrder.CareTeamId, &teaOrder.ServiceMode, &teaOrder.ServiceOfferingId, &teaOrder.ServiceVersionId, &teaOrder.TeaTopic, &teaOrder.IsApproved, &teaOrder.ApproverUserId, &teaOrder.ApprovalRejectionReason, &teaOrder.ApprovedAt, &teaOrder.FinalScore, &teaOrder.CreatedAt, &teaOrder.UpdatedAt, &teaOrder.DeletedAt)
		if err != nil {
			return nil, err
		}
		teaOrders = append(teaOrders, teaOrder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teaOrders, nil
}

// 根据状态获取茶订单数量
func GetTeaOrderCountByStatus(ctx context.Context, status string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE status = $1`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, status).Scan(&count)
	return count, err
}

// 获取待审批订单数量（用于徽章提示）
func GetPendingTeaOrderCount(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE status = $1`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, TeaOrderStatusPending).Scan(&count)
	return count, err
}

// Create 创建新的茶订单记录；带服务版本时同时生成订单上场名单快照。
func (t *TeaOrder) Create(ctx context.Context) (err error) {
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

	var serviceOfferingId interface{}
	var serviceVersionId interface{}
	if t.ServiceOfferingId > 0 {
		serviceOfferingId = t.ServiceOfferingId
	}
	if t.ServiceVersionId > 0 {
		serviceVersionId = t.ServiceVersionId
		var offeringId, teamId, status, currentVersionId int
		err = tx.QueryRowContext(ctx, `SELECT o.id, o.team_id, o.status, COALESCE(o.current_version_id, 0)
			FROM team_service_offering_versions v
			JOIN team_service_offerings o ON o.id = v.service_offering_id
			WHERE v.id = $1 AND o.deleted_at IS NULL`, t.ServiceVersionId).Scan(&offeringId, &teamId, &status, &currentVersionId)
		if err != nil {
			return fmt.Errorf("服务版本不存在或所属服务已删除: %w", err)
		}
		if t.ServiceOfferingId > 0 && t.ServiceOfferingId != offeringId {
			return errors.New("服务项目与服务版本不匹配")
		}
		if t.PayeeTeamId != teamId {
			return errors.New("服务版本所属团队与订单解题方团队不匹配")
		}
		if status != int(PublishedTeamServiceOfferingStatus) || currentVersionId != t.ServiceVersionId {
			return errors.New("只能使用解题方当前已上架的服务版本")
		}
		t.ServiceOfferingId = offeringId
		serviceOfferingId = offeringId
	}

	statement := `INSERT INTO tea_orders (objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`
	err = tx.QueryRowContext(ctx, statement, t.ObjectiveId, t.ProjectId, t.UserId, t.Status, t.VerifyTeamId, t.PayerTeamId, t.PayeeTeamId, t.CareTeamId, t.ServiceMode, serviceOfferingId, serviceVersionId).Scan(&t.Id)
	if err != nil {
		return err
	}
	if t.ServiceVersionId > 0 {
		if err = createTeaOrderMemberSnapshots(ctx, tx, t.Id, t.ServiceVersionId, t.PayeeTeamId); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// createTeaOrderMemberSnapshots 根据服务版本要求，为每个岗位选择当前活跃且技能达标的成员。
func createTeaOrderMemberSnapshots(ctx context.Context, tx *sql.Tx, teaOrderId, versionId, payeeTeamId int) error {
	rows, err := tx.QueryContext(ctx, `SELECT id, skill_id, team_id, role_name, required_level,
		required_count, responsibility_weight
		FROM service_version_skill_requirements
		WHERE service_version_id = $1 ORDER BY id`, versionId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var requirementId, skillId, teamId, requiredLevel, requiredCount, responsibilityWeight int
		var roleName string
		if err := rows.Scan(&requirementId, &skillId, &teamId, &roleName, &requiredLevel,
			&requiredCount, &responsibilityWeight); err != nil {
			return err
		}
		if teamId != payeeTeamId {
			return errors.New("服务版本技能要求包含非解题方团队技能")
		}

		candidates, err := tx.QueryContext(ctx, `SELECT tm.id, tm.user_id, tm.role, su.level
			FROM team_members tm
			JOIN skill_users su ON su.user_id = tm.user_id
			WHERE tm.team_id = $1 AND tm.status = $2 AND tm.deleted_at IS NULL
			  AND su.skill_id = $3 AND su.level >= $4
			  AND su.status IN ($5, $6) AND su.deleted_at IS NULL
			ORDER BY su.level DESC, tm.id
			LIMIT $7`, teamId, TeamMemberStatusActive, skillId, requiredLevel,
			NormalSkillUserStatus, StrongSkillUserStatus, requiredCount)
		if err != nil {
			return err
		}

		selected := 0
		for candidates.Next() {
			var teamMemberId, userId, teamRole, skillLevel int
			if err := candidates.Scan(&teamMemberId, &userId, &teamRole, &skillLevel); err != nil {
				candidates.Close()
				return err
			}
			_, err = tx.ExecContext(ctx, `INSERT INTO tea_order_members
				(tea_order_id, requirement_id, team_id, team_member_id, user_id, skill_id,
				 service_role, team_role_snapshot, skill_level_snapshot, responsibility_weight,
				 participation_status, joined_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'assigned', CURRENT_TIMESTAMP)`,
				teaOrderId, requirementId, teamId, teamMemberId, userId, skillId, roleName,
				teamRole, skillLevel, responsibilityWeight)
			if err != nil {
				candidates.Close()
				return err
			}
			selected++
		}
		if err := candidates.Err(); err != nil {
			candidates.Close()
			return err
		}
		candidates.Close()
		if selected < requiredCount {
			return fmt.Errorf("服务版本岗位 %q 没有足够的合适成员，需要%d人，实际%d人", roleName, requiredCount, selected)
		}
	}
	return rows.Err()
}

// GetTeaOrderMembers 获取订单上场名单快照。
func GetTeaOrderMembers(ctx context.Context, teaOrderId int) ([]*TeaOrderMember, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := DB.QueryContext(ctx, `SELECT id, uuid, tea_order_id, COALESCE(requirement_id, 0),
		team_id, team_member_id, user_id, skill_id, service_role, team_role_snapshot,
		skill_level_snapshot, responsibility_weight, participation_status, joined_at, left_at
		FROM tea_order_members WHERE tea_order_id = $1 ORDER BY id`, teaOrderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*TeaOrderMember, 0)
	for rows.Next() {
		member := &TeaOrderMember{}
		if err := rows.Scan(&member.Id, &member.Uuid, &member.TeaOrderId, &member.RequirementId,
			&member.TeamId, &member.TeamMemberId, &member.UserId, &member.SkillId, &member.ServiceRole,
			&member.TeamRoleSnapshot, &member.SkillLevelSnapshot, &member.ResponsibilityWeight,
			&member.ParticipationStatus, &member.JoinedAt, &member.LeftAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

// GetByIdOrUUID 根据ID或UUID获取茶订单记录
func (t *TeaOrder) GetByIdOrUUID(ctx context.Context) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if t.Id <= 0 && t.Uuid == "" {
		return errors.New("invalid TeaOrder ID or UUID")
	}
	statement := `SELECT id, uuid, objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE (id = $1 OR uuid = $2) AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, t.Id, t.Uuid).Scan(&t.Id, &t.Uuid, &t.ObjectiveId, &t.ProjectId, &t.UserId, &t.Status, &t.VerifyTeamId, &t.PayerTeamId, &t.PayeeTeamId, &t.CareTeamId, &t.ServiceMode, &t.ServiceOfferingId, &t.ServiceVersionId, &t.TeaTopic, &t.IsApproved, &t.ApproverUserId, &t.ApprovalRejectionReason, &t.ApprovedAt, &t.FinalScore, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	return err
}

// GetTeaOrderByProjectIdAndObjectiveId 根据项目ID和茶围目标ID获取茶订单记录
func GetTeaOrderByProjectIdAndObjectiveId(ctx context.Context, projectId int, objectiveId int) (*TeaOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE project_id = $1 AND objective_id = $2 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	teaOrder := &TeaOrder{}
	err = stmt.QueryRowContext(ctx, projectId, objectiveId).Scan(&teaOrder.Id, &teaOrder.Uuid, &teaOrder.ObjectiveId, &teaOrder.ProjectId, &teaOrder.UserId, &teaOrder.Status, &teaOrder.VerifyTeamId, &teaOrder.PayerTeamId, &teaOrder.PayeeTeamId, &teaOrder.CareTeamId, &teaOrder.ServiceMode, &teaOrder.ServiceOfferingId, &teaOrder.ServiceVersionId, &teaOrder.TeaTopic, &teaOrder.IsApproved, &teaOrder.ApproverUserId, &teaOrder.ApprovalRejectionReason, &teaOrder.ApprovedAt, &teaOrder.FinalScore, &teaOrder.CreatedAt, &teaOrder.UpdatedAt, &teaOrder.DeletedAt)
	if err != nil {
		return nil, err
	}
	return teaOrder, nil
}

// Update 更新茶订单记录
func (t *TeaOrder) Update() error {
	statement := `UPDATE tea_orders SET status = $2, service_mode = $3, tea_topic = $4, is_approved = $5, approver_user_id = $6,
		approval_rejection_reason = $7, approved_at = $8, final_score = $9, updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(t.Id, t.Status, t.ServiceMode, t.TeaTopic, t.IsApproved, t.ApproverUserId,
		t.ApprovalRejectionReason, t.ApprovedAt, t.FinalScore, time.Now())
	return err
}

// Delete 软删除茶订单记录
func (t *TeaOrder) Delete() error {
	statement := `UPDATE tea_orders SET deleted_at = $2 WHERE id = $1`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now()
	_, err = stmt.Exec(t.Id, now)
	if err == nil {
		t.DeletedAt = &now
	}
	return err
}

// GetTeaOrdersByPayerTeamId 查询某团队作为需求方（出题方）的茶订单，按创建时间倒序，分页
func GetTeaOrdersByPayerTeamId(ctx context.Context, teamId int, page int, pageSize int) ([]*TeaOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE payer_team_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, teamId, pageSize, page*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teaOrders := make([]*TeaOrder, 0)
	for rows.Next() {
		teaOrder := &TeaOrder{}
		err := rows.Scan(&teaOrder.Id, &teaOrder.Uuid, &teaOrder.ObjectiveId, &teaOrder.ProjectId, &teaOrder.UserId, &teaOrder.Status, &teaOrder.VerifyTeamId, &teaOrder.PayerTeamId, &teaOrder.PayeeTeamId, &teaOrder.CareTeamId, &teaOrder.ServiceMode, &teaOrder.ServiceOfferingId, &teaOrder.ServiceVersionId, &teaOrder.TeaTopic, &teaOrder.IsApproved, &teaOrder.ApproverUserId, &teaOrder.ApprovalRejectionReason, &teaOrder.ApprovedAt, &teaOrder.FinalScore, &teaOrder.CreatedAt, &teaOrder.UpdatedAt, &teaOrder.DeletedAt)
		if err != nil {
			return nil, err
		}
		teaOrders = append(teaOrders, teaOrder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teaOrders, nil
}

// GetTeaOrdersByPayeeTeamId 查询某团队作为解题方的茶订单，按创建时间倒序，分页
func GetTeaOrdersByPayeeTeamId(ctx context.Context, teamId int, page int, pageSize int) ([]*TeaOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT id, uuid, objective_id, project_id, user_id, status, verify_team_id, payer_team_id, payee_team_id, care_team_id, service_mode, service_offering_id, service_version_id, tea_topic, is_approved, approver_user_id, approval_rejection_reason, approved_at, final_score, created_at, updated_at, deleted_at FROM tea_orders WHERE payee_team_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, teamId, pageSize, page*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	teaOrders := make([]*TeaOrder, 0)
	for rows.Next() {
		teaOrder := &TeaOrder{}
		err := rows.Scan(&teaOrder.Id, &teaOrder.Uuid, &teaOrder.ObjectiveId, &teaOrder.ProjectId, &teaOrder.UserId, &teaOrder.Status, &teaOrder.VerifyTeamId, &teaOrder.PayerTeamId, &teaOrder.PayeeTeamId, &teaOrder.CareTeamId, &teaOrder.ServiceMode, &teaOrder.ServiceOfferingId, &teaOrder.ServiceVersionId, &teaOrder.TeaTopic, &teaOrder.IsApproved, &teaOrder.ApproverUserId, &teaOrder.ApprovalRejectionReason, &teaOrder.ApprovedAt, &teaOrder.FinalScore, &teaOrder.CreatedAt, &teaOrder.UpdatedAt, &teaOrder.DeletedAt)
		if err != nil {
			return nil, err
		}
		teaOrders = append(teaOrders, teaOrder)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teaOrders, nil
}

// GetTeaOrderCountByPayerTeamId 统计某团队作为需求方（出题方）的茶订单数量
func GetTeaOrderCountByPayerTeamId(ctx context.Context, teamId int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE payer_team_id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, teamId).Scan(&count)
	return count, err
}

// GetTeaOrderCountByPayeeTeamId 统计某团队作为解题方的茶订单数量
func GetTeaOrderCountByPayeeTeamId(ctx context.Context, teamId int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	statement := `SELECT COUNT(*) FROM tea_orders WHERE payee_team_id = $1 AND deleted_at IS NULL`
	stmt, err := DB.PrepareContext(ctx, statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var count int
	err = stmt.QueryRowContext(ctx, teamId).Scan(&count)
	return count, err
}

// CreatedDateTime 格式化创建时间
func (t *TeaOrder) CreatedDateTime() string {
	return t.CreatedAt.Format(FMT_DATE_TIME_CN)
}
