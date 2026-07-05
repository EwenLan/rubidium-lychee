package main

import (
	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/protocol"
)

// applyClear handles CLEAR action (main cart clears obstacle).
// Cost: 1 good fruit (consumed at completion). Readout: 6 frames (scout -3, min 2).
// Must be at target or adjacent. Does NOT count as T04.
func (e *engine) applyClear(act protocol.Action) {
	target := act.TargetNodeID
	if _, ok := e.obstacles[target]; !ok {
		return
	}
	if target != e.cartNode && !e.isAdjacent(e.cartNode, target) {
		return
	}
	if e.goodFruit < 1 {
		return
	}
	processRound := 6
	if e.consumeScoutMarker(e.cartNode) {
		processRound = max(2, processRound-3)
	}
	e.processType = "CLEAR_OBSTACLE"
	e.processTarget = target
	e.processRemain = processRound
	e.cartState = protocol.StateProcessing
}

// applyMove handles MOVE action, blocked by active obstacles and enemy guards.
func (e *engine) applyMove(act protocol.Action) {
	target := act.TargetNodeID
	node := e.gameMap.Node(e.cartNode)
	if node != nil && node.ProcessType != "" && node.ProcessType != "VERIFY" && !e.processedHere {
		return // PROCESS_REQUIRED
	}
	neighbors := e.gameMap.Neighbors(e.cartNode)
	var edge *game.AdjEntry
	for i := range neighbors {
		if neighbors[i].To == target {
			edge = &neighbors[i]
			break
		}
	}
	if edge == nil {
		return // not adjacent
	}
	if _, ok := e.obstacles[target]; ok {
		return // blocked by obstacle
	}
	if g, ok := e.guards[target]; ok && g.Active && g.OwnerTeamID == "BLUE" {
		return // blocked by enemy guard
	}
	e.cartTarget = target
	e.cartEdgeID = edge.EdgeID
	e.cartRouteType = edge.RouteType
	e.moveRemaining = (edge.Distance + e.speed - 1) / e.speed
	e.moveTotalFrames = e.moveRemaining
	e.cartState = protocol.StateMoving
}

// applyProcess handles PROCESS action (fixed process nodes, not VERIFY).
func (e *engine) applyProcess(act protocol.Action) {
	node := e.gameMap.Node(e.cartNode)
	if node == nil || node.ProcessType == "" || node.ProcessType == "VERIFY" || e.processedHere {
		return
	}
	processRound := node.ProcessRound
	if e.consumeScoutMarker(e.cartNode) {
		processRound = max(2, processRound-3)
	}
	e.processType = node.ProcessType
	e.processRemain = processRound
	e.cartState = protocol.StateProcessing
}

// applyVerifyGate handles VERIFY_GATE action (only in RUSH phase).
func (e *engine) applyVerifyGate(act protocol.Action) {
	if e.cartNode != e.gameMap.GateID || e.phase != protocol.PhaseRush || e.verified {
		return
	}
	processRound := 6
	if e.consumeScoutMarker(e.cartNode) {
		processRound = max(2, processRound-3)
	}
	e.processType = "VERIFY"
	e.processRemain = processRound
	e.cartState = protocol.StateVerifying
}

// applyDeliver handles DELIVER action.
func (e *engine) applyDeliver(act protocol.Action) {
	if e.cartNode != e.gameMap.TerminalID || !e.verified || e.goodFruit <= 0 || e.freshness <= 0 {
		return
	}
	e.delivered = true
	e.cartState = protocol.StateDelivered
}
