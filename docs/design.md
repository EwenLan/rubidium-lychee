# 小分队策略设计文档

## 架构概览

```
main.go                         入口
internal/
  protocol/                     消息结构体（已有）
  transport/                    TCP 框帧（已有）
  game/                         地图模型 + 状态解析 + 基础 Dijkstra（已有）
  log/                          日志 + 回放（已有）
  strategy/
    strategy.go                 Decide 主入口，集成主车队 + 小分队决策
    path.go                     自适应寻路（每帧重规划，感知障碍/设卡/残留税）
    prediction.go               主车队轨迹预测（长 horizon，到宫门/终点）
    opponent.go                 对手建模（位置/设卡/预测路径）
    squad.go                    小分队意图系统 + 反应式填充
    types.go                    共享类型（在途小分队、标记、意图）
cmd/mockserver/
  main.go                       入口（live/replay 模式）
  engine.go                     前向模拟引擎（扩展）
  obstacle.go                   障碍生成与管理
  guard.go                      设卡管理
  squad.go                      小分队调度/落地模拟
  scout.go                      探路标记管理
  fixture.go                    静态地图数据（已有）
```

## 每帧决策流程

```
inquire 到达
  ↓
BuildState(state, playerID)
  ↓
Strategy.Decide(state, gameMap)
  ├─ updateInternalState(state)        // 更新在途小分队、标记、对手模型
  ├─ planPath(state, gameMap, target)  // 自适应寻路，返回节点序列
  ├─ predictArrivals(state, path)      // 预测路径上各节点到达帧
  ├─ decideMainCart(state, path)       // 主车队动作（MOVE/PROCESS/VERIFY/DELIVER/CLEAR...）
  └─ decideSquad(state, path, arrivals) // 小分队动作（SCOUT/CLEAR/REINFORCE/WEAKEN）
  ↓
返回 []Action（最多 1 主车队 + 1 小分队）
```

## 模块设计

### 1. 自适应寻路 `strategy/path.go`

**目标**：每帧从当前主车队位置到目标（宫门或终点）重新规划路径，感知障碍、敌方设卡、残留税。

**接口**：
```go
type PathPlan struct {
    Nodes    []string  // 节点序列（含起点和终点）
    Distance int       // 总距离
    Blockers []Blocker // 路径上的阻挡项（障碍/设卡），含位置和类型
}

type Blocker struct {
    NodeID    string
    Type      string // "obstacle" / "enemy_guard" / "residue"
    TimeTax   int    // 通过所需额外帧数（FORCED_PASS 成本）
    CanClear  bool   // 是否可清除（障碍 true，设卡 false）
}

func (s *Strategy) planPath(state *game.State, gameMap *game.GameMap, from, to string) *PathPlan
```

**算法**：
1. 构建加权图：每个边权重 = 距离
2. 对每个节点，检查 `state.Nodes[nodeID]`：
   - `hasObstacle == true`：节点标记为 blocked，路径不能经过（除非起点就是该节点）
   - `guard.active && guard.ownerTeamID != self.teamID`：节点标记为 blocked
   - `obstacleResidue` 存在且未过期：经过该节点 +6 帧权重
3. Dijkstra 在非 blocked 节点子图上找最短路
4. 若无可达路径：放宽约束，允许 FORCED_PASS（计算时间税），选总成本最低的路径
5. 返回 `PathPlan`，含路径节点序列和阻挡项列表（供小分队策略决定是否清除/攻坚）

**关键约束**：
- 不硬编码节点 ID，所有信息从 `state.Nodes[]` 和 `gameMap.Adjacency` 派生
- 每帧重新规划（主车队位置、障碍、设卡都可能变化）

### 2. 轨迹预测 `strategy/prediction.go`

**目标**：给定当前路径，预测主车队到达路径上各节点的帧数（长 horizon，直到宫门/终点）。

**接口**：
```go
type ArrivalPrediction struct {
    NodeArrival map[string]int  // nodeID → 预计到达帧
    GateFrame   int             // 到达宫门帧（0 if 不在路径上）
    TerminalFrame int           // 到达终点帧
}

func (s *Strategy) predictArrivals(state *game.State, gameMap *game.GameMap, path *PathPlan) *ArrivalPrediction
```

**算法**：
1. 从当前帧 + 当前位置出发
2. 沿路径逐边累加：
   - 边移动帧数 = `ceil(distance × routeCostCoeff / speed)`，speed 由 buffs 决定
   - 到达处理点：+ `processRound`（若有探路标记则 -3，最低 2）
   - 到达障碍节点：+ FORCED_PASS 时间税（若路径规划选择了强行）
3. 输出每个节点的预计到达帧

**速度估算**：
- 基础每帧移动量：1000
- 快马：1200，短程马：1150，疾行令：1300
- 天气：暴雨命中水路 ×1.35，山雾命中山路 ×1.1
- 实际每帧移动量 = `floor(base × 1000 / weatherCoeff)`

