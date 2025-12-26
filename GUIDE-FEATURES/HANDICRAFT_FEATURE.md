# 手艺管理功能

## 功能概述

手艺管理功能是技能和法力的综合应用，允许用户在项目中创建、管理和展示手工艺作业。手艺结合了技能操作和创意思维，是项目实施的重要组成部分。

## 文件结构

### 数据模型
- `DAO/handicraft.go` - 手艺数据模型和数据库操作方法
- `DAO/skill.go` - 技能数据模型（支持手艺）
- `DAO/magic.go` - 法力数据模型（支持手艺）
- `DAO/evidence.go` - 凭据数据模型（支持手艺）

### 路由处理
- `Route/route_action_handicraft.go` - 手艺相关的HTTP请求处理

### 模板文件
- `templates/handicraft.new.go.html` - 新建手艺页面
- `templates/handicraft.detail.go.html` - 手艺详情页面
- `templates/handicraft.list.go.html` - 手艺列表页面
- `templates/component_handicraft_bean.go.html` - 手艺组件模板

### 数据库
- `sql/handicrafts_table.sql` - 手艺表创建脚本
- `sql/setup_db_tables.sql` - 已更新包含手艺表

## 功能特性

### 手艺分类
1. **轻体力** - 普通人都可以完成，如：喂水、简单清洁
2. **中等体力** - 介于轻重之间，如：更换水桶、搬家具
3. **重体力** - 需要较高强度体能，如：搬运洗衣机
4. **轻巧力** - 需要精细手艺，如：刺绣、精细木工
5. **中巧力** - 中等体力+中等技能，如：家电维修
6. **重巧力** - 特定体能+载重力，如：铁艺制作

### 手艺状态
- **未开始** - 初始状态
- **进行中** - 正在执行
- **已暂停** - 中途暂停
- **已完成** - 顺利结束
- **已放弃** - 因故未完成

### 二维难度系统
- **技能难度** (1-5): 操作技术要求
- **创意难度** (1-5): 思维创新要求
- **综合难度**: 自动计算整体难度
- **难度特征**: 智能分析难度类型

### 关联系统
- **项目关联**: 每个手艺都属于特定项目
- **技能关联**: 可关联多个相关技能
- **法力关联**: 可关联多个相关法力
- **凭据关联**: 可上传相关证据材料

### 人员管理
- **记录人**: 创建手艺记录的用户
- **策动人**: 发起手艺的用户
- **主理人**: 执行手艺的主要负责人
- **协助者**: 参与协助的用户列表

## 页面功能

### 1. 新建手艺 (`/v1/handicraft/new?project_uuid=xxx`)
- 基于项目创建手艺
- 设置手艺基本信息和分级
- 选择相关技能和法力
- 配置人员角色

### 2. 手艺详情 (`/v1/handicraft/detail?uuid=xxx`)
- 展示手艺完整信息
- 显示难度分析和特征
- 展示关联的技能、法力和凭据
- 提供项目跳转链接

### 3. 手艺列表 (`/v1/handicrafts/list`)
- 展示所有手艺记录
- 使用组件化展示
- 支持状态和难度筛选

## 使用方法

### 1. 数据库初始化
```sql
-- 执行手艺表创建脚本
\i sql/handicrafts_table.sql
```

### 2. 访问功能
- 新建手艺: `http://localhost:8000/v1/handicraft/new?project_uuid=xxx`
- 手艺列表: `http://localhost:8000/v1/handicrafts/list`
- 手艺详情: `http://localhost:8000/v1/handicraft/detail?uuid=xxx`

### 3. 路由注册
在 `main.go` 中已注册以下路由：
```go
mux.HandleFunc("/v1/handicraft/new", route.HandleNewHandicraft)
mux.HandleFunc("/v1/handicraft/detail", route.HandleHandicraftDetail)
mux.HandleFunc("/v1/handicrafts/list", route.HandleHandicraftList)
```

## 数据结构

### Handicraft 结构体
```go
type Handicraft struct {
    Id              int
    Uuid            string
    RecorderUserId  int                 // 记录人ID
    Name            string
    Nickname        string
    Description     string
    ProjectId       int                 // 项目ID
    InitiatorId     int                 // 策动人ID
    OwnerId         int                 // 主理人ID
    Category        HandicraftCategory  // 分类
    Status          HandicraftStatus    // 状态
    SkillDifficulty int                 // 技能难度(1-5)
    MagicDifficulty int                 // 创意难度(1-5)
    CreatedAt       time.Time
    UpdatedAt       *time.Time
    DeletedAt       *time.Time          // 软删除
}
```

## 与其他功能的关系

| 功能 | 关系 | 说明 |
|------|------|------|
| 技能(Skill) | 支撑关系 | 手艺需要技能支持 |
| 法力(Magic) | 支撑关系 | 手艺需要法力支持 |
| 项目(Project) | 归属关系 | 手艺属于特定项目 |
| 凭据(Evidence) | 证明关系 | 凭据证明手艺完成 |

## 业务流程

1. **创建阶段**: 在项目中创建手艺，选择相关技能和法力
2. **执行阶段**: 更新手艺状态，记录执行过程
3. **完成阶段**: 上传凭据证据，标记完成状态
4. **评估阶段**: 分析难度特征，总结经验教训

## 扩展功能

可以进一步扩展的功能：
1. 手艺模板系统
2. 手艺评估和评分
3. 手艺学习路径规划
4. 手艺协作和分工
5. 手艺成果展示

## 注意事项

1. 手艺必须关联到具体项目
2. 支持软删除，删除的手艺不会物理删除
3. 难度系统采用二维评估，更加精确
4. 模板使用Bootstrap 3样式，保持界面一致性
5. 手艺状态管理支持完整的生命周期