# Linux Spec 在线编辑器 - 实施计划

创建时间: 2026-03-08 23:46
状态: 规划中

## 技术栈确认

**前端**：Vue 3 + TypeScript + Vite + Element Plus + Pinia
**后端**：Go + Hertz + GORM
**编辑器**：Monaco Editor（VS Code 引擎）

---

## 功能模块

### 1. 文件浏览模块
- 树形结构浏览仓库
- 过滤显示 .spec 文件
- 支持展开/折叠
- 显示文件元信息（大小、修改时间）

**技术选型**：
- 前端：Element Plus `el-tree` 组件
- 后端：复用现有 `/api/system/files` API（需增强过滤功能）

---

### 2. Monaco 编辑器集成
- 集成 Monaco Editor
- RPM Spec 语法高亮
- 代码折叠
- 行号显示
- 撤销/重做

**技术选型**：
- 依赖：`@monaco-editor/react`
- 自定义语言定义：RPM Spec

---

### 3. 实时 Linting
- 基于规则的检查
- 错误/警告/提示
- 实时反馈
- 问题面板

**规则分类**：
- **语法规则**：必填字段、格式验证
- **最佳实践**：BuildRoot、Requires、Provides 等
- **自定义规则**：用户定义的检查规则

**技术选型**：
- 前端：Monaco diagnostic API
- 后端：规则引擎（Go 实现）

---

### 4. 规则管理
- 预设规则库（rpmlint 规则子集）
- 自定义规则创建
- 规则启用/禁用
- 规则优先级

**数据结构**：
```go
type LintRule struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Category    string `json:"category"` // syntax, best_practice, custom
    Severity    string `json:"severity"` // error, warning, info
    Pattern     string `json:"pattern"`  // regex or rule definition
    Enabled     bool   `json:"enabled"`
    Priority    int    `json:"priority"`
}
```

---

### 5. 智能保存
- 保存前验证
- 自动 Git commit
- Commit message 模板
- 回滚机制

**保存流程**：
1. 用户点击保存
2. 运行 linting
3. 如果有 error 级别问题，阻止保存
4. 如果只有 warning，提示用户确认
5. 验证通过，保存文件
6. 自动 commit：`chore(spec): update {filename}`

---

## 后端 API 设计

### 文件操作
```
GET  /api/specs/tree          - 获取 .spec 文件树
GET  /api/specs/:path/content - 获取文件内容
PUT  /api/specs/:path/content - 保存文件内容
```

### Linting
```
POST /api/specs/lint          - Lint 文件内容
GET  /api/specs/rules         - 获取规则列表
PUT  /api/specs/rules/:id     - 更新规则配置
POST /api/specs/rules         - 创建自定义规则
```

### Git 操作
```
POST /api/specs/:path/commit  - Commit 文件变更
```

---

## 前端组件设计

### 页面结构
```
/spec-editor
├── Sidebar (文件树)
├── Editor (Monaco)
├── Problems Panel (问题列表)
└── Toolbar (规则管理、保存、commit)
```

### Vue 组件
```
views/
  SpecEditor.vue          # 主页面
components/
  spec/
    FileTree.vue          # 文件树
    SpecMonaco.vue        # Monaco 编辑器
    ProblemsPanel.vue     # 问题面板
    RuleManager.vue       # 规则管理
    CommitDialog.vue      # Commit 对话框
```

---

## 实施阶段

### 阶段 1：基础框架（2-3 小时）
- [ ] 创建前端页面和路由
- [ ] 集成 Monaco Editor
- [ ] 实现 RPM Spec 语法高亮
- [ ] 创建后端 API 框架

### 阶段 2：文件浏览（1-2 小时）
- [ ] 实现文件树 API
- [ ] 前端文件树组件
- [ ] 文件加载和保存

### 阶段 3：Linting 引擎（3-4 小时）
- [ ] 设计规则数据结构
- [ ] 实现规则引擎
- [ ] 预设规则库（10-15 条）
- [ ] 前端 linting 集成

### 阶段 4：规则管理（2-3 小时）
- [ ] 规则管理 UI
- [ ] 自定义规则创建
- [ ] 规则持久化

### 阶段 5：智能保存（1-2 小时）
- [ ] 保存前验证
- [ ] 自动 commit
- [ ] Commit message 模板

### 阶段 6：测试与优化（2-3 小时）
- [ ] 功能测试
- [ ] 性能优化
- [ ] 文档更新

**预计总时长**：11-17 小时（1.5-2 个工作日）

---

## 风险与依赖

**依赖**：
- ✅ Monaco Editor 支持 Vue 3
- ✅ Element Plus 提供树形组件
- ⚠️ 需要研究 RPM Spec 语法规范
- ⚠️ 需要实现完整的规则引擎

**风险**：
- Monaco Editor 包体积较大（~500KB gzip），可能影响加载速度
- 规则引擎复杂度可能超预期
- 大文件性能问题

**缓解措施**：
- Monaco Editor 异步加载
- 规则引擎先实现核心功能，后续迭代
- 文件大小限制（建议 < 1MB）

---

## 下一步

确认需求后，开始阶段 1：基础框架搭建。

**需要确认**：
1. 是否需要支持多个仓库？还是只针对当前仓库？
2. 规则引擎是前端实现还是后端实现？
3. 是否需要支持批量操作（批量 lint、批量 commit）？
4. Commit message 是否需要自定义模板？
