package data

import "time"

// 作业场所安全隐患，
// 直接责任方默认是归属场所管理茶团（相对的“risk风险”默认责任是作业执行团队方）
// 识别安全隐患的能力也是一种magic
type Hazard struct {
	Id     int
	Uuid   string
	UserId int //记录人id

	Name        string //隐患名称
	Nickname    string //隐患别名
	Keywords    string //隐患关键词
	Description string //隐患描述
	Source      string //隐患来源

	// 分级管理
	Severity int            // 隐患严重度（1-5级）
	Category HazardCategory // 隐患类型枚举（电气/机械/化学等）

	CreatedAt time.Time
	UpdatedAt *time.Time
}
type HazardCategory int

const (
	HazardCategoryElectrical = iota + 1 //电气
	HazardCategoryMechanical            //机械
	HazardCategoryChemical              //化学
	HazardCategoryBiological            //生物
	HazardCategoryErgonomic             //工效学，人机工程学
	HazardCategoryOther                 //其他
)

func (h *Hazard) CategoryName() string {
	switch h.Category {
	case HazardCategoryElectrical:
		return "电气"
	case HazardCategoryMechanical:
		return "机械"
	case HazardCategoryChemical:
		return "化学"
	case HazardCategoryBiological:
		return "生物"
	case HazardCategoryErgonomic:
		return "工效学"
	case HazardCategoryOther:
		return "其他"
	default:
		return "未知"
	}
}

const (
	HazardSeverityNegligible = 1 // 可忽略
	HazardSeverityLow        = 2 // 低风险
	HazardSeverityMedium     = 3 // 中风险
	HazardSeverityHigh       = 4 // 高风险
	HazardSeverityCritical   = 5 // 危急
)

func (h *Hazard) SeverityName() string {
	switch h.Severity {
	case HazardSeverityNegligible:
		return "可忽略"
	case HazardSeverityLow:
		return "低风险"
	case HazardSeverityMedium:
		return "中风险"
	case HazardSeverityHigh:
		return "高风险"
	case HazardSeverityCritical:
		return "危急"
	default:
		return "未知"
	}
}

// 安全防范措施
type SafetyMeasure struct {
	Id       int
	Uuid     string
	HazardId int // 关联的隐患ID
	UserId   int // 负责人ID

	Title       string // 措施标题
	Description string // 措施描述
	Priority    int    // 优先级（1-5）
	Status      int    // 状态（计划中/执行中/已完成）

	PlannedDate   *time.Time // 计划执行时间
	CompletedDate *time.Time // 完成时间
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}

const (
	MeasureStatusPlanned    = 1 // 计划中
	MeasureStatusInProgress = 2 // 执行中
	MeasureStatusCompleted  = 3 // 已完成
	MeasureStatusCancelled  = 4 // 已取消
)

