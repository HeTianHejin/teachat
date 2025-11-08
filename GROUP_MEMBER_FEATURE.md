# 集团成员管理功能说明

## 功能概述

集团（Group）成员管理功能允许集团管理者邀请团队（Team）加入集团，以及移除集团中的成员团队。这个功能参考了team的邀请流程，实现了类似的邀请-确认机制。

## 功能特点

### 1. 增加成员（邀请团队加入集团）

**流程：**
1. 集团管理者在集团详情页面点击"增加成员"
2. 搜索目标团队
3. 在团队详情页面或搜索结果中点击"邀请加入集团"
4. 填写邀请函内容，包括：
   - 邀请词（2-239字符）
   - 团队角色（如：核心成员、协作团队等）
   - 团队等级（2-5级，1级为最高管理团队）
5. 发送邀请函
6. 团队的创建人或CEO收到邀请通知
7. 团队管理者查看邀请函详情
8. 团队管理者决定接受或拒绝邀请
9. 如果接受，团队自动加入集团

**权限要求：**
- 只有集团创建人或最高管理团队成员可以邀请团队
- 只有团队创建人或CEO可以处理邀请函

### 2. 移除成员

**流程：**
1. 集团管理者在集团详情页面点击"移除成员"
2. 查看当前集团的所有成员团队列表
3. 选择要移除的团队，点击"移除"按钮
4. 确认后，该团队从集团中移除（软删除）

**权限要求：**
- 只有集团创建人或最高管理团队成员可以移除成员
- 不能移除最高管理团队（FirstTeam）

**限制：**
- 最高管理团队不能被移除
- 移除操作采用软删除，可以恢复

## 数据库设计

### 集团邀请函表（group_invitations）

```sql
CREATE TABLE group_invitations (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE,
    group_id              INTEGER REFERENCES groups(id),
    team_id               INTEGER REFERENCES teams(id),
    invite_word           TEXT,
    role                  VARCHAR(255),
    level                 INTEGER DEFAULT 2,
    status                INTEGER DEFAULT 0,
    author_user_id        INTEGER REFERENCES users(id),
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**状态说明：**
- 0: 待处理
- 1: 已查看
- 2: 已接受
- 3: 已拒绝
- 4: 已过期

### 集团邀请函回复表（group_invitation_replies）

```sql
CREATE TABLE group_invitation_replies (
    id                    SERIAL PRIMARY KEY,
    uuid                  VARCHAR(64) NOT NULL UNIQUE,
    invitation_id         INTEGER REFERENCES group_invitations(id),
    user_id               INTEGER REFERENCES users(id),
    reply_word            TEXT,
    created_at            TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 路由说明

### 集团成员管理路由

- `GET /v1/group/member_add?uuid={group_uuid}` - 显示增加成员页面（搜索团队）
- `GET /v1/group/member_remove?uuid={group_uuid}` - 显示移除成员页面
- `POST /v1/group/member_remove` - 处理移除成员请求

### 集团邀请函路由

- `GET /v1/group/member_invite?uuid={group_uuid}&team_uuid={team_uuid}` - 显示邀请表单
- `POST /v1/group/member_invite` - 发送邀请函
- `GET /v1/group/member_invitation?id={invitation_uuid}` - 查看邀请函详情
- `POST /v1/group/member_invitation` - 处理邀请函回复

## 模板文件

- `group.member_add.go.html` - 增加成员页面（搜索团队）
- `group.member_remove.go.html` - 移除成员页面
- `group.member_invite.go.html` - 邀请团队加入集团表单
- `group.member_invitation_read.go.html` - 邀请函查看和回复页面

## 使用示例

### 1. 邀请团队加入集团

```
1. 访问集团详情页面：/v1/group/detail?id={group_uuid}
2. 点击"增加成员"按钮
3. 搜索目标团队
4. 在团队详情页面点击"邀请加入集团"
5. 填写邀请函并发送
```

### 2. 处理邀请函

```
1. 团队CEO或创建人收到消息通知
2. 访问邀请函详情：/v1/group/member_invitation?id={invitation_uuid}
3. 查看邀请内容
4. 选择接受或拒绝
5. 填写回复内容并提交
```

### 3. 移除团队

```
1. 访问集团详情页面：/v1/group/detail?id={group_uuid}
2. 点击"移除成员"按钮
3. 在成员列表中找到要移除的团队
4. 点击"移除"按钮
5. 确认操作
```



## 注意事项

1. **权限控制**：
   - 集团管理权限：创建人或最高管理团队成员
   - 邀请函处理权限：团队创建人或CEO

2. **数据完整性**：
   - 邀请函发送前检查团队是否已是集团成员
   - 移除操作采用软删除，保留历史记录

3. **消息通知**：
   - 发送邀请函后，向团队CEO发送消息通知
   - 处理邀请函后，更新消息计数

4. **用户体验**：
   - 邀请函状态实时更新
   - 已处理的邀请函不能重复处理
   - 提供清晰的操作反馈

## 未来改进

1. 批量邀请功能
2. 邀请函过期机制
3. 团队主动申请加入集团
4. 集团成员等级调整
5. 集团成员角色变更通知
