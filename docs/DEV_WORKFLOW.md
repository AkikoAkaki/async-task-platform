# 开发流程（Design → Delivery → Release）

> 适用范围：本仓库（Distributed Delay Queue）后续迭代的**完整研发流程**。
> 目标：把“从设计规划到里程碑完成并发布”的每一步落到可执行清单与命令，便于你（以及未来的协作者/AI）按同一套规则持续交付。

---

## 0. 你需要先理解的约定（项目现状对齐）

### 0.1 项目当前形态（MVP）
- 接口层：gRPC + Protobuf，服务为 `DelayQueueService`（见 `api/proto/queue.proto` 与文档 `docs/API.md`）。
- 存储层：Redis Sorted Set（ZSet）+ Lua，保证“到期任务批量弹出”的原子性（见 `internal/storage/redis/*` 与 `docs/adr/001-architecture-and-storage.md`）。
- 运行形态：本地通过 Docker Compose 启 Redis，Server/Worker 通过 `make run-server`、`make run-worker` 启动（见 `Makefile` 与 `README.md`）。

### 0.2 研发协作约定（分支 / 版本 / Changelog）
- 分支策略：Git Flow（见 `docs/adr/002-gitflow-and-versioning.md`；当前为 Draft，但其约定已经写入 `DEVELOPMENT_LOG.md`）。
- 版本策略：SemVer（语义化版本）：`MAJOR.MINOR.PATCH`。
- Tag 规则：`vX.Y.Z`（注：tag 带 `v` 前缀；而 `CHANGELOG.md` 的标题使用不带 `v` 的 `[X.Y.Z]`）。
- Changelog 工具：`git-chglog`，配置为 `.chglog/config.yml`，模板为 `.chglog/CHANGELOG.tpl.md`。
- CI 约束：`release/*` 分支会跑 `release-readiness` gate（见 `.github/workflows/ci.yaml`），它会检查：
  - 分支命名必须 `release/vX.Y.Z`
  - `CHANGELOG.md` 必须包含 `[X.Y.Z]`
  - `DEVELOPMENT_LOG.md` 必须提到目标 tag（例如 `v0.1.0`）
  - 生成 `git-chglog --next-tag vX.Y.Z` 预览工件

### 0.3 工业级基线（你必须内化的 5 大特征）
- 高可用：单节点故障不影响整体服务；对关键组件预留降级或主备/选举方案。
- 一致性：并发写入不乱序；对 CAP 做显式取舍，当前队列在分区时默认偏 AP，后续特性需声明 CP/AP 选择。
- 可观测：必须能用指标/日志/Tracing 定位问题，无需读代码；后续改动要考虑埋点和链路追踪。
- 可扩展：流量翻倍可通过水平扩容或分片撑住（如 topic 分片、Consistent Hashing）。
- 成本意识：存储/计算方案需有成本对比，能解释“为什么这个最划算”。

---

## 1. 环境准备（一次性）

### 1.1 必装工具（Windows 11 优先）
按 `docs/DEV_SETUP.md` 准备：
- Go 1.21+
- Docker Desktop（WSL2 backend）
- make（Chocolatey 或 Scoop）

> 注意：以仓库实际目录为准，Compose 文件在 `deploy/docker-compose.yaml`（与 `Makefile` 一致）。

### 1.2 初始化项目配置
1. 复制配置模板（本地文件不提交）：
   - `config/config.example.yaml` → `config/config.yaml`
2. 启动 Redis：
   - `make up`
3. 启动 Server（示例流程）：
   - `make run-server`
4. （可选）启动 Worker：
   - `make run-worker`

### 1.3 常用命令（你会反复用）
- 格式化：`make fmt`
- 静态检查：`make lint`
- 测试：`make test`
- 生成 proto：`make proto`

---

## 2. 从“需求”到“计划”（每个功能/修复都要走）

本项目建议使用“轻量 PRD + 可验证验收条件 + ADR（必要时）”的组合。

