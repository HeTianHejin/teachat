# 技能管理功能

## 功能概述

技能管理功能允许用户记录、管理和展示个人技能信息。技能分为软技能和硬技能两大类，每个技能都有体力要求、掌握难度和技能等级等属性。

## 文件结构

### 数据模型
- `DAO/skill.go` - 技能数据模型和数据库操作方法

### 路由处理
- `Route/route_action_skill.go` - 技能相关的HTTP请求处理

### 模板文件
- `templates/skill.new.go.html` - 新建技能页面
- `templates/skill.detail.go.html` - 技能详情页面
- `templates/skill.user_list.go.html` - 用户个人技能列表页面
- `templates/component_skill_bean.go.html` - 技能组件模板

### 数据库
- `sql/skills_table.sql` - 技能表创建脚本
- `sql/setup_db_tables.sql` - 已更新包含技能表

## 功能特性

### 技能分类
1. **通用软技能** - 如沟通、健康与情绪管理等
2. **通用硬技能** - 可以设立试卷考试的科目技能，如驾驶车辆、控制计算机等

### 技能属性
- **体力要求等级** (1-5): 极低、较低、中等、较高、极高
- **掌握难度等级** (1-5): 极易、较易、中等、较难、极难
- **技能等级** (1-5): 入门、初级、中级、高级、专家

### 页面功能
1. **新建技能** (`/v1/skill/new`)
   - 填写技能基本信息
   - 设置技能分级属性
   - 表单验证和错误处理

2. **技能详情** (`/v1/skill/detail?id=123`)
   - 展示技能完整信息
   - 可视化进度条显示各项属性
   - 技能等级说明

3. **技能列表** (`/v1/skills/user_list`)
   - 展示所有技能
   - 使用组件化展示
   - 支持添加新技能

## 使用方法

### 1. 数据库初始化
```sql
-- 执行技能表创建脚本
\i sql/skills_table.sql
```

### 2. 访问功能
- 新建技能: `http://localhost:8000/v1/skill/new`
- 技能列表: `http://localhost:8000/v1/skills/user_list`
- 技能详情: `http://localhost:8000/v1/skill/detail?id=1`

### 3. 路由注册
在 `main.go` 中已注册以下路由：
```go
mux.HandleFunc("/v1/skill/new", route.HandleNewSkill)
mux.HandleFunc("/v1/skill/detail", route.HandleSkillDetail)
mux.HandleFunc("/v1/skills/user_list", route.HandleSkillsUserList)
```

## 数据结构

### Skill 结构体
```go
type Skill struct {
    Id              int
    Uuid            string
    Name            string
    Nickname        string
    Description     string
    StrengthLevel   StrengthLevel   // 体力耗费等级(1-5)
    DifficultyLevel DifficultyLevel // 掌握难度等级(1-5)
    Category        SkillCategory   // 分类
    Level           int             // 等级
    CreatedAt       time.Time
    UpdatedAt       *time.Time
    DeletedAt       *time.Time      // 软删除
}
```

## 扩展功能

可以进一步扩展的功能：
1. 技能搜索和筛选
2. 技能标签系统
3. 技能学习进度跟踪
4. 技能认证和评估
5. 技能分享和交流

## 注意事项

1. 所有技能操作都需要用户登录
2. 支持软删除，删除的技能不会从数据库中物理删除
3. 技能等级和难度使用枚举类型，确保数据一致性
4. 模板使用Bootstrap 3样式，保持界面一致性