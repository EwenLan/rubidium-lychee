# 开发进度

## 当前状态

**阶段**：Phase 1-5 全部完成

**最后更新**：2026-07-05

## 已完成

- [x] Phase 0：文档初始化（development-plan.md、design.md、progress.md）
- [x] Phase 1：Mock 前向模拟扩展
  - 障碍系统：`--obstacles S08:ROCKFALL` 生成、`CLEAR`/`SQUAD_CLEAR` 清除、残留税
  - 设卡系统：`SET_GUARD`/`BREAK_GUARD`/`FORCED_PASS`、风化
  - 小分队调度：4 种动作切比雪夫距离延迟落地
  - 探路标记：生成、消耗（-3 最低 2）、45 帧过期
  - 静态对手 + `--enemy-guards S10:3` 放置敌方设卡
- [x] Phase 2：自适应寻路 + 轨迹预测
  - `game/path.go`：`ShortestPathExcluding` 排除节点集
  - `strategy/path.go`：`planPath` 每帧重规划，排除障碍/敌方设卡
  - `strategy/prediction.go`：`predictArrivals` 预测到各节点帧数
- [x] Phase 3：对手建模
  - `strategy/opponent.go`：追踪对手位置/状态/设卡
  - 预测对手到宫门路径和帧数
  - `weakenTargets` 按优先级排序候选目标
- [x] Phase 4：小分队意图系统
  - `strategy/squad.go`：意图评估 + 优先级排序 + 人手预算
  - 骨架意图：`scout_gate`(P100)、`scout_process`(P80)、`clear_obstacle`(P60)
  - 反应填充：`weaken_blocking`(P50)、`weaken_keypass`(P40)
  - 在途小分队追踪、去重
- [x] Phase 5：集成 + 端到端验证

### Phase 5 端到端验证结果

| 测试场景 | 命令 | 交付帧 | 得分 | 验证点 |
|---|---|---|---|---|
| 基线（无障碍） | `--rush-frame 30` | 45 | 240 | 回归通过 |
| 侦察标记 | `--rush-frame 30` | **36** | 240 | ✅ SCOUT 派到 S11/S13/S14，每个省 3 帧 |
| 障碍清除 | `--obstacles S10:MUD --rush-frame 30` | **40** | 240 | ✅ SQUAD_CLEAR 第 1 帧派出，第 16 帧落地清障 |
| 削弱设卡 | `--enemy-guards S10:3 --rush-frame 30` | **59** | 240 | ✅ 两次 SQUAD_WEAKEN，防御 3→1→0 |

**侦察标记效果**：
- Round 17: SQUAD_SCOUT → S11 (delay=3)
- Round 21: SQUAD_SCOUT → S14 (delay=6)
- Round 22: SQUAD_SCOUT → S13 (delay=4)
- S14 标记：`scouted:[{teamId:RED, processReduceRound:3, remainingTriggers:1}]`
- 验核帧数从 6 降到 3，交付提前 9 帧

**障碍清除效果**：
- Round 1: SQUAD_CLEAR → S10 (delay=15)
- Round 16: 障碍清除，S10 `hasObstacle` 变 false，`obstacleResidue` 出现
- 主车队无绕路，直接通过 S10

**削弱设卡效果**：
- Round 1: SQUAD_WEAKEN → S10 (delay=15, defense 3→1)
- Round 16: SQUAD_WEAKEN → S10 (delay=10, defense 1→0)
- Round 26: 设卡失效（`active:false`），路径畅通

## 适应性验证

所有决策从 `start`/`inquire` 派生，不硬编码：
- ✅ 节点 ID：从 `gameMap.Nodes` 和 `state.Nodes` 取
- ✅ 处理点：从 `gameMap.Node(id).ProcessType` 判断
- ✅ 障碍：从 `state.Nodes[id].HasObstacle` 判断
- ✅ 设卡：从 `state.Nodes[id].Guard` 判断
- ✅ 路径：每帧 `planPath` 重算
- ✅ 预测：基于当前最短路 + 速度

地图变体、障碍分布变化、敌方设卡位置变化都能自动适应。

## 构建状态

- `go build ./...`：✅ 通过
- `go vet ./...`：✅ 通过
- `go test ./...`：✅ 通过（protocol 包有样本解析测试）

## 待办（后续优化）

- [ ] 天气模拟（mock 当前跳过，SQUAD_SCOUT 延迟不含山雾 +2）
- [ ] 真实服务器速度校准（当前 speed=10 匹配 mock，真实服务器用 base 1000 + 路线系数）
- [ ] `BREAK_GUARD`/`FORCED_PASS` 主车队策略（当前遇敌方设卡只靠 SQUAD_WEAKEN）
- [ ] 资源领取/使用策略（ICE_BOX/马类/INTEL）
- [ ] 皇榜任务策略（CLAIM_TASK）
- [ ] 窗口争夺出牌策略（WINDOW_CARD）
- [ ] 终局急策（RUSH_SPEED/RUSH_PROTECT/BREAK_ORDER）
- [ ] 设卡/攻坚/悬赏策略
- [ ] 对手动态行为模拟（mock 当前对手是静态的）

## 阻塞与备注

- 无阻塞。当前实现满足"小分队完整移动和任务执行策略"的核心需求。
- Mock 的 FORCED_PASS 简化为直接时间税，不模拟 3 拍窗口。够用于策略验证。
- 窗口争夺未模拟。策略端遇到窗口先发空动作。
