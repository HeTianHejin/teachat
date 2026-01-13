# TeamMember.Role 从 String 改为 Int 的迁移计划

## 1. 角色映射定义

```go
// $事业茶团角色
const (
    RoleUnknown = 0  // 未知
    RoleCEO     = 1  // CEO
    RoleCTO     = 2  // CTO
    RoleCMO     = 3  // CMO
    RoleCFO     = 4  // CFO
    RoleTaster  = 5  // 品茶师
)

// 角色名称映射
var RoleNameMap = map[int]string{
    RoleUnknown: "未知",
    RoleCEO:     "CEO",
    RoleCTO:     "CTO",
    RoleCMO:     "CMO",
    RoleCFO:     "CFO",
    RoleTaster:  "品茶师",
}

// 角色排序级别（用于排序）
var RoleLevelMap = map[int]int{
    RoleCEO:    1,  // 最高级别
    RoleCTO:    2,
    RoleCMO:    3,
    RoleCFO:    4,
    RoleTaster: 5,  // 普通成员
}
```

## 2. 需要修改的文件

### 2.1 数据库迁移 SQL

创建文件：`sql/migrations/migrate_role_to_int.sql`

```sql
-- 第一步：添加新的 role_int 列
ALTER TABLE team_members ADD COLUMN role_int INTEGER DEFAULT 0;

-- 第二步：根据现有的 role 字段填充 role_int
UPDATE team_members SET role_int = CASE
    WHEN role = 'CEO' THEN 1
    WHEN role = 'CTO' THEN 2
    WHEN role = 'CMO' THEN 3
    WHEN role = 'CFO' THEN 4
    WHEN role = 'taster' THEN 5
    ELSE 0
END;

-- 第三步：删除旧的 role 列，重命名 role_int 为 role
ALTER TABLE team_members DROP COLUMN role;
ALTER TABLE team_members RENAME COLUMN role_int TO role;

-- 第四步：为 team_member_role_notices 表做相同操作
ALTER TABLE team_member_role_notices ADD COLUMN member_current_role_int INTEGER DEFAULT 0;
UPDATE team_member_role_notices SET member_current_role_int = CASE
    WHEN member_current_role = 'CEO' THEN 1
    WHEN member_current_role = 'CTO' THEN 2
    WHEN member_current_role = 'CMO' THEN 3
    WHEN member_current_role = 'CFO' THEN 4
    WHEN member_current_role = 'taster' THEN 5
    ELSE 0
END;
ALTER TABLE team_member_role_notices DROP COLUMN member_current_role;
ALTER TABLE team_member_role_notices RENAME COLUMN member_current_role_int TO member_current_role;

ALTER TABLE team_member_role_notices ADD COLUMN new_role_int INTEGER DEFAULT 0;
UPDATE team_member_role_notices SET new_role_int = CASE
    WHEN new_role = 'CEO' THEN 1
    WHEN new_role = 'CTO' THEN 2
    WHEN new_role = 'CMO' THEN 3
    WHEN new_role = 'CFO' THEN 4
    WHEN new_role = 'taster' THEN 5
    ELSE 0
END;
ALTER TABLE team_member_role_notices DROP COLUMN new_role;
ALTER TABLE team_member_role_notices RENAME COLUMN new_role_int TO new_role;

-- 第五步：为 team_member_resignations 表做相同操作
ALTER TABLE team_member_resignations ADD COLUMN member_current_role_int INTEGER DEFAULT 0;
UPDATE team_member_resignations SET member_current_role_int = CASE
    WHEN member_current_role = 'CEO' THEN 1
    WHEN member_current_role = 'CTO' THEN 2
    WHEN member_current_role = 'CMO' THEN 3
    WHEN member_current_role = 'CFO' THEN 4
    WHEN member_current_role = 'taster' THEN 5
    ELSE 0
END;
ALTER TABLE team_member_resignations DROP COLUMN member_current_role;
ALTER TABLE team_member_resignations RENAME COLUMN member_current_role_int TO member_current_role;

-- 第六步：更新索引（因为列类型变化）
DROP INDEX IF EXISTS idx_team_members_role;
CREATE INDEX idx_team_members_role ON team_members(role);
```

