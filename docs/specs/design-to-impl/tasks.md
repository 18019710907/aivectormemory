# 设计稿落地实现 - 任务文档

## Phase 1：认证基础

### 1.1 数据库 users 表
- [x] 1.1.1 Python 端：在 MemoryManager 初始化中新增 users 表 migration（CREATE TABLE IF NOT EXISTS）
- [x] 1.1.2 Go 端：在 desktop/internal/db/connection.go 初始化中新增 users 表 migration

### 1.2 Desktop 认证后端（Go）
- [x] 1.2.1 新建 desktop/internal/auth/auth.go：Register 函数（bcrypt hash + INSERT）
- [x] 1.2.2 auth.go：Login 函数（bcrypt verify + 生成 token + 更新 last_login）
- [x] 1.2.3 auth.go：Verify 函数（token → username）+ Logout 函数（清除 session）
- [x] 1.2.4 app.go：新增 Register/Login/Logout/GetCurrentUser 四个绑定方法

### 1.3 Desktop 认证前端（Vue）
- [x] 1.3.1 新建 desktop/frontend/src/stores/auth.ts：isLoggedIn、username、login/logout/register actions
- [x] 1.3.2 新建 desktop/frontend/src/views/AuthView.vue：登录/注册双表单（对齐设计稿 auth.html 简化版）
- [x] 1.3.3 修改 router/index.ts：新增 /auth 路由 + beforeEach 认证守卫
- [x] 1.3.4 修改 App.vue：启动时调用 GetCurrentUser 恢复登录状态

