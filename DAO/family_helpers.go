package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// 默认超时时间
const defaultTimeout = 5 * time.Second

// getContext 获取带超时的 context
func getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), defaultTimeout)
}

// wrapError 统一错误包装
func wrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("%s: record not found", operation)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// scanFamilies 通用的批量 Family 扫描函数
func scanFamilies(rows *sql.Rows) ([]Family, error) {
	defer rows.Close()
	var families []Family
	for rows.Next() {
		var f Family
		if err := rows.Scan(&f.Id, &f.Uuid, &f.AuthorId, &f.Name, &f.Introduction,
			&f.IsMarried, &f.HasChild, &f.HusbandFromFamilyId, &f.WifeFromFamilyId,
			&f.Status, &f.CreatedAt, &f.UpdatedAt, &f.Logo, &f.IsOpen); err != nil {
			return nil, err
		}
		families = append(families, f)
	}
	return families, rows.Err()
}

// queryFamiliesByUserRole 通用查询函数：根据用户ID和角色查询家庭
func queryFamiliesByUserRole(ctx context.Context, userID int, roles []int) ([]Family, error) {
	query := `SELECT f.id, f.uuid, f.author_id, f.name, f.introduction, f.is_married, 
		f.has_child, f.husband_from_family_id, f.wife_from_family_id, f.status, 
		f.created_at, f.updated_at, f.logo, f.is_open 
		FROM family_members fm 
		LEFT JOIN families f ON fm.family_id = f.id 
		WHERE fm.user_id = $1 AND fm.role = ANY($2)
		ORDER BY fm.created_at DESC`
	
	rows, err := db.QueryContext(ctx, query, userID, pq.Array(roles))
	if err != nil {
		return nil, wrapError("queryFamiliesByUserRole", err)
	}
	return scanFamilies(rows)
}

// validateFamily 验证 Family 数据
func (f *Family) Validate() error {
	if f.Name == "" {
		return fmt.Errorf("family name cannot be empty")
	}
	if f.AuthorId <= 0 {
		return fmt.Errorf("invalid author id: %d", f.AuthorId)
	}
	return nil
}

// validateFamilyMember 验证 FamilyMember 数据
func (fm *FamilyMember) Validate() error {
	if fm.FamilyId <= 0 {
		return fmt.Errorf("invalid family id: %d", fm.FamilyId)
	}
	if fm.UserId <= 0 {
		return fmt.Errorf("invalid user id: %d", fm.UserId)
	}
	if fm.Role < 0 || fm.Role > 5 {
		return fmt.Errorf("invalid role: %d", fm.Role)
	}
	return nil
}