### 3. 对手建模 `strategy/opponent.go`

**目标**：实时追踪对手状态，预测其路径，识别削弱目标。

**接口**：
```go
type OpponentModel struct {
    CurrentNodeID  string
    State          string
    Verified       bool
    Delivered      bool
    Guards         map[string]*GuardInfo  // nodeID → 对手设卡信息
    PredictedPath  []string               // 预测对手到宫门的路径
    PredictedGate  int                    // 预测对手到达宫门帧
}

type GuardInfo struct {
    NodeID        string
    Defense       int
    MaxDefense    int
    AgeRound      int
    CompleteRound int
}

func (s *Strategy) updateOpponent(state *game.State, gameMap *game.GameMap)
func (s *Strategy) weakenTargets(state *game.State) []string  // 候选削弱目标，按优先级排序
```

**更新逻辑**：
1. 从 `state.Opponent` 读取位置、状态、验核/交付标志
2. 从 `state.Nodes[]` 扫描所有 `guard.active && ownerTeamID == opponent.teamID` 的节点，记录到 `Guards`
3. 用 `gameMap.ShortestPath`（忽略障碍/设卡，因为对手的路径我们看不到那些细节）预测对手到宫门的最短路
4. 估算对手到达宫门帧（用同样的速度公式）

**削弱目标选择**：
优先级从高到低：
1. 敌方设卡在我方主路径上（阻挡我方）—— 最优先
2. 敌方设卡在关键关隘（S10 武关等 `nodeType == KEY_PASS`）
3. 敌方设卡在对手主路径上（保护对手自己）—— 干扰对手
4. 敌方设卡即将触发悬赏（`ageRound` 接近 30/60）—— 阻止悬赏

### 4. 小分队意图系统 `strategy/squad.go`

**目标**：每帧评估所有意图的触发条件，选最高优先级的执行；若无骨架意图触发，用反应式填充。

**意图定义**：
```go
type SquadIntent struct {
    Name     string
    Priority int
    Action   string  // SQUAD_SCOUT / SQUAD_CLEAR / SQUAD_REINFORCE / SQUAD_WEAKEN
    Target   string  // 目标节点
}
```

**骨架意图**（每帧重判触发条件）：

| 意图 | 优先级 | 触发条件 | 动作 |
|---|---:|---|---|
| `scout_gate` | 100 | 主车队预测 ≤15 帧到宫门 AND 宫门无活跃标记 AND 未验核 | SQUAD_SCOUT → 宫门 |
| `scout_process` | 80 | 主车队预测 ≤10 帧到处理点 AND 该点无活跃标记 | SQUAD_SCOUT → 处理点 |
| `clear_obstacle` | 60 | 主路径有障碍 AND 无更优绕路 AND 主车队距离障碍 > SQUAD_CLEAR 延迟 | SQUAD_CLEAR → 障碍节点 |

**反应式填充**（骨架未触发时评估）：

| 意图 | 优先级 | 触发条件 | 动作 |
|---|---:|---|---|
| `weaken_blocking` | 50 | 敌方设卡在我方主路径上 | SQUAD_WEAKEN → 该节点 |
| `weaken_keypass` | 40 | 敌方设卡在关键关隘 AND 我方人手充足 | SQUAD_WEAKEN → 该节点 |
| `reinforce_under_attack` | 45 | 己方设卡被攻坚（对手相邻 + 对手有 BREAK_GUARD 能力） | SQUAD_REINFORCE → 该节点 |
| `reinforce_bounty` | 35 | 己方设卡 `ageRound` 接近 30/60 且防守值低 | SQUAD_REINFORCE → 该节点 |

**决策流程**：
1. 评估所有意图的触发条件
2. 按优先级排序，取最高
3. 检查人手预算（SCOUT=1, 其他=2）
4. 检查每帧 1 个小分队动作限制
5. 检查目标有效性（节点存在、障碍/设卡仍在）
6. 若通过，提交动作，记录到 `inFlightSquads`

**在途小分队管理**：
```go
type SquadDispatch struct {
    SubmitFrame   int
    ArrivalFrame  int
    Action        string
    TargetNodeID  string
    Cost          int  // 消耗的人手
}
```
- 提交时：加入 `inFlightSquads`，扣减 `squadAvailable`
- 每帧：检查 `state.Round >= ArrivalFrame`，从 `inFlightSquads` 移除（落地结果由 inquire 反映）

**探路标记管理**：
```go
type ScoutMarker struct {
    NodeID        string
    GeneratedFrame int
    ExpiryFrame    int  // GeneratedFrame + 45
}
```
- 来源：`state.Nodes[nodeID].scouted[]` 中 `teamID == self.teamID` 的标记
- 每帧从 inquire 同步（不自行维护过期，以服务端为准）

### 5. Mock 前向模拟引擎 `cmd/mockserver/engine.go`

