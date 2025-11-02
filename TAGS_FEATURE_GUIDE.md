# 团队和集团分类标签功能说明

## 功能概述

为团队（Team）和集团（Group）添加分类标签（Tags）功能，帮助用户快速找到同行业、同兴趣的组织。

## 设计思路

### 标签格式
- 使用逗号分隔的字符串存储多个标签
- 示例：`"诗词书法,文化艺术"` 或 `"家电维修,上门服务"`

### 数据库字段
```sql
-- 团队表添加tags字段
ALTER TABLE teams ADD COLUMN tags VARCHAR(200) DEFAULT '';

-- 集团表添加tags字段  
ALTER TABLE groups ADD COLUMN tags VARCHAR(200) DEFAULT '';
```

## 标签示例

### 文化艺术类
- 诗词书法
- 绘画摄影
- 音乐舞蹈
- 戏曲曲艺

### 技术服务类
- 家电维修
- 电脑维护
- 水电安装
- 装修设计

### 教育培训类
- 语言培训
- 职业技能
- 兴趣爱好
- 学历教育

### 商业贸易类
- 电子商务
- 批发零售
- 进出口贸易
- 物流配送

### 专业服务类
- 法律咨询
- 财务会计
- 人力资源
- 市场营销

### 医疗健康类
- 中医养生
- 健身运动
- 心理咨询
- 营养膳食

### 科技研发类
- 软件开发
- 硬件制造
- 人工智能
- 物联网

## 功能实现

### 1. 数据结构修改

**Team 结构（DAO/team.go）**
```go
type Team struct {
    Id           int
    Uuid         string
    Name         string
    Mission      string
    FounderId    int
    Class        int
    Abbreviation string
    Logo         string
    Tags         string // 新增：分类标签，逗号分隔
    CreatedAt    time.Time
    UpdatedAt    *time.Time
    DeletedAt    *time.Time
}
```

**Group 结构（DAO/group.go）**
```go
type Group struct {
    Id           int
    Uuid         string
    Name         string
    Abbreviation string
    Mission      string
    FounderId    int
    FirstTeamId  int
    Class        int
    Logo         string
    Tags         string // 新增：分类标签，逗号分隔
    CreatedAt    time.Time
    UpdatedAt    *time.Time
    DeletedAt    *time.Time
}
```

### 2. 标签辅助函数

```go
// GetTags 获取标签数组
func (team *Team) GetTags() []string {
    if team.Tags == "" {
        return []string{}
    }
    tags := strings.Split(team.Tags, ",")
    result := make([]string, 0, len(tags))
    for _, tag := range tags {
        trimmed := strings.TrimSpace(tag)
        if trimmed != "" {
            result = append(result, trimmed)
        }
    }
    return result
}

// SetTags 设置标签
func (team *Team) SetTags(tags []string) {
    team.Tags = strings.Join(tags, ",")
}

// HasTag 检查是否包含某个标签
func (team *Team) HasTag(tag string) bool {
    tags := team.GetTags()
    for _, t := range tags {
        if t == tag {
            return true
        }
    }
    return false
}
```

### 3. 搜索功能

```go
// SearchTeamsByTag 根据标签搜索团队
func SearchTeamsByTag(tag string) ([]Team, error) {
    query := `SELECT id, uuid, name, mission, founder_id, created_at, class, 
              abbreviation, logo, tags, updated_at 
              FROM teams 
              WHERE tags LIKE $1 AND deleted_at IS NULL 
              ORDER BY created_at DESC`
    
    rows, err := db.Query(query, "%"+tag+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    teams := make([]Team, 0)
    for rows.Next() {
        var team Team
        err = rows.Scan(&team.Id, &team.Uuid, &team.Name, &team.Mission,
            &team.FounderId, &team.CreatedAt, &team.Class, &team.Abbreviation,
            &team.Logo, &team.Tags, &team.UpdatedAt)
        if err != nil {
            return nil, err
        }
        teams = append(teams, team)
    }
    return teams, nil
}

// SearchGroupsByTag 根据标签搜索集团
func SearchGroupsByTag(tag string) ([]Group, error) {
    query := `SELECT id, uuid, name, abbreviation, mission, founder_id, 
              first_team_id, class, logo, tags, created_at, updated_at 
              FROM groups 
              WHERE tags LIKE $1 AND deleted_at IS NULL 
              ORDER BY created_at DESC`
    
    rows, err := db.Query(query, "%"+tag+"%")
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    groups := make([]Group, 0)
    for rows.Next() {
        var group Group
        err = rows.Scan(&group.Id, &group.Uuid, &group.Name, &group.Abbreviation,
            &group.Mission, &group.FounderId, &group.FirstTeamId, &group.Class,
            &group.Logo, &group.Tags, &group.CreatedAt, &group.UpdatedAt)
        if err != nil {
            return nil, err
        }
        groups = append(groups, group)
    }
    return groups, nil
}
```

