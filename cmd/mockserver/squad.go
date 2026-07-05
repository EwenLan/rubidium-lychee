package main

import "rubidium-lychee/internal/protocol"

// applySquadAction dispatches a squad action with Chebyshev-distance delay.
// Validates target conditions and 人手 budget. RUSH phase forbids new squads.
func (e *engine) applySquadAction(act protocol.Action) {
	if e.phase == protocol.PhaseRush {
		return
	}
	cost := squadCost(act.Action)
	if e.squadAvailable < cost {
		return
	}
	if e.gameMap.Node(act.TargetNodeID) == nil {
		return
	}
	switch act.Action {
	case protocol.ActionSquadClear:
		if _, ok := e.obstacles[act.TargetNodeID]; !ok {
			return
		}
	case protocol.ActionSquadReinforce:
		g, ok := e.guards[act.TargetNodeID]
		if !ok || !g.Active || g.OwnerTeamID != "RED" || g.Defense <= 0 {
			return
		}
	case protocol.ActionSquadWeaken:
		g, ok := e.guards[act.TargetNodeID]
		if !ok || !g.Active || g.OwnerTeamID != "BLUE" || g.Defense <= 0 {
			return
		}
	case protocol.ActionSquadScout:
		// any valid node
	}
	delay := e.computeSquadDelay(act.TargetNodeID)
	e.squadAvailable -= cost
	e.inFlightSquads = append(e.inFlightSquads, squadDispatch{
		ArrivalFrame:  e.round + delay,
		Action:        act.Action,
		TargetNodeID:  act.TargetNodeID,
		OwnerPlayerID: e.playerID,
		Cost:          cost,
	})
}

func squadCost(action string) int {
	if action == protocol.ActionSquadScout {
		return 1
	}
	return 2
}

// computeSquadDelay returns the delay in frames: min(15, max(3, ceil(D/3)))
// where D is the Chebyshev distance from the main cart's current node.
// (If the cart is moving, the task book uses the edge start node; in our
// model cartNode already holds the edge start during MOVING.)
func (e *engine) computeSquadDelay(targetNodeID string) int {
	from := e.gameMap.Node(e.cartNode)
	to := e.gameMap.Node(targetNodeID)
	if from == nil || to == nil {
		return 15
	}
	dx := absDiff(from.X, to.X)
	dy := absDiff(from.Y, to.Y)
	d := dx
	if dy > d {
		d = dy
	}
	if d > 45 {
		d = 45
	}
	delay := (d + 2) / 3 // ceil(d/3)
	if delay < 3 {
		delay = 3
	}
	if delay > 15 {
		delay = 15
	}
	return delay
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

// processSquadLandings applies effects of squads arriving this frame.
func (e *engine) processSquadLandings() {
	remaining := e.inFlightSquads[:0]
	for _, sd := range e.inFlightSquads {
		if sd.ArrivalFrame > e.round {
			remaining = append(remaining, sd)
			continue
		}
		e.landSquad(sd)
	}
	e.inFlightSquads = remaining
}

func (e *engine) landSquad(sd squadDispatch) {
	switch sd.Action {
	case protocol.ActionSquadScout:
		e.scoutMarkers[sd.TargetNodeID] = append(e.scoutMarkers[sd.TargetNodeID], scoutMarker{
			GeneratedFrame: e.round,
			ExpiryFrame:    e.round + 45,
		})
	case protocol.ActionSquadClear:
		if _, ok := e.obstacles[sd.TargetNodeID]; ok {
			delete(e.obstacles, sd.TargetNodeID)
			e.obstacleResidue[sd.TargetNodeID] = e.round + 30
		}
	case protocol.ActionSquadReinforce:
		if g, ok := e.guards[sd.TargetNodeID]; ok && g.OwnerTeamID == "RED" && g.Active && g.Defense > 0 {
			g.Defense = min(g.Defense+2, g.MaxDefense)
		}
	case protocol.ActionSquadWeaken:
		if g, ok := e.guards[sd.TargetNodeID]; ok && g.OwnerTeamID == "BLUE" && g.Active && g.Defense > 0 {
			g.Defense = max(0, g.Defense-2)
			if g.Defense == 0 {
				g.Active = false
			}
		}
	}
}
