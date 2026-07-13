package dao

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// IndustryTag 职业团队行业分类白名单标签
// 参考《国民经济行业分类》门类，用于约束职业团队的 tags 必须从该白名单中选取
type IndustryTag struct {
	Id          int
	Name        string // 标签名称，如"信息传输、软件和信息技术服务业"
	Category    string // 门类代码，如"I"
	Description string // 说明
	CreatedAt   time.Time
}

// Create 创建行业标签
func (tag *IndustryTag) Create() error {
	statement := `INSERT INTO industry_tags (name, category, description, created_at)
	              VALUES ($1, $2, $3, $4)
	              RETURNING id, created_at`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRow(tag.Name, tag.Category, tag.Description, time.Now()).
		Scan(&tag.Id, &tag.CreatedAt)
}

// Get 根据 ID 获取行业标签
func (tag *IndustryTag) Get() error {
	return DB.QueryRow("SELECT id, name, category, description, created_at FROM industry_tags WHERE id = $1", tag.Id).
		Scan(&tag.Id, &tag.Name, &tag.Category, &tag.Description, &tag.CreatedAt)
}

// GetByName 根据名称获取行业标签
func (tag *IndustryTag) GetByName() error {
	return DB.QueryRow("SELECT id, name, category, description, created_at FROM industry_tags WHERE name = $1", tag.Name).
		Scan(&tag.Id, &tag.Name, &tag.Category, &tag.Description, &tag.CreatedAt)
}

// GetAllIndustryTags 获取全部行业分类标签，按门类代码、名称排序
func GetAllIndustryTags() ([]IndustryTag, error) {
	rows, err := DB.Query("SELECT id, name, category, description, created_at FROM industry_tags ORDER BY category, name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]IndustryTag, 0)
	for rows.Next() {
		var tag IndustryTag
		if err = rows.Scan(&tag.Id, &tag.Name, &tag.Category, &tag.Description, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

// GetIndustryTagNames 获取全部行业标签名称集合
func GetIndustryTagNames() ([]string, error) {
	tags, err := GetAllIndustryTags()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names, nil
}

// IsValidIndustryTag 判断单个标签是否在行业分类白名单中
func IsValidIndustryTag(name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, nil
	}
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM industry_tags WHERE name = $1", name).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ValidateProfessionalTags 校验职业团队标签字符串
// 职业团队标签不能为空，且每个标签必须在 industry_tags 白名单中
// 返回错误信息，若校验通过返回 nil
func ValidateProfessionalTags(tags string) error {
	parts := SplitTags(tags)
	if len(parts) == 0 {
		return fmt.Errorf("职业团队必须至少填写一个行业分类标签")
	}

	invalid := make([]string, 0)
	for _, tag := range parts {
		ok, err := IsValidIndustryTag(tag)
		if err != nil {
			return fmt.Errorf("校验行业标签失败: %w", err)
		}
		if !ok {
			invalid = append(invalid, tag)
		}
	}
	if len(invalid) > 0 {
		return fmt.Errorf("以下标签不在行业分类白名单中：%s", strings.Join(invalid, ", "))
	}
	return nil
}

// NormalizeTags 将标签字符串规范化为逗号分隔（去重、去空白、去空值）
func NormalizeTags(tags string) string {
	parts := SplitTags(tags)
	seen := make(map[string]struct{}, len(parts))
	unique := make([]string, 0, len(parts))
	for _, p := range parts {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		unique = append(unique, p)
	}
	return strings.Join(unique, ",")
}

// DeleteIndustryTag 根据名称删除行业标签（用于管理后台）
func DeleteIndustryTag(name string) error {
	_, err := DB.Exec("DELETE FROM industry_tags WHERE name = $1", name)
	return err
}

// CountIndustryTags 统计行业标签数量
func CountIndustryTags() (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM industry_tags").Scan(&count)
	return count, err
}

// EnsureIndustryTagExists 根据名称查找或创建行业标签
func EnsureIndustryTagExists(name, category, description string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("行业标签名称不能为空")
	}

	tag := IndustryTag{Name: name}
	err := tag.GetByName()
	if err == nil {
		return tag.Id, nil
	}
	if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) && err != sql.ErrNoRows {
		return 0, err
	}

	tag.Category = category
	tag.Description = description
	if err := tag.Create(); err != nil {
		return 0, err
	}
	return tag.Id, nil
}
