# 法力管理功能

## 功能概述

法力管理功能允许用户记录、管理和展示个人的思维能力和智慧技能。法力分为理性和感性两大类，每个法力都有智力要求、掌握难度和法力等级等属性。

## 文件结构

### 数据模型
- `DAO/magic.go` - 法力数据模型和数据库操作方法

### 路由处理
- `Route/route_action_magic.go` - 法力相关的HTTP请求处理

### 模板文件
- `templates/magic.new.go.html` - 新建法力页面
- `templates/magic.detail.go.html` - 法力详情页面
- `templates/magic.list.go.html` - 法力列表页面
- `templates/component_magic_bean.go.html` - 法力组件模板

### 数据库
- `sql/magics_table.sql` - 法力表创建脚本
- `sql/setup_db_tables.sql` - 已更新包含法力表

## 功能特性

### 法力分类
1. **理性** - 逻辑思维、分析推理等
2. **感性** - 创意思维、艺术感知等

### 法力属性
- **智力要求等级** (1-5): 极低、低、中等、高、极高智力需求
- **掌握难度等级** (1-5): 极易、较易、中等、较难、极难
- **法力等级** (1-5): 入门、初级、中级、高级、专家

### 智力等级详细说明
- **极低智力需求**: 机械性重复操作，无需复杂思考
- **低智力需求**: 基础理解，简单问题解决
- **中等智力需求**: 系统学习，多因素分析
- **高智力需求**: 深度专业知识，创造性思维
- **极高智力需求**: 顶尖专业水平，突破性思维

### 页面功能
1. **新建法力** (`/v1/magic/new`)
   - 填写法力基本信息
   - 设置法力分级属性
   - 表单验证和错误处理

2. **法力详情** (`/v1/magic/detail?id=123`)
   - 展示法力完整信息
   - 可视化进度条显示各项属性
   - 智力等级详细说明

3. **法力列表** (`/v1/magic/list`)
   - 展示所有法力
   - 使用组件化展示
   - 支持添加新法力

## 使用方法

### 1. 数据库初始化
```sql
-- 执行法力表创建脚本
\i sql/magics_table.sql
```

### 2. 访问功能
- 新建法力: `http://localhost:8000/v1/magic/new`
- 法力列表: `http://localhost:8000/v1/magic/list`
- 法力详情: `http://localhost:8000/v1/magic/detail?id=1`

### 3. 路由注册
在 `main.go` 中已注册以下路由：
```go
mux.HandleFunc("/v1/magic/new", route.HandleNewMagic)
mux.HandleFunc("/v1/magic/detail", route.HandleMagicDetail)
mux.HandleFunc("/v1/magic/list", route.HandleMagicList)
```

## 数据结构

### Magic 结构体
```go
type Magic struct {
    Id                int
    Uuid              string
    UserId            int                 // 创建者用户ID
    Name              string
    Nickname          string
    Description       string
    IntelligenceLevel IntelligenceLevel   // 智力耗费等级(1-5)
    DifficultyLevel   DifficultyLevel     // 掌握难度等级(1-5)
    Category          MagicCategory       // 分类：理性/感性
    Level             int                 // 等级
    CreatedAt         time.Time
    UpdatedAt         *time.Time
    DeletedAt         *time.Time          // 软删除
}
```

## 与技能管理的区别

| 特征 | 技能(Skill) | 法力(Magic) |
|------|-------------|-------------|
| 主要属性 | 体力要求 | 智力要求 |
| 分类 | 软技能/硬技能 | 理性/感性 |
| 侧重点 | 操作能力 | 思维能力 |
| 应用场景 | 具体技术操作 | 创意和分析 |

## 扩展功能

可以进一步扩展的功能：
1. 法力搜索和筛选
2. 法力组合和协同
3. 法力学习路径规划
4. 法力评估和认证
5. 法力分享和交流

## 注意事项

1. 所有法力操作都需要用户登录
2. 支持软删除，删除的法力不会从数据库中物理删除
3. 法力等级和难度使用枚举类型，确保数据一致性
4. 模板使用Bootstrap 3样式，保持界面一致性
5. 智力等级包含详细的描述信息，帮助用户理解不同等级的含义