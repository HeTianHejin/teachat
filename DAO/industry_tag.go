package dao

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// 行业分类层级（参考《国民经济行业分类》GB/T 4754—2017）
// 门类 20 个、大类 97 个、中类 473 个、小类 1380 个
const (
	IndustryTagLevelCategory = 1 // 门类
	IndustryTagLevelDivision = 2 // 大类
	IndustryTagLevelGroup    = 3 // 中类
	IndustryTagLevelClass    = 4 // 小类
)

// IndustryTagMinProfessionalLevel 职业团队行业标签必须选到的最浅层级（大类）
const IndustryTagMinProfessionalLevel = IndustryTagLevelDivision

// industryTagColumns 统一查询列，保证各处 Scan 顺序一致
const industryTagColumns = "id, code, name, level, parent_id, category, description, created_at"

// IndustryTag 职业团队行业分类白名单标签（自引用树）
// 参考《国民经济行业分类》GB/T 4754—2017，用于约束职业团队的 tags 必须从该白名单中选取
type IndustryTag struct {
	Id          int
	Code        string // 完整标准代码：门类 'C'，大类 'C13'，中类 'C131'，小类 'C1311'
	Name        string // 标签名称，如"制造业"
	Level       int    // 层级：1=门类 2=大类 3=中类 4=小类
	ParentId    int    // 父节点 id；门类为 0（数据库存储 NULL）
	Category    string // 所属门类代码，如"C"
	Description string // 说明
	CreatedAt   time.Time
}

// IndustryTagLevelName 返回层级的中文名称
func IndustryTagLevelName(level int) string {
	switch level {
	case IndustryTagLevelCategory:
		return "门类"
	case IndustryTagLevelDivision:
		return "大类"
	case IndustryTagLevelGroup:
		return "中类"
	case IndustryTagLevelClass:
		return "小类"
	default:
		return "未知"
	}
}

// industryTagParentValue 把 0 值父节点转换为 SQL NULL
func industryTagParentValue(parentId int) sql.NullInt64 {
	if parentId <= 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(parentId), Valid: true}
}

// scanIndustryTag 按 industryTagColumns 的顺序扫描一行
func scanIndustryTag(scanner interface{ Scan(dest ...any) error }) (IndustryTag, error) {
	var (
		tag    IndustryTag
		parent sql.NullInt64
		cate   sql.NullString
		desc   sql.NullString
	)
	if err := scanner.Scan(&tag.Id, &tag.Code, &tag.Name, &tag.Level, &parent,
		&cate, &desc, &tag.CreatedAt); err != nil {
		return tag, err
	}
	if parent.Valid {
		tag.ParentId = int(parent.Int64)
	}
	if cate.Valid {
		tag.Category = cate.String
	}
	if desc.Valid {
		tag.Description = desc.String
	}
	return tag, nil
}

