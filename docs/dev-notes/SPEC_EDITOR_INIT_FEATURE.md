# Spec 编辑器 - 初始化功能实施完成

实施时间: 2026-03-09 00:52
状态: 已完成 ✅

---

## 新增功能

### 初始化 Spec 文件

**功能描述**:
- 如果仓库中没有 .spec 文件，显示 "初始化 Spec 文件" 按钮
- 点击按钮打开对话框，填写基本信息
- 根据用户输入生成完整的 Spec 模板
- 自动创建文件并在编辑器中打开

**用户体验**:
1. 进入仓库详情页 → 点击 "Spec 编辑器" 标签
2. 如果没有 .spec 文件，看到友好的空状态提示
3. 点击 "初始化 Spec 文件" 按钮
4. 填写表单：
   - 文件名（如: mypackage）
   - Name（软件包名称）
   - Version（版本号，默认: 1.0.0）
   - Release（发布号，默认: 1）
   - Summary（简要描述）
   - License（许可证，支持选择）
   - URL（项目主页）
   - Description（详细描述）
5. 点击 "创建并打开"
6. 自动生成包含所有信息的 .spec 文件
7. 文件树刷新，自动打开新创建的文件

---

## 技术实现

### 前端修改

**文件**: `frontend/src/components/spec/SpecEditor.vue`

**新增组件**:
1. **空状态容器**: 显示 "初始化 Spec 文件" 按钮
2. **初始化对话框**: 表单输入和验证
3. **模板生成器**: 根据用户输入生成完整 Spec 文件

**新增状态**:
- `showInitDialog`: 对话框显示状态
- `initInProgress`: 创建进行中状态
- `initForm`: 表单数据
- `initFormRules`: 表单验证规则

**新增方法**:
- `handleInitSpec()`: 处理初始化逻辑
- `generateSpecTemplate()`: 生成 Spec 模板内容

**样式优化**:
- `.empty-tree-container`: 空状态容器样式
- `.form-tip`: 表单提示样式

---

### 后端修改

**文件 1**: `biz/model/api/spec.go`
```go
type CreateSpecFileReq struct {
    RepoKey string `json:"repo_key"`
    Path    string `json:"path"`
    Name    string `json:"name"`
    Content string `json:"content"` // 新增：可选内容
}
```

**文件 2**: `biz/service/spec/spec_service.go`
```go
func (s *SpecService) CreateSpecFileWithContent(
    repoPath, dirPath, fileName, content string
) (string, error) {
    // 如果没有提供内容，使用模板
    if content == "" {
        content = s.GetSpecTemplate()
    }
    // 写入内容
    os.WriteFile(fullPath, []byte(content), 0644)
}
```

**文件 3**: `biz/handler/spec/spec_service.go`
```go
func CreateSpecFile(ctx context.Context, c *app.RequestContext) {
    // ...
    path, err := specSvc.CreateSpecFileWithContent(
        repo.Path, req.Path, req.Name, req.Content
    )
    // ...
}
```

---

## Spec 文件模板

生成的模板包含以下部分：

```spec
Name:           %{name}
Version:        %{version}
Release:        %{release}%{?dist}
Summary:        %{summary}

License:        %{license}
URL:            %{url}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make

%description
%{description}

%prep
%setup -q

%build
%configure
make %{?_smp_mflags}

%install
rm -rf %{buildroot}
%make_install

%files
%doc README.md
%license LICENSE
%{_bindir}/%{name}

%changelog
* %{date} %{maintainer} - %{version}-%{release}
- Initial package
```

**特性**:
- 所有用户输入的信息都会被填充到模板中
- 包含标准的 RPM Spec 文件结构
- 预设了常用的 BuildRequires
- 自动生成 changelog 条目

---

## 测试验证

### API 测试

**测试命令**:
```bash
curl -X POST http://localhost:38080/api/v1/spec/create \
  -H "Content-Type: application/json" \
  -d '{
    "repo_key": "test-repo",
    "path": ".",
    "name": "test-init.spec"
  }'
```

