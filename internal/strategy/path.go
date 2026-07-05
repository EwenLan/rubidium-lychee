package strategy

import (
	"rubidium-lychee/internal/game"
)

// planPath computes the shortest path from `from` to `to`, excluding nodes
// with active obstacles or enemy guards. If no clear path exists, falls back
// to the shortest path including blockers (returned in Blockers) so the
// strategy can decide to clear/break/force-pass.
//
// Adaptive: re-plans every frame from current state, so map variations,
// obstacle changes, and guard changes are all handled automatically.
func (s *Strategy) planPath(state *game.State, gameMap *game.GameMap, from, to string) *PathPlan {
	blocked := make(map[string]bool)
	for nodeID, ns := range state.Nodes {
		if ns.HasObstacle {
			blocked[nodeID] = true
		}
		if ns.Guard != nil && ns.Guard.Active && ns.Guard.OwnerTeamID != s.selfTeamID {
			blocked[nodeID] = true
		}
	}
	// Never block the current node (we're already there).
	delete(blocked, from)

	// Try clear path first.
	if path, dist := gameMap.ShortestPathExcluding(from, to, blocked); path != nil {
		return &PathPlan{Nodes: path, Distance: dist}
	}

	// Fallback: shortest path including blockers.
	path, dist := gameMap.ShortestPath(from, to)
	if path == nil {
		return nil
	}
	var blockers []Blocker
	for _, nodeID := range path {
		if nodeID == from {
			continue
		}
		ns := state.Nodes[nodeID]
		if ns == nil {
			continue
		}
		if ns.HasObstacle {
			blockers = append(blockers, Blocker{NodeID: nodeID, Type: "obstacle"})
		}
		if ns.Guard != nil && ns.Guard.Active && ns.Guard.OwnerTeamID != s.selfTeamID {
			blockers = append(blockers, Blocker{NodeID: nodeID, Type: "enemy_guard", Defense: ns.Guard.Defense})
		}
	}
	return &PathPlan{Nodes: path, Distance: dist, Blockers: blockers}
}
