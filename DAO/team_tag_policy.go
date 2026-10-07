package dao

import (
	"fmt"
	"strings"
)

// 业余团队自由标签的限制
// 业余团队体验优先：超出限制的部分静默裁剪，不返回错误
const (
	AmateurTagMaxCount = 5  // 最多保留 5 个标签
	AmateurTagMaxRunes = 12 // 单个标签最多 12 个字符
)

// NormalizeAmateurTags 规范化业余团队的自由标签：去空白、去空值、去重、限长限数。
// 体验优先：不做校验、不返回错误，超出限制的部分直接丢弃。
func NormalizeAmateurTags(raw string) string {
	parts := SplitTags(raw)
	result := make([]string, 0, AmateurTagMaxCount)
	seen := make(map[string]struct{}, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// 单个标签限长
		if runes := []rune(p); len(runes) > AmateurTagMaxRunes {
			p = string(runes[:AmateurTagMaxRunes])
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		result = append(result, p)
		if len(result) >= AmateurTagMaxCount {
			break
		}
	}

	return strings.Join(result, ",")
}

// NormalizeAndValidateTeamTags 团队标签统一入口，按 team.nature 分派到两条轨道：
//
//   - 业余团队（TeamNatureAmateur）：用户自由填写，体验优先，
//     静默裁剪，永不返回错误，只写 teams.tags。
//   - 职业团队（TeamNatureProfessional）：必须提交行业白名单标签 id（明牌实码），
//     校验失败返回可直接展示给用户的中文错误；同时产出 teams.primary_industry_tag_id
//     与 teams.industry_path（代码路径，如 "C/C13"）用于前缀上卷搜索。
//
// 返回值：写入 teams.tags 的展示快照、主行业标签 id、行业代码路径。
func NormalizeAndValidateTeamTags(nature int, freeTags string, industryTagID int) (tags string, primaryIndustryTagID int, industryPath string, err error) {
	if nature != TeamNatureProfessional {
		// 业余团队：自由标签，不做校验
		return NormalizeAmateurTags(freeTags), 0, "", nil
	}

	// 职业团队：校验白名单 id（必须存在且至少到大类）
	tag, err := ValidateProfessionalIndustryTagID(industryTagID)
	if err != nil {
		return "", 0, "", err
	}

	// 展示快照用行业名称路径，如 "制造业/农副食品加工业"
	namePath, err := GetIndustryTagNamePath(tag.Id)
	if err != nil {
		return "", 0, "", fmt.Errorf("生成行业名称路径失败: %w", err)
	}

	// 搜索用代码路径，如 "C/C13"
	codePath, err := GetIndustryTagPath(tag.Id)
	if err != nil {
		return "", 0, "", fmt.Errorf("生成行业代码路径失败: %w", err)
	}

	return namePath, tag.Id, codePath, nil
}
