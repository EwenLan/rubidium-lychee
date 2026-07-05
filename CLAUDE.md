# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository purpose

This repo holds a competition client for 《一骑红尘：荔枝争运战》 ("Lychee Transport Battle"), a two-player TCP strategy game. The participant writes a client that connects to a server-provided match server, controls a cart carrying lychees from Lingnan (S01) to Chang'an Xingqing Palace (S15), and competes on final score.

The repo currently contains only specs — there is no source code yet. The two authoritative spec documents live in `docs/`:

- `docs/一骑红尘：荔枝争运战 参赛选手任务书.md` — game rules, scoring, action catalog, submission rules.
- `docs/一骑红尘：荔枝争运战 通信协议.md` — wire protocol, message schemas, error codes, field-by-field reference.

When rules and protocol disagree, treat those two files as authoritative, not this file. The `.gitignore` is the Go template, which hints at the intended implementation language, but the competition also permits C/C++ (GCC/G++ 12.3.1), Java (JDK 21), Python 3.12.9, and Node.js 16.15.0.

## Submission & runtime constraints (任务书 ch. 10)

The participant submits a single ZIP whose root directly contains an executable `start.sh` taking exactly three positional args:

```
./start.sh <playerId> <host> <port>
```

Hardcoding `playerId`, host, port, or team (RED/BLUE) is forbidden — the same ZIP may be instantiated as either side. The runtime provides Go 1.22.5, GCC/G++ 12.3.1, JDK 21, Python 3.12.9, Node.js 16.15.0; **no internet access, no `apt/yum/pip/npm install`, no writes to system directories**. Any third-party deps must ship inside the ZIP. Go/C/C++ may ship static binaries; interpreted languages must vendor their deps. There is no in-match reconnection — a dropped TCP connection is treated as offline and accumulates toward forfeit.

## Wire protocol (通信协议 ch. 1)

TCP byte-stream, framed as **5-digit decimal length prefix + UTF-8 JSON body**. Half-packet and sticky-packet handling are mandatory: buffer bytes, parse by length prefix, then UTF-8 decode, then JSON parse. Max body 99999 bytes. Field names and enum values are case-sensitive.

## Message flow (通信协议 ch. 2)

```
client: registration {playerId, playerName, version}
server: start {matchId, players[], nodes[], edges[], resources[], taskTemplates[], map.gameplay}
client: ready {matchId, round: 1, playerId}
loop each settlement frame:
  server: inquire {round: N, phase, players[], nodes[], tasks[], contests[], events[], actionResults[], ...}
  client: action {matchId, round: N, playerId, actions: [...]}   # actions: [] is the valid heartbeat
server: over {winnerPlayerId, resultType, players[].totalScore, ...}
```

`action.round` must equal the just-received `inquire.round`. Even with no action, send `actions: []` — silence counts as a missed frame and accumulates toward forfeit (10 frames = warning, 60 = retire). Don't trust `actionResults.accepted=true` alone — verify via `events[]` and the next frame's state (通信协议 ch. 10 has worked examples for MOVE, CLAIM_RESOURCE, and a rejected DELIVER).

## Game-loop mental model (任务书 ch. 1-7)

- 600 settlement frames max. Score = delivery + tasks + bounty + freshness/good-fruit quality − penalties (任务书 7.2).
- Fixed path: S01 start → … → S14 palace gate (verify) → S15 terminal (deliver). S14 only opens during 宫宴冲刺 (RUSH phase), triggered from frame 390 if a cart is near, forced at frame 450.
- Each frame, a team may submit at most: 1 main-cart action, 1 squad action, 1 window-card, 1 rush tactic (whole-game total: 1). Exceeding any bucket rejects the whole bucket and counts as one illegal action.
- Main cart is a state machine: `IDLE / MOVING / WAITING / PROCESSING / CONTESTING / RESTING / FORCED_PASSING / VERIFYING / DELIVERED / RETIRED` (任务书 3.1, 8.2). Most actions only valid in `IDLE`; `MOVING/WAITING` allow only `WAIT`, `MOVE` to current target, horse resources, or rush tactics. Always read `inquire.players[].state` before deciding.
- Windows (contests) are 3-beat simultaneous-play resolves for shared resources, tasks, dock processing, gate verify, forced-pass, and obstacle clears. Cards cost resources/guard-points; cost table in 任务书 5.4.3, win-matrix in 5.4.4. Two consecutive draws on the same object trigger a cooldown.

## Strategy gotchas that span both docs

- The map Excel is for human reading only. Always derive adjacency from `start.edges[]` / `inquire.edges[]`. `routePaths[]` and `map.layers` are display-only (通信协议 ch. 5).
- The full future weather schedule is **not** public — only the next 30-frame forecast. Read `inquire.weather` each frame. Weather affects movement permille and freshness multiplier (任务书 2.5).
- Freshness thresholds at 90/80/…/10 each convert 1 good-fruit → bad-fruit the first time crossed; at 0, all remaining good-fruit spoil. Plan `ICE_BOX` use around these thresholds (任务书 3.2).
- T04 (clear-obstacle task) is the only way to clear an obstacle *and* score — `CLEAR` and `SQUAD_CLEAR` clear the obstacle but don't score. After any clear, non-clearing teams pay a 6-frame residual tax for 30 frames (任务书 6.1.2).
- `claimedTask`/`processType` field is the server's settlement effect, not a client action — always send `CLAIM_TASK` with the `taskId` from `inquire.tasks[]` (通信协议 ch. 7 tasks field note).
