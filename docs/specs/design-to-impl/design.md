# 设计稿落地实现 - 设计文档

## 1. 架构概览

```
┌─────────────────────────────────────────────────────┐
│                  共享数据层                           │
│  ~/.aivectormemory/memory.db (SQLite + sqlite-vec)  │
│  新增表: users (本地认证)                            │
└───────────┬─────────────────────────┬───────────────┘
            │                         │
    ┌───────▼────────┐       ┌───────▼─────────┐
    │   Desktop 端    │       │   Web 看板       │
    │  Wails + Vue 3  │       │  Flask + 原生 JS │
    │  Go 后端直连 DB  │       │  Python API      │
    └────────────────┘       └─────────────────┘
```

两端共用同一个数据库，认证数据存在 `users` 表中。

## 2. 本地认证系统

### 2.1 数据库设计

在 `memory.db` 中新增 `users` 表：

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    last_login TEXT
);
```

- 密码用 bcrypt 哈希存储
- 不需要 email、2FA、邀请码、API 密钥等远程服务功能
- 设计稿 auth.html 中的远程功能（邮箱验证、优惠码、余额、订阅、API 密钥）不实现

### 2.2 认证流程

```
注册: username + password → bcrypt hash → INSERT users
登录: username + password → bcrypt verify → 生成 session token
验证: 请求携带 token → 校验有效性 → 放行/拒绝
```

**Session Token**：
- Desktop 端：Wails Go 后端维护内存 session，前端 localStorage 存 token
- Web 看板：Flask 用 cookie/header 传递 token，后端内存校验

### 2.3 Auth 页面（简化版）

只实现两个表单（设计稿中远程相关功能全部去掉）：

| 表单 | 字段 | 说明 |
|------|------|------|
| **登录** | 用户名 + 密码 | 验证通过跳转 projects 页 |
| **注册** | 用户名 + 密码 + 确认密码 | 创建成功自动登录 |

未登录状态：路由守卫拦截，重定向到 auth 页面。

### 2.4 实现位置

| 端 | 认证后端 | 认证前端 |
|----|---------|---------|
| Desktop | `desktop/internal/auth/` 新建包：Register/Login/Verify | `desktop/frontend/src/views/AuthView.vue` 新建 |
| Web 看板 | `aivectormemory/web/routes/auth.py` 新建 | `static/app.js` 中新增 auth 渲染逻辑 |

## 3. Desktop 端改造

### 3.1 现状 vs 目标

| 项目 | 现状 | 目标 |
|------|------|------|
| 路由 | `/` → ProjectSelect，`/project/*` → 各页面 | 新增 `/auth` → AuthView |
| 布局 | ProjectLayout 包裹子页面 | 不变，auth 页面独立于 ProjectLayout |
| 样式 | variables.css + light.css（`data-theme`） | 更新 CSS 变量对齐设计稿色值 |
| 侧边栏 | Sidebar.vue 已有导航结构 | 更新导航项文字/图标对齐设计稿 |
| Projects 页 | ProjectSelect.vue（卡片列表） | 更新样式对齐设计稿 glass 效果 |

### 3.2 新增/修改文件

```
desktop/frontend/src/
├── views/
│   └── AuthView.vue          ← 新增：登录/注册页
├── router/
│   └── index.ts              ← 修改：新增 /auth 路由 + 认证守卫
├── stores/
│   └── auth.ts               ← 新增：认证状态管理
├── styles/
│   └── variables.css         ← 修改：CSS 变量对齐设计稿
│   └── light.css             ← 修改：亮色主题对齐设计稿

desktop/
├── app.go                    ← 修改：新增 Register/Login/Logout/GetCurrentUser 方法
├── internal/
│   └── auth/                 ← 新增目录
│       └── auth.go           ← 注册/登录/session 管理
```

### 3.3 路由守卫

```typescript
router.beforeEach((to) => {
  const authStore = useAuthStore()
  if (to.path !== '/auth' && !authStore.isLoggedIn) {
    return '/auth'
  }
})
```

### 3.4 Go 后端新增方法

```go
// desktop/internal/auth/auth.go
func Register(db *sql.DB, username, password string) error
func Login(db *sql.DB, username, password string) (token string, err error)
func Verify(token string) (username string, err error)
func Logout(token string)

// desktop/app.go 新增绑定
func (a *App) Register(username, password string) error
func (a *App) Login(username, password string) (map[string]string, error)
func (a *App) Logout() error
func (a *App) GetCurrentUser() (map[string]string, error)
```

## 4. Web 看板改造

### 4.1 现状 vs 目标

| 项目 | 现状 | 目标 |
|------|------|------|
| 页面 | 项目选择 + 7 个 tab | 新增 auth 页面（登录/注册切换） |
| 路由 | SPA tab 切换 | 新增 auth 状态判断 |
| 样式 | style.css（`data-theme`） | 更新 CSS 对齐设计稿 |
| API | token 参数认证（可选） | 新增 /api/auth/* 端点 |

### 4.2 新增/修改文件

```
aivectormemory/web/
├── routes/
│   └── auth.py               ← 新增：register/login/logout/me 端点
├── static/
│   ├── index.html            ← 修改：新增 auth 页面容器
│   ├── app.js                ← 修改：新增 auth 渲染 + 登录状态管理
│   └── style.css             ← 修改：新增 auth 样式 + 对齐设计稿
├── api.py                    ← 修改：注册 auth 路由
└── app.py                    ← 修改：session 中间件
```

### 4.3 API 端点

```
POST /api/auth/register    { username, password }  → { token, username }
POST /api/auth/login        { username, password }  → { token, username }
POST /api/auth/logout       {}                      → { ok }
GET  /api/auth/me           (需 token)              → { username }
```

### 4.4 前端认证流程

```javascript
// app.js
function checkAuth() {
  const token = localStorage.getItem('avm-token')
  if (!token) { showAuthPage(); return }
  api('auth/me').then(user => {
    window.currentUser = user
    loadProjects()
  }).catch(() => {
    localStorage.removeItem('avm-token')
    showAuthPage()
  })
}
```

## 5. 样式对齐策略

### 5.1 设计稿色值体系

设计稿使用 HSL 色值，与现有两端的 CSS 变量命名不同但功能对应：

| 设计稿变量 | Desktop 对应 | Web 看板对应 | 值 (暗色) |
|-----------|-------------|-------------|-----------|
| `--bg` | `--bg-primary` | `--bg-primary` | `hsl(240 5% 12%)` |
| `--bg-sidebar` | `--bg-sidebar` | `--bg-sidebar` | `hsl(240 5% 10%)` |
| `--bg-card` | `--bg-surface` | `--bg-surface` | `hsl(240 5% 16%)` |
| `--text` | `--text-primary` | `--text-primary` | `hsl(0 0% 98%)` |
| `--text-muted` | `--text-secondary` | `--text-secondary` | `hsl(240 5% 64.9%)` |
| `--primary` | `--accent` | `--accent` | `hsl(210 100% 54%)` |
| `--border` | `--border` | `--border` | `hsl(240 5% 24%)` |

### 5.2 策略

- **不改变量名**：两端保持各自已有的 CSS 变量命名
- **只改色值**：将两端的变量值更新为设计稿的 HSL 色值
- **新增缺失变量**：设计稿有但两端没有的变量追加定义
- Glass 效果、动画等设计稿特有样式按需添加

## 6. 页面对齐清单

### 6.1 Desktop 端（Vue 3）

| 页面 | 现有 View | 改动量 | 说明 |
|------|----------|--------|------|
| auth | 无 | **新建** | AuthView.vue |
| projects | ProjectSelect.vue | **样式更新** | glass 卡片效果 |
| stats | StatsView.vue | **样式更新** | 统计卡片样式 |
| session | StatusView.vue | **样式更新** | 脉冲动画等 |
| issues | IssuesView.vue | **样式更新** | 表格/徽章样式 |
| tasks | TasksView.vue | **样式更新** | 树形/进度条 |
| project-memories | MemoriesView.vue | **样式更新** | 卡片列表 |
| user-memories | MemoriesView.vue | **样式更新** | 同上 |
| tags | TagsView.vue | **样式更新** | 标签颜色 |
| settings | SettingsView.vue + MaintenanceView.vue | **合并** | 维护内容并入设置 |

### 6.2 Web 看板（Flask + 原生 JS）

| 页面 | 现有实现 | 改动量 | 说明 |
|------|---------|--------|------|
| auth | 无 | **新建** | index.html + app.js 新增 |
| projects | 有 | **样式更新** | glass 卡片 |
| stats | 有 | **样式更新** | 统计卡片 |
| session | 有 | **样式更新** | 脉冲动画 |
| issues | 有 | **样式更新** | 表格样式 |
| tasks | 有 | **样式更新** | 树形层级 |
| project-memories | 有 | **样式更新** | 卡片列表 |
| user-memories | 有 | **样式更新** | 同上 |
| tags | 有 | **样式更新** | 标签颜色 |
| settings | 无维护部分 | **新增** | 维护工具 + 样式更新 |

## 7. 执行顺序

### Phase 1：认证基础（Desktop + Web）
1. 数据库：`users` 表 + migration
2. Desktop auth 后端（Go）
3. Desktop AuthView + 路由守卫
4. Web auth 后端（Python）
5. Web auth 前端 + 路由守卫

### Phase 2：Desktop 样式对齐
6. CSS 变量更新（variables.css + light.css）
7. Sidebar 更新
8. ProjectSelect 样式
9. 各 View 样式逐个对齐

### Phase 3：Web 看板样式对齐
10. CSS 变量更新（style.css）
11. 侧边栏样式更新
12. 各页面渲染逻辑 + 样式逐个对齐

### Phase 4：收尾
13. Settings + Maintenance 合并（Desktop）
14. Settings + 维护工具（Web）
15. 集成测试