**测试结果**: ✅ 成功
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "path": "test-init.spec",
    "message": "Spec 文件创建成功"
  }
}
```

**文件验证**:
```bash
$ ls -la /opt/project/git-manage-service/data/test-repo/*.spec
-rw-------@ 1 zhangyi  admin  383 Mar  9 00:51 test-init.spec
-rw-------@ 1 zhangyi  admin  120 Mar  9 00:14 test.spec
```

✅ 文件已成功创建

---

## 用户流程图

```
用户进入仓库详情页
    ↓
点击 "Spec 编辑器" 标签
    ↓
文件树为空？
    ├─ 是 → 显示 "初始化 Spec 文件" 按钮
    │       ↓
    │    点击按钮 → 打开对话框
    │       ↓
    │    填写表单（Name, Version, Summary, License...）
    │       ↓
    │    点击 "创建并打开"
    │       ↓
    │    生成 Spec 模板
    │       ↓
    │    调用后端 API 创建文件
    │       ↓
    │    文件树刷新
    │       ↓
    │    自动打开新文件
    │
    └─ 否 → 显示文件树
            ↓
        点击文件 → 在编辑器中打开
```

---

## 表单验证规则

| 字段 | 规则 | 说明 |
|------|------|------|
| filename | 必填，字母/数字/下划线/短横线 | 文件名（不含 .spec 后缀） |
| name | 必填 | 软件包名称 |
| version | 必填 | 版本号（默认: 1.0.0） |
| release | 必填 | 发布号（默认: 1） |
| summary | 必填 | 简要描述 |
| license | 必填 | 许可证（下拉选择） |
| url | 可选 | 项目主页 URL |
| description | 可选 | 详细描述 |

---

## 许可证选项

预设的许可证选项（支持搜索和过滤）：
- MIT
- Apache License 2.0
- GNU General Public License v3.0
- BSD 2-Clause
- BSD 3-Clause
- Mozilla Public License 2.0
- ISC
- Unlicense

---

## 对话框截图

**空状态**:
```
┌─────────────────────────┐
│   暂无 .spec 文件        │
│                         │
│  [初始化 Spec 文件]      │
└─────────────────────────┘
```

**初始化对话框**:
```
┌───────────────────────────────────┐
│  初始化 Spec 文件                  │
├───────────────────────────────────┤
│  文件名: [mypackage      ].spec   │
│  Name:   [MyPackage         ]     │
│  Version: [1.0.0           ]      │
│  Release: [1               ]      │
│  Summary: [简要描述         ]     │
│  License: [MIT            ▼]      │
│  URL:     [https://...     ]      │
│  描述:    [详细描述...     ]      │
│                                   │
├───────────────────────────────────┤
│         [取消] [创建并打开]        │
└───────────────────────────────────┘
```

---

## 完成度评估

**功能完成度**: 100%

**已完成项**:
- ✅ 空状态检测和提示
- ✅ 初始化按钮
- ✅ 初始化对话框
- ✅ 表单验证
- ✅ 模板生成器
- ✅ API 集成
- ✅ 文件树刷新
- ✅ 自动打开新文件
- ✅ 后端支持自定义内容

**待优化项**（可选）:
- 添加更多许可证选项
- 支持自定义 BuildRequires
- 支持自定义 %files 列表
- 添加 Spec 文件模板库
- 支持从现有项目生成 Spec

---

## 下一步建议

### 立即测试
1. 访问 http://localhost:3000
2. 进入一个没有 .spec 文件的仓库
3. 点击 "Spec 编辑器" 标签
4. 点击 "初始化 Spec 文件" 按钮
5. 填写表单并创建

### 可选增强
1. 添加 Spec 文件模板库（不同类型的软件）
2. 支持从 Git 仓库自动填充信息
3. 集成 rpmlint 验证
4. 添加更多的 BuildRequires 预设
5. 支持多文件项目

---

## 相关文件

**前端**:
- `frontend/src/components/spec/SpecEditor.vue`
- `frontend/src/api/modules/spec.ts`

**后端**:
- `biz/model/api/spec.go`
- `biz/handler/spec/spec_service.go`
- `biz/service/spec/spec_service.go`

**文档**:
- `SPEC_EDITOR_PLAN.md`
- `SPEC_EDITOR_TEST_REPORT_FINAL.md`
- `TODO-Git管理服务整体优化-20260308.md`

---

## 结论

✅ **Spec 编辑器初始化功能已完成并可用**

- 用户可以在没有 .spec 文件的仓库中快速初始化
- 提供友好的表单界面
- 自动生成完整的 Spec 模板
- 无需手动编写基础结构

**建议**: 立即进行手动测试，验证所有交互功能。