### 2.1 需求输入（必须落到文字）
每个需求至少要写清：
1. **背景 / 痛点**：为什么要做。
2. **目标**：要达到什么业务/技术目标。
3. **非目标**：明确不做什么（防止 scope creep）。
4. **用户行为/调用方式**：例如是 client 调 `Enqueue` 还是 worker 调 `Retrieve`。
5. **验收标准（Acceptance Criteria）**：必须可验证、可测试。

建议把这些写成一个 Issue（或一段 markdown），并给出最小验收例子：
- 输入（请求）是什么
- 输出（响应/副作用）是什么
- 失败场景的错误码/错误信息是什么

### 2.2 计划拆分（你后续开发的“最小闭环”）
把需求拆成可以独立合并的子任务（每个子任务都能在 PR 中被 review）：
- API 变更（proto/handler/service）
- Domain 变更（内部 service/model）
- Storage 变更（JobStore 接口/Redis 实现/Lua 脚本）
- 配置变更（`config/*.yaml` 与读取逻辑）
- 文档变更（`docs/API.md`、`docs/ARCHITECTURE.md`、ADR）
- 测试（单测/集成测试）

> 原则：不要把“API 变更 + 大重构 + 新存储后端 + 新部署方式”揉进一个 PR。

### 2.3 Definition of Ready（开始写代码前的检查）
在开始编码前，确保以下问题都有答案：
- 这个改动是否影响 gRPC 合同（proto）？
- 是否引入破坏性变更（Breaking Change）？
- 是否需要 ADR（架构/流程/存储语义变化）？
- 是否会影响 release（需要更新 `CHANGELOG.md` / `DEVELOPMENT_LOG.md`）？
- 是否有明确的错误码映射策略（参考 `docs/API.md` 的 guidelines）？

---

## 3. 设计阶段（Design）

### 3.1 什么时候必须写 ADR
当满足任意一条时，建议新增/更新 ADR：
- 改变存储语义（例如：从 JSON member → Protobuf member；或引入 ID 索引）
- 改变调度/投递语义（例如：从 at-least-once → exactly-once 的承诺）
- 引入新组件/服务边界（例如：拆出独立 scheduler 服务）
- 引入新的基础设施依赖（例如：Sentinel/Cluster/消息总线）

ADR 最小结构：Context / Decision / Consequences。
- 新 ADR 放在 `docs/adr/`，按编号递增。
- 状态建议：Draft → Accepted（合入后）→ Superseded（被替代时）。

### 3.2 API 设计（proto）
如果要改 `api/proto/queue.proto`：
- 优先“向后兼容”的 additive 变更（加字段、加新 RPC）。
- 避免改字段类型/语义；如必须改，视为 Breaking Change。
- 删除字段前先 `reserved`（proto 的最佳实践；并在变更说明中解释）。

同步更新：
- `docs/API.md`：补充新的 RPC/字段说明与示例。
- （若影响流程）`docs/ARCHITECTURE.md`：更新 runtime flow。

### 3.3 数据与存储设计（JobStore）
改动存储逻辑时，先对齐现有抽象：
- 接口：`internal/storage/interface.go`
- Redis 实现：`internal/storage/redis/store.go`
- Lua：`internal/storage/redis/script.go`

设计必须回答：
- 任务的唯一标识是什么（`Task.id`）？是否允许客户端传入？
- 任务的排序依据是什么（`execute_time`）？精度如何（秒/毫秒）？
- `GetReady` 的一致性/幂等性如何保证？
- 是否存在重复投递？如果存在，worker 如何处理？

### 3.4 设计取舍（Trade-offs）必答题
- 技术选型要能辩论：为何用 Redis ZSet+Lua 而不是 Timing Wheel/Kafka；若改协议，gRPC 相对 REST 的体积/类型安全收益。
- CAP 取舍要明确：队列在分区时是偏 AP 还是 CP，不同场景（限流 vs 支付）取向不同。
- 分片与扩容：当 Redis 容量/吞吐不足时，如何按 topic 或哈希做分片，以及数据迁移策略（Consistent Hashing）。
- 幂等与锁：高并发下怎么防止重复执行（去重键、乐观锁、分布式锁、续期/超时释放）。

