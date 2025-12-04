# EulerGuard 3.0 技术规划书

**AI 原生内核安全智能体 (AI-Native Kernel Security Agent)**

---

## 📑 目录

| 章节 | 内容 | 状态 |
|------|------|------|
| [一、产品愿景](#一产品愿景) | 产品定位、核心设计哲学、三层架构 | 📋 |
| [二、Phase 0: 代码重构](#二phase-0-代码重构) | Package 结构、事件存储、初始化流程 | ⏳ |
| [三、Phase 0.5: BPF 与 Prompt 重构](#三phase-05-bpf-与-prompt-重构) | 事件结构优化、性能提升、Prompt 设计 | ⏳ |
| [四、Phase 1: 全息遥测仓库](#四phase-1-全息遥测仓库) | TimeRingBuffer、倒排索引、进程画像 | ⏳ |
| [五、Phase 2: 双模执行引擎](#五phase-2-双模执行引擎) | Shadow Mode、模拟引擎 | ⏳ |
| [六、Phase 3: AI 接口层](#六phase-3-ai-接口层) | 意图解析、规则生成、Sentinel | ⏳ |
| [七、Phase 4: AI 原生前端](#七phase-4-ai-原生前端) | Omnibox、Observatory、Policy Studio | ⏳ |
| [八、总结与展望](#八总结与展望) | 核心创新点总结 | 📋 |
| [附录 A: 后端 API 总览](#附录-a-后端-api-总览) | 完整 API 列表与页面依赖 | 📋 |

**状态说明**: 📋 规划中 | ⏳ 待开发 | 🚧 开发中 | ✅ 完成

---

## 一、产品愿景

### 1.1 产品定位
EulerGuard 不是一个"带 AI 功能的安全工具"，而是一个 **"AI 原生的内核安全智能体"**。

**传统安全工具** vs **AI 原生安全智能体**：
| 维度 | 传统工具 | EulerGuard 3.0 |
|------|---------|----------------|
| 规则来源 | 人工编写 | AI 生成 + 人工审核 |
| 交互方式 | GUI 配置 | 自然语言对话 |
| 响应模式 | 被动告警 | 主动洞察 + 自动建议 |
| 学习能力 | 无 | 持续学习行为基线 |
| 决策辅助 | 无 | 模拟预览 + 风险评估 |

### 1.2 核心设计哲学

**第一性原则：AI First（AI 优先）**
- 所有用户交互优先通过自然语言完成
- AI 不是"附加功能"，而是系统的"神经中枢"
- 人类专家的角色从"编写规则"转变为"审核 AI 建议"

**三层架构**：
```
┌─────────────────────────────────────────────────────────────────┐
│                    🧠 AI Intelligence Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ NL Parser   │  │ Rule Gen    │  │ Anomaly Detection       │ │
│  │ 自然语言解析 │  │ 规则生成     │  │ 异常检测                │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ Sentinel    │  │ Simulation  │  │ Context Reasoning       │ │
│  │ 主动巡检     │  │ 影响预测     │  │ 上下文推理              │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    📊 Data & Storage Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ TimeRing    │  │ Indexer     │  │ Process Profile         │ │
│  │ 时序存储     │  │ 倒排索引     │  │ 进程画像                │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    ⚙️ Kernel Enforcement Layer                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐ │
│  │ eBPF/LSM    │  │ Rule Engine │  │ Shadow Mode             │ │
│  │ 内核探针     │  │ 规则引擎     │  │ 影子模式                │ │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

1. **内核执行层 (Kernel Enforcement Layer)**
   - eBPF/LSM 探针：实时采集内核事件
   - 规则引擎：执行 Enforce/Shadow 规则
   - 这一层是"肌肉"，负责执行

2. **数据存储层 (Data & Storage Layer)**
   - 时序数据库 + 倒排索引：为 AI 提供记忆
   - 进程画像：为 AI 提供上下文
   - 这一层是"记忆"，支撑 AI 分析

3. **AI 智能层 (AI Intelligence Layer)** ⭐ 核心
   - 自然语言解析：理解用户意图
   - 规则生成：将意图转化为策略
   - 异常检测：发现未知威胁
   - 主动巡检：持续优化建议
   - 这一层是"大脑"，驱动整个系统

---

## 二、Phase 0: 代码重构

> **目标**: 重构现有代码结构，使其直接服务于后续阶段的功能开发，避免过度设计。

### 2.1 重新组织 Package 结构
- **当前问题**：package 职责不清晰，`ui` 包混合了 Web 服务器、事件桥接、统计等多种职责
- **重构方案**：按功能模块重新组织
  - **`pkg/ebpf/`**：eBPF 程序加载与链接（保持不变）
  - **`pkg/events/`**：事件定义、解码、基础类型（保持不变）
  - **`pkg/storage/`**：**新增**，为第一阶段做准备
    - 将 `ui.Stats` 中的事件存储逻辑抽离
    - 预留 `ringbuffer.go` 和 `indexer.go` 的位置（第一阶段实现）
    - `stats.go`：统计聚合（从 `ui.Stats` 简化而来，只保留计数和告警）
  - **`pkg/rules/`**：规则引擎（保持不变，第二阶段扩展）
  - **`pkg/proc/`**：进程树管理（保持不变，第一阶段添加 `profile.go`）
  - **`pkg/workload/`**：工作负载注册（保持不变）
  - **`pkg/ai/`**：AI 服务（保持不变，第三阶段扩展）
  - **`pkg/simulation/`**：**新增空目录**，第二阶段实现
  - **`pkg/api/`**：**新增空目录**，第三阶段实现语义查询
  - **`pkg/server/`**：Web 服务器（从 `ui` 重命名，只保留 HTTP 路由和 WebSocket）
  - **`pkg/cli/`**：CLI 模式（保持不变）
  - **`pkg/config/`**：配置管理（保持不变，仅优化常量管理）
  - **`pkg/types/`**：共享类型定义（保持不变）
  - **`pkg/metrics/`**：速率统计（保持不变，CLI 专用）
  - **`pkg/output/`**：输出和日志（保持不变，CLI 专用）
  - **`pkg/profiler/`**：学习模式分析（保持不变，CLI 和 Web 都使用）
  - **`pkg/utils/`**：工具函数（保持不变，被多个包使用）
  - **`pkg/core/`**：**新增**，统一初始化流程
- **价值**：清晰的模块划分，每个 package 职责单一，便于后续扩展

### 2.2 彻底重构事件存储，删除旧实现
- **当前痛点**：
  - `ui.Stats` 混合了统计计数、事件存储、前端事件转换等多种职责
  - `RecentExecs`、`RecentFiles`、`RecentConnects` 使用简单切片，容量只有 50 条
  - 这些旧实现无法支撑第一阶段的需求，必须完全删除
- **重构方案**：
  - **删除 `ui.Stats` 中的事件存储相关代码**：
    - 删除 `RecentExecs()`, `RecentFiles()`, `RecentConnects()` 方法
    - 删除 `recentExecs`, `recentFiles`, `recentConnects` 字段
    - 删除 `RecordExec()`, `RecordFileEvent()`, `RecordConnectEvent()` 中的切片追加逻辑
  - **创建 `pkg/storage/` 包**：
    - `storage/stats.go`：只保留统计计数（execCount, fileCount, connectCount, alerts）
      - **职责**：无状态的统计聚合，不涉及事件存储
      - **接口**：`RecordExec()`, `RecordFile()`, `RecordConnect()`, `AddAlert()`, `Counts()`, `Rates()`
    - `storage/store.go`：定义存储接口
      - **职责**：定义 `EventStore` 接口，不包含实现
      - **接口**：`Append()`, `Query()`, `Latest()` 等
    - `storage/ringbuffer.go`：**第一阶段实现**，实现 `EventStore` 接口
    - `storage/indexer.go`：**第一阶段实现**，倒排索引
    - **职责分离**：`stats.go` 和 `store.go` 完全独立，`stats.go` 不依赖存储实现
  - **事件转换逻辑**：
    - 从 `ui` 包移到 `events/transform.go`
    - 前端事件类型定义移到 `types/events.go`（与后端事件类型分离）
  - **简化 `server.Bridge`**（原 `ui.Bridge`）：
    - 只负责事件转发和告警生成
    - 事件存储直接调用 `storage` 包接口
- **删除的文件/代码**：
  - `ui/stats.go` 中的 `RecentExecs()`, `RecentFiles()`, `RecentConnects()` 方法
  - `ui/stats.go` 中的 `recentExecs`, `recentFiles`, `recentConnects` 字段及相关逻辑
  - `ui/stats.go` 中的 `maxRecent` 常量
- **价值**：彻底清理旧代码，为第一阶段的 `TimeRingBuffer` 和 `Indexer` 实现扫清障碍

### 2.3 彻底简化 `tracer` 包，删除无用代码
- **当前痛点**：
  - `tracer.Core` 承担过多职责：eBPF 加载、规则加载、BPF Map 填充、事件分发
  - `PopulateMonitoredFiles` 和 `PopulateBlockedPorts` 逻辑复杂，与核心职责无关
  - `ReloadRules()` 方法混合了规则加载和 BPF Map 更新
- **重构方案**：
  - **删除 `tracer.Core` 结构体**，拆分为更小的组件：
    - `ebpf/loader.go`：eBPF 程序加载与链接（已存在，优化）
    - `ebpf/maps.go`：**新增**，BPF Map 填充逻辑（从 `tracer` 移入）
    - `tracer/reader.go`：只负责 RingBuffer Reader 的封装
  - **删除 `tracer.Core.Init()`**：
    - 改为 `ebpf.Load()` 和 `ebpf.Attach()` 等独立函数
    - 规则加载由 `rules` 包负责
    - BPF Map 填充由 `ebpf/maps.go` 负责
    - 初始化逻辑统一到 `core.Bootstrap()`
  - **删除 `tracer.Core.ReloadRules()`**：
    - 规则重载逻辑移到调用方（`cli` 和 `server` 包）
    - 使用 `ebpf/maps.go` 中的函数更新 BPF Map
  - **简化 `tracer.DispatchEvent()`**：
    - 只负责事件解码和分发
    - 事件存储由 `storage` 包处理
  - **`tracer.EventLoop()` 保持不变**：
    - 只负责从 RingBuffer 读取并调用 `DispatchEvent()`
    - 不依赖 `tracer.Core` 结构体
    - 改为接收独立的组件参数（`*ringbuf.Reader`, `*events.HandlerChain` 等）
  - **`tracer` 包中的其他函数**：
    - `AttachLSMHooks()`：移到 `ebpf/attach.go` 或保留在 `tracer/`（如果与事件处理相关）
    - `CloseLinks()`：移到 `ebpf/attach.go` 或保留在 `tracer/`
    - `LoadRules()`：删除，由 `rules` 包或调用方负责
    - `PopulateMonitoredFiles()`, `PopulateBlockedPorts()`：移到 `ebpf/maps.go`
    - `RepopulateMonitoredFiles()`, `RepopulateBlockedPorts()`：移到 `ebpf/maps.go`
    - `extractParentFilename()`, `newPIDResolver()`：移到 `ebpf/maps.go` 或 `ebpf/utils.go`
- **删除的代码**：
  - `tracer/core.go` 中的 `Core` 结构体（完全删除）
  - `tracer/core.go` 中的 `Init()`, `ReloadRules()`, `Close()` 方法
  - `tracer/core.go` 中的 `PopulateMonitoredFiles`, `PopulateBlockedPorts` 等函数
  - `tracer/core.go` 中的 `LoadRules()` 函数（移到 `rules` 包或调用方）
  - `ui.App.core` 字段（改为使用 `core.Bootstrap()` 返回的组件）
  - `cli.CLI.Core` 字段（改为使用 `core.Bootstrap()` 返回的组件）
- **更新的代码**：
  - `ui.App.Run()`：改为调用 `core.Bootstrap()`，不再调用 `tracer.Init()`
  - `cli.RunCLI()`：改为调用 `core.Bootstrap()`，不再调用 `tracer.Init()`
  - `tracer.EventLoop()`：改为接收独立的组件参数，不依赖 `Core` 结构体
- **价值**：`tracer` 包职责单一，代码更清晰，便于测试和维护

### 2.4 统一初始化流程，消除重复代码
- **当前痛点**：
  - `ui.App.Run()` 和 `cli.RunCLI()` 有大量重复的初始化代码
  - 组件初始化顺序不清晰
- **重构方案**：
  - **创建 `pkg/core/bootstrap.go`**（简单实用，不过度设计）：
    - `Bootstrap()` 函数：统一的初始化流程
    - 返回 `*CoreComponents` 结构体，包含所有核心组件
    - 分阶段初始化，清晰的错误处理
  - **`CoreComponents` 结构**：
    ```go
    type CoreComponents struct {
        EBpfObjs    *ebpf.LSMObjects
        EBpfLinks   []link.Link
        Reader      *ringbuf.Reader
        ProcessTree *proc.ProcessTree
        WorkloadReg *workload.Registry
        RuleEngine  *rules.Engine
        Rules       []types.Rule
    }
    ```
  - **`cli` 和 `server` 包**：
    - 都调用 `core.Bootstrap()` 初始化核心组件
    - 然后各自添加特定逻辑（CLI 的 printer，Web 的 stats/bridge）
  - **`cmd/cli.go` 和 `cmd/web.go`**：
    - 保持不变，只调用 `cli.RunCLI()` 和 `server.RunWebServer()`
    - 初始化逻辑已移到 `core.Bootstrap()`
- **价值**：消除重复代码，初始化流程清晰，便于维护

### 2.5 优化配置管理，集中常量定义
- **当前痛点**：
  - 硬编码常量散布在代码中（如 `maxRecent: 50`, `maxAlerts: 100`）
  - 配置项与常量混用
- **重构方案**：
  - **`pkg/config/constants.go`**：
    - 定义所有默认值和常量
    - 如：`DefaultRecentEventsCapacity = 50`（第一阶段将改为 10000+）
    - 如：`DefaultMaxAlerts = 100`
  - **配置结构优化**（可选，不强制）：
    - 在 `config.Options` 中添加 `Storage` 子结构
    - 为第一阶段的存储配置预留位置
- **价值**：常量集中管理，便于后续调整和扩展

### 2.6 重构事件处理流程，删除冗余逻辑
- **当前痛点**：
  - `HandlerChain` 设计简单，但事件处理逻辑分散
  - `ui.Bridge` 和 `cli.alertHandler` 有重复的事件处理逻辑
  - 事件转换逻辑混在 UI 层
- **重构方案**：
  - **统一事件处理流程**：
    - `tracer.DispatchEvent()` 中：事件解码 → 存储到 `storage` → 分发到 HandlerChain
    - 所有事件必须经过存储层，不再有"可选"的存储逻辑
  - **简化 HandlerChain**：
    - `server.Bridge`：只负责告警生成和前端通知
    - `cli.alertHandler`：只负责 CLI 输出
    - 删除重复的事件处理逻辑
  - **删除无用代码**：
    - `ui.Bridge` 中的事件转换逻辑（移到 `events/transform.go`）
    - `ui.Bridge` 中的 `SetRuleEngine()`, `SetWorkloadRegistry()` 等 Setter（改为构造函数注入）
    - `ui.Stats` 中的 `PublishEvent()`, `SubscribeEvents()` 等发布订阅逻辑（如果不需要）
- **价值**：事件处理流程清晰，删除冗余代码，为存储层集成做好准备

### 2.7 彻底清理无用代码
- **删除的文件**：
  - `pkg/ui/` 整个目录（重命名为 `pkg/server/`，删除旧实现）
  - `pkg/tracer/core.go`（功能拆分到其他包，但保留 `EventLoop` 和 `DispatchEvent` 函数）
- **删除的函数/方法**：
  - `ui.Stats.RecentExecs()`, `RecentFiles()`, `RecentConnects()`
  - `ui.Stats.RecordExec()`, `RecordFileEvent()`, `RecordConnectEvent()` 中的切片逻辑
  - `tracer.Core.Init()`, `ReloadRules()`, `Close()`
  - `tracer.PopulateMonitoredFiles()`, `PopulateBlockedPorts()`, `Repopulate*()`
  - `tracer.LoadRules()`（移到调用方或 `rules` 包）
  - `ui.Bridge.SetRuleEngine()`, `SetWorkloadRegistry()`, `SetProfiler()`（改为构造函数注入）
  - `ui.App.core` 字段（改为使用 `core.Bootstrap()` 返回的组件）
  - `cli.CLI.Core` 字段（改为使用 `core.Bootstrap()` 返回的组件）
- **删除的字段/变量**：
  - `ui.Stats` 中的 `recentExecs`, `recentFiles`, `recentConnects`, `recentMu`, `maxRecent`
  - `tracer.Core` 结构体中的所有字段（结构体已删除）
  - 未使用的导入和变量
- **保留的代码**：
  - `tracer.EventLoop()` 和 `tracer.DispatchEvent()`：保留，但改为接收独立参数
  - `tracer.AttachLSMHooks()` 和 `tracer.CloseLinks()`：保留，移到 `ebpf/` 包或保留在 `tracer/`
  - `profiler` 包：保持不变，CLI 和 Web 都使用
  - `metrics` 和 `output` 包：保持不变，CLI 专用
- **代码优化**：
  - **函数拆分**：将长函数拆分为小函数
  - **错误处理**：统一使用 `fmt.Errorf` 包装错误，添加上下文
  - **GoDoc 注释**：为所有公开接口添加注释
- **价值**：代码库更精简，只保留必要的代码，降低维护成本

### 2.8 测试说明
- **单元测试**：
  - `pkg/core/bootstrap_test.go`：测试 `Bootstrap()` 函数的初始化流程
  - `pkg/ebpf/maps_test.go`：测试 BPF Map 填充逻辑（使用 Mock eBPF Map）
  - `pkg/storage/stats_test.go`：测试统计计数功能
  - `pkg/events/transform_test.go`：测试事件转换逻辑
- **集成测试**：
  - `pkg/core/bootstrap_integration_test.go`：测试完整的初始化流程（需要 root 权限）
  - `pkg/server/bridge_test.go`：测试事件桥接功能（Mock 依赖）
- **测试覆盖率目标**：
  - 新增代码覆盖率 > 80%
  - 关键路径（初始化、事件处理）覆盖率 > 90%
- **测试工具**：
  - 使用 `testify` 进行断言和 Mock
  - 使用 `testcontainers` 或 Mock 对象模拟 eBPF 环境

### 2.9 模块化清晰度确认
- **分层架构**（清晰的依赖方向，避免循环依赖）：
  - **基础设施层**（最底层，不依赖业务逻辑）：
    - `pkg/ebpf/`：eBPF 程序加载、链接、Map 管理
    - `pkg/events/`：事件定义、解码（纯数据结构）
    - `pkg/types/`：共享类型定义（纯数据结构）
    - `pkg/utils/`：工具函数（无状态函数）
    - `pkg/config/`：配置管理（纯配置）
  - **核心业务层**（依赖基础设施层）：
    - `pkg/storage/`：事件存储和统计（依赖 `events`, `types`）
    - `pkg/rules/`：规则引擎（依赖 `events`, `types`）
    - `pkg/proc/`：进程树管理（依赖 `events`, `types`）
    - `pkg/workload/`：工作负载注册（依赖 `proc`, `types`）
    - `pkg/tracer/`：事件分发（依赖 `events`, `storage`, `proc`, `workload`）
  - **服务层**（依赖核心业务层）：
    - `pkg/ai/`：AI 服务（依赖 `storage`, `proc`, `rules`, `types`）
    - `pkg/simulation/`：模拟引擎（依赖 `storage`, `rules`）
    - `pkg/api/`：语义查询（依赖 `storage`, `ai`）
  - **应用层**（依赖所有下层）：
    - `pkg/core/`：统一初始化（依赖所有核心业务层）
    - `pkg/server/`：Web 服务器（依赖 `core`, `storage`, `rules`, `ai`, `api`）
    - `pkg/cli/`：CLI 模式（依赖 `core`, `output`, `metrics`, `profiler`）
  - **辅助包**（被应用层使用）：
    - `pkg/metrics/`：速率统计（CLI 专用）
    - `pkg/output/`：输出和日志（CLI 专用）
    - `pkg/profiler/`：学习模式分析（CLI 和 Web 都使用）
- **职责边界清晰**：
  - ✅ **`storage`**：只负责事件存储和统计，不涉及业务逻辑
  - ✅ **`rules`**：只负责规则匹配，不涉及存储和事件分发
  - ✅ **`tracer`**：只负责事件解码和分发，不涉及存储实现
  - ✅ **`proc`**：只负责进程树管理，不涉及事件处理
  - ✅ **`server`**：只负责 HTTP/WebSocket，不涉及核心业务逻辑
  - ✅ **`cli`**：只负责 CLI 输出，不涉及核心业务逻辑
- **依赖关系验证**：
  - ✅ **无循环依赖**：依赖方向单向（基础设施 → 核心业务 → 服务 → 应用）
  - ✅ **接口隔离**：`storage` 通过接口暴露，`rules` 通过接口暴露
  - ✅ **最小依赖**：每个包只依赖必要的包，不引入不必要的依赖
- **模块边界明确**：
  - **`storage` 包内部职责分离**：
    - `stats.go`：统计计数（无状态聚合，独立于存储实现）
    - `store.go`：存储接口定义（纯接口，无实现）
    - `ringbuffer.go`：存储实现（第一阶段，实现 `EventStore` 接口）
    - `indexer.go`：索引实现（第一阶段，与 `ringbuffer` 协同工作）
    - **边界**：`stats.go` 不依赖 `store.go`，两者完全独立
  - **`events` 包职责明确**：
    - `types.go`：事件类型定义（纯数据结构）
    - `decoder.go`：事件解码（从字节流解码）
    - `handler.go`：事件处理器接口（`EventHandler`）
    - `transform.go`：事件格式转换（后端事件 → 前端事件，不涉及业务逻辑）
  - **`tracer` 包简化后职责单一**：
    - `EventLoop()`：事件循环（从 RingBuffer 读取）
    - `DispatchEvent()`：事件分发（解码 → 存储 → HandlerChain）
    - **不包含**：存储实现、规则加载、BPF Map 管理等职责
  - **`ebpf` 包职责明确**：
    - `loader.go`：eBPF 程序加载
    - `attach.go`：LSM Hook 链接（从 `tracer` 移入）
    - `maps.go`：BPF Map 填充和管理（从 `tracer` 移入）
    - **不包含**：事件处理、规则匹配等业务逻辑
  - **`core` 包职责明确**：
    - `bootstrap.go`：统一初始化流程
    - **只负责**：组件初始化和组装，不包含业务逻辑
    - **依赖方向**：依赖所有核心业务层，但不被业务层依赖
- **依赖关系图**（确保无循环依赖）：
  ```
  应用层: server, cli
    ↓
  服务层: api, simulation, ai
    ↓
  核心业务层: tracer, storage, rules, proc, workload
    ↓
  基础设施层: ebpf, events, types, utils, config
  ```
  - **依赖规则**：只能依赖下层，不能依赖上层或同层
  - **例外**：`tracer` 可以依赖 `storage`（同层，但 `storage` 不依赖 `tracer`）
  - **验证**：通过 `go mod graph` 或静态分析工具验证无循环依赖

**重构与后续阶段的对接关系**
- **第零阶段 → 第一阶段**：
  - `pkg/storage/` 包已创建，`store.go` 定义接口
  - 第一阶段直接实现 `ringbuffer.go` 和 `indexer.go`，实现 `store.go` 中的接口
  - `proc/profile.go` 在第一阶段添加，无需重构准备
  - **旧代码已完全删除**，无向后兼容负担
- **第零阶段 → 第二阶段**：
  - `pkg/rules/` 包结构清晰，直接扩展 `engine.go` 支持 Shadow Mode
  - `pkg/simulation/` 目录已创建，第二阶段直接实现 `runner.go`
  - 模拟引擎依赖 `storage` 包的 `TimeRingBuffer`（第一阶段已实现）
- **第零阶段 → 第三阶段**：
  - `pkg/api/` 目录已创建，第三阶段直接实现 `query.go`
  - `pkg/ai/sentinel.go` 直接添加，依赖第一阶段和第二阶段的功能
  - 语义查询依赖 `storage` 包的 `Indexer`（第一阶段已实现）
- **重构原则**：
  - ✅ **彻底删除**：旧实现完全删除，不保留临时代码
  - ✅ **直接对接**：重构后的结构直接服务于后续阶段
  - ✅ **清晰模块化**：每个包职责单一，依赖关系清晰，无循环依赖
  - ❌ **不过度抽象**：不创建复杂的 DI 容器，不定义过多接口
  - ❌ **不提前实现**：不实现第一阶段的功能，只做结构准备
  - ❌ **不保留无用代码**：所有无用代码必须删除

---

## 三、Phase 0.5: BPF 与 Prompt 重构

> **目标**: 重构 BPF 数据采集以支持 AI 原生功能，最大化性能，重新设计 Prompt 体系以支撑意图解析、规则生成等核心 AI 能力。

### 3.1 BPF 数据采集现状分析

**当前采集的事件类型**：
| 事件类型 | LSM Hook | 采集字段 | 用途 |
|---------|----------|---------|------|
| `exec_event` | `bprm_check_security` | PID, PPID, CgroupID, Comm, PComm, Filename, Blocked | 进程执行监控 |
| `file_open_event` | `file_open` | PID, CgroupID, Flags, Ino, Dev, Filename, Blocked | 文件访问监控 |
| `connect_event` | `socket_connect` | PID, CgroupID, Family, Port, AddrV4/V6, Blocked | 网络连接监控 |

**当前问题**：
1. **缺少时间戳**：事件没有内核时间戳，依赖用户态填充，精度不足
2. **缺少 UID/GID**：无法实现"非 root 进程"等条件判断
3. **缺少完整路径**：文件事件只有 parent/filename，无法获取完整绝对路径
4. **缺少命令行参数**：exec 事件没有 argv，AI 分析时缺少上下文
5. **缺少进程 Comm**：file_open 和 connect 事件没有 comm 字段，需要额外查询
6. **RingBuffer 大小固定**：256KB 可能在高负载下丢失事件
7. **事件结构冗余**：packed 结构未对齐，CPU 读取效率低

### 3.2 BPF 数据采集重构方案（性能优先）

**2.1 统一事件结构（`bpf/main.bpf.c`）**

```c
#define TASK_COMM_LEN 16
#define PATH_MAX_LEN  256
#define ARGV0_LEN     128

// ═══════════════════════════════════════════════════════════════════
// 统一事件头部（8 字节对齐，优化 CPU 缓存）
// ═══════════════════════════════════════════════════════════════════
struct event_header {
    u64 timestamp_ns;      // 8B: 内核单调时钟 (bpf_ktime_get_ns)
    u64 cgroup_id;         // 8B: Cgroup ID
    u32 pid;               // 4B: 进程 ID
    u32 tid;               // 4B: 线程 ID
    u32 uid;               // 4B: 用户 ID
    u32 gid;               // 4B: 组 ID
    u8  type;              // 1B: 事件类型
    u8  blocked;           // 1B: 是否被拦截
    u8  _pad[6];           // 6B: 对齐填充
    char comm[TASK_COMM_LEN]; // 16B: 进程名
};  // Total: 56 bytes, 8-byte aligned

// ═══════════════════════════════════════════════════════════════════
// 进程执行事件
// ═══════════════════════════════════════════════════════════════════
struct exec_event {
    struct event_header hdr;           // 56B
    u32 ppid;                          // 4B: 父进程 ID
    u8  _pad[4];                       // 4B: 对齐
    char pcomm[TASK_COMM_LEN];         // 16B: 父进程名
    char filename[PATH_MAX_LEN];       // 256B: 执行文件路径
    char argv0[ARGV0_LEN];             // 128B: 第一个命令行参数
};  // Total: 464 bytes

// ═══════════════════════════════════════════════════════════════════
// 文件访问事件
// ═══════════════════════════════════════════════════════════════════
struct file_event {
    struct event_header hdr;           // 56B
    u64 ino;                           // 8B: inode 号
    u64 dev;                           // 8B: 设备号
    u32 flags;                         // 4B: 打开标志
    u8  _pad[4];                       // 4B: 对齐
    char filename[PATH_MAX_LEN];       // 256B: 文件名（parent/name 格式）
};  // Total: 336 bytes

// ═══════════════════════════════════════════════════════════════════
// 网络连接事件
// ═══════════════════════════════════════════════════════════════════
struct connect_event {
    struct event_header hdr;           // 56B
    u32 addr_v4;                       // 4B: IPv4 地址
    u16 family;                        // 2B: 地址族
    u16 port;                          // 2B: 端口
    u8  addr_v6[16];                   // 16B: IPv6 地址
};  // Total: 80 bytes
```

**2.2 高性能数据采集**

```c
// ═══════════════════════════════════════════════════════════════════
// BPF Maps 优化配置
// ═══════════════════════════════════════════════════════════════════

// 主事件缓冲区：2MB，支持高吞吐
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 2 * 1024 * 1024);  // 2MB
} events SEC(".maps");

// Per-CPU 进程上下文缓存（避免重复查询）
struct process_ctx {
    u32 ppid;
    u32 uid;
    u32 gid;
    char comm[TASK_COMM_LEN];
    char pcomm[TASK_COMM_LEN];
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_PERCPU_HASH);
    __uint(max_entries, 16384);  // 每 CPU 16K 条目
    __type(key, u32);            // PID
    __type(value, struct process_ctx);
} process_cache SEC(".maps");

// Per-CPU scratch buffer（避免栈溢出）
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, u32);
    __type(value, struct exec_event);  // 使用最大的事件结构
} scratch SEC(".maps");

// PID 过滤表（白名单/黑名单）
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, u32);   // PID
    __type(value, u8);  // 1=skip, 0=trace
} pid_filter SEC(".maps");
```

**2.3 内联优化函数**

```c
// ═══════════════════════════════════════════════════════════════════
// 高性能辅助函数
// ═══════════════════════════════════════════════════════════════════

// 填充事件头部（所有事件共用）
static __always_inline void fill_event_header(
    struct event_header *hdr, 
    u8 type,
    struct task_struct *task
) {
    hdr->timestamp_ns = bpf_ktime_get_ns();
    hdr->type = type;
    hdr->blocked = 0;
    
    u64 pid_tgid = bpf_get_current_pid_tgid();
    hdr->pid = pid_tgid >> 32;
    hdr->tid = (u32)pid_tgid;
    
    u64 uid_gid = bpf_get_current_uid_gid();
    hdr->uid = (u32)uid_gid;
    hdr->gid = uid_gid >> 32;
    
    hdr->cgroup_id = bpf_get_current_cgroup_id();
    bpf_get_current_comm(&hdr->comm, sizeof(hdr->comm));
}

// 快速 PID 过滤（内核线程 + 白名单）
static __always_inline bool should_skip_pid(u32 pid) {
    if (pid == 0) return true;  // 跳过内核线程
    
    u8 *skip = bpf_map_lookup_elem(&pid_filter, &pid);
    return skip && *skip == 1;
}

// 从缓存获取进程上下文（减少重复读取）
static __always_inline struct process_ctx* get_process_ctx(u32 pid) {
    return bpf_map_lookup_elem(&process_cache, &pid);
}

// 更新进程缓存
static __always_inline void update_process_cache(
    u32 pid, 
    u32 ppid,
    u32 uid,
    u32 gid,
    const char *comm,
    const char *pcomm
) {
    struct process_ctx ctx = {};
    ctx.ppid = ppid;
    ctx.uid = uid;
    ctx.gid = gid;
    __builtin_memcpy(ctx.comm, comm, TASK_COMM_LEN);
    if (pcomm) {
        __builtin_memcpy(ctx.pcomm, pcomm, TASK_COMM_LEN);
    }
    bpf_map_update_elem(&process_cache, &pid, &ctx, BPF_ANY);
}
```

**2.4 优化后的 LSM Hook 实现**

```c
SEC("lsm/bprm_check_security")
int BPF_PROG(lsm_bprm_check, struct linux_binprm *bprm)
{
    u32 pid = bpf_get_current_pid_tgid() >> 32;
    if (should_skip_pid(pid)) return 0;
    
    // 使用 per-CPU scratch buffer 避免栈溢出
    u32 key = 0;
    struct exec_event *event = bpf_map_lookup_elem(&scratch, &key);
    if (!event) return 0;
    
    struct task_struct *task = (struct task_struct *)bpf_get_current_task_btf();
    fill_event_header(&event->hdr, EVENT_TYPE_EXEC, task);
    
    // 获取父进程信息
    struct task_struct *parent = BPF_CORE_READ(task, real_parent);
    event->ppid = BPF_CORE_READ(parent, tgid);
    BPF_CORE_READ_STR_INTO(&event->pcomm, parent, comm);
    
    // 获取执行文件路径
    struct file *file = BPF_CORE_READ(bprm, file);
    if (file) {
        struct dentry *dentry = BPF_CORE_READ(file, f_path.dentry);
        // ... 路径提取逻辑（复用 check_file_action）
    }
    
    // 获取 argv[0]（如果可用）
    unsigned long argv = BPF_CORE_READ(bprm, p);
    if (argv) {
        bpf_probe_read_user_str(event->argv0, ARGV0_LEN, (void *)argv);
    }
    
    // 检查规则匹配
    u8 action = check_file_action(/* ... */);
    if (action == ACTION_BLOCK) {
        event->hdr.blocked = 1;
    }
    
    // 零拷贝提交到 ringbuf
    struct exec_event *rb_event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (rb_event) {
        __builtin_memcpy(rb_event, event, sizeof(*event));
        bpf_ringbuf_submit(rb_event, 0);  // 无需 FORCE_WAKEUP，用户态轮询
    }
    
    // 更新进程缓存
    update_process_cache(pid, event->ppid, event->hdr.uid, event->hdr.gid,
                        event->hdr.comm, event->pcomm);
    
    return event->hdr.blocked ? -EPERM : 0;
}
```

**2.5 用户态高性能处理**

```go
// pkg/events/decoder.go - 统一解码器

// 事件头部大小
const EventHeaderSize = 56

// 解码事件头部（所有事件共用）
func DecodeHeader(data []byte) (EventHeader, error) {
    if len(data) < EventHeaderSize {
        return EventHeader{}, ErrTooSmall
    }
    
    return EventHeader{
        TimestampNs: binary.LittleEndian.Uint64(data[0:8]),
        CgroupID:    binary.LittleEndian.Uint64(data[8:16]),
        PID:         binary.LittleEndian.Uint32(data[16:20]),
        TID:         binary.LittleEndian.Uint32(data[20:24]),
        UID:         binary.LittleEndian.Uint32(data[24:28]),
        GID:         binary.LittleEndian.Uint32(data[28:32]),
        Type:        EventType(data[32]),
        Blocked:     data[33] == 1,
        Comm:        extractCString(data[40:56]),
    }, nil
}

// 批量处理优化
func (r *RingBufReader) ReadBatch(maxEvents int) ([]Event, error) {
    events := make([]Event, 0, maxEvents)
    
    for i := 0; i < maxEvents; i++ {
        record, err := r.reader.Read()
        if err != nil {
            if errors.Is(err, ringbuf.ErrClosed) {
                break
            }
            continue
        }
        
        event, err := DecodeEvent(record.RawSample)
        if err == nil {
            events = append(events, event)
        }
    }
    
    return events, nil
}
```

**2.6 性能优化汇总**

| 优化项 | 优化前 | 优化后 | 收益 |
|-------|-------|-------|------|
| RingBuffer | 256KB | 2MB | 8x 缓冲容量 |
| 事件结构 | packed 未对齐 | 8字节对齐 | CPU 缓存友好 |
| 进程查询 | 每次重新读取 | Per-CPU LRU 缓存 | 减少内核读取 |
| 栈使用 | 大结构体在栈上 | Per-CPU scratch | 避免栈溢出 |
| PID 过滤 | 用户态过滤 | BPF 层预过滤 | 减少事件传输 |
| 批量处理 | 单事件处理 | 批量读取 | 减少系统调用 |

**2.7 配置化设计**

```go
// pkg/config/bpf.go

type BPFOptions struct {
    RingBufferSize    int    `yaml:"ring_buffer_size"`    // 默认 2MB
    ProcessCacheSize  int    `yaml:"process_cache_size"`  // 默认 16384
    EnableArgv        bool   `yaml:"enable_argv"`         // 是否采集 argv
    BatchSize         int    `yaml:"batch_size"`          // 批量处理大小
    SkipKernelThreads bool   `yaml:"skip_kernel_threads"` // 跳过内核线程
}

var DefaultBPFOptions = BPFOptions{
    RingBufferSize:    2 * 1024 * 1024,
    ProcessCacheSize:  16384,
    EnableArgv:        true,
    BatchSize:         100,
    SkipKernelThreads: true,
}
```

### 3.3 Prompt 设计完善

**3.1 当前 Prompt 问题**

现有 Prompt 设计仅支持诊断和聊天，缺少：
1. **意图解析 Prompt**：识别用户自然语言意图
2. **规则生成 Prompt**：从描述生成 YAML 规则
3. **事件解释 Prompt**：解释特定事件的安全含义
4. **上下文分析 Prompt**：分析进程/工作负载行为

**3.2 新增 Prompt 模板体系**

**目录结构**：
```
pkg/ai/
├── prompt/
│   ├── templates.go       # 模板定义和管理
│   ├── intent.go          # 意图解析 Prompt
│   ├── rulegen.go         # 规则生成 Prompt
│   ├── explain.go         # 事件解释 Prompt
│   ├── analyze.go         # 上下文分析 Prompt
│   └── sentinel.go        # Sentinel 巡检 Prompt
├── prompt.go              # 现有 Prompt（保留，用于诊断/聊天）
└── ...
```

**3.3 意图解析 Prompt（`pkg/ai/prompt/intent.go`）**

```go
const IntentSystemPrompt = `You are EulerGuard's intent parser. Parse user's natural language input and extract structured intent.

Available intent types:
- create_rule: User wants to create a security rule
- query_events: User wants to search/filter events  
- explain_event: User wants to understand why something happened
- analyze_process: User wants to analyze a process/workload
- promote_rule: User wants to promote a shadow rule
- navigation: User wants to navigate to a page

Output JSON only, no explanation:
{
  "type": "<intent_type>",
  "confidence": <0.0-1.0>,
  "params": { ... },
  "ambiguous": <true|false>,
  "clarification": "<if ambiguous, what to ask>"
}`

const IntentUserTemplate = `Current context:
- Page: {{.CurrentPage}}
- Selected: {{.SelectedItem}}
- Recent actions: {{.RecentActions}}

User input: "{{.Input}}"

Parse the intent:`
```

**3.4 规则生成 Prompt（`pkg/ai/prompt/rulegen.go`）**

```go
const RuleGenSystemPrompt = `You are EulerGuard's rule generator. Generate YAML security rules from natural language descriptions.

Rule schema:
- name: kebab-case unique identifier
- description: Human-readable description
- match: Conditions (process, filename, dest_port, cgroup, uid, etc.)
- action: "block" or "monitor"
- severity: "critical", "high", "warning", "info"

Guidelines:
1. Be specific: avoid overly broad rules that cause false positives
2. Use shadow mode for new rules (mode: shadow)
3. Include relevant context in description
4. Consider common legitimate use cases

Output YAML only, wrapped in yaml code block.`

const RuleGenUserTemplate = `Context:
- Existing rules: {{.ExistingRuleNames}}
- Recent blocked events: {{.RecentBlocked}}
- Target workload: {{.TargetWorkload}}

User request: "{{.Description}}"

Generate rule:`
```

**3.5 事件解释 Prompt（`pkg/ai/prompt/explain.go`）**

```go
const ExplainSystemPrompt = `You are EulerGuard's security analyst. Explain security events in clear, actionable terms.

When explaining an event:
1. What happened (technical details)
2. Why it was flagged/blocked (rule that matched)
3. Is it likely malicious or benign? (with reasoning)
4. What action should be taken?

Be concise but thorough. Use markdown formatting.`

const ExplainUserTemplate = `Event details:
- Type: {{.EventType}}
- Process: {{.ProcessName}} (PID: {{.PID}})
- Parent: {{.ParentName}}
- Target: {{.Target}}
- Action taken: {{.Action}}
- Rule matched: {{.RuleName}}

Process history (last 5 events):
{{range .ProcessHistory}}
- {{.Timestamp}}: {{.Description}}
{{end}}

Related processes (same cgroup):
{{range .RelatedProcesses}}
- {{.Comm}} ({{.EventCount}} events)
{{end}}

User question: "{{.Question}}"

Explain:`
```

**3.6 Sentinel 巡检 Prompt（`pkg/ai/prompt/sentinel.go`）**

```go
const SentinelShadowPrompt = `Analyze this shadow rule's performance and recommend whether to promote it.

Rule: {{.RuleName}}
Observation period: {{.ObservationHours}} hours
Total hits: {{.TotalHits}}
Hit breakdown:
{{range .HitsByProcess}}
- {{.ProcessName}}: {{.Count}} hits
{{end}}

Sample matched events:
{{range .SampleEvents}}
- {{.Timestamp}}: {{.ProcessName}} → {{.Target}}
{{end}}

Criteria for promotion:
1. Sufficient observation time (>24h recommended)
2. Consistent hit pattern (not just noise)
3. No obvious false positives (legitimate services)
4. Hits indicate real security value

Output JSON:
{
  "recommend": "promote" | "keep_shadow" | "delete",
  "confidence": <0.0-1.0>,
  "reasoning": "<explanation>",
  "concerns": ["<any concerns>"]
}`

const SentinelAnomalyPrompt = `Analyze this process for anomalous behavior.

Process: {{.ProcessName}} (PID: {{.PID}})
Workload: {{.CgroupPath}}
Running since: {{.StartTime}}

Baseline (normal behavior):
- Avg file opens/min: {{.BaselineFileRate}}
- Avg network conns/min: {{.BaselineNetRate}}
- Common file patterns: {{.BaselineFiles}}

Current (last 5 minutes):
- File opens/min: {{.CurrentFileRate}}
- Network conns/min: {{.CurrentNetRate}}
- Unusual files accessed: {{.UnusualFiles}}
- Unusual connections: {{.UnusualConnections}}

Determine if this is:
1. Normal operational variation
2. Legitimate but unusual (e.g., config reload)
3. Potentially malicious behavior

Output JSON:
{
  "assessment": "normal" | "unusual_benign" | "suspicious" | "malicious",
  "confidence": <0.0-1.0>,
  "reasoning": "<explanation>",
  "recommended_action": "<what to do>"
}`
```

**3.7 Prompt 效率优化**

```go
// pkg/ai/prompt/templates.go

// 1. Token 预算管理
type TokenBudget struct {
    SystemPrompt  int  // 系统提示词预算
    Context       int  // 上下文预算
    UserInput     int  // 用户输入预算
    Response      int  // 响应预算
}

var DefaultBudgets = map[string]TokenBudget{
    "intent":   {500, 200, 100, 200},    // 快速响应，小上下文
    "rulegen":  {800, 500, 200, 500},    // 中等复杂度
    "explain":  {600, 800, 100, 600},    // 需要较多上下文
    "sentinel": {400, 1000, 0, 400},     // 数据密集型
}

// 2. 上下文压缩
func CompressContext(ctx *PromptContext, budget int) string {
    // 优先保留：最近告警 > 拦截事件 > 普通事件
    // 按重要性裁剪，确保不超过 token 预算
}

// 3. 缓存常用 Prompt 片段
var promptCache = sync.Map{} // 缓存编译后的模板

// 4. 批量请求合并（用于 Sentinel）
func BatchAnalyze(items []AnalysisItem) []AnalysisResult {
    // 将多个小分析合并为一个请求，减少 API 调用
}
```

### 3.4 测试说明

**BPF 测试**：
- `bpf/main_test.go`：使用 `cilium/ebpf/cmd/bpf2go` 测试 BPF 程序加载
- `pkg/events/decoder_test.go`：测试新事件格式的解码
- 性能基准：测量事件吞吐量（目标：> 100k events/sec）

**Prompt 测试**：
- `pkg/ai/prompt/intent_test.go`：测试意图解析准确率
- `pkg/ai/prompt/rulegen_test.go`：测试规则生成的格式正确性
- `pkg/ai/prompt/benchmark_test.go`：测试 Token 使用效率

---

---

## 四、Phase 1: 全息遥测仓库

> **目标**: 打造一个高性能的内存时序数据底座，支持毫秒级的复杂查询与回溯。这是 AI "拥有记忆"的前提。

### 4.1 实现 `TimeRingBuffer` (高性能时序环形缓冲)
- **对接第零阶段**：`pkg/storage/store.go` 已定义接口，现在实现具体存储
- **实现 (`pkg/storage/ringbuffer.go`)**：
  - **数据结构**：
    - 基于定长数组 + 原子游标的环形缓冲区
    - 容量：**10,000+** 条（配置项 `Storage.RingBufferCapacity`，默认 10000）
    - 存储统一的事件类型 `storage.Event`（包含 `Type`, `Timestamp`, `Data` 等）
  - **零拷贝优化**：
    - 使用指针存储事件对象（`*storage.Event`）
    - 事件对象在堆上分配，避免栈拷贝
  - **并发安全**：
    - 写入使用原子操作更新游标
    - 读取使用 `RWMutex` 保护范围查询
  - **接口实现**：
    - 实现 `storage/store.go` 中定义的 `EventStore` 接口
    - `Append(event *Event)`：追加事件
    - `Query(start, end time.Time) []*Event`：时间范围查询
    - `Latest(n int) []*Event`：获取最近 N 条
- **集成点**：
  - 在 `tracer.DispatchEvent()` 中，事件解码后立即调用 `storage.Append()`
  - 删除所有旧的事件存储逻辑（第零阶段已清理）
- **价值**：为 AI 分析提供足够的历史上下文，支持时间窗口查询

### 4.2 构建"倒排索引" (Inverted Indexing)
- **实现 (`pkg/storage/indexer.go`)**：
  - **索引结构**：
    - `pidIndex map[uint32][]*Event`：PID 到事件列表
    - `cgroupIndex map[uint64][]*Event`：CgroupID 到事件列表
    - `typeIndex map[EventType][]*Event`：事件类型到事件列表
    - `processIndex map[string][]*Event`：进程名到事件列表（用于 "redis" 等查询）
  - **实时维护**：
    - 在 `TimeRingBuffer.Append()` 时，同步更新所有索引
    - 使用 `sync.Map` 或 `RWMutex` 保护并发访问
    - 索引只存储指针，不复制事件数据
  - **查询接口**：
    - `QueryByPID(pid uint32) []*Event`
    - `QueryByCgroup(cgroupID uint64) []*Event`
    - `QueryByType(eventType EventType) []*Event`
    - `QueryByProcess(processName string) []*Event`
    - `QueryByFilter(filter Filter) []*Event`：组合查询（用于语义查询）
  - **索引清理**：
    - 当 `TimeRingBuffer` 覆盖旧数据时，同步清理索引中的过期指针
    - 使用时间戳判断是否过期
- **集成点**：
  - `Indexer` 与 `TimeRingBuffer` 紧密耦合，在 `storage/store.go` 中统一管理
  - 第三阶段的语义查询直接调用 `Indexer` 的查询接口
- **价值**：查询复杂度从 **O(N)** 降为 **O(1)**，支持 AI 快速检索

### 4.3 进程画像快照 (Live Process Profile)
- **实现 (`pkg/proc/profile.go`)**：
  - **数据结构**：
    ```go
    type ProcessProfile struct {
        PID      uint32
        Static   StaticProfile    // 静态信息
        Dynamic  DynamicProfile   // 动态统计
        Baseline *BaselineProfile // 基线（可选）
    }
    type StaticProfile struct {
        StartTime    time.Time
        CommandLine  string
        Genealogy    []uint32  // 父进程链
    }
    type DynamicProfile struct {
        FileOpenCount    int64
        NetConnectCount  int64
        LastFileOpen     time.Time
        LastConnect      time.Time
        // 过去 5 分钟的统计
    }
    ```
  - **维护机制**：
    - 在 `ProcessTree.AddProcess()` 时创建 `ProcessProfile`
    - 在事件存储时，同步更新对应 PID 的 `DynamicProfile`
    - 使用 `sync.Map` 存储 `map[uint32]*ProcessProfile`
  - **查询接口**：
    - `GetProfile(pid uint32) (*ProcessProfile, bool)`
    - `GetAnomalousProcesses() []*ProcessProfile`：检测行为突变
    - `GetProcessTree(pid uint32) (*ProcessNode, error)`：获取进程树
  - **集成点**：
    - 与 `ProcessTree` 集成，在进程创建时初始化
    - 与 `storage` 集成，在事件存储时更新统计
    - 第三阶段的 Sentinel 调用 `GetAnomalousProcesses()` 检测异常
- **价值**：为 AI 提供进程行为上下文，支持异常检测

### 4.4 API 暴露（第一阶段）
**事件查询 API**：
```go
// pkg/server/api.go

// 事件列表（分页）
GET /api/events?page=1&limit=50&type=exec&process=nginx
Response: { "events": [...], "total": 1000, "page": 1 }

// 事件详情
GET /api/events/{id}
Response: { "event": {...} }

// 时间范围查询
GET /api/events/range?start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z
Response: { "events": [...] }
```

**进程画像 API**：
```go
// 获取进程画像
GET /api/process/{pid}/profile
Response: {
  "pid": 12345,
  "static": { "startTime": "...", "commandLine": "...", "genealogy": [...] },
  "dynamic": { "fileOpenCount": 100, "netConnectCount": 50, ... },
  "anomalyScore": 0.3
}

// 获取进程树
GET /api/process/{pid}/tree
Response: { "ancestors": [...], "children": [...] }

// 获取进程相关事件
GET /api/process/{pid}/events?limit=100
Response: { "events": [...] }
```

**统计 API**：
```go
// 总体统计
GET /api/stats
Response: { "execCount": 1000, "fileCount": 500, "connectCount": 300, "alertCount": 10 }

// 速率统计
GET /api/stats/rates
Response: { "execRate": 10.5, "fileRate": 5.2, "connectRate": 3.1 }
```

### 4.5 测试说明

- **单元测试**：
  - `pkg/storage/ringbuffer_test.go`：
    - 测试 `Append()` 的并发安全性（使用 `go test -race`）
    - 测试 `Query()` 的时间范围查询准确性
    - 测试环形缓冲的覆盖行为（容量满时覆盖旧数据）
    - 测试 `Latest()` 返回最近 N 条的正确性
  - `pkg/storage/indexer_test.go`：
    - 测试索引的实时维护（添加事件后索引立即更新）
    - 测试各种查询接口（`QueryByPID`, `QueryByCgroup` 等）
    - 测试组合查询 `QueryByFilter` 的准确性
    - 测试索引清理（过期数据从索引中移除）
  - `pkg/proc/profile_test.go`：
    - 测试 `ProcessProfile` 的创建和更新
    - 测试 `GetAnomalousProcesses()` 的异常检测逻辑
    - 测试 `DynamicProfile` 的统计准确性
- **集成测试**：
  - `pkg/storage/integration_test.go`：
    - 测试 `TimeRingBuffer` 与 `Indexer` 的协同工作
    - 测试高并发场景下的性能和正确性
    - 测试事件存储与查询的端到端流程
  - `pkg/tracer/dispatch_integration_test.go`：
    - 测试事件解码 → 存储 → 分发的完整流程
    - 验证事件确实被存储到 `TimeRingBuffer`
- **性能测试**：
  - `pkg/storage/ringbuffer_bench_test.go`：
    - 基准测试：`Append()` 的吞吐量（目标：> 100k events/sec）
    - 基准测试：`Query()` 的延迟（目标：< 1ms for 10k events）
  - `pkg/storage/indexer_bench_test.go`：
    - 基准测试：索引查询性能（目标：O(1) 复杂度）
- **测试覆盖率目标**：
  - `storage` 包覆盖率 > 85%
  - `proc/profile.go` 覆盖率 > 80%

---

## 五、Phase 2: 双模执行引擎

> **目标**: 让规则引擎支持"测试服"逻辑，这是系统工具安全落地的核心机制。
### 5.1 影子模式 (Shadow Mode)
- **规则属性扩展 (`pkg/types/rules.go`)**：
  - `Rule` 结构体增加 `Mode` 字段：
    ```go
    type RuleMode string
    const (
        ModeEnforce RuleMode = "enforce"  // 强制拦截
        ModeShadow  RuleMode = "shadow"   // 影子观察
    )
    type Rule struct {
        // ... 现有字段
        Mode RuleMode `yaml:"mode,omitempty"` // 默认为 "enforce"
    }
    ```
- **规则引擎扩展 (`pkg/rules/engine.go`)**：
  - **执行逻辑修改**：
    - `MatchExec()`, `MatchFile()`, `MatchConnect()` 返回结果增加 `Mode` 信息
    - 当 `Mode == Shadow` 且规则命中时：
      - **不返回 `-EPERM`**（在 eBPF 层面放行）
      - 生成 `ShadowHit` 事件，推送到 `ShadowBuffer`
  - **ShadowBuffer 实现 (`pkg/rules/shadow.go`)**：
    - 使用 `TimeRingBuffer` 的简化版本存储影子命中事件
    - 记录：规则名、命中时间、事件详情、是否误报（由 AI 判断）
    - 查询接口：`GetHits(ruleName string, timeWindow time.Duration) []ShadowHit`
- **集成点**：
  - 在 `server.Bridge` 和 `cli.alertHandler` 中，根据 `Mode` 决定是否真正拦截
  - Shadow 模式的命中事件不触发告警，只记录到 `ShadowBuffer`
- **价值**：AI 生成的规则默认进入 Shadow 模式，通过实战数据验证准确性

### 5.2 虚拟回放引擎 (Simulation Engine)
- **实现 (`pkg/simulation/runner.go`)**：
  - **接口设计**：
    ```go
    type SimulationRequest struct {
        Rules      []types.Rule
        RuleNames  []string    // 使用现有规则（与 Rules 二选一）
        TimeWindow TimeWindow  // 时间窗口
    }
    type SimulationReport struct {
        TotalEvents       int
        WouldBlock        int
        ShadowHits        int
        AffectedProcesses []AffectedProcess
        AIAnalysis        string  // AI 分析摘要
    }
    type AffectedProcess struct {
        PID   uint32
        Name  string
        Hits  int
    }
    func RunSimulation(req SimulationRequest, store storage.EventStore) (*SimulationReport, error)
    ```
  - **执行逻辑**：
    1. 从 `TimeRingBuffer` 中拉取 `TimeWindow` 范围内的所有事件
    2. 对每个事件，使用临时规则引擎进行匹配
    3. 统计命中次数、拦截次数、影响的进程
    4. 返回 `SimulationReport`
  - **优化**：
    - 使用 `Indexer` 快速过滤相关事件（如只查询特定 PID 的事件）
    - 并行处理多个事件（如果数据量大）
- **集成点**：
  - 第三阶段的 AI 规则生成会自动调用模拟引擎
  - 第四阶段的前端 Policy Studio 调用此接口显示模拟结果
- **价值**：在部署规则前预览效果，降低误拦截风险

### 5.3 API 暴露（第二阶段）

**规则管理 API**：
```go
// pkg/server/api.go

// 获取规则列表（按模式分组）
GET /api/rules
Response: { 
  "enforce": [...],  // 强制拦截的规则
  "shadow": [...],   // 影子模式的规则
  "draft": [...]     // 草稿规则
}

// 创建规则
POST /api/rules
Body: { "rule": {...}, "mode": "shadow" | "enforce" | "draft" }
Response: { "rule": {...}, "created": true }

// 更新规则
PUT /api/rules/{name}
Body: { "rule": {...} }

// 删除规则
DELETE /api/rules/{name}

// Shadow 规则转正
POST /api/rules/{name}/promote
Response: { "success": true, "rule": {...}, "previousMode": "shadow" }

// 获取 Shadow 命中统计
GET /api/rules/{name}/shadow-stats
Response: { 
  "hits": 156, 
  "observationHours": 72,
  "hitsByProcess": [{ "name": "nginx", "count": 100 }],
  "recentHits": [...]
}
```

**模拟引擎 API**：
```go
// 运行模拟
POST /api/simulation/run
Body: {
  "rules": [...],        // 规则列表（可选，使用临时规则）
  "ruleNames": [...],    // 现有规则名（可选）
  "timeWindow": { "start": "...", "end": "..." }
}
Response: {
  "totalEvents": 1000,
  "wouldBlock": 50,
  "shadowHits": 100,
  "affectedProcesses": [{ "pid": 123, "name": "nginx", "hits": 10 }]
}

// 规则对比模拟
POST /api/simulation/compare
Body: { "ruleA": {...}, "ruleB": {...}, "timeWindow": {...} }
Response: { 
  "comparison": { 
    "onlyA": 10, 
    "onlyB": 5, 
    "both": 100,
    "differences": [...]
  } 
}
```

### 5.4 测试说明

- **单元测试**：
  - `pkg/types/rules_test.go`：
    - 测试 `Rule.Mode` 字段的序列化/反序列化（YAML）
    - 测试默认值（未指定时默认为 `ModeEnforce`）
  - `pkg/rules/engine_test.go`：
    - 测试 Shadow 模式的匹配逻辑（命中但不拦截）
    - 测试 Enforce 模式的正常拦截逻辑
    - 测试规则引擎对两种模式的处理差异
  - `pkg/rules/shadow_test.go`：
    - 测试 `ShadowBuffer` 的存储和查询
    - 测试 `GetHits()` 的时间窗口过滤
    - 测试 Shadow 命中事件的记录格式
  - `pkg/simulation/runner_test.go`：
    - 测试 `RunSimulation()` 的基本功能
    - 测试模拟报告的准确性（统计数字正确）
    - 测试时间窗口过滤的正确性
    - 测试空规则列表的处理
- **集成测试**：
  - `pkg/rules/shadow_integration_test.go`：
    - 测试 Shadow 规则从匹配到记录的完整流程
    - 测试 Shadow 模式不影响实际系统行为（不拦截）
  - `pkg/simulation/integration_test.go`：
    - 测试模拟引擎与 `TimeRingBuffer` 的集成
    - 测试使用真实历史数据运行模拟
    - 验证模拟结果与实际规则执行的一致性
- **测试覆盖率目标**：
  - `rules` 包覆盖率 > 85%
  - `simulation` 包覆盖率 > 80%

---

## 六、Phase 3: AI 接口层

> **目标**: 封装 AI 能力，构建完整的 AI 原生后端 API 体系，使 AI 成为系统的"神经中枢"。

### 6.1 AI 核心服务 (`pkg/ai/`)

**1.1 意图解析服务 (`pkg/ai/intent.go`)** - **新增，支持 Omnibox**
- **实现**：
  ```go
  type Intent struct {
      Type       IntentType `json:"type"`       // create_rule, query_events, explain, analyze, etc.
      Confidence float64    `json:"confidence"` // 0-1 置信度
      Params     any        `json:"params"`     // 意图参数（结构化）
      Preview    *Preview   `json:"preview"`    // 预览（如生成的规则 YAML）
      Warnings   []string   `json:"warnings"`   // AI 警告
  }
  
  type IntentType string
  const (
      IntentCreateRule   IntentType = "create_rule"   // 创建规则
      IntentQueryEvents  IntentType = "query_events"  // 查询事件
      IntentExplainEvent IntentType = "explain_event" // 解释事件
      IntentAnalyzeProc  IntentType = "analyze_process" // 分析进程
      IntentPromoteRule  IntentType = "promote_rule"  // 转正规则
      IntentNavigation   IntentType = "navigation"    // 导航到页面
  )
  
  func ParseIntent(ctx context.Context, input string, context *RequestContext) (*Intent, error)
  ```
- **Prompt 设计**：
  ```
  You are EulerGuard's AI assistant for kernel security.
  Parse the user's intent from their natural language input.
  
  User Input: "{{input}}"
  Current Context: {{context}}
  
  Classify the intent and extract structured parameters.
  Output JSON format: { "type": "...", "confidence": 0.95, "params": {...} }
  ```
- **上下文感知**：
  - `RequestContext` 包含当前页面、选中的事件/规则、最近操作等
  - AI 根据上下文推断用户意图

**1.2 规则生成服务 (`pkg/ai/rulegen.go`)** - **扩展**
- **实现**：
  ```go
  type RuleGenRequest struct {
      Description string          `json:"description"` // 自然语言描述
      Context     *RequestContext `json:"context"`     // 上下文
      Examples    []types.Rule    `json:"examples"`    // 现有规则作为参考
  }
  
  type RuleGenResponse struct {
      Rule       types.Rule `json:"rule"`       // 生成的规则
      YAML       string     `json:"yaml"`       // YAML 格式
      Reasoning  string     `json:"reasoning"`  // AI 推理过程
      Confidence float64    `json:"confidence"` // 置信度
      Warnings   []string   `json:"warnings"`   // 潜在风险警告
      Simulation *SimulationReport `json:"simulation"` // 预模拟结果
  }
  
  func GenerateRule(ctx context.Context, req *RuleGenRequest) (*RuleGenResponse, error)
  ```
- **集成模拟引擎**：
  - 生成规则后，自动调用 `simulation.RunSimulation()` 进行预评估
  - 将模拟结果附加到响应中

**1.3 事件解释服务 (`pkg/ai/explain.go`)** - **新增**
- **实现**：
  ```go
  type ExplainRequest struct {
      EventID   string `json:"event_id"`   // 事件 ID
      EventData any    `json:"event_data"` // 事件详情（备选）
      Question  string `json:"question"`   // 用户问题（可选）
  }
  
  type ExplainResponse struct {
      Explanation   string           `json:"explanation"`    // 自然语言解释
      RootCause     string           `json:"root_cause"`     // 根本原因
      MatchedRule   *types.Rule      `json:"matched_rule"`   // 触发的规则
      RelatedEvents []storage.Event  `json:"related_events"` // 相关事件
      SuggestedActions []Action      `json:"suggested_actions"` // 建议操作
  }
  
  func ExplainEvent(ctx context.Context, req *ExplainRequest) (*ExplainResponse, error)
  ```
- **关联分析**：
  - 使用 `storage.Indexer` 查询同 PID 的相关事件
  - 使用 `proc.ProcessProfile` 获取进程上下文
  - AI 综合分析后生成解释

**1.4 上下文分析服务 (`pkg/ai/analyze.go`)** - **新增**
- **实现**：
  ```go
  type AnalyzeRequest struct {
      Type string `json:"type"` // "process", "workload", "rule"
      ID   string `json:"id"`   // PID, CgroupID, RuleName
  }
  
  type AnalyzeResponse struct {
      Summary        string           `json:"summary"`         // 摘要
      Anomalies      []Anomaly        `json:"anomalies"`       // 异常点
      BaselineStatus string           `json:"baseline_status"` // 基线状态
      Recommendations []Recommendation `json:"recommendations"` // 建议
      RelatedInsights []Insight       `json:"related_insights"` // 相关洞察
  }
  
  func Analyze(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error)
  ```
- **分析类型**：
  - `process`：进程画像分析（基于 `proc.ProcessProfile`）
  - `workload`：工作负载分析（基于 `workload.Registry`）
  - `rule`：规则效果分析（基于 `rules.ShadowBuffer`）

### 6.2 语义查询接口 (Semantic Query Layer)

**实现 (`pkg/api/query.go`)**：
- **自然语言解析**：
  - 使用 `ai.ParseIntent()` 识别查询意图
  - 转换为结构化的 `QueryFilter`
  ```go
  type QueryFilter struct {
      Types       []string    // "exec", "file", "connect"
      Processes   []string    // 进程名列表
      Actions     []string    // "block", "shadow", "allow"
      PIDs        []uint32    // PID 列表
      CgroupIDs   []uint64    // CgroupID 列表
      TimeWindow  TimeWindow  // 时间窗口
      Correlation bool        // 是否关联同 PID 事件
  }
  ```
- **查询执行**：
  - 调用 `storage.Indexer.QueryByFilter(filter)` 获取事件
  - 支持组合查询（多条件 AND/OR）
  - 支持相关性排序（按 AI 评估的重要性）

**API 端点**：
- `POST /api/query`：结构化查询
- `POST /api/query/semantic`：自然语言查询（调用 AI）
- `GET /api/events`：事件列表（分页）
- `GET /api/events/{id}`：事件详情

### 6.3 后台主动巡检 (Background Sentinel)

**实现 (`pkg/ai/sentinel.go`)**：
- **启动机制**：
  - 在 `core.Bootstrap()` 中初始化 Sentinel
  - 启动独立的 goroutine，使用 `Ticker`（默认每分钟）执行巡检
  
- **巡检任务**：

  | 任务 | 触发频率 | 数据源 | 输出 |
  |------|---------|-------|------|
  | Shadow 规则转正建议 | 每 5 分钟 | `rules.ShadowBuffer` | `ShadowPromotionInsight` |
  | 进程行为异常检测 | 每 1 分钟 | `proc.ProcessProfile` | `AnomalyInsight` |
  | 规则优化建议 | 每 30 分钟 | `rules.Engine` | `OptimizationInsight` |
  | 每日安全摘要 | 每天 | `storage.Stats` | `DailyReportInsight` |

- **Insight 类型定义**：
  ```go
  type Insight struct {
      ID          string       `json:"id"`
      Type        InsightType  `json:"type"`
      Title       string       `json:"title"`
      Summary     string       `json:"summary"`
      Confidence  float64      `json:"confidence"`
      Severity    Severity     `json:"severity"`
      Data        any          `json:"data"`        // 类型相关数据
      Actions     []Action     `json:"actions"`     // 可执行操作
      CreatedAt   time.Time    `json:"created_at"`
  }
  
  type Action struct {
      Label    string         `json:"label"`    // "转正", "调查", "忽略"
      ActionID string         `json:"action_id"` // "promote", "investigate", "dismiss"
      Params   map[string]any `json:"params"`   // 操作参数
  }
  ```

- **通知推送**：
  - WebSocket 端点：`WS /api/ai/sentinel/stream`
  - 支持重连和消息回溯

### 6.4 规则管理 API（扩展）

**Shadow 规则管理**：
```go
// pkg/server/api.go

// 获取规则列表
GET /api/rules
Response: { "rules": [...], "shadowRules": [...], "draftRules": [...] }

// 创建规则
POST /api/rules
Body: { "rule": {...}, "mode": "shadow" | "enforce" | "draft" }

// 更新规则
PUT /api/rules/{name}
Body: { "rule": {...} }

// 删除规则
DELETE /api/rules/{name}

// **新增：规则转正**
POST /api/rules/{name}/promote
Response: { "success": true, "rule": {...} }

// **新增：获取 Shadow 命中统计**
GET /api/rules/{name}/shadow-stats
Response: { "hits": 156, "falsePositives": 2, "observationHours": 72 }

// **新增：AI 规则审查**
POST /api/rules/{name}/ai-review
Response: { "issues": [...], "suggestions": [...], "score": 85 }
```

### 6.5 进程画像 API（新增）

```go
// pkg/server/api.go

// 获取进程画像
GET /api/process/{pid}/profile
Response: {
  "pid": 12345,
  "static": { "startTime": "...", "commandLine": "...", "genealogy": [...] },
  "dynamic": { "fileOpenCount": 100, "netConnectCount": 50, ... },
  "baseline": { "normalFileRate": 10, "normalNetRate": 5, ... },
  "anomalyScore": 0.3
}

// 获取进程树
GET /api/process/{pid}/tree
Response: { "ancestors": [...], "children": [...] }

// 获取进程相关事件
GET /api/process/{pid}/events?limit=100
Response: { "events": [...] }
```

### 6.6 模拟引擎 API（扩展）

```go
// pkg/server/api.go

// 运行模拟
POST /api/simulation/run
Body: {
  "rules": [...],        // 规则列表（可选，使用临时规则）
  "ruleNames": [...],    // 现有规则名（可选）
  "timeWindow": { "start": "...", "end": "..." }
}
Response: {
  "totalEvents": 1000,
  "wouldBlock": 50,
  "shadowHits": 100,
  "affectedProcesses": [{ "pid": 123, "name": "nginx", "hits": 10 }],
  "aiAnalysis": "该规则会影响 nginx 的正常日志写入..."
}

// **新增：规则对比模拟**
POST /api/simulation/compare
Body: { "ruleA": {...}, "ruleB": {...}, "timeWindow": {...} }
Response: { "comparison": { "onlyA": 10, "onlyB": 5, "both": 100 } }
```

### 6.7 AI 统一 API 端点汇总

| 端点 | 方法 | 描述 | 依赖 |
|------|------|------|------|
| `/api/ai/intent` | POST | 意图解析（Omnibox 核心） | ai.ParseIntent |
| `/api/ai/generate-rule` | POST | AI 规则生成 | ai.GenerateRule, simulation |
| `/api/ai/explain` | POST | 事件解释 | ai.ExplainEvent, storage |
| `/api/ai/analyze` | POST | 上下文分析 | ai.Analyze, proc, workload |
| `/api/ai/sentinel/stream` | WS | Sentinel 洞察流 | ai.Sentinel |
| `/api/ai/sentinel/insights` | GET | 历史洞察列表 | ai.Sentinel |
| `/api/ai/sentinel/action` | POST | 执行洞察建议的操作 | 各操作对应的服务 |

### 6.8 Prompt 模板管理 (`pkg/ai/prompt/`)

**目录结构**：
```
pkg/ai/prompt/
├── intent.go        // 意图解析 Prompt
├── rulegen.go       // 规则生成 Prompt
├── explain.go       // 事件解释 Prompt
├── analyze.go       // 上下文分析 Prompt
├── sentinel.go      // Sentinel 分析 Prompt
└── templates/       // Prompt 模板文件
    ├── intent.tmpl
    ├── rulegen.tmpl
    └── ...
```

**Prompt 版本管理**：
- 每个 Prompt 有版本号
- 支持 A/B 测试不同 Prompt 效果
- 记录 Prompt 调用日志用于优化

### 6.9 测试说明

**单元测试**：
- `pkg/ai/intent_test.go`：
  - 测试意图解析的准确性
  - 测试上下文对意图识别的影响
  - 测试低置信度情况的处理
- `pkg/ai/rulegen_test.go`：
  - 测试规则生成的正确性
  - 测试生成规则的 YAML 格式
  - 测试警告信息的生成
- `pkg/ai/explain_test.go`：
  - 测试事件解释的完整性
  - 测试关联事件的查询
- `pkg/ai/analyze_test.go`：
  - 测试进程分析的准确性
  - 测试异常检测的逻辑
- `pkg/ai/sentinel_test.go`：
  - 测试各巡检任务的执行
  - 测试 Insight 的生成格式
  - 测试 WebSocket 推送

**集成测试**：
- `pkg/api/ai_integration_test.go`：
  - 测试 AI API 的端到端流程
  - 测试意图解析 → 执行 → 结果的完整链路
- `pkg/ai/sentinel_integration_test.go`：
  - 测试 Sentinel 与所有数据源的集成
  - 测试长时间运行的稳定性

**API 测试**：
- `pkg/server/api_test.go`：
  - 测试所有 AI API 端点
  - 测试错误处理和边界情况
  - 测试并发请求的处理

**测试覆盖率目标**：
- `ai` 包覆盖率 > 80%
- `api` 包覆盖率 > 85%

---

## 七、Phase 4: AI 原生前端

> **目标**: 打造 AI 原生的安全工作台，让"与 AI 对话"成为主要交互方式，而非传统的点击配置。

### 7.1 AI 原生设计理念

**核心理念：Conversation-Driven Security（对话驱动安全）**

传统安全工具的交互模式：
```
用户 → 点击菜单 → 填写表单 → 提交配置 → 等待生效
```

EulerGuard AI 原生交互模式：
```
用户 → 自然语言描述意图 → AI 理解并生成方案 → 用户审核确认 → 自动部署
```

**AI 原生的 5 个设计原则**：

1. **Conversation First（对话优先）**
   - 任何操作都可以通过自然语言完成
   - Omnibox 是系统的"灵魂入口"，不是可选功能
   - 传统 GUI 作为"AI 建议的可视化"，而非主要交互方式

2. **AI as Co-pilot（AI 作为副驾驶）**
   - AI 不仅响应请求，还主动提供建议
   - 每个页面都有 AI 上下文感知，随时可以询问
   - AI 解释每个操作的影响和风险

3. **Trust but Verify（信任但验证）**
   - AI 生成的规则默认进入 Shadow 模式
   - 所有 AI 建议都有置信度评分
   - 提供模拟预览，让用户看到"如果部署会发生什么"

4. **Progressive Automation（渐进式自动化）**
   - 初期：AI 建议 + 人工审核
   - 中期：高置信度建议自动进入 Shadow
   - 后期：经过验证的 Shadow 规则自动转正（可配置）

5. **Explainable AI（可解释的 AI）**
   - 每条规则都有 AI 的推理过程
   - 每次拦截都有 AI 的风险评估
   - 每个异常都有 AI 的上下文分析

### 7.2 AI 原生页面架构

**重新定义页面结构**（以 AI 交互为中心）：

```
┌─────────────────────────────────────────────────────────────────┐
│  🧠 [AI Omnibox - "Ask anything..."]              [Cmd/Ctrl+K]  │
├─────────┬───────────────────────────────────────────────────────┤
│         │                                                       │
│ 导航栏   │  主内容区                                              │
│         │  ┌─────────────────────────────────────────────────┐  │
│ ┌─────┐ │  │                                                 │  │
│ │ 🎯  │ │  │   [当前页面内容]                                 │  │
│ │ ✨  │ │  │                                                 │  │
│ │ 🔍  │ │  │                                                 │  │
│ │ 🤖  │ │  └─────────────────────────────────────────────────┘  │
│ └─────┘ │                                                       │
│         │  ┌─────────────────────────────────────────────────┐  │
│         │  │ 💬 AI Context Bar - 当前上下文的 AI 助手         │  │
│         │  │ "我注意到 nginx 今天有 23 次异常连接..."         │  │
│         │  └─────────────────────────────────────────────────┘  │
├─────────┴───────────────────────────────────────────────────────┤
│  [Status: 🟢 AI Ready | 📊 1.2k events/s | 🛡️ 5 blocked today] │
└─────────────────────────────────────────────────────────────────┘
```

**关键 UI 元素**：
- **AI Omnibox（顶部）**：全局 AI 入口，支持自然语言一切操作
- **AI Context Bar（底部悬浮）**：当前页面上下文的 AI 洞察，可展开对话
- **AI Hints（页面内嵌）**：各组件内的 AI 提示和建议

### 7.3 页面详细设计

##### **3.1 AI Omnibox（全局 AI 入口）- 系统灵魂**
- **定位**：不只是搜索框，而是"与系统对话的窗口"
- **触发方式**：`Cmd/Ctrl + K` 或点击顶部 AI 图标
- **交互模式**：
  ```
  ┌──────────────────────────────────────────────────────────────┐
  │ 🧠 Ask EulerGuard anything...                                │
  │ ──────────────────────────────────────────────────────────── │
  │                                                              │
  │ 💡 Try:                                                      │
  │   • "Block all outbound connections from nginx to port 3306" │
  │   • "Why was this process blocked?"                          │
  │   • "Show me suspicious file access in the last hour"        │
  │   • "Create a whitelist for my Redis container"              │
  │                                                              │
  │ ──────────────────────────────────────────────────────────── │
  │ 📝 Recent:                                                   │
  │   "阻止 nginx 访问敏感文件"  →  Created rule (Shadow)        │
  │   "分析最近的告警"          →  Found 3 anomalies             │
  └──────────────────────────────────────────────────────────────┘
  ```
- **AI 意图识别**：
  | 用户输入 | AI 识别意图 | 执行动作 |
  |---------|-----------|---------|
  | "阻止 nginx 访问 /etc/passwd" | 创建规则 | 生成规则 YAML → 显示预览 → 部署确认 |
  | "为什么 java 被拦截了" | 解释事件 | 查询相关事件 → AI 分析原因 → 展示解释 |
  | "最近有什么可疑活动" | 安全分析 | 调用 Sentinel → 展示洞察列表 |
  | "redis 的行为正常吗" | 基线对比 | 获取进程画像 → AI 分析偏差 |
- **组件**：
  - `components/global/AIomnibox.vue`：AI 对话式搜索框
  - `components/global/IntentPreview.vue`：意图识别预览
  - `components/global/ActionConfirm.vue`：操作确认面板

##### **3.2 Observatory（观测站）- AI 驱动的态势感知**
- **AI 原生特性**：
  - **AI 健康评估**：不只是数字，而是 AI 的综合判断
    ```
    ┌─────────────────────────────────────────────────────┐
    │  🛡️ System Health: 87/100                          │
    │  ─────────────────────────────────────────────────  │
    │  AI Assessment:                                     │
    │  "系统整体健康，但 nginx 容器近 2 小时有异常         │
    │   文件访问模式，建议关注。"                          │
    │                                                     │
    │  [🔍 详细分析] [✨ 一键优化]                         │
    └─────────────────────────────────────────────────────┘
    ```
  - **AI 威胁摘要**：用自然语言描述威胁态势
  - **一键询问**：任何数据点都可以问 AI "这是什么意思？"
- **核心组件**：
  - `HealthScore.vue`：AI 健康评分 + 自然语言解释
  - `ThreatSummary.vue`：AI 威胁摘要（非传统图表）
  - `AIInsightCards.vue`：AI 主动推送的洞察卡片
  - `QuickAsk.vue`：快速询问 AI 的悬浮按钮

##### **3.3 Policy Studio（策略工坊）- AI 辅助规则创作**
- **AI 原生特性**：
  - **自然语言规则创建**：
    ```
    ┌─────────────────────────────────────────────────────────────┐
    │ ✨ Describe your security intent:                           │
    │ ┌─────────────────────────────────────────────────────────┐ │
    │ │ 阻止所有非 root 用户执行的进程访问 /etc/shadow          │ │
    │ └─────────────────────────────────────────────────────────┘ │
    │                                                             │
    │ 🧠 AI Understanding:                                        │
    │ • Target: File access to /etc/shadow                        │
    │ • Condition: Process not running as root (UID != 0)         │
    │ • Action: Block                                             │
    │ • Confidence: 95%                                           │
    │                                                             │
    │ 📄 Generated Rule:                                          │
    │ ┌─────────────────────────────────────────────────────────┐ │
    │ │ name: block-non-root-shadow-access                      │ │
    │ │ match:                                                  │ │
    │ │   filename: /etc/shadow                                 │ │
    │ │   uid_not: 0                                            │ │
    │ │ action: block                                           │ │
    │ └─────────────────────────────────────────────────────────┘ │
    │                                                             │
    │ [▶️ 模拟预览] [🌙 部署为 Shadow] [🛡️ 立即生效]              │
    └─────────────────────────────────────────────────────────────┘
    ```
  - **AI 规则审查**：对手动编写的规则提供 AI 审查
  - **模拟解读**：AI 解释模拟结果的含义
- **三栏布局（重新设计）**：
  ```
  ┌──────────────┬────────────────────────────┬─────────────────┐
  │  规则列表     │     AI 辅助编辑器          │   AI 分析面板   │
  │  ───────     │     ──────────             │   ──────────    │
  │  [Enforce]   │  ┌──────────────────────┐  │  📊 模拟结果    │
  │  • rule-1    │  │ 自然语言输入...       │  │  • 命中: 156   │
  │  • rule-2    │  └──────────────────────┘  │  • 拦截: 12     │
  │              │  ↓ AI 生成                 │                 │
  │  [Shadow]    │  ┌──────────────────────┐  │  🧠 AI 分析     │
  │  • rule-3    │  │ YAML Editor          │  │  "该规则可能    │
  │  • rule-4 ⭐ │  │ (Monaco)             │  │   影响 nginx    │
  │              │  └──────────────────────┘  │   的正常运行"   │
  │  [Draft]     │                            │                 │
  │  • rule-5    │  🤖 AI Suggestions:        │  [Ask AI...]    │
  │              │  "建议添加 UID 条件..."    │                 │
  └──────────────┴────────────────────────────┴─────────────────┘
  ```

##### **3.4 Investigation（调查台）- AI 辅助威胁狩猎**
- **AI 原生特性**：
  - **自然语言查询**：
    ```
    ┌─────────────────────────────────────────────────────────────┐
    │ 🔍 "Show me all network connections from containers         │
    │     that also accessed sensitive files in the last hour"    │
    └─────────────────────────────────────────────────────────────┘
    
    🧠 AI translated to:
    • Event type: connect + file
    • Filter: cgroup != host, filename contains "/etc/passwd|shadow"
    • Time: last 1 hour
    • Correlation: same PID
    ```
  - **AI 上下文解释**：选中任何事件，AI 解释其含义
  - **AI 推荐调查路径**：
    ```
    💡 AI suggests investigating:
    1. "This PID also made 47 DNS queries to unusual domains"
    2. "The parent process has been running for only 3 minutes"
    3. "Similar pattern seen from 2 other containers"
    ```
- **布局**：
  ```
  ┌─────────────────────────────────────────────────────────────────┐
  │ 🔍 [自然语言查询栏] "redis 为什么被拦截..."                      │
  ├─────────────────────────────────────────────────────────────────┤
  │  Timeline: ──●────●──●───●●───●────●──→                        │
  ├───────────────────────────────────────┬─────────────────────────┤
  │  Event List                           │  AI Context Panel      │
  │  ────────────                         │  ───────────────       │
  │  ▶ 10:32:15 nginx → /etc/passwd ⛔    │  🧠 为什么被拦截？      │
  │    10:32:14 nginx → connect:3306      │  "该进程尝试读取敏感    │
  │    10:32:10 nginx exec                │   系统文件，触发了      │
  │                                       │   rule-shadow-access"  │
  │                                       │                        │
  │                                       │  💡 相关发现            │
  │                                       │  • 同 PID 3 分钟内有   │
  │                                       │    47 次异常网络连接    │
  │                                       │                        │
  │                                       │  [创建规则] [深入调查]  │
  └───────────────────────────────────────┴─────────────────────────┘
  ```

##### **3.5 Sentinel（哨兵中心）- AI 主动洞察**
- **定位**：AI 的"主动输出"展示，而非被动查询
- **AI 原生特性**：
  - **洞察流**：AI 主动发现并推送的安全洞察
  - **一键行动**：每个洞察都有 AI 建议的下一步操作
  - **对话深入**：点击任何洞察可与 AI 深入讨论
- **洞察类型**：
  ```
  ┌─────────────────────────────────────────────────────────────────┐
  │ 🤖 AI Sentinel - Active Insights                               │
  ├─────────────────────────────────────────────────────────────────┤
  │                                                                 │
  │ ⭐ Shadow Rule Ready for Promotion                              │
  │ ┌─────────────────────────────────────────────────────────────┐ │
  │ │ Rule: block-redis-external-connect                          │ │
  │ │ Status: 156 hits, 0% false positive                         │ │
  │ │ AI Confidence: 98%                                          │ │
  │ │ AI says: "该规则已观察 72 小时，无误报，建议转正"            │ │
  │ │ [✅ Promote] [📊 View Details] [💬 Ask AI]                   │ │
  │ └─────────────────────────────────────────────────────────────┘ │
  │                                                                 │
  │ ⚠️ Anomaly Detected                                             │
  │ ┌─────────────────────────────────────────────────────────────┐ │
  │ │ Process: java (PID: 12345)                                  │ │
  │ │ Anomaly: File access frequency +300% from baseline          │ │
  │ │ AI says: "该进程突然大量读取配置文件，可能是配置热重载，     │ │
  │ │          也可能是异常行为，建议确认"                         │ │
  │ │ [🔍 Investigate] [✅ Mark as Normal] [🛡️ Create Rule]        │ │
  │ └─────────────────────────────────────────────────────────────┘ │
  │                                                                 │
  │ 💡 Optimization Suggestion                                      │
  │ ┌─────────────────────────────────────────────────────────────┐ │
  │ │ AI noticed: "3 条规则有重叠，可以合并优化"                   │ │
  │ │ [👀 View] [✨ Auto-merge] [❌ Ignore]                        │ │
  │ └─────────────────────────────────────────────────────────────┘ │
  │                                                                 │
  └─────────────────────────────────────────────────────────────────┘
  ```

##### **3.6 AI Context Bar（全局 AI 上下文栏）**
- **定位**：始终可见的 AI 助手入口
- **功能**：
  - 显示当前页面上下文的 AI 洞察
  - 一键展开与 AI 对话
  - 显示 AI 正在后台分析的状态
- **示例**：
  ```
  ┌─────────────────────────────────────────────────────────────────┐
  │ 💬 AI: "我注意到 nginx 容器今天有 23 次异常连接尝试..."        │
  │        [展开详情] [创建规则] [忽略]                     [💬 对话]│
  └─────────────────────────────────────────────────────────────────┘
  ```

##### **3.7 Settings（系统设置）**
- **AI 配置**：
  - AI Provider 选择（Ollama/OpenAI）
  - 模型选择和参数
  - 自动化级别（手动审核/半自动/全自动）
- **通知设置**
- **数据保留策略**

### 7.4 AI 原生的 API 设计

**核心 AI API 端点**：

```typescript
// AI 意图解析（Omnibox 核心）
POST /api/ai/intent
Request:  { "input": "阻止 nginx 访问敏感文件", "context": {...} }
Response: { 
  "intent": "create_rule",
  "confidence": 0.95,
  "params": { "process": "nginx", "action": "block", "target": "sensitive_files" },
  "preview": { "yaml": "...", "simulation": {...} }
}

// AI 规则生成
POST /api/ai/generate-rule
Request:  { "description": "...", "context": {...} }
Response: { 
  "rule": { "yaml": "..." },
  "reasoning": "AI 推理过程...",
  "confidence": 0.92,
  "warnings": ["可能影响 3 个进程"]
}

// AI 事件解释
POST /api/ai/explain
Request:  { "eventId": "...", "question": "为什么被拦截" }
Response: { 
  "explanation": "该进程尝试...",
  "relatedEvents": [...],
  "suggestedActions": [...]
}

// AI 上下文分析
POST /api/ai/analyze
Request:  { "type": "process", "id": "12345" }
Response: { 
  "summary": "该进程行为分析...",
  "anomalies": [...],
  "recommendations": [...]
}

// Sentinel 洞察流（WebSocket）
WS /api/ai/sentinel/stream
Message: { 
  "type": "shadow_promotion" | "anomaly" | "optimization",
  "title": "...",
  "summary": "...",
  "confidence": 0.95,
  "actions": [{ "label": "转正", "action": "promote", "params": {...} }]
}
```

### 7.5 AI 原生的前端组件架构

**AI 相关组件**（新增）：
```
components/
├── ai/                              # AI 核心组件
│   ├── AIomnibox.vue                # 全局 AI 对话框（Cmd+K）
│   ├── IntentPreview.vue            # AI 意图识别预览
│   ├── ActionConfirm.vue            # AI 操作确认面板
│   ├── AIContextBar.vue             # 底部 AI 上下文栏
│   ├── AIExplanation.vue            # AI 解释气泡
│   ├── AIConfidenceBadge.vue        # AI 置信度徽章
│   ├── StreamingResponse.vue        # AI 流式输出组件
│   └── QuickAsk.vue                 # 快速询问 AI 按钮
│
├── global/
│   ├── TopBar.vue                   # 顶栏（含 AI Omnibox 触发器）
│   └── StatusBar.vue                # 状态栏（含 AI 状态指示）
```

**AI Composables**（新增）：
```typescript
// composables/useAI.ts - AI 核心功能
export function useAI() {
  const parseIntent = (input: string) => { /* 调用 /api/ai/intent */ }
  const generateRule = (description: string) => { /* 调用 /api/ai/generate-rule */ }
  const explainEvent = (eventId: string, question: string) => { /* 调用 /api/ai/explain */ }
  const analyzeContext = (type: string, id: string) => { /* 调用 /api/ai/analyze */ }
  return { parseIntent, generateRule, explainEvent, analyzeContext }
}

// composables/useOmnibox.ts - Omnibox 状态管理
export function useOmnibox() {
  const isOpen = ref(false)
  const input = ref('')
  const intent = ref<Intent | null>(null)
  const toggle = () => { isOpen.value = !isOpen.value }
  const executeIntent = () => { /* 根据 intent 执行操作 */ }
  return { isOpen, input, intent, toggle, executeIntent }
}

// composables/useSentinel.ts - Sentinel 洞察订阅
export function useSentinel() {
  const insights = ref<Insight[]>([])
  const subscribe = () => { /* WebSocket 订阅 /api/ai/sentinel/stream */ }
  const executeAction = (insight: Insight, action: Action) => { /* 执行洞察建议的操作 */ }
  return { insights, subscribe, executeAction }
}
```

### 7.6 AI 原生交互流程示例

**场景 1：通过自然语言创建规则**
```
用户: [Cmd+K] "阻止所有容器访问宿主机的 Docker socket"

AI 解析:
┌────────────────────────────────────────────────────────────┐
│ 🧠 I understand you want to:                              │
│                                                            │
│ ✅ Create a BLOCK rule                                     │
│ 📦 Target: All containers (cgroup != host)                 │
│ 🔒 Protect: /var/run/docker.sock                           │
│                                                            │
│ Confidence: 97%                                            │
│                                                            │
│ Generated Rule Preview:                                    │
│ ┌────────────────────────────────────────────────────────┐ │
│ │ name: block-container-docker-socket                    │ │
│ │ match:                                                 │ │
│ │   filename: /var/run/docker.sock                       │ │
│ │   cgroup_not: host                                     │ │
│ │ action: block                                          │ │
│ │ severity: critical                                     │ │
│ └────────────────────────────────────────────────────────┘ │
│                                                            │
│ 📊 Simulation: Would block 0 events in last hour           │
│               (No containers attempted this access)        │
│                                                            │
│ [🌙 Deploy as Shadow] [🛡️ Deploy as Enforce] [✏️ Edit]     │
└────────────────────────────────────────────────────────────┘

用户: [点击 "Deploy as Shadow"]

系统: ✅ Rule deployed in Shadow mode. Monitoring for hits...
```

**场景 2：AI 主动洞察 → 用户行动**
```
Sentinel 推送:
┌────────────────────────────────────────────────────────────┐
│ ⭐ Shadow Rule Ready for Promotion                         │
│                                                            │
│ Rule: block-container-docker-socket                        │
│ Observation: 72 hours                                      │
│ Hits: 23 (all from suspicious container 'rogue-pod')       │
│ False Positives: 0%                                        │
│                                                            │
│ 🧠 AI Analysis:                                            │
│ "所有命中都来自同一个未知容器，该容器还尝试了其他敏感        │
│  文件访问。建议立即转正此规则并调查该容器。"                │
│                                                            │
│ [✅ Promote to Enforce] [🔍 Investigate Container]         │
│ [💬 Ask AI for more context]                               │
└────────────────────────────────────────────────────────────┘

用户: [点击 "Investigate Container"]

→ 跳转 Investigation 页面，自动填入查询：
  "Show all events from container 'rogue-pod' in the last 72 hours"
```

### 7.7 侧边栏与导航（更新）

**新侧边栏结构**：
```typescript
const navItems = [
  { icon: Target, label: 'Observatory', route: '/', section: 'core' },
  { icon: Wand2, label: 'Policy Studio', route: '/policy', section: 'core' },
  { icon: Search, label: 'Investigation', route: '/investigation', section: 'core' },
  { icon: Bot, label: 'Sentinel', route: '/sentinel', section: 'core' },
  { icon: Settings, label: 'Settings', route: '/settings', section: 'system' },
]
```

**Omnibox 快捷键提示**：顶栏显示 `⌘K` 或 `Ctrl+K`

### 7.8 技术栈

**保留**：
- Vue 3 + TypeScript
- Vue Router
- Vite
- Lucide Icons

**新增**：
- **Monaco Editor**：Policy Studio 的 YAML 编辑器
- **D3.js**：Investigation 的 Event Timeline
- **VueUse**：组合式工具函数（`@vueuse/core`）

**状态管理**：
- 继续使用 Composables 模式（`useXxx.ts`）
- AI 相关状态集中在 `useAI.ts` 和 `useOmnibox.ts`

### 7.9 样式系统

**保留现有设计语言**：
- 暗色主题（`--bg-void`, `--bg-surface` 等）
- 色彩系统（status colors, accent colors）
- 字体系统（Inter + JetBrains Mono）
- 圆角/阴影/过渡动画

**新增 AI 相关 Design Tokens**：
```css
:root {
  /* AI Omnibox */
  --ai-omnibox-bg: rgba(0, 0, 0, 0.9);
  --ai-omnibox-border: var(--accent-primary);
  --ai-glow: 0 0 20px rgba(139, 92, 246, 0.3);
  
  /* AI Confidence Indicator */
  --ai-confidence-high: var(--status-safe);      /* > 90% */
  --ai-confidence-medium: var(--status-warning); /* 70-90% */
  --ai-confidence-low: var(--status-critical);   /* < 70% */
  
  /* AI Context Bar */
  --ai-context-bar-bg: var(--bg-surface);
  --ai-thinking-pulse: rgba(139, 92, 246, 0.5);
  
  /* Insight Cards */
  --insight-shadow-bg: rgba(139, 92, 246, 0.1);
  --insight-anomaly-bg: rgba(245, 158, 11, 0.1);
  --insight-report-bg: rgba(59, 130, 246, 0.1);
  
  /* Policy Studio */
  --editor-line-highlight: rgba(255, 255, 255, 0.05);
  --simulation-success: var(--status-safe);
  --simulation-warning: var(--status-warning);
}
```

### 7.10 开发计划（前端部分）

**Phase 4.1：AI 核心基础设施**
- [ ] 实现 AI Omnibox 组件（全局入口）
- [ ] 实现 AI Context Bar（底部悬浮）
- [ ] 创建 `useAI.ts` composable
- [ ] 集成 AI 流式响应组件
- [ ] 重构顶栏（集成 Omnibox 触发器）

**Phase 4.2：Observatory 实现（AI 驱动）**
- [ ] AI 健康评估组件（自然语言解释）
- [ ] AI 威胁摘要（非传统图表）
- [ ] Sentinel Insights 预览卡片
- [ ] 响应式布局

**Phase 4.3：Policy Studio 实现（AI 辅助）**
- [ ] 自然语言规则创建界面
- [ ] AI 规则生成 + 推理展示
- [ ] Monaco YAML Editor
- [ ] AI 模拟结果解读
- [ ] Shadow/Enforce 部署流程

**Phase 4.4：Investigation 实现（AI 辅助）**
- [ ] 自然语言查询栏
- [ ] AI 查询翻译展示
- [ ] Event Timeline（D3.js）
- [ ] AI 上下文解释面板
- [ ] AI 推荐调查路径

**Phase 4.5：Sentinel 实现（AI 主动）**
- [ ] Insights Feed（洞察流）
- [ ] 各类洞察卡片（Shadow/Anomaly/Optimization）
- [ ] 一键行动按钮
- [ ] Ask AI 深入对话
- [ ] WebSocket 实时推送

**Phase 4.6：打磨与优化**
- [ ] AI 响应骨架屏/加载状态
- [ ] AI 错误处理和降级
- [ ] 性能优化（懒加载/代码分割）
- [ ] 响应式适配

### 7.11 与后端阶段的对接关系

| 前端功能 | 依赖的后端阶段 | 依赖的 API |
|---------|--------------|-----------|
| AI Omnibox 意图解析 | 第三阶段 | `POST /api/ai/intent` |
| AI 规则生成 | 第三阶段 | `POST /api/ai/generate-rule` |
| AI 事件解释 | 第三阶段 | `POST /api/ai/explain` |
| Sentinel 洞察流 | 第三阶段 | `WS /api/ai/sentinel/stream` |
| Event Timeline | 第一阶段 | `GET /api/events` |
| Process Profile | 第一阶段 | `GET /api/process/{pid}/profile` |
| Simulation Panel | 第二阶段 | `POST /api/simulation/run` |
| Shadow Rule 管理 | 第二阶段 | `POST /api/rules/{name}/promote` |

**开发策略**：
- **AI 功能优先**：先实现 AI 相关组件的 UI 和交互
- **Mock AI 响应**：使用预设的 Mock 数据模拟 AI 响应
- **渐进式接入**：后端 AI API 就绪后逐步切换

### 7.12 前端文件结构（AI 原生）

**新的目录结构**：
```
frontend/src/
├── App.vue                          # 主应用组件
├── main.ts                          # 入口文件
├── style.css                        # 全局样式
│
├── assets/
│   └── styles/
│       ├── reset.css
│       └── variables.css            # 含 AI 相关 Design Tokens
│
├── components/
│   ├── ai/                          # 🧠 AI 核心组件（新增）
│   │   ├── AIomnibox.vue            # AI 全局对话框（Cmd+K）
│   │   ├── IntentPreview.vue        # AI 意图识别预览
│   │   ├── ActionConfirm.vue        # AI 操作确认面板
│   │   ├── AIContextBar.vue         # 底部 AI 上下文栏
│   │   ├── AIExplanation.vue        # AI 解释气泡/面板
│   │   ├── AIConfidenceBadge.vue    # AI 置信度徽章
│   │   ├── StreamingResponse.vue    # AI 流式输出组件
│   │   ├── AIThinking.vue           # AI 思考中动画
│   │   └── QuickAsk.vue             # 快速询问 AI 按钮
│   │
│   ├── global/                      # 全局组件
│   │   ├── StatusBar.vue            # 底部状态栏
│   │   └── KeyboardShortcuts.vue    # 快捷键提示
│   │
│   ├── layout/                      # 布局组件
│   │   ├── Sidebar.vue              # 侧边栏
│   │   └── TopBar.vue               # 顶栏（含 AI Omnibox 触发器）
│   │
│   ├── observatory/                 # 观测站组件
│   │   ├── AIHealthScore.vue        # AI 健康评估（含自然语言解释）
│   │   ├── AIThreatSummary.vue      # AI 威胁摘要
│   │   ├── DefenseStats.vue         # 主动防御统计
│   │   └── SentinelPreview.vue      # 哨兵洞察预览
│   │
│   ├── policy/                      # 策略工坊组件
│   │   ├── RuleList.vue             # 规则列表
│   │   ├── AIRuleCreator.vue        # AI 规则创建器
│   │   ├── NaturalLanguageInput.vue # 自然语言输入
│   │   ├── AIRulePreview.vue        # AI 生成规则预览
│   │   ├── YamlEditor.vue           # YAML 编辑器（Monaco）
│   │   ├── AISimulationPanel.vue    # AI 模拟分析面板
│   │   └── DeployConfirm.vue        # 部署确认
│   │
│   ├── investigation/               # 调查台组件
│   │   ├── AIQueryBar.vue           # AI 查询栏
│   │   ├── QueryTranslation.vue     # AI 查询翻译展示
│   │   ├── EventTimeline.vue        # 事件时间轴
│   │   ├── EventList.vue            # 事件列表
│   │   ├── AIContextPanel.vue       # AI 上下文分析面板
│   │   └── AIInvestigationPath.vue  # AI 推荐调查路径
│   │
│   ├── sentinel/                    # 哨兵中心组件
│   │   ├── InsightsFeed.vue         # 洞察流
│   │   ├── InsightCard.vue          # 洞察卡片基类
│   │   ├── ShadowPromotionCard.vue  # Shadow 转正卡片
│   │   ├── AnomalyCard.vue          # 异常检测卡片
│   │   ├── OptimizationCard.vue     # 优化建议卡片
│   │   └── DeepAskAI.vue            # 深入询问 AI
│   │
│   └── common/                      # 通用组件
│       ├── Card.vue
│       ├── Badge.vue
│       ├── Modal.vue
│       └── LoadingSpinner.vue
│
├── composables/                     # 组合式函数
│   ├── useAI.ts                     # 🧠 AI 核心功能
│   ├── useOmnibox.ts                # Omnibox 状态管理
│   ├── useSentinel.ts               # Sentinel 洞察订阅
│   ├── useSimulation.ts             # 模拟引擎
│   ├── useInvestigation.ts          # 调查台状态
│   ├── useEvents.ts                 # 事件订阅
│   └── useAlerts.ts                 # 告警订阅
│
├── lib/
│   ├── api.ts                       # API 封装
│   ├── ai-api.ts                    # 🧠 AI API 封装
│   └── monaco.ts                    # Monaco Editor 配置
│
├── pages/
│   ├── Observatory.vue              # 观测站
│   ├── PolicyStudio.vue             # 策略工坊
│   ├── Investigation.vue            # 调查台
│   ├── Sentinel.vue                 # 哨兵中心
│   └── Settings.vue                 # 系统设置
│
├── router/
│   └── index.ts                     # 路由配置
│
└── types/
    ├── ai.ts                        # 🧠 AI 相关类型
    ├── events.ts                    # 事件类型
    ├── rules.ts                     # 规则类型
    └── sentinel.ts                  # Sentinel 类型
```

### 7.13 迁移策略

**删除的旧文件/目录**：
```
# 旧页面 → 新页面
- pages/Dashboard.vue        → Observatory.vue
- pages/LiveStream.vue       → Investigation.vue
- pages/Alerts.vue           → Investigation.vue
- pages/Rules.vue            → PolicyStudio.vue
- pages/Workloads.vue        → Investigation.vue (过滤器)
- pages/Profiler.vue         → PolicyStudio.vue (AI 生成白名单)
- pages/KernelXRay.vue       → Settings.vue 或删除
- pages/AIChat.vue           → Sentinel.vue + AIomnibox.vue

# 旧组件 → 新组件
- components/alerts/         → investigation/
- components/charts/         → observatory/
- components/kernel/         → 删除
- components/profiler/       → 删除（AI 重新实现）
- components/rules/          → policy/
- components/stream/         → investigation/
- components/topology/       → 删除
- components/ai/             → components/ai/（重写为 AI 原生）
```

**保留并复用的代码**：
1. **样式系统**：`variables.css`、全局 CSS 变量、暗色主题
2. **虚拟滚动**：`vue-virtual-scroller` 相关实现
3. **API 封装**：`lib/api.ts` 基础架构

**增量开发路径**：
1. 先保留旧页面，新建新页面（并行存在）
2. 新页面开发完成后，更新路由指向新页面
3. 确认新页面稳定后，删除旧页面和组件

### 7.14 测试说明

**AI 组件测试**（Vitest + Vue Test Utils）：
```
tests/components/
├── ai/
│   ├── AIomnibox.spec.ts            # 测试意图解析和预览
│   ├── StreamingResponse.spec.ts    # 测试流式输出渲染
│   └── AIConfidenceBadge.spec.ts    # 测试置信度显示
├── sentinel/
│   ├── InsightsFeed.spec.ts         # 测试洞察流订阅
│   └── InsightCard.spec.ts          # 测试洞察卡片交互
└── policy/
    ├── AIRuleCreator.spec.ts        # 测试 AI 规则生成流程
    └── AISimulationPanel.spec.ts    # 测试模拟结果展示
```

**E2E 测试**（Playwright）：
```typescript
// tests/e2e/ai-rule-creation.spec.ts
test('AI 规则创建端到端流程', async ({ page }) => {
  // 1. 打开 Omnibox
  await page.keyboard.press('Meta+k')
  
  // 2. 输入自然语言
  await page.fill('[data-testid="omnibox-input"]', '阻止 nginx 执行 curl')
  
  // 3. 验证意图解析
  await expect(page.locator('[data-testid="intent-preview"]')).toContainText('创建规则')
  
  // 4. 确认执行
  await page.click('[data-testid="action-confirm"]')
  
  // 5. 验证规则预览
  await expect(page.locator('[data-testid="rule-preview"]')).toBeVisible()
  
  // 6. 运行模拟
  await page.click('[data-testid="run-simulation"]')
  
  // 7. 验证模拟结果
  await expect(page.locator('[data-testid="simulation-result"]')).toBeVisible()
})
```

**API Mock**：
- 使用 MSW (Mock Service Worker) 模拟后端 AI API
- 预设的 AI 响应数据用于前端独立开发和测试
- 支持模拟流式响应 (SSE/WebSocket)

---

## 八、总结与展望

EulerGuard 3.0 的 AI 原生前端重构将安全工具从"仪表板驱动"转变为"对话驱动"。核心创新点：

1. **AI Omnibox**：统一入口，自然语言操作，告别繁琐的表单和导航
2. **意图驱动**：AI 理解用户意图，主动生成操作方案
3. **Sentinel 洞察**：AI 后台持续分析，主动推送建议
4. **Shadow 模式可视化**：规则部署前充分模拟，降低风险
5. **对话式安全**：从"我需要学习如何使用"到"我告诉系统我想要什么"

这套前端设计与后端的三层架构（内核执行 → 数据存储 → AI 智能）紧密配合，共同构建了一个真正 AI 原生的内核安全平台。

---

## 附录 A: 后端 API 总览

> **说明**: 以下是后端各阶段需要暴露的完整 API 列表，用于支撑第四阶段的 AI 原生前端。

### A.1 API 分类汇总

| 类别 | 端点数量 | 主要阶段 |
|------|---------|---------|
| 事件查询 | 5 | Phase 1 |
| 进程画像 | 3 | Phase 1 |
| 规则管理 | 6 | Phase 2 |
| 模拟引擎 | 2 | Phase 2 |
| AI 核心 | 6 | Phase 3 |
| Sentinel | 3 | Phase 3 |
| **总计** | **25** | - |

### A.2 完整 API 列表

```
# ═══════════════════════════════════════════════════════════════════
# Phase 1 API - 全息遥测仓库
# ═══════════════════════════════════════════════════════════════════

# 事件查询
GET    /api/events                        # 事件列表（分页、过滤）
GET    /api/events/{id}                   # 事件详情
GET    /api/events/range                  # 时间范围查询
GET    /api/stats                         # 总体统计
GET    /api/stats/rates                   # 速率统计

# 进程画像
GET    /api/process/{pid}/profile         # 进程画像
GET    /api/process/{pid}/tree            # 进程树
GET    /api/process/{pid}/events          # 进程相关事件

# ═══════════════════════════════════════════════════════════════════
# Phase 2 API - 双模执行引擎
# ═══════════════════════════════════════════════════════════════════

# 规则管理
GET    /api/rules                         # 规则列表（按模式分组）
POST   /api/rules                         # 创建规则
PUT    /api/rules/{name}                  # 更新规则
DELETE /api/rules/{name}                  # 删除规则
POST   /api/rules/{name}/promote          # Shadow 规则转正
GET    /api/rules/{name}/shadow-stats     # Shadow 命中统计

# 模拟引擎
POST   /api/simulation/run                # 运行模拟
POST   /api/simulation/compare            # 规则对比模拟

# ═══════════════════════════════════════════════════════════════════
# Phase 3 API - AI 接口层
# ═══════════════════════════════════════════════════════════════════

# AI 核心（支撑 Omnibox 和 AI 交互）
POST   /api/ai/intent                     # 意图解析（Omnibox 核心）
POST   /api/ai/generate-rule              # AI 规则生成
POST   /api/ai/explain                    # 事件解释
POST   /api/ai/analyze                    # 上下文分析
POST   /api/ai/review                     # AI 规则审查

# 语义查询
POST   /api/query/semantic                # 自然语言查询

# Sentinel 洞察
WS     /api/ai/sentinel/stream            # Sentinel 洞察流（WebSocket）
GET    /api/ai/sentinel/insights          # 历史洞察列表
POST   /api/ai/sentinel/action            # 执行洞察建议的操作
```

### A.3 前端页面与 API 对应关系

| 前端页面 | 核心 API | 依赖阶段 |
|---------|---------|---------|
| **AI Omnibox** | `/api/ai/intent`, `/api/ai/generate-rule` | Phase 3 |
| **Observatory** | `/api/stats`, `/api/ai/sentinel/insights` | Phase 1+3 |
| **Policy Studio** | `/api/rules/*`, `/api/simulation/*`, `/api/ai/generate-rule` | Phase 2+3 |
| **Investigation** | `/api/events/*`, `/api/query/semantic`, `/api/ai/explain` | Phase 1+3 |
| **Sentinel** | `/api/ai/sentinel/*`, `/api/rules/{name}/promote` | Phase 2+3 |