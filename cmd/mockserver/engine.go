package main

import (
	"encoding/json"
	"fmt"

	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/protocol"
)

// engine is a forward-simulation match engine for testing squad strategy.
// It simulates main cart movement, process, verify, deliver, clear, set_guard,
// break_guard, forced_pass, and all 4 squad actions with delay/landing,
// plus scout markers and guard weathering.
//
// NOT simulated: window contests (3-beat), tasks, bounties, resources,
// weather effects, rush tactics. FORCED_PASS applies time tax directly
// without a 3-beat window.
type engine struct {
	matchID   string
	playerID  int
	round     int
	phase     string
	rushFrame int
	speed     int
	gameMap   *game.GameMap

	// Main cart (self, RED).
	cartNode        string
	cartState       string // IDLE/MOVING/PROCESSING/VERIFYING/FORCED_PASSING/DELIVERED
	cartTarget      string
	cartEdgeID      string
	cartRouteType   string
	moveRemaining   int
	processRemain   int
	processType     string // TRANSFER/BOARD/.../VERIFY/CLEAR_OBSTACLE/SET_GUARD
	processTarget   string
	processedHere   bool
	verified        bool
	delivered       bool
	freshness       float64
	goodFruit       int
	squadAvailable  int
	guardActionPoint int
	extraGoodFruit  int // pending SET_GUARD extra investment

	// Obstacles: nodeID → type (active only).
	obstacles map[string]string

	// Guards: nodeID → state.
	guards map[string]*guardState

	// In-flight squads.
	inFlightSquads []squadDispatch

	// Scout markers: nodeID → FIFO.
	scoutMarkers map[string][]scoutMarker

	// Obstacle residue: nodeID → expiry frame.
	obstacleResidue map[string]int

	// Static opponent (BLUE) for testing opponent modeling.
	opponentNode      string
	opponentState     string
	opponentFreshness float64
	opponentGoodFruit int

	// Movement progress tracking (for edgeProgressMs/edgeTotalMs fields).
	moveTotalFrames int

	// Last frame's main-cart action (for actionResults[]).
	lastAction         string
	lastActionAccepted bool
	lastActionResult   string
	lastActionError    string

	// Events accumulated during the frame (included in next inquire).
	events []protocol.Event

	// Player names (for player state).
	playerName   string
	opponentName string
}

type guardState struct {
	OwnerTeamID    string
	Defense        int
	InitialDefense int
	MaxDefense     int
	CompleteRound  int
	AgeRound       int
	Active         bool
}

type squadDispatch struct {
	ArrivalFrame  int
	Action        string
	TargetNodeID  string
	OwnerPlayerID int
	Cost          int
}

type scoutMarker struct {
	GeneratedFrame int
	ExpiryFrame    int
}

func newEngine(matchID string, playerID int, gm *game.GameMap, rushFrame, speed int) *engine {
	return &engine{
		matchID:          matchID,
		playerID:         playerID,
		phase:            protocol.PhaseNormal,
		rushFrame:        rushFrame,
		speed:            speed,
		cartNode:         gm.StartID,
		cartState:        protocol.StateIdle,
		freshness:        100.0,
		goodFruit:        100,
		squadAvailable:   8,
		guardActionPoint: 4,
		gameMap:          gm,
		obstacles:        make(map[string]string),
		guards:           make(map[string]*guardState),
		scoutMarkers:     make(map[string][]scoutMarker),
		obstacleResidue:  make(map[string]int),
		opponentNode:      gm.StartID,
		opponentState:     protocol.StateIdle,
		opponentFreshness: 100.0,
		opponentGoodFruit: 100,
		playerName:        "rubidium-lychee",
		opponentName:      "mock-opponent",
	}
}

// generateObstacles places obstacles at the given nodes.
func (e *engine) generateObstacles(spec map[string]string) {
	for nodeID, obType := range spec {
		e.obstacles[nodeID] = obType
	}
}