func (m *SafetyMeasure) StatusName() string {
	switch m.Status {
	case MeasureStatusPlanned:
		return "计划中"
	case MeasureStatusInProgress:
		return "执行中"
	case MeasureStatusCompleted:
		return "已完成"
	case MeasureStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

func (m *SafetyMeasure) PriorityName() string {
	switch m.Priority {
	case 1:
		return "低"
	case 2:
		return "较低"
	case 3:
		return "中等"
	case 4:
		return "较高"
	case 5:
		return "高"
	default:
		return "未知"
	}
}

// 获取隐患的所有防范措施
func (h *Hazard) GetSafetyMeasures() ([]SafetyMeasure, error) {
	rows, err := Db.Query("SELECT id, uuid, hazard_id, user_id, title, description, priority, status, planned_date, completed_date, created_at, updated_at FROM safety_measures WHERE hazard_id = $1 ORDER BY priority DESC, created_at DESC", h.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measures []SafetyMeasure
	for rows.Next() {
		var m SafetyMeasure
		err := rows.Scan(&m.Id, &m.Uuid, &m.HazardId, &m.UserId, &m.Title, &m.Description, &m.Priority, &m.Status, &m.PlannedDate, &m.CompletedDate, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		measures = append(measures, m)
	}
	return measures, nil
}

// 创建安全防范措施
func (m *SafetyMeasure) Create() error {
	statement := "INSERT INTO safety_measures (uuid, hazard_id, user_id, title, description, priority, status, planned_date, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	m.Uuid = Random_UUID()
	err = stmt.QueryRow(m.Uuid, m.HazardId, m.UserId, m.Title, m.Description, m.Priority, m.Status, m.PlannedDate, time.Now()).Scan(&m.Id, &m.CreatedAt)
	return err
}

// 更新安全防范措施
func (m *SafetyMeasure) Update() error {
	statement := "UPDATE safety_measures SET title = $2, description = $3, priority = $4, status = $5, planned_date = $6, completed_date = $7, updated_at = $8 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Id, m.Title, m.Description, m.Priority, m.Status, m.PlannedDate, m.CompletedDate, time.Now())
	return err
}

// 根据ID获取安全防范措施
func GetSafetyMeasureById(id int) (SafetyMeasure, error) {
	var m SafetyMeasure
	err := Db.QueryRow("SELECT id, uuid, hazard_id, user_id, title, description, priority, status, planned_date, completed_date, created_at, updated_at FROM safety_measures WHERE id = $1", id).Scan(&m.Id, &m.Uuid, &m.HazardId, &m.UserId, &m.Title, &m.Description, &m.Priority, &m.Status, &m.PlannedDate, &m.CompletedDate, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

// 删除安全防范措施
func (m *SafetyMeasure) Delete() error {
	statement := "DELETE FROM safety_measures WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(m.Id)
	return err
}

// 创建隐患
func (h *Hazard) Create() error {
	statement := "INSERT INTO hazards (uuid, user_id, name, nickname, keywords, description, source, severity, category, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	h.Uuid = Random_UUID()
	err = stmt.QueryRow(h.Uuid, h.UserId, h.Name, h.Nickname, h.Keywords, h.Description, h.Source, h.Severity, h.Category, time.Now()).Scan(&h.Id, &h.CreatedAt)
	return err
}

// 更新隐患
func (h *Hazard) Update() error {
	statement := "UPDATE hazards SET name = $2, nickname = $3, keywords = $4, description = $5, source = $6, severity = $7, category = $8, updated_at = $9 WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(h.Id, h.Name, h.Nickname, h.Keywords, h.Description, h.Source, h.Severity, h.Category, time.Now())
	return err
}

// 根据ID获取隐患
func GetHazardById(id int) (Hazard, error) {
	var h Hazard
	err := Db.QueryRow("SELECT id, uuid, user_id, name, nickname, keywords, description, source, severity, category, created_at, updated_at FROM hazards WHERE id = $1", id).Scan(&h.Id, &h.Uuid, &h.UserId, &h.Name, &h.Nickname, &h.Keywords, &h.Description, &h.Source, &h.Severity, &h.Category, &h.CreatedAt, &h.UpdatedAt)
	return h, err
}

// 删除隐患（同时删除相关的防范措施）
func (h *Hazard) Delete() error {
	// 先删除相关的防范措施
	_, err := Db.Exec("DELETE FROM safety_measures WHERE hazard_id = $1", h.Id)
	if err != nil {
		return err
	}
	// 再删除隐患本身
	statement := "DELETE FROM hazards WHERE id = $1"
	stmt, err := Db.Prepare(statement)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(h.Id)
	return err
}

// 获取所有隐患
func GetAllHazards() ([]Hazard, error) {
	rows, err := Db.Query("SELECT id, uuid, user_id, name, nickname, keywords, description, source, severity, category, created_at, updated_at FROM hazards ORDER BY severity DESC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hazards []Hazard
	for rows.Next() {
		var h Hazard
		err := rows.Scan(&h.Id, &h.Uuid, &h.UserId, &h.Name, &h.Nickname, &h.Keywords, &h.Description, &h.Source, &h.Severity, &h.Category, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			return nil, err
		}
		hazards = append(hazards, h)
	}
	return hazards, nil
}
