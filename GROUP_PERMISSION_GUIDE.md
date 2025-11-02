# 集团创建权限说明

## 权限规则

### 谁可以代表团队创建集团？

**只有以下两类人员有权限：**

1. **团队创建人（Founder）**
   - 团队的原始创建者
   - 拥有最高权限

2. **团队CEO**
   - 由创建人指定的首席执行官
   - 代表团队进行重大决策

### 为什么这样设计？

这个设计符合现实社会中的组织层级和权力结构：

1. **符合社会印象**
   - 在现实中，只有公司的创始人或CEO才能代表公司做出成立集团这样的重大决策
   - 普通员工或中层管理者无权做出这种战略级别的决定

2. **责任明确**
   - 创建集团是重大决策，需要有明确的责任人
   - 创建人和CEO是团队的最高决策者，承担相应责任

3. **防止滥用**
   - 避免普通成员随意创建集团
   - 保证集团创建的严肃性和规范性

## 权限检查实现

### 前端控制

在团队管理页面（team.manage.go.html）：

```html
{{ if or $.IsFounder $.IsCEO }}
  <!-- 显示可点击的"创建集团"链接 -->
  <li><a href="/v1/group/new?team_id={{ .Team.Uuid }}">创建集团</a></li>
{{ else }}
  <!-- 显示灰色不可点击状态 -->
  <li><a href="#" title="只有团队创建人或CEO才能创建集团" 
         style="color: #ccc; cursor: not-allowed;">创建集团</a></li>
{{ end }}
```

### 后端验证

#### GET /v1/group/new

```go
// 如果有team_id参数，检查用户权限
if teamId != "" {
    team, err := data.GetTeamByUUID(teamId)
    // ...
    
    // 检查权限：必须是创建人或CEO
    isCEO := false
    if sessUser.Id != team.FounderId {
        ceo, err := team.MemberCEO()
        if err == nil && ceo.UserId == sessUser.Id {
            isCEO = true
        }
        if !isCEO {
            report(w, r, "你好，只有团队创建人或CEO才能代表团队创建集团。")
            return
        }
    }
}
```

#### POST /v1/group/create

```go
// 检查权限：用户必须是该团队的创建人或CEO
team, err := data.GetTeam(firstTeamId)
// ...

isCEO := false
if sessUser.Id != team.FounderId {
    ceo, err := team.MemberCEO()
    if err == nil && ceo.UserId == sessUser.Id {
        isCEO = true
    }
    if !isCEO {
        report(w, r, "你好，只有团队创建人或CEO才能代表团队创建集团。")
        return
    }
}
```

## 用户体验

### 有权限的用户（创建人/CEO）

1. 在团队管理页面看到可点击的"创建集团"链接
2. 点击后直接进入创建集团表单
3. 该团队被自动预选为最高管理团队
4. 提交表单后成功创建集团

### 无权限的用户（普通成员/其他管理员）

1. 在团队管理页面看到灰色的"创建集团"文字
2. 鼠标悬停时显示提示："只有团队创建人或CEO才能创建集团"
3. 无法点击，无法访问创建页面
4. 即使通过URL直接访问，也会被后端拦截并提示无权限

## 安全性

### 多层防护

1. **前端UI控制**：无权限用户看不到可点击的链接
2. **GET请求验证**：访问创建页面时验证权限
3. **POST请求验证**：提交表单时再次验证权限

### 防止绕过

即使用户：
- 通过浏览器开发工具修改前端代码
- 直接构造URL访问创建页面
- 直接发送POST请求

后端都会进行权限验证，拒绝无权限的操作。

## 角色说明

### 团队创建人（Founder）
- 团队的原始创建者
- 永久拥有最高权限
- 可以指定和撤销CEO

### CEO（首席执行官）
- 由创建人指定
- 代表团队进行日常管理和重大决策
- 可以管理其他核心成员（CTO、CMO、CFO）

### 其他核心成员（CTO/CMO/CFO）
- 由CEO指定
- 负责特定领域的管理
- 无权创建集团

### 普通成员（Taster/品茶师）
- 团队的普通成员
- 无管理权限
- 无权创建集团

## 总结

这个权限设计：
- ✅ 符合现实社会的组织结构
- ✅ 责任明确，防止滥用
- ✅ 多层验证，安全可靠
- ✅ 用户体验友好，提示清晰
