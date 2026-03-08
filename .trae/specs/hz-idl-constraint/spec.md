# Git Manage Service - Hz IDL 约束规范

## Overview
- **Summary**: 本规范旨在通过使用 Hz 工具约束 IDL 定义，确保所有 HTTP 接口都通过 IDL 进行规范化管理，提高代码一致性和可维护性。
- **Purpose**: 解决当前代码库中很多接口没有使用 Hz 约束 IDL 的问题，建立统一的接口定义和代码生成机制。
- **Target Users**: 开发团队成员，特别是负责 API 开发和维护的工程师。

## Goals
- 所有 HTTP 接口必须通过 IDL (Proto 文件) 定义
- 使用 Hz 工具生成 HTTP 代码，包括路由和处理器结构
- 替换手动路由注册为 Hz 生成的路由
- 确保所有处理器实现使用 Hz 生成的结构
- 建立统一的接口文档和代码生成流程

## Non-Goals (Out of Scope)
- 不修改现有的业务逻辑实现
- 不改变现有的 API 接口定义和行为
- 不涉及 RPC 接口的修改（Kitex 部分保持不变）

## Background & Context
- 当前代码库已经有完整的 IDL 目录结构和 proto 文件
- 存在代码生成脚本 `gen.sh`，但 Hz 生成的代码没有被充分利用
- 大部分路由是手动注册的，处理器实现也是手动的
- 需要建立统一的接口约束机制，确保接口定义和实现的一致性

## Functional Requirements
- **FR-1**: 所有 HTTP 接口必须在 IDL 文件中定义
- **FR-2**: 使用 Hz 工具生成 HTTP 代码，包括路由和处理器结构
- **FR-3**: 集成 Hz 生成的路由到项目中
- **FR-4**: 确保所有处理器实现使用 Hz 生成的结构
- **FR-5**: 建立统一的代码生成和更新流程

## Non-Functional Requirements
- **NFR-1**: 代码生成过程必须自动化，可通过脚本执行
- **NFR-2**: 生成的代码必须符合项目的代码风格和规范
- **NFR-3**: 集成过程不能破坏现有的功能
- **NFR-4**: 文档必须清晰，便于团队成员理解和遵循

## Constraints
- **Technical**: 使用 CloudWeGo Hertz 框架和 Hz 工具
- **Business**: 保持现有 API 接口的兼容性
- **Dependencies**: 依赖 protoc、hz 等工具

## Assumptions
- 所有现有的 API 接口都可以在 IDL 中定义
- Hz 工具可以正确生成所需的代码
- 团队成员已经熟悉基本的 protobuf 语法

## Acceptance Criteria

### AC-1: IDL 定义完整性
- **Given**: 现有所有 HTTP 接口
- **When**: 检查 IDL 文件
- **Then**: 所有接口都在 IDL 文件中定义
- **Verification**: `human-judgment`

### AC-2: Hz 代码生成
- **Given**: 完整的 IDL 文件
- **When**: 运行代码生成脚本
- **Then**: 成功生成 Hz 代码，包括路由和处理器结构
- **Verification**: `programmatic`

### AC-3: 路由集成
- **Given**: 生成的 Hz 路由代码
- **When**: 集成到项目中
- **Then**: 所有接口可以通过生成的路由访问
- **Verification**: `programmatic`

### AC-4: 处理器实现
- **Given**: 生成的 Hz 处理器结构
- **When**: 更新处理器实现
- **Then**: 所有处理器使用生成的结构，功能正常
- **Verification**: `programmatic`

### AC-5: 代码生成流程
- **Given**: 项目结构和脚本
- **When**: 执行代码生成脚本
- **Then**: 脚本可以正确生成和更新代码
- **Verification**: `programmatic`

## Open Questions
- [ ] 是否需要对现有的 IDL 文件进行修改以符合 Hz 的要求？
- [ ] 如何处理自定义路由和中间件？
- [ ] 如何确保生成的代码与现有代码风格一致？