### 3.5 工程化起手式（写第一行业务代码前）
- 数据库/存储迁移（Schema as Code）：若引入持久化存储，先确定迁移工具（如 Flyway/Liquibase），所有表结构变更脚本化。
- API First：改协议前先写接口文档/契约（proto 已是契约，新增 HTTP/OpenAPI 也遵循先 Spec 后实现）。
- 本地环境容器化：依赖尽量用 Compose 拉起（当前 Redis 已有），新增依赖也写进 `deploy/docker-compose.yaml`。
- 配置外置：通过环境变量覆盖默认配置，避免把密钥写进仓库；本地可用 `.env`（记得 .gitignore）。
- 敏感信息防护：启用 git-secrets/pre-commit 规则；依赖安全扫描（CI 里已有 Trivy，可补充 SCA 如 Snyk/OWASP DepCheck）。

---

## 4. 开发实施（Implement）

### 4.1 分支策略（Git Flow，按项目约定执行）
- 日常开发分支（从 `develop` 拉出）：
  - `feature/<slug>`：新功能
  - `bugfix/<slug>`：非紧急缺陷
- 发布分支（从 `develop` 拉出）：
  - `release/vX.Y.Z`
- 线上紧急修复（从 `main` 拉出）：
  - `hotfix/vX.Y.Z`

> 如果当前仓库还没有 `develop`，建议你先创建并推送一次，让流程和 ADR 对齐；否则就临时以 `main` 扮演集成分支，但要在文档/ADR 中明确“过渡期策略”。

