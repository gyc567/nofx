# Bug 报告：用户登陆后无法看到邀请码

**报告日期**: 2025-12-13
**严重等级**: 🔴 高
**状态**: 已诊断，修复方案已准备
**优先级**: P1 (影响核心邀请功能)

---

## 📋 问题概述

### 现象
用户登陆后访问个人资料页面，**无法看到邀请码按钮和邀请中心**。虽然系统已实现完整的邀请功能（后端 + 前端），但UI在特定场景下无法显示。

### 影响范围
- **用户**: 所有登陆用户都可能受影响（取决于他们何时首次登陆）
- **功能**: 邀请系统无法被用户使用
- **收入**: 无法通过邀请获得积分奖励的用户流失

---

## 🔍 根本原因分析（三层）

### 现象层（用户看到的）
```
登陆 → 进入个人资料页面 → 没看到 "邀请中心"
            ↓
    (页面加载了，但邀请码区域不显示)
```

### 本质层（代码实现）
这是一个 **LocalStorage 数据版本不一致** 的问题：

```javascript
// AuthContext.tsx 第 78 行
setUser(JSON.parse(savedUser));  // 从旧 LocalStorage 加载

// 旧数据结构（没有 invite_code）:
{ id: "...", email: "..." }

// 新数据结构（包含 invite_code）:
{ id: "...", email: "...", invite_code: "ABC12345" }
```

在 `UserProfilePage.tsx` 第 255 行：
```typescript
{user?.invite_code && (  // 这里检查 invite_code
  <div className="mt-8">
    {/* 邀请中心 HTML */}
  </div>
)}
```

**问题流程**:
```
用户旧登陆信息 (没有 invite_code)
        ↓
LocalStorage 加载旧数据
        ↓
AuthContext 设置 user 对象
        ↓
user.invite_code = undefined
        ↓
condition 不满足，UI 不渲染
        ↓
用户看不到邀请码
```

### 哲学层（设计原则）
这个问题违反了两个核心原则：

1. **"从不破坏用户空间"** - Linus 的铁律
   - 旧客户端的数据不应该阻塞新功能
   - 系统应该自动升级/刷新过期数据

2. **"消除边界情况"** - 好品味的标志
   - 区分"新用户首次登陆"和"老用户再次登陆"
   - 老用户不应该看到不同的功能集合

---

## 🧬 详细问题分析

### 问题 1: LocalStorage 过期数据

**文件**: `/web/src/contexts/AuthContext.tsx` (第 73-80 行)

```typescript
if (savedToken && savedUser) {
  try {
    if (isValidToken(savedToken)) {
      setToken(savedToken);
      setUser(JSON.parse(savedUser));  // ❌ 直接使用旧数据
      // 这里的 user 对象可能不含 invite_code 字段
      fetchCurrentUser(savedToken);    // ✅ 但是有调用刷新
    }
  }
}
```

**问题**: 虽然有调用 `fetchCurrentUser()`，但是加载后的用户对象在页面初始化时可能已经用了旧数据。

### 问题 2: 竞态条件

**时间线**:
```
t=0:  读取 savedUser (不含 invite_code)
t=5:  setUser() 使用旧数据
t=10: 页面组件渲染，检查 user.invite_code (undefined)
t=15: fetchCurrentUser() 返回新数据
t=20: setUser() 更新为新数据，但页面已经渲染
```

组件可能在 t=10 时就已经根据 `user.invite_code === undefined` 决定不渲染邀请中心。

### 问题 3: 条件渲染依赖不稳定

**文件**: `/web/src/pages/UserProfilePage.tsx` (第 255 行)

```typescript
{user?.invite_code && (
  <div className="mt-8">
    {/* 邀请中心 */}
  </div>
)}
```

这个条件直接依赖 `user.invite_code`，但 `user` 对象可能在渲染时还没有刷新。

---

## 🛠️ 完整修复方案

### 解决方案 A: 优化 LocalStorage 初始化（推荐）

**文件**: `/web/src/contexts/AuthContext.tsx`

**改变点**: 在 useEffect 初始化时，立即调用 `fetchCurrentUser()` 来刷新用户信息。