// placeEnemyGuard places an enemy (BLUE) guard at nodeID with given defense.
func (e *engine) placeEnemyGuard(nodeID string, defense int) {
	node := e.gameMap.Node(nodeID)
	maxDef := 6
	if node != nil && node.NodeType == "KEY_PASS" {
		maxDef = 7
	}
	if nodeID == e.gameMap.GateID {
		maxDef = 4
	}
	e.guards[nodeID] = &guardState{
		OwnerTeamID:    "BLUE",
		Defense:        defense,
		InitialDefense: defense,
		MaxDefense:     maxDef,
		CompleteRound:  1,
		AgeRound:       0,
		Active:         true,
	}
}

func (e *engine) nextRound() {
	e.round++
	if e.phase == protocol.PhaseNormal && e.round >= e.rushFrame {
		e.phase = protocol.PhaseRush
	}
}

// preFrame processes state changes that happen at the start of a frame,
// before sending inquire: squad landings, marker expiry, guard weathering.
func (e *engine) preFrame() {
	e.processSquadLandings()
	e.expireScoutMarkers()
	e.weatherGuards()
}

// applyActions processes the client's action list, respecting per-category
// limits (1 main cart + 1 squad per frame; excess rejects the whole category).
func (e *engine) applyActions(actions []protocol.Action) {
	if e.delivered {
		return
	}
	// Reset per-frame state.
	e.events = nil
	e.lastAction = ""
	e.lastActionAccepted = false
	e.lastActionResult = ""
	e.lastActionError = ""

	var mainAct, squadAct *protocol.Action
	var mainCount, squadCount int
	for i := range actions {
		a := actions[i]
		if isSquadAction(a.Action) {
			squadCount++
			if squadCount == 1 {
				squadAct = &actions[i]
			}
		} else {
			mainCount++
			if mainCount == 1 {
				mainAct = &actions[i]
			}
		}
	}
	if mainCount > 1 {
		mainAct = nil
		e.lastAction = actions[0].Action
		e.lastActionAccepted = false
		e.lastActionResult = "INVALID_ACTION"
		e.lastActionError = "INVALID_ACTION_CONFLICT"
	}
	if squadCount > 1 {
		squadAct = nil
	}
	if mainAct != nil {
		prevState := e.cartState
		e.applyMainCartAction(*mainAct)
		e.lastAction = mainAct.Action
		e.lastActionAccepted = true
		// Detect if the action took effect (state changed or action was WAIT).
		if mainAct.Action == protocol.ActionWait ||
			e.cartState != prevState ||
			(mainAct.Action == protocol.ActionDeliver && e.delivered) {
			e.lastActionResult = "ACCEPTED"
		} else {
			e.lastActionResult = "ACTION_REJECTED"
			e.lastActionError = "ACTION_REJECTED"
		}
	}
	if squadAct != nil {
		e.applySquadAction(*squadAct)
	}
}

func isSquadAction(action string) bool {
	switch action {
	case protocol.ActionSquadScout, protocol.ActionSquadClear,
		protocol.ActionSquadReinforce, protocol.ActionSquadWeaken:
		return true
	}
	return false
}

// applyMainCartAction dispatches main cart actions.
func (e *engine) applyMainCartAction(act protocol.Action) {
	if e.cartState != protocol.StateIdle && act.Action != protocol.ActionWait {
		return
	}
	switch act.Action {
	case protocol.ActionWait:
		// no-op
	case protocol.ActionMove:
		e.applyMove(act)
	case protocol.ActionProcess:
		e.applyProcess(act)
	case protocol.ActionVerifyGate:
		e.applyVerifyGate(act)
	case protocol.ActionDeliver:
		e.applyDeliver(act)
	case protocol.ActionClear:
		e.applyClear(act)
	case protocol.ActionSetGuard:
		e.applySetGuard(act)
	case protocol.ActionBreakGuard:
		e.applyBreakGuard(act)
	case protocol.ActionForcedPass:
		e.applyForcedPass(act)
	}
}

