package strategy

import (
	"rubidium-lychee/internal/game"
)

// predictArrivals estimates the arrival frame at each node along the planned
// path, based on edge distances and a simplified speed model.
//
// Speed model: calibrated for the mock (speed=10 distance units/frame).
// TODO: for the real server, derive speed from state.Self.Buffs and weather
// (base 1000, FAST_HORSE 1200, SHORT_HORSE 1150, RUSH_SPEED 1300; weather
// divides by 1000/1350/1100). The 45-frame scout marker window tolerates
// prediction error, so exact calibration is not critical for squad timing.
func (s *Strategy) predictArrivals(state *game.State, gameMap *game.GameMap, plan *PathPlan) *ArrivalPrediction {
	pred := &ArrivalPrediction{NodeArrival: make(map[string]int)}
	if plan == nil || len(plan.Nodes) == 0 {
		return pred
	}
	currentFrame := state.Round
	currentNode := plan.Nodes[0]
	pred.NodeArrival[currentNode] = currentFrame

	speed := 10 // mock default

	for i := 1; i < len(plan.Nodes); i++ {
		nextNode := plan.Nodes[i]
		edge := findEdge(gameMap, currentNode, nextNode)
		if edge == nil {
			break
		}
		moveFrames := (edge.Distance + speed - 1) / speed
		currentFrame += moveFrames

		// Add process time at intermediate process nodes (not VERIFY — that's
		// handled separately at the gate).
		node := gameMap.Node(nextNode)
		if node != nil && node.ProcessType != "" && node.ProcessType != "VERIFY" {
			processTime := node.ProcessRound
			if hasOwnScoutMarker(state, nextNode) {
				processTime = max(2, processTime-3)
			}
			currentFrame += processTime
		}

		pred.NodeArrival[nextNode] = currentFrame
		currentNode = nextNode
	}

	if f, ok := pred.NodeArrival[gameMap.GateID]; ok {
		pred.GateFrame = f
	}
	if f, ok := pred.NodeArrival[gameMap.TerminalID]; ok {
		pred.TerminalFrame = f
	}
	return pred
}

func findEdge(gameMap *game.GameMap, from, to string) *game.AdjEntry {
	neighbors := gameMap.Neighbors(from)
	for i := range neighbors {
		if neighbors[i].To == to {
			return &neighbors[i]
		}
	}
	return nil
}

func hasOwnScoutMarker(state *game.State, nodeID string) bool {
	ns := state.Nodes[nodeID]
	if ns == nil || state.Self == nil {
		return false
	}
	for _, m := range ns.Scouted {
		if m.TeamID == state.Self.TeamID && m.RemainingTriggers > 0 {
			return true
		}
	}
	return false
}