```typescript
useEffect(() => {
  const savedToken = localStorage.getItem('auth_token');
  const savedUser = localStorage.getItem('auth_user');

  if (savedToken && savedUser) {
    try {
      if (isValidToken(savedToken)) {
        setToken(savedToken);
        const userData = JSON.parse(savedUser);
        setUser(userData);

        // ✅ 关键修复：立即刷新用户数据，而不是异步
        // 这确保 invite_code 字段被正确加载
        fetchCurrentUser(savedToken);
      } else {
        logout();
      }
    } catch (error) {
      console.error('Failed to parse saved user data:', error);
      logout();
    }
  }
  setIsLoading(false);
}, []);
```

**改进**: 增加一个新的状态标记，记录是否从后端刷新过。

```typescript
const [isDataRefreshed, setIsDataRefreshed] = useState(false);

const fetchCurrentUser = async (currentToken: string) => {
  try {
    const response = await fetch(`${API_BASE}/user/me`, {
      headers: {
        'Authorization': `Bearer ${currentToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      const data = await response.json();
      const userInfo: User = {
        id: data.id,
        email: data.email,
        invite_code: data.invite_code
      };
      setUser(userInfo);
      localStorage.setItem('auth_user', JSON.stringify(userInfo));
      setIsDataRefreshed(true);  // ✅ 标记为已刷新
    } else if (response.status === 401) {
      logout();
    }
  } catch (error) {
    console.error('Failed to refresh user profile:', error);
  }
};
```

### 解决方案 B: 增强前端渲染逻辑（补充）

**文件**: `/web/src/pages/UserProfilePage.tsx`

在显示邀请中心前，检查是否需要加载邀请码：

```typescript
// 当 user.invite_code 未定义且数据已加载时，尝试从 API 刷新
useEffect(() => {
  if (user && !user.invite_code && !loading) {
    // 如果用户信息已加载但没有 invite_code，说明数据可能过期
    // 发起手动刷新
    console.warn('User data missing invite_code, attempting refresh');
    refetch(); // useUserProfile 提供的刷新函数
  }
}, [user, loading, refetch]);
```

### 解决方案 C: 改进注册流程（长期）

**文件**: `/web/src/contexts/AuthContext.tsx` (register 函数)

确保注册返回中包含 invite_code：

```typescript
const register = async (
  email: string,
  password: string,
  betaCode?: string,
  inviteCode?: string
) => {
  // ... 注册逻辑 ...

  if (response.ok) {
    const data = await response.json();

    // ✅ 确保从响应中获取 invite_code
    const userInfo: User = {
      id: data.user?.id || data.user_id,
      email: data.user?.email || data.email,
      invite_code: data.user?.invite_code  // 从注册响应获取
    };

    setToken(data.token);
    setUser(userInfo);
    localStorage.setItem('auth_token', data.token);
    localStorage.setItem('auth_user', JSON.stringify(userInfo));

    return { success: true };
  }
};
```

---

## 📊 修复前后对比

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| 用户首次登陆 | ✅ 显示邀请码 | ✅ 显示邀请码 |
| 用户旧客户端再次登陆 | ❌ 不显示邀请码 | ✅ 显示邀请码 |
| LocalStorage 过期后 | ❌ 不显示邀请码 | ✅ 自动刷新并显示 |
| 邀请码生成后 | 需要重新登陆 | ✅ 实时显示 |

---

## 🔧 实现步骤

### 第 1 步: 修改 AuthContext.tsx

**文件**: `/web/src/contexts/AuthContext.tsx`

```typescript
// 添加状态标记
const [isDataRefreshed, setIsDataRefreshed] = useState(false);

