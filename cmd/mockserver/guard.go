package main

import "rubidium-lychee/internal/protocol"

// applySetGuard handles SET_GUARD action.
// Readout: 4 frames. Defense = min(maxDefense, 2 + extraGoodFruit*2).
// S15 forbidden. Process must be done if at process node.
func (e *engine) applySetGuard(act protocol.Action) {
	if act.TargetNodeID != e.cartNode {
		return
	}
	if e.cartNode == e.gameMap.TerminalID {
		return
	}
	extra := act.ExtraGoodFruit
	if extra < 0 || extra > 2 {
		return
	}
	if g, ok := e.guards[e.cartNode]; ok && g.Active {
		return
	}
	node := e.gameMap.Node(e.cartNode)
	if node != nil && node.ProcessType != "" && node.ProcessType != "VERIFY" && !e.processedHere {
		return
	}
	cost := 0
	if node != nil && node.NodeType == "KEY_PASS" {
		cost = 1
	}
	if e.cartNode == e.gameMap.GateID {
		cost = 1
	}
	if e.goodFruit < cost+extra {
		return
	}
	e.goodFruit -= cost
	e.extraGoodFruit = extra
	e.processType = "SET_GUARD"
	e.processRemain = 4
	e.cartState = protocol.StateProcessing
}

// applyBreakGuard handles BREAK_GUARD action.
// No readout; immediate resolution. attackValue = goodFruit*2 + badFruit*3.
func (e *engine) applyBreakGuard(act protocol.Action) {
	target := act.TargetNodeID
	if !e.isAdjacent(e.cartNode, target) {
		return
	}
	g, ok := e.guards[target]
	if !ok || !g.Active || g.OwnerTeamID == "RED" {
		return
	}
	goodIn := act.GoodFruit
	badIn := act.BadFruit
	if goodIn < 0 || goodIn > 2 || badIn < 0 || badIn > 2 {
		return
	}
	if e.goodFruit < goodIn {
		return
	}
	e.goodFruit -= goodIn
	attackValue := goodIn*2 + badIn*3
	if act.RushTactic == protocol.RushTacticBreakOrder {
		attackValue += 3
	}
	if attackValue >= g.Defense {
		g.Defense = 0
		g.Active = false
	} else {
		g.Defense -= attackValue
	}
}

// applyForcedPass handles FORCED_PASS action.
// Mock simplification: applies time tax directly, no 3-beat window.
func (e *engine) applyForcedPass(act protocol.Action) {
	target := act.TargetNodeID
	if !e.isAdjacent(e.cartNode, target) {
		return
	}
	hasObstacle := e.obstacles[target] != ""
	g, hasGuard := e.guards[target]
	hasEnemyGuard := hasGuard && g.Active && g.OwnerTeamID == "BLUE"
	if !hasObstacle && !hasEnemyGuard {
		return
	}
	var timeTax int
	if hasEnemyGuard {
		node := e.gameMap.Node(target)
		base := 10
		maxTax := 40
		if node != nil && node.NodeType == "KEY_PASS" {
			base = 15
			maxTax = 50
		}
		if target == e.gameMap.GateID {
			base = 12
			maxTax = 32
		}
		timeTax = min(maxTax, base+g.Defense*5)
	} else {
		timeTax = 8 // obstacle
	}
	neighbors := e.gameMap.Neighbors(e.cartNode)
	for i := range neighbors {
		if neighbors[i].To == target {
			e.cartTarget = target
			e.cartEdgeID = neighbors[i].EdgeID
			e.cartRouteType = neighbors[i].RouteType
			e.moveRemaining = timeTax
			e.moveTotalFrames = timeTax
			e.cartState = protocol.StateForcedPassing
			return
		}
	}
}

// weatherGuards reduces guard defense every 30 frames (simplified; ignores
// the 45-frame first weathering for key passes with defense >= 4).
func (e *engine) weatherGuards() {
	for _, g := range e.guards {
		if !g.Active {
			continue
		}
		g.AgeRound++
		if g.AgeRound > 0 && g.AgeRound%30 == 0 {
			g.Defense = max(0, g.Defense-1)
			if g.Defense == 0 {
				g.Active = false
			}
		}
	}
}