### 1.4 Web 认证后端（Python）
- [x] 1.4.1 新建 aivectormemory/web/routes/auth.py：register 端点（bcrypt hash + INSERT + 返回 token）
- [x] 1.4.2 auth.py：login 端点（bcrypt verify + 生成 token + 返回）
- [x] 1.4.3 auth.py：logout 端点 + me 端点（token → 用户信息）
- [x] 1.4.4 修改 api.py：注册 /api/auth/* 路由
- [x] 1.4.5 修改 app.py：请求中间件提取 token（Authorization header 或 query param）

### 1.5 Web 认证前端（原生 JS）
- [x] 1.5.1 修改 index.html：新增 #auth-page 容器（登录/注册表单 HTML）
- [x] 1.5.2 修改 app.js：新增 showAuthPage/renderLoginForm/renderRegisterForm 函数
- [x] 1.5.3 修改 app.js：新增 checkAuth 函数 + 启动时调用，未登录显示 auth 页面
- [x] 1.5.4 修改 style.css：新增 auth 页面样式（对齐设计稿 auth.html）

## Phase 2：Desktop 样式对齐

### 2.1 基础样式
- [x] 2.1.1 更新 variables.css：暗色主题 CSS 变量值对齐设计稿色值
- [x] 2.1.2 更新 light.css：亮色主题 CSS 变量值对齐设计稿色值
- [x] 2.1.3 更新 base.css：按钮、徽章、卡片等通用组件样式对齐设计稿

### 2.2 布局组件
- [x] 2.2.1 更新 Sidebar.vue：导航项图标/文字/顺序对齐设计稿，底部用户卡片 + 主题切换
- [x] 2.2.2 更新 ProjectLayout.vue：顶栏样式（52px 高度、项目选择器位置）

### 2.3 Projects 页
- [x] 2.3.1 更新 ProjectSelect.vue：glass 效果卡片、统计数字、新增项目卡片样式对齐设计稿

### 2.4 Stats 页
- [x] 2.4.1 更新 StatsGrid.vue：4 列统计卡片样式对齐设计稿
- [x] 2.4.2 更新 BlockAlert.vue：阻塞警告样式对齐设计稿
- [x] 2.4.3 更新 VectorNetwork.vue：向量网络可视化样式对齐设计稿

### 2.5 Session 页
- [x] 2.5.1 更新 StatusView.vue：会话状态显示样式对齐设计稿（脉冲动画、状态列表）

### 2.6 Issues 页
- [x] 2.6.1 更新 IssueCard.vue：问题卡片样式对齐设计稿（表格行、状态徽章）
- [x] 2.6.2 更新 IssuesView.vue：搜索栏、筛选器、分页样式对齐设计稿

### 2.7 Tasks 页
- [x] 2.7.1 更新 TaskGroup.vue + TaskNode.vue：树形层级、进度条、勾选样式对齐设计稿

### 2.8 Memories 页
- [x] 2.8.1 更新 MemoryCard.vue：记忆卡片样式对齐设计稿（标签、时间戳）
- [x] 2.8.2 更新 MemoriesView.vue：搜索栏、分页、工具栏样式对齐设计稿

### 2.9 Tags 页
- [x] 2.9.1 更新 TagsView.vue + TagTable.vue：标签颜色变体、计数、批量操作栏样式对齐设计稿

### 2.10 Settings 页
- [x] 2.10.1 合并 MaintenanceView.vue 内容到 SettingsView.vue（健康检查、数据库统计、维护工具、备份列表）
- [x] 2.10.2 删除 MaintenanceView.vue + 移除路由中的 /project/maintenance
- [x] 2.10.3 更新 Sidebar.vue：移除维护导航项
- [x] 2.10.4 更新合并后的 SettingsView.vue 样式对齐设计稿

## Phase 3：Web 看板样式对齐

### 3.1 基础样式
- [x] 3.1.1 更新 style.css：暗色主题 CSS 变量值对齐设计稿色值
- [x] 3.1.2 更新 style.css：亮色主题 CSS 变量值对齐设计稿色值
- [x] 3.1.3 更新 style.css：按钮、徽章、卡片等通用组件样式对齐设计稿

### 3.2 布局组件
- [x] 3.2.1 更新 index.html + style.css：侧边栏导航项图标/文字/顺序对齐设计稿
- [x] 3.2.2 更新 index.html + style.css：底部用户卡片 + 主题切换对齐设计稿
- [x] 3.2.3 更新 index.html + style.css：顶栏样式对齐设计稿

### 3.3 Projects 页
- [x] 3.3.1 更新 app.js loadProjects：项目卡片 glass 效果 + 统计数字样式对齐设计稿

### 3.4 Stats 页
- [x] 3.4.1 更新 app.js loadStats + style.css：统计卡片样式对齐设计稿
- [x] 3.4.2 更新向量网络可视化样式对齐设计稿

### 3.5 Session 页
- [x] 3.5.1 更新 app.js loadStatus + style.css：会话状态样式对齐设计稿

### 3.6 Issues 页
- [x] 3.6.1 更新 app.js loadIssues + style.css：问题列表样式对齐设计稿

### 3.7 Tasks 页
- [x] 3.7.1 更新 app.js loadTasks + style.css：任务树形样式对齐设计稿

### 3.8 Memories 页
- [x] 3.8.1 更新 app.js loadMemoriesByScope + style.css：记忆卡片样式对齐设计稿

### 3.9 Tags 页
- [x] 3.9.1 更新 app.js loadTags + style.css：标签样式对齐设计稿

### 3.10 Settings 页
- [x] 3.10.1 新增 Web 看板维护工具页（健康检查、重建向量、备份）- routes/maintenance.py
- [x] 3.10.2 修改 api.py 注册维护路由
- [x] 3.10.3 更新 app.js + index.html：设置页 + 维护工具 UI 对齐设计稿

## Phase 4：收尾

### 4.1 集成测试
- [x] 4.1.1 Desktop 端：注册/登录/未登录拦截 + 各页面渲染验证
- [x] 4.1.2 Web 看板：注册/登录/未登录拦截 + 各页面渲染验证（Playwright）
- [x] 4.1.3 两端共用数据库验证：Desktop 注册的用户在 Web 看板能登录
