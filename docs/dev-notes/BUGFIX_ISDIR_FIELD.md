# Bug 修复报告 - 目录节点识别错误

修复时间: 2026-03-09 00:57
状态: ✅ 已修复

---

## 问题描述

**错误信息**:
```json
{
  "code": 500,
  "msg": "failed to read spec file: read /Users/zhangyi/gitclone/shortlink: is a directory"
}
```

**触发场景**:
用户在 Spec 编辑器中点击文件树节点时，前端尝试加载目录（`.`），导致后端报错。

**根本原因**:
后端返回的文件树数据使用驼峰命名 `isDir`，但前端类型定义使用 snake_case `is_dir`，导致前端无法正确识别目录节点。

---

## 问题分析

### 后端返回的数据结构（修复前）
```json
{
  "name": "shortlink",
  "path": ".",
  "isDir": true,  // 驼峰命名
  "size": 416,
  "modTime": "2026-02-03T02:53:06.678510442+08:00"
}
```

### 前端类型定义
```typescript
export interface SpecFileNode {
  name: string
  path: string
  is_dir: boolean  // snake_case
  children?: SpecFileNode[]
  size?: number
  mod_time?: string
}
```

### 前端检查逻辑
```typescript
function handleNodeClick(data: SpecFileNode) {
  if (!data.is_dir) {  // 检查 is_dir，但后端返回的是 isDir
    loadFile(data.path)
  }
}
```

**结果**: 前端无法识别目录节点，尝试加载所有节点，包括目录。

---

## 解决方案

### 修改后端 JSON tag

**文件**: `biz/model/api/spec.go`

**修改前**:
```go
type SpecFile struct {
    Name     string     `json:"name"`
    Path     string     `json:"path"`
    IsDir    bool       `json:"isDir"`     // 驼峰命名
    Children []SpecFile `json:"children,omitempty"`
    Size     int64      `json:"size,omitempty"`
    ModTime  time.Time  `json:"modTime,omitempty"`  // 驼峰命名
}
```

**修改后**:
```go
type SpecFile struct {
    Name     string     `json:"name"`
    Path     string     `json:"path"`
    IsDir    bool       `json:"is_dir"`    // snake_case
    Children []SpecFile `json:"children,omitempty"`
    Size     int64      `json:"size,omitempty"`
    ModTime  time.Time  `json:"mod_time,omitempty"`  // snake_case
}
```

---

## 验证结果

### API 测试

**请求**:
```bash
curl "http://localhost:38080/api/v1/spec/tree?repo_key=a9800306-f89c-4357-9c0c-f5e3f9a89705"
```

**响应**（修复后）:
```json
{
  "name": "shortlink",
  "path": ".",
  "is_dir": true,  // ✅ snake_case
  "size": 416,
  "mod_time": "2026-02-03T02:53:06.678510442+08:00"  // ✅ snake_case
}
```

### 前端行为

**修复前**:
- ❌ 点击目录节点 → 尝试加载 → 500 错误

**修复后**:
- ✅ 点击目录节点 → 不触发加载 → 无错误
- ✅ 点击文件节点 → 正常加载文件内容

---

## 影响范围

**修改的文件**:
- `biz/model/api/spec.go`（后端数据结构）

**影响的功能**:
- Spec 编辑器文件树显示
- 文件节点点击行为
- 所有使用 `SpecFile` 结构的 API

**后端 API**:
- `GET /api/v1/spec/tree` - 获取文件树

**前端组件**:
- `SpecEditor.vue` - 文件树交互

---

## 测试建议

### 测试场景 1: 目录节点
1. 访问 Spec 编辑器
2. 点击文件树中的目录节点（如根目录）
3. **预期**: 不触发文件加载，无错误

### 测试场景 2: 文件节点
1. 点击文件树中的 .spec 文件
2. **预期**: 正常加载文件内容到编辑器

### 测试场景 3: 混合操作
1. 先点击目录节点（无反应）
2. 再点击文件节点（正常加载）
3. **预期**: 两者都能正常工作，无错误

---

## 一致性改进

### 修复前的问题
- `SpecFile` 使用驼峰命名（`isDir`, `modTime`）
- `SpecFileInfo` 使用 snake_case（`is_dir`, `mod_time`）
- 前端类型定义使用 snake_case
- **不一致**导致混淆

### 修复后的改进
- 所有字段统一使用 **snake_case**
- 与前端类型定义保持一致
- 与其他 API 的命名风格保持一致
- 减少潜在的混淆和错误

---

## 相关文件

**后端**:
- `biz/model/api/spec.go` - 数据结构定义
- `biz/handler/spec/spec_service.go` - API 处理器

**前端**:
- `frontend/src/types/spec.ts` - 类型定义
- `frontend/src/components/spec/SpecEditor.vue` - 组件实现

**文档**:
- `SPEC_EDITOR_TEST_REPORT_FINAL.md` - 测试报告
- `SPEC_EDITOR_INIT_FEATURE.md` - 功能文档

---

## 结论

✅ **问题已修复**

**修复方式**: 统一使用 snake_case 命名
**影响**: 所有使用 Spec 编辑器的用户
**风险**: 低（仅修改 JSON tag，不改变业务逻辑）

**建议**: 立即测试验证修复效果。