**目标**：mock 真实模拟游戏规则，使小分队动作的延迟、落地、效果都能被验证。

**扩展点**：

#### 障碍系统 `obstacle.go`
```go
type ObstacleState struct {
    NodeID string
    Type   string  // ROCKFALL / FLOOD / MUD / DOCK_BLOCK / PASS_CROWD / BROKEN_BRIDGE / LANDSLIDE
    Active bool
}
```
- 开局：从 `gameplay.obstacleCandidateNodeIds` 中随机选 1-2 个生成障碍
- `CLEAR` 动作：主车队在障碍节点或相邻，6 帧读条后清除，消耗 1 好果
- `SQUAD_CLEAR`：延迟到达后清除，不算 T04，创建残留税
- 残留税：非清障方 30 帧内经过 +6 帧

#### 设卡系统 `guard.go`
```go
type GuardState struct {
    NodeID         string
    OwnerTeamID    string
    Defense        int
    InitialDefense int
    MaxDefense     int
    CompleteRound  int
    AgeRound       int
    Active         bool
}
```
- `SET_GUARD`：在当前节点，4 帧读条后建立，防守值 = min(maxDefense, 2 + extraGoodFruit×2)
- `BREAK_GUARD`：攻坚，攻坚值 = goodFruit×2 + badFruit×3，≥ 防守值则清零
- `FORCED_PASS`：通行窗口 + 时间税，到达后通过
- 风化：每 30 帧（关键关隘 ≥4 时 45 帧）防守值 -1

#### 小分队调度 `squad.go`
```go
type SquadDispatch struct {
    ArrivalFrame  int
    Action        string
    TargetNodeID  string
    OwnerPlayerID int
}
```
- 提交时：计算延迟 = `min(15, max(3, ceil(D/3)))`，D = Chebyshev 距离
- 山雾区域 SQUAD_SCOUT +2（可选）
- 加入 `inFlightSquads`，扣 `squadAvailable`
- 每帧 `round == arrivalFrame` 时落地：
  - SCOUT：添加探路标记到目标节点
  - CLEAR：清除目标障碍（若仍在）
  - REINFORCE：己方设卡 +2（若仍 active 且 defense > 0）
  - WEAKEN：敌方设卡 -2（若仍 active 且 defense > 0）

#### 探路标记 `scout.go`
```go
type ScoutMarker struct {
    TeamID         string
    GeneratedFrame int
    ExpiryFrame    int  // Generated + 45
}
```
- 来源：SQUAD_SCOUT 落地 或 主车队 USE_RESOURCE(INTEL)
- 消耗：处理类动作（CLAIM_RESOURCE/CLAIM_TASK/PROCESS/VERIFY_GATE/CLEAR）完成时，若有标记则 -3 帧（最低 2），消耗最早生成的 1 个
- 过期：`round > expiryFrame` 时移除

#### 模拟对手（简化）
mock 当前只有 1 个客户端。为测试对手建模，mock 可选模拟一个静态对手（停在 S01，无设卡）。完整对手模拟（移动、设卡）不在本次范围。

## 数据流

### 客户端每帧
```
inquire → BuildState → Strategy.Decide
  → updateInternalState (sync inFlightSquads, scoutMarkers, opponent)
  → planPath (adaptive Dijkstra with obstacle/guard avoidance)
  → predictArrivals (long-horizon frame estimation)
  → decideMainCart (MOVE/PROCESS/VERIFY_GATE/DELIVER/CLEAR/...)
  → decideSquad (intent evaluation + reactive fill)
  → return [mainAction?, squadAction?]
```

### Mock 每帧
```
nextRound (round++, phase update)
processSquadLandings (apply effects of squads arriving this frame)
buildInquire (reflect current engine state)
send inquire
recv action
  → applyMainCartAction (MOVE/PROCESS/.../CLEAR/SET_GUARD/BREAK_GUARD/FORCED_PASS)
  → applySquadAction (dispatch with delay, consume 人手)
tick (advance main cart, weather guards, expire markers)
checkGameOver
```

## 适应性保证

所有模块从 `start`/`inquire` 派生决策，不硬编码：
- 节点 ID：从 `gameMap.Nodes` 和 `state.Nodes` 取
- 处理点：从 `gameMap.Node(id).ProcessType` 判断
- 障碍：从 `state.Nodes[id].HasObstacle` 判断
- 设卡：从 `state.Nodes[id].Guard` 判断
- 路径：每帧 `planPath` 重算
- 速度：从 `state.Self.Buffs` 和 `state.Weather` 计算

地图变体、天气变体、障碍分布变化都能自动适应。

## 性能考量

- 节点数 ≤15，Dijkstra 复杂度 O(V²) ≈ 225 操作，每帧 < 1ms
- 意图评估 ≤10 个意图，每个 O(1) 检查，< 1ms
- 总决策时间远低于 500ms 帧时限