// tick advances the main cart state by one frame.
func (e *engine) tick() {
	// Freshness decay (simplified: based on cart state + route type).
	if !e.delivered && e.cartState != protocol.StateDelivered {
		decay := 0.05
		if e.cartState == protocol.StateMoving || e.cartState == protocol.StateForcedPassing {
			switch e.cartRouteType {
			case "ROAD":
				decay = 0.055
			case "WATER":
				decay = 0.045
			case "MOUNTAIN":
				decay = 0.07
			case "BRANCH":
				decay = 0.065
			}
		}
		prevFresh := e.freshness
		e.freshness -= decay
		if e.freshness < 0 {
			e.freshness = 0
		}
		if e.freshness < prevFresh {
			e.addEvent("FRESHNESS_DROP", map[string]any{
				"playerId": e.playerID,
				"before":   prevFresh,
				"after":    e.freshness,
				"loss":     decay,
			})
		}
	}

	switch e.cartState {
	case protocol.StateMoving:
		e.moveRemaining--
		if e.moveRemaining <= 0 {
			arrivedAt := e.cartTarget
			e.addEvent("NODE_ENTER", map[string]any{
				"playerId": e.playerID,
				"nodeId":   arrivedAt,
			})
			e.cartNode = e.cartTarget
			e.cartState = protocol.StateIdle
			e.cartTarget = ""
			e.cartEdgeID = ""
			e.cartRouteType = ""
			e.processedHere = false
		} else if e.moveTotalFrames > 0 {
			elapsed := e.moveTotalFrames - e.moveRemaining
			totalMs := e.edgeTotalMs()
			e.addEvent("MOVE_PROGRESS", map[string]any{
				"playerId":    e.playerID,
				"fromNodeId":  e.cartNode,
				"toNodeId":    e.cartTarget,
				"routeEdgeId": e.cartEdgeID,
				"progress":    float64(elapsed) / float64(e.moveTotalFrames),
				"edgeProgressMs": elapsed * 1000,
				"edgeTotalMs":  totalMs,
			})
		}
	case protocol.StateProcessing:
		e.processRemain--
		if e.processRemain <= 0 {
			e.completeProcess()
		}
	case protocol.StateVerifying:
		e.processRemain--
		if e.processRemain <= 0 {
			e.verified = true
			e.cartState = protocol.StateIdle
			e.processType = ""
			e.addEvent("VERIFY_GATE_COMPLETE", map[string]any{
				"playerId": e.playerID,
				"nodeId":   e.gameMap.GateID,
			})
		}
	case protocol.StateForcedPassing:
		e.moveRemaining--
		if e.moveRemaining <= 0 {
			arrivedAt := e.cartTarget
			e.addEvent("NODE_ENTER", map[string]any{
				"playerId": e.playerID,
				"nodeId":   arrivedAt,
			})
			e.cartNode = e.cartTarget
			e.cartState = protocol.StateIdle
			e.cartTarget = ""
			e.cartEdgeID = ""
			e.cartRouteType = ""
			e.processedHere = false
		}
	}
}

// completeProcess handles completion of PROCESS / CLEAR_OBSTACLE / SET_GUARD.
func (e *engine) completeProcess() {
	switch e.processType {
	case "CLEAR_OBSTACLE":
		delete(e.obstacles, e.processTarget)
		e.obstacleResidue[e.processTarget] = e.round + 30
		e.goodFruit--
	case "SET_GUARD":
		node := e.gameMap.Node(e.cartNode)
		maxDef := 6
		if node != nil && node.NodeType == "KEY_PASS" {
			maxDef = 7
		}
		if e.cartNode == e.gameMap.GateID {
			maxDef = 4
		}
		defense := 2 + e.extraGoodFruit*2
		if defense > maxDef {
			defense = maxDef
		}
		e.guards[e.cartNode] = &guardState{
			OwnerTeamID:    "RED",
			Defense:        defense,
			InitialDefense: defense,
			MaxDefense:     maxDef,
			CompleteRound:  e.round,
			AgeRound:       0,
			Active:         true,
		}
		e.goodFruit -= e.extraGoodFruit
		e.extraGoodFruit = 0
	}
	e.processType = ""
	e.processTarget = ""
	e.processedHere = true
	e.cartState = protocol.StateIdle
}

func (e *engine) isOver(maxRounds int) bool {
	return e.delivered || e.round >= maxRounds
}

// isAdjacent reports whether b is a neighbor of a.
func (e *engine) isAdjacent(a, b string) bool {
	for _, n := range e.gameMap.Neighbors(a) {
		if n.To == b {
			return true
		}
	}
	return false
}

