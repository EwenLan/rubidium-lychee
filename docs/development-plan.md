# 小分队策略开发计划

## 目标

采用 **C 混合式**（意图式骨架 + 反应式填充）实现小分队完整的移动和任务执行策略，并同步扩展 mock 服务器（前向模拟）以支持本地验证。

## 核心需求

| 需求 | 说明 |
|---|---|
| 方案 | C 混合式：意图式骨架（每帧重判触发条件）+ 反应式填充 |
| 对手建模 | 实时追踪对手位置和设卡，动态选削弱目标 |
| 预测 horizon | 远——每次 inquire 重新规划主车队到宫门/终点的完整路径 |
| Mock 扩展 | 前向模拟：mock 真实模拟游戏规则推进，而非脚本回放 |
| 文档 | docs/ 下维护开发计划、设计文档、进度，实时更新 |

## 阶段划分

### Phase 0：文档与任务初始化
- 创建 `docs/development-plan.md`（本文）
- 创建 `docs/design.md`（架构与模块设计）
- 创建 `docs/progress.md`（进度跟踪）
- 建立任务列表

### Phase 1：Mock 前向模拟扩展
扩展 `cmd/mockserver/engine.go`，使其真实模拟游戏规则，支撑小分队策略验证。

子任务：
1. 障碍系统：开局生成障碍、`CLEAR`/`SQUAD_CLEAR` 清除、残留通行税
2. 设卡系统：`SET_GUARD`/`BREAK_GUARD`/`FORCED_PASS`、风化衰减
3. 小分队调度：4 种动作的提交→延迟→落地全流程
4. 探路标记：生成、消耗（处理 -3 最低 2）、45 帧过期
5. 增援/削弱：`SQUAD_REINFORCE`/`SQUAD_WEAKEN` 修改设卡防守值
6. 天气模拟（可选）：暴雨/山雾/酷暑的基础效果，影响 SQUAD_SCOUT 延迟

### Phase 2：自适应寻路与轨迹预测
- `strategy/path.go`：每帧根据当前地图状态（障碍/设卡/残留税）重新规划路径
- `strategy/prediction.go`：基于当前路径 + 速度（buffs/天气）预测到各关键节点的到达帧

### Phase 3：对手建模
- `strategy/opponent.go`：追踪对手位置、状态、设卡
- 预测对手到宫门的路径
- 识别削弱目标（敌方设卡在我方路径或关键节点上）

### Phase 4：小分队意图系统
- `strategy/squad.go`：意图定义、优先级、触发条件
- 骨架意图：
  - `scout_gate`：主车队 ≤15 帧到宫门且无活跃标记 → SCOUT
  - `scout_process`：主车队 ≤10 帧到处理点且无标记 → SCOUT
  - `clear_obstacle`：主路径有障碍且无更优绕路 → SQUAD_CLEAR
- 反应式填充：
  - `reinforce_guard`：己方关键设卡被攻坚 → REINFORCE
  - `weaken_guard`：敌方设卡挡路或挡关键节点 → WEAKEN
  - `scout_resource`：路过资源点且主车队将领取 → SCOUT（低优先级）

### Phase 5：集成与端到端验证
- 将小分队决策接入 `strategy.Decide`（与主车队动作同帧提交，不同类别）
- 端到端测试：mock + 客户端，验证小分队动作生效、标记消耗、障碍清除
- 回归测试：原有送货策略不退化
- 性能验证：每帧决策 < 500ms

### Phase 6：文档收尾
- 更新 `docs/progress.md` 为最终状态
- 更新 `docs/design.md` 反映实际实现
- 更新 `CLAUDE.md` 反映新模块

## 风险与未知

| 风险 | 缓解 |
|---|---|
| Mock 前向模拟工作量大 | 按 Phase 1 子任务增量实现，每个子任务可独立验证 |
| 意图触发条件复杂 | 先实现骨架意图（scout_gate/scout_process/clear_obstacle），反应式填充后做 |
| 对手路径预测不准 | 用最短路 + 当前速度作粗估，不追求精确 |
| 性能（每帧重规划路径） | 节点数 ≤15，Dijkstra 复杂度足够低 |
| 设卡/攻坚逻辑复杂 | Mock 先支持基础 SET/BREAK/FORCED_PASS，风化和悬赏后做 |

## 验收标准

1. 客户端在扩展后的 mock 上能完成送货，且小分队动作被正确模拟
2. 验核前 S14 有探路标记，验核帧数从 6 降到 3
3. 路上有障碍时，客户端能 SQUAD_CLEAR 或绕路
4. 对手设卡时，客户端能 SQUAD_WEAKEN 或绕路
5. `go build`/`go vet`/`go test` 全通过
6. docs/ 下文档与实现一致