### 4. 更新方法

```go
// UpdateTags 更新团队标签
func (team *Team) UpdateTags() error {
    statement := `UPDATE teams SET tags = $1, updated_at = $2 WHERE id = $3`
    stmt, err := db.Prepare(statement)
    if err != nil {
        return err
    }
    defer stmt.Close()
    _, err = stmt.Exec(team.Tags, time.Now(), team.Id)
    return err
}

// UpdateTags 更新集团标签
func (group *Group) UpdateTags() error {
    statement := `UPDATE groups SET tags = $1, updated_at = $2 WHERE id = $3`
    stmt, err := db.Prepare(statement)
    if err != nil {
        return err
    }
    defer stmt.Close()
    _, err = stmt.Exec(group.Tags, time.Now(), group.Id)
    return err
}
```

## 用户界面

### 1. 创建/编辑表单

```html
<div class="form-group">
    <label for="tags">分类标签（用逗号分隔，如：诗词书法,文化艺术）</label>
    <input type="text" class="form-control" name="tags" id="tags" 
           placeholder="请输入分类标签，多个标签用逗号分隔" 
           maxlength="200" />
    <p class="help-block">标签帮助其他用户快速找到同行业的团队</p>
</div>
```

### 2. 详情页显示

```html
{{ if .Team.Tags }}
<div class="tags">
    <i class="glyphicon glyphicon-tags"></i>
    {{ range .Team.GetTags }}
    <span class="label label-info">{{ . }}</span>
    {{ end }}
</div>
{{ end }}
```

### 3. 搜索页面

```html
<form action="/v1/search/by_tag" method="get">
    <div class="input-group">
        <input type="text" class="form-control" name="tag" 
               placeholder="输入标签搜索同行..." />
        <span class="input-group-btn">
            <button class="btn btn-default" type="submit">
                <i class="glyphicon glyphicon-search"></i> 搜索
            </button>
        </span>
    </div>
</form>

<!-- 热门标签 -->
<div class="hot-tags">
    <h4>热门标签</h4>
    <a href="/v1/search/by_tag?tag=诗词书法" class="btn btn-sm btn-default">诗词书法</a>
    <a href="/v1/search/by_tag?tag=家电维修" class="btn btn-sm btn-default">家电维修</a>
    <a href="/v1/search/by_tag?tag=软件开发" class="btn btn-sm btn-default">软件开发</a>
    <!-- 更多热门标签... -->
</div>
```

## 使用场景

### 场景1：创建团队时添加标签
1. 用户创建新团队
2. 在表单中输入标签："家电维修,上门服务"
3. 保存后，其他用户可以通过这些标签找到该团队

### 场景2：搜索同行团队
1. 用户想找家电维修相关的团队
2. 在搜索框输入"家电维修"
3. 系统返回所有包含该标签的团队列表

### 场景3：浏览热门标签
1. 用户访问搜索页面
2. 看到热门标签列表
3. 点击感兴趣的标签，查看相关团队

## 优势

1. **快速定位**：通过标签快速找到同行业组织
2. **精准匹配**：标签比名称搜索更精准
3. **分类清晰**：标签提供清晰的分类体系
4. **易于扩展**：可以随时添加新标签
5. **用户友好**：简单直观的标签系统

## 注意事项

1. **标签长度限制**：建议单个标签不超过8个字符
2. **标签数量限制**：建议每个团队/集团不超过5个标签
3. **标签规范**：建议使用常见的行业术语
4. **重复标签**：系统应该去重，避免重复标签
5. **标签审核**：可以考虑对标签进行审核，避免不当内容

## 数据库迁移

```sql
-- 为现有团队表添加tags字段
ALTER TABLE teams ADD COLUMN IF NOT EXISTS tags VARCHAR(200) DEFAULT '';

-- 为现有集团表添加tags字段
ALTER TABLE groups ADD COLUMN IF NOT EXISTS tags VARCHAR(200) DEFAULT '';

-- 创建标签索引以提高搜索性能
CREATE INDEX IF NOT EXISTS idx_teams_tags ON teams USING gin(to_tsvector('simple', tags));
CREATE INDEX IF NOT EXISTS idx_groups_tags ON groups USING gin(to_tsvector('simple', tags));
```

## 未来扩展

1. **标签统计**：显示每个标签的使用次数
2. **标签推荐**：根据团队内容自动推荐标签
3. **标签云**：可视化展示热门标签
4. **标签关联**：显示相关标签
5. **标签管理**：管理员可以管理标签库

## 总结

标签功能是一个简单但非常实用的功能，可以大大提升用户查找同行组织的效率。通过合理的标签设计和搜索功能，用户可以快速建立行业联系，促进协作。