// buildInquire constructs the inquire from current engine state, matching
// the real server's message structure (all fields populated).
func (e *engine) buildInquire() protocol.InquireMessage {
	return protocol.InquireMessage{
		MatchID:       e.matchID,
		RulesVersion:  "4.1",
		Round:         e.round,
		Tick:          e.round - 1,
		Phase:         e.phase,
		Players:       []protocol.PlayerState{e.buildSelfPlayerState(), e.buildOpponentPlayerState()},
		Nodes:         e.buildNodeStates(),
		Edges:         basicMapEdges(),
		Weather:       protocol.Weather{Active: []protocol.WeatherEvent{}, Forecast: []protocol.WeatherEvent{}},
		Tasks:         []protocol.Task{},
		Bounties:      []protocol.Bounty{},
		Contests:      []protocol.Contest{},
		Events:        e.events,
		ActionResults: e.buildActionResults(),
		ScorePreview:  map[string]int{"RED": e.selfScore(), "BLUE": 0},
	}
}

// buildSelfPlayerState returns the RED player state with all fields populated.
func (e *engine) buildSelfPlayerState() protocol.PlayerState {
	ps := protocol.PlayerState{
		PlayerID:            e.playerID,
		Camp:                0,
		TeamID:              "RED",
		PlayerName:          e.playerName,
		Online:              true,
		State:               e.cartState,
		CurrentNodeID:       e.cartNode,
		Freshness:           e.freshness,
		GoodFruit:           e.goodFruit,
		FrozenGoodFruit:     0,
		BadFruit:            0,
		SquadAvailable:      e.squadAvailable,
		SquadInFlight:       len(e.inFlightSquads),
		GuardActionPoint:    e.guardActionPoint,
		Verified:            e.verified,
		Delivered:           e.delivered,
		Retired:             false,
		RetiredRound:        0,
		MissingActionRounds: 0,
		IllegalActionCount:  0,
		PenaltyScore:        0,
		BreakOrderReady:     false,
		RushTacticUsedCount: 0,
		Buffs:               []protocol.Buff{},
		CurrentProcess:      nil,
		Resources:           map[string]int{},
		TotalScore:          e.selfScore(),
		TaskScore:           0,
		BountyScore:         0,
		ScoreDetail:         protocol.ScoreDetail{},
		MoveDirection:       "NONE",
		MoveProgress:        0,
		MoveProgressRound:   0,
		CurrentEdgeCost:     0,
		EdgeProgressMs:      0,
		EdgeProgressPermille: 0,
		EdgeTotalMs:         0,
	}
	if e.cartState == protocol.StateMoving || e.cartState == protocol.StateForcedPassing {
		ps.NextNodeID = e.cartTarget
		ps.RouteEdgeID = e.cartEdgeID
		ps.RouteType = e.cartRouteType
		ps.MoveDirection = "FORWARD"
		if e.moveTotalFrames > 0 {
			elapsed := e.moveTotalFrames - e.moveRemaining
			totalMs := e.edgeTotalMs()
			ps.EdgeTotalMs = totalMs
			ps.EdgeProgressMs = elapsed * 1000
			if totalMs > 0 {
				ps.EdgeProgressPermille = ps.EdgeProgressMs * 1000 / totalMs
				ps.MoveProgress = float64(ps.EdgeProgressMs) / float64(totalMs)
			}
			ps.MoveProgressRound = elapsed
		}
	}
	if e.cartState == protocol.StateProcessing || e.cartState == protocol.StateVerifying {
		ps.CurrentProcess = &protocol.CurrentProcess{
			Action:    e.processType,
			TargetNodeID: e.processTarget,
			Type:      e.processType,
			StartedRound: e.round,
			TotalRound:   e.processRemain,
			RemainRound:  e.processRemain,
			RemainingRound: e.processRemain,
		}
	}
	return ps
}