### 2.2 DAO 文件修改

#### DAO/team.go

**修改 1：常量定义**
```go
// 删除旧的字符串常量
// const (
//     RoleCEO    = "CEO"
//     RoleCTO    = "CTO"
//     RoleCMO    = "CMO"
//     RoleCFO    = "CFO"
//     RoleTaster = "taster"
// )

// 新的 int 类型常量
const (
    RoleUnknown = 0  // 未知
    RoleCEO     = 1  // CEO
    RoleCTO     = 2  // CTO
    RoleCMO     = 3  // CMO
    RoleCFO     = 4  // CFO
    RoleTaster  = 5  // 品茶师
)

// 角色名称映射
var RoleNameMap = map[int]string{
    RoleUnknown: "未知",
    RoleCEO:     "CEO",
    RoleCTO:     "CTO",
    RoleCMO:     "CMO",
    RoleCFO:     "CFO",
    RoleTaster:  "品茶师",
}
```

**修改 2：TeamMember 结构体**
```go
type TeamMember struct {
    Id        int
    Uuid      string
    TeamId    int
    UserId    int
    Role      int              // 改为 int 类型
    Status    TeamMemberStatus
    CreatedAt time.Time
    UpdatedAt *time.Time
    DeletedAt *time.Time
}
```

**修改 3：添加方法**
```go
// GetRoleName 返回角色名称
func (member *TeamMember) GetRoleName() string {
    return RoleNameMap[member.Role]
}

// GetRoleLevel 返回角色级别（用于排序）
func (member *TeamMember) GetRoleLevel() int {
    return RoleLevelMap[member.Role]
}

// IsCoreMember 检查是否为核心成员（CEO/CTO/CMO/CFO）
func (member *TeamMember) IsCoreMember() bool {
    return member.Role >= RoleCEO && member.Role <= RoleCFO
}

// IsCEO 检查是否为 CEO
func (member *TeamMember) IsCEO() bool {
    return member.Role == RoleCEO
}
```

**修改 4：TeamMemberRoleNotice 结构体**
```go
type TeamMemberRoleNotice struct {
    Id                int
    Uuid              string
    TeamId            int
    CeoId             int
    MemberId          int
    MemberCurrentRole int  // 改为 int 类型
    NewRole           int  // 改为 int 类型
    Title             string
    Content           string
    Status            int
    CreatedAt         time.Time
    UpdatedAt         *time.Time
}
```

**修改 5：TeamMemberResignation 结构体**
```go
type TeamMemberResignation struct {
    Id               int
    Uuid             string
    TeamId           int
    CeoUserId        int
    CoreMemberUserId int
    MemberId         int
    MemberUserId     int
    MemberCurrentRole int  // 改为 int 类型
    Title            string
    Content          string
    Status           int
    CreatedAt        time.Time
    UpdatedAt        *time.Time
}
```

**修改 6：SQL 查询中的字符串比较改为数字比较**

```go
// 旧代码
// rows, err := DB.Query("... WHERE team_members.role = 'CEO' ...")

// 新代码
rows, err := DB.Query("... WHERE team_members.role = $1 ...", RoleCEO)

// 旧代码
// (team_members.role = 'CEO' or team_members.role = 'CTO' or team_members.role = 'CMO' or team_members.role = 'CFO')

// 新代码
// team_members.role IN ($1, $2, $3, $4)
rows, err := DB.Query("... WHERE team_members.role IN ($1, $2, $3, $4) ...",
    RoleCEO, RoleCTO, RoleCMO, RoleCFO)
```

