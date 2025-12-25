# 茶叶支付系统实现文档

## 概述

茶叶支付系统为星际茶棚提供了完整的虚拟货币转账功能，以茶叶为基本单位（1克），支持用户间的安全转账交易。

## 已完成功能

### 1. 数据库设计

#### 核心表结构
- **tea_accounts**: 用户茶叶账户表
  - 存储用户茶叶余额（精确到毫克）
  - 账户状态管理（正常/冻结）
  - 冻结原因记录

- **tea_transfers**: 茶叶转账记录表
  - 转账发起和确认的双阶段机制
  - 支持备注和拒绝原因
  - 自动过期机制

- **tea_transactions**: 交易流水表
  - 详细记录所有资金变动
  - 支持多种交易类型
  - 完整的余额变动追踪

#### 业务约束
- 金额不能为负数
- 账户状态枚举约束
- 外键关联完整性
- 自动时间戳更新

### 2. 后端实现

#### DAO层 (`DAO/tea_account.go`)
- **TeaAccount**: 账户管理
  - 获取账户信息
  - 创建账户
  - 状态更新（冻结/解冻）
  - 系统余额调整

- **TeaTransfer**: 转账管理
  - 创建转账记录
  - 确认接收转账
  - 拒绝转账
  - 获取转账历史

- **TeaTransaction**: 流水管理
  - 记录所有交易
  - 多维度查询支持
  - 关联转账追踪

#### 路由层 (`Route/route_tea_user_account.go`)
- **账户相关API**:
  - `GET /v1/tea/user/account` - 获取账户信息
  - `POST /v1/tea/user/account/freeze` - 冻结账户（管理员）
  - `POST /v1/tea/user/account/unfreeze` - 解冻账户（管理员）

- **转账相关API**:
  - `POST /v1/tea/user/transfer/new` - 发起转账
  - `POST /v1/tea/user/transfer/confirm` - 确认接收
  - `POST /v1/tea/user/transfer/reject` - 拒绝转账
  - `GET /v1/tea/user/transfers/pending` - 待确认转账列表
  - `GET /v1/tea/user/transfers/history` - 转账历史

- **流水相关API**:
  - `GET /v1/tea/user/transactions` - 交易流水

### 3. 安全机制

#### 事务安全
- 使用数据库事务确保转账原子性
- 转账时锁定账户记录防止并发问题
- 余额检查防止超支

#### 状态管理
- 转账四状态管理：pending, confirmed, rejected, expired
- 自动过期机制
- 账户冻结功能

#### 权限控制
- 用户身份验证
- 管理员权限检查
- 转账权限验证

### 4. 业务逻辑

#### 转账流程
1. **发起转账**: 验证余额，创建待确认记录
2. **确认接收**: 扣除转出方余额，增加接收方余额
3. **拒绝转账**: 更新转账状态，资金不变动
4. **过期处理**: 超时自动失效

#### 余额管理
- 精确到毫克的余额计算
- 实时余额更新
- 完整的变动记录

## API接口详情

### 获取账户信息
```
GET /v1/tea/user/account
响应:
{
  "success": true,
  "message": "获取账户信息成功",
  "data": {
    "uuid": "xxx",
    "user_id": 1,
    "balance_grams": 100.500,
    "status": "normal",
    "created_at": "2024-01-01 12:00:00"
  }
}
```

### 发起转账
```
POST /v1/tea/user/transfer/new
请求:
{
  "to_user_id": 2,
  "amount_grams": 10.5,
  "notes": "茶叶转账测试",
  "expire_hours": 24
}
响应:
{
  "success": true,
  "message": "转账发起成功",
  "data": {
    "uuid": "xxx",
    "from_user_id": 1,
    "to_user_id": 2,
    "amount_grams": 10.5,
    "status": "pending",
    "expires_at": "2024-01-02 12:00:00"
  }
}
```

### 确认接收
```
POST /v1/tea/user/transfer/confirm
请求:
{
  "transfer_uuid": "xxx"
}
响应:
{
  "success": true,
  "message": "转账确认成功"
}
```

### 拒绝转账
```
POST /v1/tea/user/transfer/reject
请求:
{
  "transfer_uuid": "xxx",
  "reason": "不需要这笔茶叶"
}
响应:
{
  "success": true,
  "message": "转账拒绝成功"
}
```

## 部署说明

### 数据库更新
```sql
-- 执行茶叶支付系统表创建
psql -h localhost -U postgres -d teachat -f sql/tea_payment_system.sql
```

## 后续扩展建议

### 1. 增强功能
- 团队账户系统
- 批量转账功能
- 定时转账
- 转账模板

### 2. 集成场景
- 茶叶商城
- 打赏功能
- 服务购买
- 活动奖励

### 3. 管理功能
- 统计报表
- 异常监控
- 限额管理
- 黑名单机制

### 4. 用户体验
- 移动端适配
- 推送通知
- 交易提醒
- 快捷操作

## 技术特点

1. **高精度**: 使用DECIMAL(15,3)确保金额精确到毫克
2. **高并发**: 使用数据库锁和事务处理并发转账
3. **可扩展**: 模块化设计，易于添加新功能
4. **安全性**: 完整的权限控制和状态管理
5. **可维护**: 详细的日志记录和错误处理

茶叶支付系统已完成基础功能实现，可以支持星际茶棚的用户间茶叶转账需求。