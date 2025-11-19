# 家庭名称占位符机制

## 设计理念

使用占位符"*"代替配偶姓名，防止用户冒用他人姓名（特别是名人明星）作为家庭名称。

## 核心机制

### 1. 创建家庭时使用占位符

**格式**: `用户名&*`

**示例**:
- 张三创建家庭 → `张三&*`
- 李四创建家庭 → `李四&*`
- Alice创建家庭 → `Alice&*`

### 2. 配偶确认后自动更新

当配偶确认加入家庭时，系统自动将占位符替换为实际姓名。

**流程**:
```
1. 张三创建家庭: "张三&*"
2. 张三邀请李四加入
3. 李四确认加入
4. 系统自动更新: "张三&李四"
```

### 3. 显示名称处理

在界面显示时，占位符"*"显示为"待确认"。

**显示效果**:
- 数据库存储: `张三&*`
- 界面显示: `张三&待确认`

## 防止滥用场景

### 场景1：冒用名人
```
❌ 错误做法（旧方案）:
用户可以自定义名称为 "张三&范冰冰"
→ 可能冒犯范冰冰本人

✅ 正确做法（新方案）:
用户创建家庭: "张三&*"
→ 只有范冰冰本人确认后才会显示真实姓名
```

### 场景2：恶意关联
```
❌ 错误做法:
用户A自定义名称为 "用户A&用户B"
→ 用户B可能不知情

✅ 正确做法:
用户A创建: "用户A&*"
→ 必须用户B确认后才显示 "用户A&用户B"
```

### 场景3：虚假宣传
```
❌ 错误做法:
普通用户自定义 "张三&马云"
→ 误导他人以为与马云有关系

✅ 正确做法:
只能显示 "张三&*" 或 "张三&待确认"
→ 明确表示配偶未确认
```

## 代码实现

### 创建家庭
```go
// Route/route_family.go - NewFamilyPost()
new_family.Name = s_u.Name + "&*"  // 使用占位符
new_family.PerspectiveUserId = s_u.Id
```

### 配偶确认加入
```go
// Route/route_family_member.go - FamilyMemberSignInReply()
if family_member.Role == 1 || family_member.Role == 2 {
    // 父母角色确认加入时，自动更新家庭名称
    family.UpdateFamilyNameWithSpouse(family_member.UserId)
}
```

### 自动更新名称
```go
// DAO/family_name_update.go
func (f *Family) UpdateFamilyNameWithSpouse(spouseUserId int) error {
    // 检查是否包含占位符
    if !strings.Contains(f.Name, "*") {
        return nil
    }
    
    // 获取父母成员
    parentMembers, _ := f.ParentMembers()
    
    // 找出男女主人
    var husband, wife *User
    for _, member := range parentMembers {
        memberUser, _ := GetUser(member.UserId)
        if member.Role == FamilyMemberRoleHusband {
            husband = &memberUser
        } else if member.Role == FamilyMemberRoleWife {
            wife = &memberUser
        }
    }
    
    // 生成新名称
    if husband != nil && wife != nil {
        f.Name = husband.Name + "&" + wife.Name
    }
    
    return f.Update()
}
```

### 显示名称
```go
// DAO/family_name_update.go
func (f *Family) GetDisplayName() string {
    if strings.Contains(f.Name, "*") {
        return strings.Replace(f.Name, "*", "待确认", -1)
    }
    return f.Name
}
```

## 用户界面

### 创建家庭页面
```html
<input type="text" readonly value="张三&*" />
<small>
    ☆ 系统自动生成名称，"*"代表配偶占位符
    当配偶确认加入后，系统将自动更新为实际名称（例：张三&李四）
    这样可以防止冒用他人姓名
</small>
```

### 家庭详情页面
```
家庭名称: 张三&待确认
状态: 单身
说明: 等待配偶确认加入
```

### 茶议发布者背景显示
```
发布者: 张三 (张三&待确认, $某公司)
       ↑          ↑
     用户名    家庭背景
```

## 特殊情况处理

### 1. 单身家庭
```
创建: "张三&*"
显示: "张三&待确认"
说明: 单身状态，暂无配偶
```

### 2. 同性家庭
```
Alice创建: "Alice&*"
Bob确认加入: "Alice&Bob"
说明: 系统自动识别两位女主人或两位男主人
```

### 3. 离婚后
```
原家庭: "张三&李四"
离婚后: 软删除原家庭
新家庭: "张三&*" (重新开始)
```

### 4. 再婚
```
张三的新家庭: "张三&*"
王五确认加入: "张三&王五"
说明: 每个新家庭都从占位符开始
```

## 优势

### 1. 隐私保护 ✅
- 防止冒用他人姓名
- 保护名人隐私
- 避免恶意关联

### 2. 真实性保证 ✅
- 必须双方确认
- 自动更新名称
- 明确显示状态

### 3. 用户体验 ✅
- 简单明了
- 自动化处理
- 友好提示

### 4. 系统安全 ✅
- 防止滥用
- 数据一致性
- 可追溯性

## 与其他功能的配合

### 1. 视角字段
```
张三的视角: "张三&*" (PerspectiveUserId = 张三ID)
李四的视角: "李四&张三" (PerspectiveUserId = 李四ID)
```

### 2. 家庭关联
```
通过family_relations建立关联后
两个视角的家庭名称可能不同，但指向同一个家庭
```

### 3. 消息通知
```
发送者: 张三 (张三&待确认)
接收者可以看到发送者的家庭状态
```

## 测试场景

### 测试1: 创建单身家庭
```
1. 用户登录
2. 创建家庭
3. 验证名称为 "用户名&*"
4. 界面显示 "用户名&待确认"
```

### 测试2: 配偶确认
```
1. 用户A创建家庭 "A&*"
2. 用户A邀请用户B
3. 用户B确认加入
4. 验证名称自动更新为 "A&B"
```

### 测试3: 防止冒用
```
1. 用户尝试输入自定义名称
2. 系统忽略输入，使用占位符
3. 验证无法冒用他人姓名
```

## 总结

占位符机制完美解决了：
- ✅ 防止冒用名人姓名
- ✅ 保护用户隐私
- ✅ 确保关系真实性
- ✅ 提供良好用户体验
- ✅ 维护系统安全

这是一个简单而有效的解决方案！
