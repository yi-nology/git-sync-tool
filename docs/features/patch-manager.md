# Patch 管理

![文档截图](images/docs/docs-patch-manager.png)

## 概述

Patch 管理功能帮助开发者生成、管理和应用 Git Patch 文件，支持批量操作和自动化工作流。

> **GitHub 仓库**: [yi-nology/git-manage-service](https://github.com/yi-nology/git-manage-service)

## 功能特性

### 1. Patch 列表管理

- **状态可视化**：清晰显示已应用、待应用、冲突状态
- **进度追踪**：显示应用进度条
- **批量操作**：支持批量应用多个 Patch

### 2. Patch 生成

支持两种生成方式：

#### 方式一：分支/Tag/Commit 范围

选择基准（起点）和目标（终点），自动生成差异 Patch。

#### 方式二：选择 Commits

从提交历史中选择特定 Commits，支持多选。

### 3. Patch 应用

- **自动检测**：应用前自动检查是否可以应用
- **冲突提示**：如有冲突显示详细信息
- **自动提交**：可选自动提交到 Git

## 生成 Patch

### 打开生成对话框

1. 进入仓库详情页
2. 点击 **"Patch 管理"** 标签
3. 点击 **"生成 Patch"** 按钮

### 选择生成方式

#### 方式一：范围模式

| 字段 | 说明 | 示例 |
|------|------|------|
| 基准（起点） | 分支、Tag 或 Commit | `main`, `v1.0.0`, `abc123` |
| 目标（终点） | 分支、Tag 或 Commit | `feature-branch`, `v2.0.0` |

#### 方式二：Commit 选择

- 从最近 50 条 Commits 中选择
- 支持多选（Ctrl/Cmd + 点击）
- 显示 Commit 简要信息

### 保存选项

| 选项 | 说明 |
|------|------|
| 保存到项目 | 保存到仓库的 patches 目录 |
| 文件名 | 描述性名称，如 `feature-login` |
| 保存路径 | 默认 `patches/`，可自定义 |
| 立即提交到 Git | 自动 commit 新创建的 Patch |

### 命名规则

系统自动生成序号前缀：
- 第一个 Patch：`001-description.patch`
- 第二个 Patch：`002-description.patch`
- 以此类推...

## 应用 Patch

### 单个应用

1. 找到待应用的 Patch
2. 点击 **"应用"** 按钮
3. 查看应用前检查结果
4. 填写提交消息（可选）
5. 点击 **"应用"** 确认

### 批量应用

1. 点击 **"批量应用"** 按钮
2. 确认应用数量
3. 系统按顺序依次应用

### 应用状态

| 状态 | 说明 |
|------|------|
| ✅ 已应用 | Patch 已成功应用到代码 |
| ⏳ 待应用 | Patch 等待应用 |
| ⚠️ 冲突 | Patch 存在冲突，无法应用 |

## Patch 操作

### 查看 Patch

点击 **"查看"** 按钮查看 Patch 内容：

```diff
From: abc123...
Subject: [PATCH] feat: add login feature

---
 src/login.ts | 50 +++++++++++++++++++++++++++++++++++++++++++++++++
 1 file changed, 50 insertions(+)

diff --git a/src/login.ts b/src/login.ts
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/src/login.ts
@@ -0,0 +1,50 @@
+// Login implementation
...
```

### 下载 Patch

点击 **"下载"** 按钮下载 `.patch` 文件到本地。

### 删除 Patch

1. 点击 **"删除"** 按钮
2. 确认删除操作
3. Patch 文件从仓库中移除

## 进度追踪

### 进度条

顶部进度条显示整体应用进度：

```
已应用 3 / 共 5 个 patch
```

- 绿色：正常进度
- 红色：存在冲突

### 快捷按钮

| 按钮 | 功能 |
|------|------|
| 应用下一个 | 快速应用下一个待应用的 Patch |
| 批量应用 | 一次应用所有待应用的 Patch |

## 最佳实践

### 1. 命名规范

使用描述性名称：
- `feature-login` - 新功能
- `fix-crash-bug` - Bug 修复
- `refactor-database` - 代码重构

### 2. 有序应用

按序号顺序应用 Patch，避免依赖问题。

### 3. 定期清理

删除已应用且不再需要的 Patch 文件。

### 4. 备份重要 Patch

下载保存重要的 Patch 文件作为备份。

## 使用场景

### 场景一：代码审查

1. 开发者生成 Patch
2. 发送给审查者
3. 审查者下载并应用
4. 检查代码变更
5. 反馈意见

### 场景二：版本迁移

1. 从稳定分支生成 Patch
2. 应用到开发分支
3. 保持功能同步

### 场景三：热修复部署

1. 修复紧急 Bug
2. 生成 Patch
3. 应用到生产环境
4. 快速部署

## 相关链接

- **GitHub**: [https://github.com/yi-nology/git-manage-service](https://github.com/yi-nology/git-manage-service)
- **Issues**: [https://github.com/yi-nology/git-manage-service/issues](https://github.com/yi-nology/git-manage-service/issues)
- **Releases**: [https://github.com/yi-nology/git-manage-service/releases](https://github.com/yi-nology/git-manage-service/releases)
- [Spec 编辑器](./spec-editor.md)
- [分支管理](./usage.md#分支管理)
- [Webhook 接口](./webhook.md)
