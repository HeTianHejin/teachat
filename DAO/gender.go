package dao

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

// Gender 可空性别声明。
//
// 数据库 users.gender/sessions.gender 列允许 NULL，且历史上 0=女、1=男。
// 为了让“未申报/不适用”（例如 AGI 代理用户）不再被 0 冒充成“女”，
// Gender 用 GenderUnknown 表示 NULL，并与数据库 NULL 相互映射。
//
// 零值仍然是 GenderFemale（历史行为），所以凡是新建 User 都必须显式赋值。
type Gender int

const (
	GenderUnknown Gender = -1 // 未申报/不适用，对应数据库 NULL
	GenderFemale  Gender = 0  // 女
	GenderMale    Gender = 1  // 男
)

// Valid 是否已申报为可识别的性别。
func (g Gender) Valid() bool {
	return g == GenderFemale || g == GenderMale
}

func (g Gender) IsFemale() bool { return g == GenderFemale }
func (g Gender) IsMale() bool   { return g == GenderMale }

// String 返回中文描述，用于日志与页面展示。
func (g Gender) String() string {
	switch g {
	case GenderFemale:
		return "女"
	case GenderMale:
		return "男"
	default:
		return "未申报性别"
	}
}

// Honorific 返回称谓。未申报时返回空字符串，调用方必须显式处理，
// 不允许再默认成“女士”或“先生”。
func (g Gender) Honorific() string {
	switch g {
	case GenderFemale:
		return "女士"
	case GenderMale:
		return "先生"
	default:
		return ""
	}
}

// IconClass 返回 bootstrap-icons 的图标类名。
func (g Gender) IconClass() string {
	switch g {
	case GenderFemale:
		return "bi-gender-female"
	case GenderMale:
		return "bi-gender-male"
	default:
		return "bi-question-circle"
	}
}

// ParseGender 解析表单或文本来源的性别值。
func ParseGender(raw string) (Gender, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "0", "f", "female", "女":
		return GenderFemale, nil
	case "1", "m", "male", "男":
		return GenderMale, nil
	default:
		return GenderUnknown, fmt.Errorf("无法识别的性别值: %q", raw)
	}
}

// Scan 实现 sql.Scanner：NULL -> GenderUnknown。
func (g *Gender) Scan(src any) error {
	if src == nil {
		*g = GenderUnknown
		return nil
	}

	var value int64
	switch v := src.(type) {
	case int64:
		value = v
	case int32:
		value = int64(v)
	case int:
		value = int64(v)
	case []byte:
		n, err := strconv.ParseInt(strings.TrimSpace(string(v)), 10, 64)
		if err != nil {
			return fmt.Errorf("无法把 %q 解析为 gender: %w", string(v), err)
		}
		value = n
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return fmt.Errorf("无法把 %q 解析为 gender: %w", v, err)
		}
		value = n
	default:
		return fmt.Errorf("不支持把 %T 扫描为 gender", src)
	}

	parsed := Gender(value)
	if !parsed.Valid() {
		parsed = GenderUnknown
	}
	*g = parsed
	return nil
}

// Value 实现 driver.Valuer：GenderUnknown -> NULL。
func (g Gender) Value() (driver.Value, error) {
	if !g.Valid() {
		return nil, nil
	}
	return int64(g), nil
}
