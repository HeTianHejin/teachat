package dao

import (
	"context"
	"database/sql"
	"time"
)

// Create 保存任务参与者快照。
func (p *HandicraftParticipant) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return DB.QueryRowContext(ctx, `INSERT INTO handicraft_participants
		(handicraft_id, tea_order_member_id, user_id, participation_role, responsibility_weight,
		 actual_contribution, score, score_reason, participation_status)
		VALUES ($1, NULLIF($2, 0), $3, $4, $5, $6, NULLIF($7, -1), $8, $9)
		RETURNING id, uuid, created_at`, p.HandicraftId, p.TeaOrderMemberId, p.UserId,
		p.ParticipationRole, p.ResponsibilityWeight, p.ActualContribution, nullableScore(p.Score),
		p.ScoreReason, p.ParticipationStatus).Scan(&p.Id, &p.Uuid, &p.CreatedAt)
}

// GetHandicraftParticipants 获取未删除的任务参与者。
func GetHandicraftParticipants(ctx context.Context, handicraftId int) ([]HandicraftParticipant, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	rows, err := DB.QueryContext(ctx, `SELECT id, uuid, handicraft_id, COALESCE(tea_order_member_id, 0),
		user_id, participation_role, responsibility_weight, actual_contribution, score, score_reason,
		participation_status, created_at, deleted_at
		FROM handicraft_participants WHERE handicraft_id = $1 AND deleted_at IS NULL ORDER BY id`, handicraftId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	participants := make([]HandicraftParticipant, 0)
	for rows.Next() {
		var participant HandicraftParticipant
		if err := rows.Scan(&participant.Id, &participant.Uuid, &participant.HandicraftId,
			&participant.TeaOrderMemberId, &participant.UserId, &participant.ParticipationRole,
			&participant.ResponsibilityWeight, &participant.ActualContribution, &participant.Score,
			&participant.ScoreReason, &participant.ParticipationStatus, &participant.CreatedAt,
			&participant.DeletedAt); err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}
	return participants, rows.Err()
}

// Create 保存带来源、维度和证据的任务评分。
func (r *HandicraftRating) Create(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return DB.QueryRowContext(ctx, `INSERT INTO handicraft_ratings
		(handicraft_id, rater_user_id, rating_type, dimension, evidence_id, target_member_id, raw_score, comment)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, 0), $7, $8)
		RETURNING id, uuid, created_at`, r.HandicraftId, r.RaterUserId, r.RatingType,
		r.Dimension, r.EvidenceId, r.TargetMemberId, r.RawScore, r.Comment).Scan(&r.Id, &r.Uuid, &r.CreatedAt)
}

// SaveMemberOrderScore 保存订单结算后的个人成绩快照。
func (s *MemberOrderScore) Save(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return DB.QueryRowContext(ctx, `INSERT INTO member_order_scores
		(tea_order_id, user_id, contribution_score, performance_score, responsibility_score,
		 evidence_score, calculation_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tea_order_id, user_id) DO UPDATE SET
		contribution_score = EXCLUDED.contribution_score,
		performance_score = EXCLUDED.performance_score,
		responsibility_score = EXCLUDED.responsibility_score,
		evidence_score = EXCLUDED.evidence_score,
		calculation_version = EXCLUDED.calculation_version,
		calculated_at = CURRENT_TIMESTAMP
		RETURNING id, uuid, calculated_at`, s.TeaOrderId, s.UserId, s.ContributionScore,
		s.PerformanceScore, s.ResponsibilityScore, s.EvidenceScore, s.CalculationVersion).Scan(&s.Id, &s.Uuid, &s.CalculatedAt)
}

func nullableScore(score sql.NullInt64) int64 {
	if !score.Valid {
		return -1
	}
	return score.Int64
}
