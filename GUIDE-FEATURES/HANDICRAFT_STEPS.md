# 手工艺记录分页创建功能

## 概述
参考see-seek的分页创建模式，为handicraft实现了5步分页创建功能。

## 实现的文件

### 路由处理文件
1. **Route/route_action_handicraft.go** - 主路由处理
   - HandleNewHandicraft (GET/POST) - 第1步：基本信息
   - HandleHandicraftDetail (GET) - 详情页面

2. **Route/route_handicraft_step.go** - 分步骤路由处理
   - HandleHandicraftStep2 (GET/POST) - 第2步：难度评估
   - HandleHandicraftStep3 (GET/POST) - 第3步：开工仪式
   - HandleHandicraftStep4 (GET/POST) - 第4步：过程记录
   - HandleHandicraftStep5 (GET/POST) - 第5步：结束仪式

### 数据层文件
3. **DAO/handicraft.go** - 添加了Create方法
   - Inauguration.Create() - 创建开工仪式记录
   - ProcessRecord.Create() - 创建过程记录
   - Ending.Create() - 创建结束仪式记录

### 模板文件
4. **templates/action.handicraft.new.go.html** - 第1步：基本信息
5. **templates/action.handicraft.step2.go.html** - 第2步：难度评估
6. **templates/action.handicraft.step3.go.html** - 第3步：开工仪式
7. **templates/action.handicraft.step4.go.html** - 第4步：过程记录
8. **templates/action.handicraft.step5.go.html** - 第5步：结束仪式
9. **templates/action.handicraft.detail.go.html** - 详情页面

### 路由注册
10. **main.go** - 更新了路由注册

## 创建流程

### 步骤1：基本信息 (/v1/handicraft/new)
- 手工艺名称（必填，2-24字）
- 昵称（可选）
- 描述（必填，17-456字）
- 分类（轻体力/中等体力/重体力/轻巧力/中巧力/重巧力）
- 自动设置：策动人、主理人、项目ID

### 步骤2：难度评估 (/v1/handicraft/step2)
- 技能操作难度（1-5星）
- 创意思维难度（1-5星）

### 步骤3：开工仪式 (/v1/handicraft/step3)
- 开工仪式名称（可选）
- 开工仪式描述（可选）
- 更新状态为"进行中"

### 步骤4：过程记录 (/v1/handicraft/step4)
- 过程记录名称（可选）
- 过程记录描述（可选）

### 步骤5：结束仪式 (/v1/handicraft/step5)
- 结束仪式名称（可选）
- 结束仪式描述（可选）
- 最终状态选择（已完成/已暂停/已放弃）
- 完成后跳转到项目详情页

## 使用方式

1. 从项目详情页点击"新建手工艺"
2. 访问 `/v1/handicraft/new?uuid=<project_uuid>`
3. 按照5个步骤依次填写信息
4. 每步可以选择性填写，灵活记录
5. 完成后自动跳转回项目详情页

## 特点

- 分步骤创建，降低单次填写负担
- 进度条显示当前步骤
- 每步都可以取消返回项目页
- 支持可选字段，灵活记录
- 自动保存状态变化
- 参考see-seek的成熟模式

## 数据库表

需要确保以下表存在：
- handicrafts - 手工艺主表
- inaugurations - 开工仪式表
- process_records - 过程记录表
- endings - 结束仪式表