### 4.2 提交信息规范（与 git-chglog 一致）
`.chglog/config.yml` 使用 `angular` 风格：
- 格式：`<type>(<scope>): <subject>`
- 允许的 type（CI 会按这些分类渲染）：
  - `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `test`
- `scope` 建议用模块名：`storage`, `api`, `scheduler`, `worker`, `server`, `docs` 等。
- 破坏性变更：在正文里写 `BREAKING CHANGE:`（会被 notes 抓取）。

示例：
- `feat(api): add batch_size cap to Retrieve`
- `fix(storage): prevent empty task payload from crashing unmarshal`
- `docs(architecture): document scheduler loop invariants`

### 4.3 代码改动的“落点”规则（按目录定位）
- gRPC 合同：`api/proto/queue.proto`
- gRPC handler / service 编排：`internal/handler/`、`internal/queue/service.go`
- 调度逻辑：`internal/scheduler/`
- 存储抽象与实现：`internal/storage/`、`internal/storage/redis/`
- 可执行入口：`cmd/server/main.go`、`cmd/worker/main.go`

原则：
- 不要在 `cmd/*` 里堆业务逻辑；把逻辑放到 `internal/*`，`cmd/*` 只做 wiring（配置加载、依赖注入、启动）。

### 4.4 修改 proto 的标准流程（必须按顺序）
1. 编辑 `api/proto/queue.proto`
2. 生成代码：`make proto`
3. 编译验证：`go test ./...`（或 `make test`）
4. 更新文档：`docs/API.md`（必要时 `docs/ARCHITECTURE.md`）
5. 更新 changelog（若影响用户/行为）：`CHANGELOG.md`

### 4.5 修改 Redis/Lua 的标准流程
1. 明确 Lua 脚本的输入输出（keys/args）与原子性边界。
2. 改 `internal/storage/redis/script.go` 与对应调用方。
3. 本地启动 Redis：`make up`
4. 跑最小验证：
   - 启动 server/worker 或使用脚本客户端（如 `scripts/test_client.go`）
5. 加测试（如果项目里已有对应测试结构，就补单测/集成测；避免只靠手测）。

### 4.6 Go 代码规范（结合字节标准）
- 格式化：提交前运行 `make fmt`（调用 goimports）；单行建议不超 120 列，必要时在逗号/运算符处换行。
- 注释：导出包/函数必须写用途与使用示例，复杂私有函数需解释设计意图；避免在代码留 TODO/FIXME，未完事项用 Issue/任务跟踪。
- 命名：语义化，少缩写；包名小写单词；单方法接口以 er 结尾，多方法接口直述职责；布尔用 is/has/should/can 开头；常量大写下划线。
- 函数：参数超 3 个用 struct 封装并自校验；返回切片时约定错误场景返回空切片而非 nil；指针接收者为主，小值可值接收。
- 错误处理：用 `fmt.Errorf("...: %w", err)` 包装；用 `errors.Is/As` 判错，禁止字符串比较；错误信息勿含敏感数据，顶层要转换为用户/调用方可读的错误码。
- 并发：禁止裸 goroutine，必须有 WaitGroup 或 context 取消；循环启动 goroutine 时传值参数；通道标注方向（<-chan/chan<-）。

### 4.7 业务健壮性与架构分层
- 幂等：为 Enqueue/任务执行设计幂等键（如 RequestID/TaskID 去重）；失败重试要可安全重入。
- 背压与削峰：在高流量场景考虑限流（令牌桶/漏桶）与队列满时策略（丢弃/死信队列）。
- 锁与一致性：按需使用乐观锁或 Redis 分布式锁，明确锁续期/超时释放策略。
- 分层/DDD：接口层做校验与 DTO 转换，应用层编排用例，领域层放纯业务规则，基础设施层实现存储/客户端；避免把业务写进 `cmd/*`。

---

## 5. 本地验证（Verify）

每次准备开 PR 前，至少保证：
1. `make fmt`
2. `make lint`
3. `make test`

如果你改了：
- proto：必须跑 `make proto` 并确保生成代码被提交。
- 文档：确保示例命令仍可运行（尤其是端口、服务名、RPC 路径）。

### 5.1 测试金字塔
- 单元测试：覆盖 Domain 规则（>80% 覆盖优先），外部依赖全部 mock。
- 集成测试：尽量用真依赖或 Testcontainers（Redis/MQ），避免过度 mock 导致行为漂移。
- 端到端/回归：关键链路（入队→到期→取回）需有自动化脚本。

### 5.2 压测与背压验证
- 使用 k6/JMeter/Locust 做基线压测，输出 QPS、P95/P99 延迟与资源占用。
- 验证限流/背压策略：队列满或消费者变慢时系统的降级/丢弃/死信表现。

### 5.3 混沌与故障演练
- 断 Redis、升高延迟，观察服务降级与重试（指数退避）是否生效。
- 模拟时钟偏移或分区，看任务触发的偏差并记录风险。

### 5.4 可观测性基线
- 规划指标：入队/取回 QPS、延迟、错误率；Lua 脚本失败率；Redis 慢查询。
- 日志：结构化 JSON，包含 TraceID；异步/并发场景确保 TraceID 传递。
- Tracing：推荐 OpenTelemetry/Jaeger，至少覆盖 gRPC 入/出与 Redis 操作。

---

## 6. PR 流程（Review & Merge）

### 6.1 PR 的最小内容
- 说明“做了什么、为什么、怎么验证”。
- 贴出关键命令的验证结果（例如 `make test`）。
- 如果涉及 API/行为变化：必须指向 `docs/API.md` / `CHANGELOG.md` 更新。
- 如果涉及架构/存储语义：必须新增/更新 ADR。

### 6.2 PR 自检清单（强制）
- [ ] 分支命名符合约定（feature/bugfix/release/hotfix）。
- [ ] 提交信息符合 angular 格式（方便 `git-chglog`）。
- [ ] `make fmt && make lint && make test` 全绿。
- [ ] 影响用户的变更已写入 `CHANGELOG.md`。
- [ ] 里程碑/发布相关变更已同步到 `DEVELOPMENT_LOG.md`（如果适用）。

---

## 7. 里程碑收敛（Milestone → Release Branch）

当一个版本目标（例如 `v0.1.0`）达到“功能基本齐备”时，进入收敛阶段。

### 7.1 切发布分支
从 `develop` 切：
- 分支名：`release/vX.Y.Z`
- 立即 push，让 CI 开始跑 `release-readiness` gate。

### 7.2 文档与版本条目必须先行
1. 更新 `CHANGELOG.md`
   - 保证存在标题：`## [X.Y.Z] - YYYY-MM-DD`（或先保持 Unreleased，但要满足 CI 对 `[X.Y.Z]` 的检测）
2. 更新 `DEVELOPMENT_LOG.md`
   - 在 Release Workflow / Milestones 或 Change Journal 里提到目标 `vX.Y.Z`

### 7.3 生成 release notes 预览（对齐 CI）
本地可执行：
- `git-chglog --next-tag vX.Y.Z > release-notes.md`

目的：
- 提前检查 commit 类型是否被正确分组
- 提前发现“漏写 changelog/漏写 type”的提交

### 7.4 发布分支的规则
- 只允许 release-blocking 的修复进入发布分支。
- 新功能回到 `develop`。

---

## 8. 发布（Release）

### 8.1 合并与打 Tag
当 release 分支 CI 全绿：
1. 将 `release/vX.Y.Z` 合并到 `main`
2. 在 `main` 上打 annotated tag：`vX.Y.Z`
3. 将 `main` 回合并到 `develop`（保持历史一致）

### 8.2 发布说明（GitHub Release）
- 以 `git-chglog` 生成的 release-notes 为骨架
- 补上人工说明：重大变更、迁移说明、关键 ADR 链接

### 8.3 CI/CD 与镜像最佳实践
- CI 最低要求：格式化、lint、测试、构建镜像全自动，主干/发布分支必须可构建。
- 安全与合规：保留 Trivy/SCA 扫描；必要时加 git-secrets 检查。
- Docker 镜像：使用多阶段构建，编译阶段用完整工具链，运行阶段用精简镜像；镜像务必打版本 Tag（vX.Y.Z），避免只用 latest。

---

## 9. 发布后（Post-release）
- 将下一版本的 `CHANGELOG.md` 预置为 `Unreleased` 区块（Keep a Changelog 的惯例）。
- 在 `DEVELOPMENT_LOG.md` 记录本次发布总结与下一里程碑计划。
- 如本次引入了临时方案（例如过渡期没有 `develop`）：发布后立刻补齐治理工作（创建 `develop`、补 ADR 状态等）。

---

## 10. 快速模板（你复制就能用）

### 10.1 新功能开发模板
1. 写 Issue：背景/目标/非目标/验收标准
2. （如需）写 ADR：Draft
3. 拉分支：`feature/<slug>`
4. 实现 + 本地验证：`make fmt && make lint && make test`
5. 更新文档：API/Architecture
6. 更新 `CHANGELOG.md`
7. 开 PR

### 10.2 发布模板（vX.Y.Z）
1. 从 `develop` 切 `release/vX.Y.Z` 并 push
2. 更新 `CHANGELOG.md` 与 `DEVELOPMENT_LOG.md`
3. `git-chglog --next-tag vX.Y.Z` 检查 release notes
4. CI 全绿后合并到 `main`
5. `main` 打 tag：`vX.Y.Z`
6. 回合并到 `develop`

---

## 11. 文档写作与汇报（字节风格）

### 11.1 设计文档（Design Doc / RFC）
- 必备：背景与问题、目标/约束（QPS、P99、CAP 取舍）、架构拓扑、存储设计、接口协议、核心算法、容错/降级、扩展方案。
- 写法：数据驱动、列出异常/边界处理（超时、重复消费、分区）、强调 Trade-off（为何选 CP/AP、Push/Pull、Redis/ZSet）。

### 11.2 复盘（Post-Mortem）
- 时间线、影响面、根因（5 Whys）、改进动作（短期/长期），遵循无责文化但要深挖根因。

### 11.3 预研/选型报告
- 现状瓶颈、业界对比（吞吐/延迟/成本/运维）、压测数据、结论与建议。适用于新存储/新追踪方案等评估。

### 11.4 性能与容量规划
- Profiling/Tracing 发现的瓶颈、优化前后对比、未来 3-6 个月容量预估与资源预算。

### 11.5 API/Onboarding 文档
- README/快速开始、SDK 使用方法（超时/重试/限流）、错误码与幂等约定。

### 11.6 周报/日报
- 周报：Progress / Risks / Next Steps，突出指标和决策依据。
- 日报：用 [状态]+[任务]+[关键细节] 记录执行与风险。

### 11.7 写作风格
- 数据说话，少“可能/大概”；用图表/日志/指标佐证。
- 结构化 Bullet，直截了当；列出里程碑规划与 Backlog（可观测性、死信、多租户等待办）。