// buildOpponentPlayerState returns the BLUE player state.
func (e *engine) buildOpponentPlayerState() protocol.PlayerState {
	return protocol.PlayerState{
		PlayerID:            9999,
		Camp:                1,
		TeamID:              "BLUE",
		PlayerName:          e.opponentName,
		Online:              true,
		State:               e.opponentState,
		CurrentNodeID:       e.opponentNode,
		Freshness:           e.opponentFreshness,
		GoodFruit:           e.opponentGoodFruit,
		SquadAvailable:      8,
		GuardActionPoint:    4,
		Verified:            false,
		Delivered:           false,
		Buffs:               []protocol.Buff{},
		Resources:           map[string]int{},
		ScoreDetail:         protocol.ScoreDetail{},
		MoveDirection:       "NONE",
	}
}

// buildActionResults returns the action result for the last frame's main-cart action.
func (e *engine) buildActionResults() []protocol.ActionResult {
	if e.lastAction == "" {
		return nil
	}
	ar := protocol.ActionResult{
		Round:     e.round,
		PlayerID:  e.playerID,
		Action:    e.lastAction,
		Accepted:  e.lastActionAccepted,
		Result:    e.lastActionResult,
		ErrorCode: e.lastActionError,
	}
	if e.lastActionError != "" {
		ar.Message = "ACTION_REJECTED"
	}
	return []protocol.ActionResult{ar}
}

// buildNodeStates assembles per-node state from the static base + engine dynamic state.
func (e *engine) buildNodeStates() []protocol.NodeState {
	base := basicMapNodeStates()
	out := make([]protocol.NodeState, len(base))
	for i, n := range base {
		ns := n // copy static fields (Name, X, Y, NodeType, processType, resourceStock, etc.)
		// Dynamic: obstacle.
		if obType, ok := e.obstacles[n.NodeID]; ok {
			ns.HasObstacle = true
			ns.ObstacleType = obType
		}
		// Dynamic: guard.
		if g, ok := e.guards[n.NodeID]; ok && g.Active {
			ns.Guard = &protocol.GuardState{
				OwnerTeamID:    g.OwnerTeamID,
				Defense:        g.Defense,
				InitialDefense: g.InitialDefense,
				MaxDefense:     g.MaxDefense,
				CompleteRound:  g.CompleteRound,
				AgeRound:       g.AgeRound,
				Active:         g.Active,
			}
		}
		// Dynamic: scout markers.
		if markers, ok := e.scoutMarkers[n.NodeID]; ok {
			for _, m := range markers {
				ns.Scouted = append(ns.Scouted, protocol.ScoutMarker{
					TeamID:             "RED",
					RemainRound:        m.ExpiryFrame - e.round,
					ProcessReduceRound: 3,
					RemainingTriggers:  1,
				})
			}
		}
		// Dynamic: obstacle residue.
		if expFrame, ok := e.obstacleResidue[n.NodeID]; ok && expFrame > e.round {
			ns.ObstacleResidue = &protocol.ObstacleResidue{
				ClearedByTeamID: "RED",
				ClearRound:      expFrame - 30,
				UntilRound:      expFrame,
				RemainRound:     expFrame - e.round,
				TaxRound:        6,
			}
		}
		out[i] = ns
	}
	return out
}

// edgeTotalMs returns the total ms for the current edge (distance × routeCoeff).
func (e *engine) edgeTotalMs() int {
	for _, a := range e.gameMap.Neighbors(e.cartNode) {
		if a.To == e.cartTarget {
			return a.Distance * routeCostCoeff(a.RouteType)
		}
	}
	return 0
}

// routeCostCoeff returns the per-distance cost coefficient by route type.
func routeCostCoeff(routeType string) int {
	switch routeType {
	case "ROAD":
		return 1380
	case "WATER":
		return 1250
	case "MOUNTAIN":
		return 1780
	case "BRANCH":
		return 1550
	}
	return 1380
}

// selfScore returns a placeholder score (delivery base only).
func (e *engine) selfScore() int {
	if e.delivered {
		return 240
	}
	return 0
}

// addEvent appends an event to the current frame's event list.
func (e *engine) addEvent(eventType string, payload map[string]any) {
	payloadBytes, _ := json.Marshal(payload)
	e.events = append(e.events, protocol.Event{
		EventID: fmt.Sprintf("EV_%03d_%03d", e.round, len(e.events)+1),
		Type:    eventType,
		Round:   e.round,
		Payload: payloadBytes,
	})
}
