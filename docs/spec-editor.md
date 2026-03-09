# Spec 编辑器

## 概述

Spec 编辑器是 Git Manage Service 提供的专业 RPM Spec 文件编辑工具，集成 Monaco Editor 和实时语法检查功能，帮助开发者高效编写和管理 RPM 打包规范文件。

> **GitHub 仓库**: [yi-nology/git-manage-service](https://github.com/yi-nology/git-manage-service)

## 功能特性

### 1. 文件树管理

- **自动扫描**：自动扫描仓库中的 `.spec` 文件
- **搜索过滤**：支持文件名搜索和过滤
- **目录结构**：显示完整的目录层级

### 2. Monaco 代码编辑器

- **语法高亮**：针对 RPM Spec 语法的定制高亮
- **代码补全**：智能补全常用宏和指令
- **行号显示**：清晰的行号和代码区域
- **实时编辑**：流畅的编辑体验

### 3. 实时语法检查

- **实时 Linting**：编辑时自动检查语法错误
- **问题面板**：显示所有错误、警告和信息
- **快速定位**：点击问题自动跳转到对应行
- **错误标记**：在编辑器中高亮显示问题位置

## 初始化 Spec 文件

### 自动引导

当进入一个没有 `.spec` 文件的仓库时，系统会自动显示空状态提示，点击 **"初始化 Spec 文件"** 按钮开始创建。

### 填写表单

| 字段 | 说明 | 示例 |
|------|------|------|
| 文件名 | Spec 文件名（不含 `.spec` 后缀） | `mypackage` |
| Name | 软件包名称 | `MyPackage` |
| Version | 版本号 | `1.0.0` |
| Release | 发布号 | `1` |
| Summary | 简要描述 | `A sample package` |
| License | 许可证（下拉选择） | `MIT`, `Apache-2.0`, `GPL-3.0` |
| URL | 项目主页 | `https://github.com/user/repo` |
| 描述 | 详细描述 | 支持多行文本 |

### 生成的模板

点击 **"创建并打开"** 后，系统会自动生成完整的 Spec 模板：

```spec
Name:           mypackage
Version:        1.0.0
Release:        1%{?dist}
Summary:        A sample package

License:        MIT
URL:            https://github.com/user/repo
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  make

%description
This is the description of mypackage

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
* Mon Mar 09 2026 Your Name <your.email@example.com> - 1.0.0-1
- Initial package
```

## 编辑操作

### 工具栏按钮

| 按钮 | 功能 | 快捷键 |
|------|------|--------|
| 🔄 刷新 | 重新加载文件树 | - |
| ✅ 检查 | 执行语法检查 | - |
| 💾 保存 | 保存当前文件 | `Ctrl+S` |
| 📝 Commit | 提交更改到 Git | - |

### 保存文件

1. 编辑文件后，状态栏显示 **"未保存"** 标签
2. 点击 **"保存"** 按钮
3. 如果存在错误级别的语法问题，系统会提示先修复

### 问题面板

底部问题面板显示所有检测结果：

- 🔴 **错误**：必须修复才能保存
- 🟡 **警告**：建议修复但不阻止保存
- ℹ️ **信息**：提示性信息

点击任意问题可跳转到对应行。

## 支持的许可证

初始化时支持选择以下许可证：

- MIT
- Apache License 2.0
- GNU General Public License v3.0
- BSD 2-Clause
- BSD 3-Clause
- Mozilla Public License 2.0
- ISC
- Unlicense

## 最佳实践

### 1. 文件命名

使用软件包名称作为文件名，如 `nginx.spec`、`python-requests.spec`。

### 2. 版本号规范

遵循 [语义化版本](https://semver.org/) 规范：
- 主版本号.次版本号.修订号
- 示例：`1.0.0`, `2.1.3`

### 3. 描述编写

- Summary：一句话简洁描述
- Description：详细说明软件功能、用途等

### 4. 定期检查

编辑完成后点击 **"检查"** 按钮，确保没有语法错误。

## 快速开始

1. 进入仓库详情页
2. 点击 **"Spec 编辑器"** 标签
3. 如果没有 spec 文件，点击 **"初始化 Spec 文件"**
4. 填写表单并创建
5. 在编辑器中修改和完善
6. 保存并提交

## 相关链接

- **GitHub**: [https://github.com/yi-nology/git-manage-service](https://github.com/yi-nology/git-manage-service)
- **Issues**: [https://github.com/yi-nology/git-manage-service/issues](https://github.com/yi-nology/git-manage-service/issues)
- **Releases**: [https://github.com/yi-nology/git-manage-service/releases](https://github.com/yi-nology/git-manage-service/releases)
- [Patch 管理](./patch-manager.md)
- [分支管理](./usage.md#分支管理)
- [部署指南](./deployment.md)