**修改 7：CoreMembers 排序查询**
```go
func (team *Team) CoreMembers() (team_members []TeamMember, err error) {
    if team.Id == TeamIdNone {
        return nil, fmt.Errorf("team not found with id: %d", team.Id)
    }
    if team.Id == TeamIdFreelancer {
        return nil, fmt.Errorf("team member cannot find with id: %d", team.Id)
    }
    rows, err := DB.Query(`
        SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at
        FROM team_members
        WHERE team_id = $1 AND role IN ($2, $3, $4, $5) AND status = $6
        ORDER BY role ASC
    `, team.Id, RoleCEO, RoleCTO, RoleCMO, RoleCFO, TeamMemberStatusActive)
    // ... 其余代码不变
}
```

**修改 8：GetTeamMemberByRole 方法**
```go
func (team *Team) GetTeamMemberByRole(role int) (team_member TeamMember, err error) {
    team_member = TeamMember{}
    err = DB.QueryRow("SELECT id, uuid, team_id, user_id, role, created_at, status, updated_at FROM team_members WHERE team_id = $1 AND role = $2", team.Id, role).
        Scan(&team_member.Id, &team_member.Uuid, &team_member.TeamId, &team_member.UserId, &team_member.Role, &team_member.CreatedAt, &team_member.Status, &team_member.UpdatedAt)
    return
}
```

**修改 9：UpdateRoleStatus 方法**
```go
func (teamMember *TeamMember) UpdateRoleStatus() (err error) {
    statement := `UPDATE team_members SET role = $2, updated_at = $3, status = $4 WHERE id = $1`
    stmt, err := DB.Prepare(statement)
    if err != nil {
        return
    }
    defer stmt.Close()
    _, err = stmt.Exec(teamMember.Id, teamMember.Role, time.Now(), teamMember.Status)
    return
}
```

### 2.3 Route 文件修改

#### Route/route_team_member.go

**修改：字符串比较改为数字比较**

```go
// 旧代码
// if t_member.Role != dao.RoleCEO && t_member.Role != "taster" {
//     t_member.Role = "taster"
// }

// 新代码
if t_member.Role != dao.RoleCEO && t_member.Role != dao.RoleTaster {
    t_member.Role = dao.RoleTaster
    if err := t_member.UpdateRoleStatus(); err != nil {
        util.Debug("Cannot update member role to taster", err)
        report(w, s_u, "你好，茶博士正在忙碌中，稍后再试。")
        return
    }
}
```

### 2.4 模板文件修改

需要在模板中添加辅助函数来将 int 转换为显示字符串。可以在 template_data.go 中添加：

```go
// 添加到 DAO/template_struct_team.go 或相关文件
func RoleName(role int) string {
    if name, ok := RoleNameMap[role]; ok {
        return name
    }
    return "未知"
}
```

然后在模板中使用：
```gohtml
<!-- 旧代码 -->
<span class="label label-primary">{{ .TeamMember.Role }}</span>

<!-- 新代码 -->
<span class="label label-primary">{{ RoleName .TeamMember.Role }}</span>
```

## 3. 迁移步骤

### 3.1 备份数据库
```bash
pg_dump teachat > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 3.2 执行数据库迁移
```bash
psql -U robin -d teachat -f sql/migrations/migrate_role_to_int.sql
```

### 3.3 修改 Go 代码
按照上述修改清单，逐个修改相关文件。

### 3.4 更新 schema.sql
将 schema.sql 中的相关表定义更新为新的类型。

### 3.5 测试
- 测试团队成员的创建、查询、更新
- 测试角色排序
- 测试角色变更通知
- 测试成员退出流程
- 测试权限验证

## 4. 注意事项

1. **数据完整性**：迁移前务必备份数据库
2. **向后兼容**：如果需要向后兼容，可以在过渡期保留旧字段
3. **索引优化**：迁移后重建相关索引
4. **模板渲染**：确保所有模板都使用 RoleName 函数转换
5. **API 变更**：如果对外提供 API，需要更新 API 文档

## 5. 优势总结

✅ 性能提升：数据库查询更快，占用空间更少
✅ 类型安全：编译时检查，避免拼写错误
✅ 易于排序：天然支持按级别排序
✅ 便于扩展：添加新角色只需增加常量
✅ 代码清晰：逻辑更直观，便于维护