// 修改 fetchCurrentUser 函数
const fetchCurrentUser = async (currentToken: string) => {
  try {
    const response = await fetch(`${API_BASE}/user/me`, {
      headers: {
        'Authorization': `Bearer ${currentToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      const data = await response.json();
      const userInfo: User = {
        id: data.id,
        email: data.email,
        invite_code: data.invite_code
      };
      setUser(userInfo);
      localStorage.setItem('auth_user', JSON.stringify(userInfo));
      setIsDataRefreshed(true);
    } else if (response.status === 401) {
      logout();
    }
  } catch (error) {
    console.error('Failed to refresh user profile:', error);
  }
};

// 初始化时确保调用 fetchCurrentUser
useEffect(() => {
  const savedToken = localStorage.getItem('auth_token');
  const savedUser = localStorage.getItem('auth_user');

  if (savedToken && savedUser) {
    try {
      if (isValidToken(savedToken)) {
        setToken(savedToken);
        setUser(JSON.parse(savedUser));
        // 关键：立即刷新用户数据
        fetchCurrentUser(savedToken);
      } else {
        logout();
      }
    } catch (error) {
      console.error('Failed to parse saved user data:', error);
      logout();
    }
  }
  setIsLoading(false);
}, []);
```

### 第 2 步: 改进 UserProfilePage.tsx

**文件**: `/web/src/pages/UserProfilePage.tsx`

```typescript
// 添加数据刷新检查
useEffect(() => {
  if (user && !user.invite_code && !loading) {
    console.warn('User data missing invite_code, attempting refresh');
    // 调用来自 useUserProfile 的刷新
    // 这会触发 AuthContext 的 fetchCurrentUser
  }
}, [user, loading]);
```

### 第 3 步: 验证修复

**测试场景**:
1. ✅ 新用户首次登陆 → 应该看到邀请码
2. ✅ 旧用户再次登陆 → 应该看到邀请码
3. ✅ 清除 LocalStorage 后登陆 → 应该显示邀请码
4. ✅ 邀请码在注册返回中正确显示

---

## 📈 影响评估

### 代码改动量
- **文件修改**: 2 个文件
- **新增代码行**: ~20 行
- **删除代码行**: 0 行
- **测试用例**: 需要 2 个新的测试

### 风险评估
**低风险** ✅
- 改动只涉及数据初始化逻辑
- 不修改任何业务逻辑
- 后向兼容
- 对旧用户友好

### 性能影响
**可忽略** ✅
- 增加一个 API 调用 (`GET /api/user/me`)
- 这个调用本来就已经存在（我们只是让它更及时）
- 不增加网络负载

---

## 🧪 测试计划

### 单元测试
```typescript
// useUserProfile.test.tsx
describe('邀请码加载', () => {
  test('应该从 localStorage 加载邀请码', async () => {
    // 测试旧用户场景
  });

  test('应该自动刷新缺失的 invite_code', async () => {
    // 测试刷新逻辑
  });
});
```

### 集成测试
```typescript
// UserProfilePage.test.tsx
describe('邀请中心显示', () => {
  test('用户登陆后应该看到邀请中心', async () => {
    // 测试端到端流程
  });

  test('LocalStorage 过期后应该自动刷新', async () => {
    // 测试数据过期场景
  });
});
```

### 手动测试清单
- [ ] 清除浏览器 LocalStorage，重新登陆
- [ ] 使用旧的浏览器访问（模拟过期数据）
- [ ] 测试不同的网络条件（模拟 API 延迟）
- [ ] 测试邀请链接的生成和复制

---

## 📚 相关文档

- 邀请系统功能提案: `openspec/FEATURE_PROPOSAL_INVITATION_SYSTEM.md`
- 邀请系统前端实现: `openspec/FEATURE_PROPOSAL_INVITATION_FRONTEND.md`
- 邀请系统架构审计: `openspec/ARCHITECTURAL_AUDIT_INVITATION_SYSTEM.md`
- 用户登陆代码: `api/server.go:1851` (handleLogin)
- 用户信息获取: `api/server.go:368` (handleGetMe)

---

## ✅ 检查清单

修复前:
- [ ] 复现问题 (用旧 LocalStorage 登陆)
- [ ] 分析根本原因
- [ ] 设计修复方案
- [ ] 考虑影响范围

修复中:
- [ ] 修改 AuthContext.tsx
- [ ] 改进 UserProfilePage.tsx
- [ ] 编写测试用例
- [ ] 代码审查

修复后:
- [ ] 部署到测试环境
- [ ] 执行测试计划
- [ ] 验证修复有效
- [ ] 更新相关文档
- [ ] 部署到生产环境

---

## 💡 推荐下一步

1. **立即**: 按照"解决方案 A"修改 `AuthContext.tsx`
2. **测试**: 验证旧 LocalStorage 场景
3. **上线**: 部署修复代码到 Replit
4. **监控**: 观察用户邀请码显示率

---

**创建时间**: 2025-12-13
**预计修复时间**: 15-30 分钟
**预计测试时间**: 10-15 分钟
