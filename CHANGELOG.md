# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2026-03-08

### Added
- **UI/UX 优化**
  - 深色模式支持（可通过主题切换按钮切换）
  - 全局快捷键支持（Cmd/Ctrl + K 搜索，Cmd/Ctrl + N 新建）
  - 统一的 Toast 通知系统（useNotification composable）
  - 通用分页功能（usePagination composable）
  - 骨架屏加载组件（TableSkeleton、AppSkeleton）
  - 深色模式 CSS 变量支持
  - 优化的滚动条样式

- **新增组件**
  - `ThemeSwitch.vue`: 深色模式切换组件
  - `AppSkeleton.vue`: 通用骨架屏组件
  - `TableSkeleton.vue`: 表格骨架屏组件

- **新增 Composables**
  - `useNotification`: 统一的消息通知管理
  - `usePagination`: 通用的分页逻辑
  - `useKeyboard`: 全局快捷键支持

- **新增 Store**
  - `useUIStore`: UI 全局状态管理（深色模式、侧边栏等）

### Changed
- **仓库列表页优化** (`RepoListPage.vue`)
  - 添加搜索/筛选功能
  - 添加分页功能（10/20/50/100 条每页）
  - 添加骨架屏加载动画
  - 优化操作按钮（使用下拉菜单，减少视觉负担）
  - 添加表格排序功能
  - 优化空状态显示
  - 响应式设计优化

- **全局样式优化** (`style.css`)
  - 添加深色模式 CSS 变量
  - 优化滚动条样式
  - 添加过渡动画

- **布局优化** (`AppLayout.vue`)
  - 添加深色模式切换按钮
  - 添加快捷键提示
  - 优化响应式布局
  - 深色模式样式支持

### Fixed
- 修复 TypeScript 类型错误
- 修复分页逻辑中的类型定义问题
- 修复组件中的未使用导入警告

### Performance
- 优化大数据集渲染性能（分页）
- 添加骨架屏减少加载焦虑

## [Previous Releases]

查看 GitHub Releases 页面获取历史版本信息。