// queryIndustryTags 执行返回行业标签行的查询
func queryIndustryTags(query string, args ...any) ([]IndustryTag, error) {
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make([]IndustryTag, 0)
	for rows.Next() {
		tag, err := scanIndustryTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

// Create 创建行业标签
func (tag *IndustryTag) Create() error {
	statement := `INSERT INTO industry_tags (code, name, level, parent_id, category, description, created_at)
	              VALUES ($1, $2, $3, $4, $5, $6, $7)
	              RETURNING id, created_at`
	stmt, err := DB.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	return stmt.QueryRow(tag.Code, tag.Name, tag.Level, industryTagParentValue(tag.ParentId),
		tag.Category, tag.Description, time.Now()).
		Scan(&tag.Id, &tag.CreatedAt)
}

// Get 根据 ID 获取行业标签，结果写回接收者
func (tag *IndustryTag) Get() error {
	found, err := GetIndustryTagByID(tag.Id)
	if err != nil {
		return err
	}
	*tag = found
	return nil
}

// GetByName 根据名称获取行业标签，结果写回接收者
func (tag *IndustryTag) GetByName() error {
	row := DB.QueryRow("SELECT "+industryTagColumns+" FROM industry_tags WHERE name = $1", tag.Name)
	found, err := scanIndustryTag(row)
	if err != nil {
		return err
	}
	*tag = found
	return nil
}

// GetIndustryTagByID 根据 ID 获取行业标签
func GetIndustryTagByID(id int) (IndustryTag, error) {
	row := DB.QueryRow("SELECT "+industryTagColumns+" FROM industry_tags WHERE id = $1", id)
	return scanIndustryTag(row)
}

// GetIndustryTagByCode 根据完整代码获取行业标签，如 'C'、'C13'、'C131'
func GetIndustryTagByCode(code string) (IndustryTag, error) {
	row := DB.QueryRow("SELECT "+industryTagColumns+" FROM industry_tags WHERE code = $1", strings.TrimSpace(code))
	return scanIndustryTag(row)
}

// GetIndustryTagChildren 获取指定父节点下的直接子节点，按代码排序；
// parentId <= 0 时返回全部门类（level=1）
func GetIndustryTagChildren(parentId int) ([]IndustryTag, error) {
	if parentId <= 0 {
		return queryIndustryTags("SELECT " + industryTagColumns +
			" FROM industry_tags WHERE parent_id IS NULL ORDER BY code")
	}
	return queryIndustryTags("SELECT "+industryTagColumns+
		" FROM industry_tags WHERE parent_id = $1 ORDER BY code", parentId)
}

// CountIndustryTagChildren 统计直接子节点数量；返回 0 表示已到叶子，前端可停止下钻
func CountIndustryTagChildren(parentId int) (int, error) {
	var (
		count int
		err   error
	)
	if parentId <= 0 {
		err = DB.QueryRow("SELECT COUNT(*) FROM industry_tags WHERE parent_id IS NULL").Scan(&count)
	} else {
		err = DB.QueryRow("SELECT COUNT(*) FROM industry_tags WHERE parent_id = $1", parentId).Scan(&count)
	}
	return count, err
}

// GetAllIndustryTags 获取全部行业分类标签，按层级、代码排序
func GetAllIndustryTags() ([]IndustryTag, error) {
	return queryIndustryTags("SELECT " + industryTagColumns + " FROM industry_tags ORDER BY level, code")
}

// GetIndustryTagsUpToLevel 获取层级不超过 maxLevel 的标签（用于门类/大类这种浅层下拉）
func GetIndustryTagsUpToLevel(maxLevel int) ([]IndustryTag, error) {
	return queryIndustryTags("SELECT "+industryTagColumns+
		" FROM industry_tags WHERE level <= $1 ORDER BY level, code", maxLevel)
}

// GetIndustryTagAncestors 返回从门类到指定节点的祖先链（含自身），按层级升序
func GetIndustryTagAncestors(id int) ([]IndustryTag, error) {
	const query = `
WITH RECURSIVE chain AS (
    SELECT ` + industryTagColumns + ` FROM industry_tags WHERE id = $1
    UNION ALL
    SELECT t.id, t.code, t.name, t.level, t.parent_id, t.category, t.description, t.created_at
    FROM industry_tags t
    JOIN chain c ON t.id = c.parent_id
)
SELECT id, code, name, level, parent_id, category, description, created_at
FROM chain
ORDER BY level`
	return queryIndustryTags(query, id)
}

// GetIndustryTagDescendants 返回指定节点的全部后代（不含自身），按层级、代码升序
func GetIndustryTagDescendants(id int) ([]IndustryTag, error) {
	const query = `
WITH RECURSIVE subtree AS (
    SELECT t.id, t.code, t.name, t.level, t.parent_id, t.category, t.description, t.created_at
    FROM industry_tags t
    WHERE t.parent_id = $1
    UNION ALL
    SELECT t.id, t.code, t.name, t.level, t.parent_id, t.category, t.description, t.created_at
    FROM industry_tags t
    JOIN subtree s ON t.parent_id = s.id
)
SELECT id, code, name, level, parent_id, category, description, created_at
FROM subtree
ORDER BY level, code`
	return queryIndustryTags(query, id)
}

// GetIndustryTagPath 返回节点的代码路径，如 "C/C13"，用于 code 前缀上卷搜索
func GetIndustryTagPath(id int) (string, error) {
	chain, err := GetIndustryTagAncestors(id)
	if err != nil {
		return "", err
	}
	codes := make([]string, 0, len(chain))
	for _, tag := range chain {
		codes = append(codes, tag.Code)
	}
	return strings.Join(codes, "/"), nil
}

// GetIndustryTagNamePath 返回节点的名称路径，如 "制造业/农副食品加工业"，用于展示
func GetIndustryTagNamePath(id int) (string, error) {
	chain, err := GetIndustryTagAncestors(id)
	if err != nil {
		return "", err
	}
	names := make([]string, 0, len(chain))
	for _, tag := range chain {
		names = append(names, tag.Name)
	}
	return strings.Join(names, "/"), nil
}

// ValidateProfessionalIndustryTagID 校验职业团队提交的行业标签 id：
// 标签必须存在，且层级不低于大类（level >= 2）
func ValidateProfessionalIndustryTagID(id int) (IndustryTag, error) {
	tag, err := GetIndustryTagByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return tag, fmt.Errorf("所选行业不存在")
		}
		return tag, fmt.Errorf("校验行业标签失败: %w", err)
	}
	if tag.Level < IndustryTagMinProfessionalLevel {
		return tag, fmt.Errorf("职业团队的行业标签至少要选到「%s」", IndustryTagLevelName(IndustryTagMinProfessionalLevel))
	}
	return tag, nil
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

// DeleteIndustryTagByCode 删除行业标签（用于管理后台；存在子节点时拒绝删除）
func DeleteIndustryTagByCode(code string) error {
	tag, err := GetIndustryTagByCode(code)
	if err != nil {
		return err
	}
	count, err := CountIndustryTagChildren(tag.Id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("行业标签 %s 下仍有子节点，请先删除子节点", tag.Code)
	}
	_, err = DB.Exec("DELETE FROM industry_tags WHERE id = $1", tag.Id)
	return err
}

// CountIndustryTags 统计行业标签数量
func CountIndustryTags() (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM industry_tags").Scan(&count)
	return count, err
}

// EnsureIndustryTagExists 根据完整代码查找或创建行业标签
func EnsureIndustryTagExists(code, name string, level, parentId int, category, description string) (int, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0, fmt.Errorf("行业标签代码不能为空")
	}

	existing, err := GetIndustryTagByCode(code)
	if err == nil {
		return existing.Id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	newTag := IndustryTag{
		Code:        code,
		Name:        name,
		Level:       level,
		ParentId:    parentId,
		Category:    category,
		Description: description,
	}
	if err := newTag.Create(); err != nil {
		return 0, err
	}
	return newTag.Id, nil
}

// IndustryTagRecord 批量导入行业标签树时的单条记录
type IndustryTagRecord struct {
	Code        string // 完整标准代码，如 'C'、'C13'、'C131'
	ParentCode  string // 上一级完整代码；门类留空
	Name        string // 名称
	Description string // 说明，可为空
}

// IndustryTagLevelFromCode 根据完整代码推断层级：
// 门类为单个字母（1 位），大类为字母+2 位数字，中类为字母+3 位，小类为字母+4 位
func IndustryTagLevelFromCode(code string) int {
	code = strings.TrimSpace(code)
	if code == "" {
		return 0
	}
	switch len(code) - 1 { // 去掉门类字母
	case 0:
		return IndustryTagLevelCategory
	case 2:
		return IndustryTagLevelDivision
	case 3:
		return IndustryTagLevelGroup
	case 4:
		return IndustryTagLevelClass
	default:
		return 0
	}
}

// BulkImportIndustryTags 按代码批量导入行业标签树（幂等，以 code 为唯一键覆盖）。
// 请按层级从浅到深传入（先门类，再大类，依此类推），以便解析父节点。
func BulkImportIndustryTags(records []IndustryTagRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 预载 已有代码 -> id，便于解析父节点
	codeToId := make(map[string]int)
	rows, err := tx.Query("SELECT code, id FROM industry_tags")
	if err != nil {
		return err
	}
	for rows.Next() {
		var (
			code string
			id   int
		)
		if err := rows.Scan(&code, &id); err != nil {
			rows.Close()
			return err
		}
		codeToId[code] = id
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	const statement = `INSERT INTO industry_tags (code, name, level, parent_id, category, description)
	                   VALUES ($1, $2, $3, $4, $5, $6)
	                   ON CONFLICT (code) DO UPDATE SET
	                       name = EXCLUDED.name,
	                       level = EXCLUDED.level,
	                       parent_id = EXCLUDED.parent_id,
	                       category = EXCLUDED.category,
	                       description = EXCLUDED.description
	                   RETURNING id`
	stmt, err := tx.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, rec := range records {
		code := strings.TrimSpace(rec.Code)
		if code == "" {
			return fmt.Errorf("行业标签代码不能为空")
		}
		level := IndustryTagLevelFromCode(code)
		if level == 0 {
			return fmt.Errorf("无法从代码 %s 推断层级", code)
		}

		var parentId int
		if parentCode := strings.TrimSpace(rec.ParentCode); parentCode != "" {
			id, ok := codeToId[parentCode]
			if !ok {
				return fmt.Errorf("行业 %s 的父级代码 %s 不存在，请先导入上一层级", code, parentCode)
			}
			parentId = id
		}

		var id int
		if err := stmt.QueryRow(code, rec.Name, level, industryTagParentValue(parentId),
			code[:1], rec.Description).Scan(&id); err != nil {
			return err
		}
		codeToId[code] = id
	}

	return tx.Commit()
}
