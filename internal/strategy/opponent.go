package strategy

import (
	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/log"
)

// OpponentModel tracks the opponent's state and identifies squad targets.
// Updated every frame from inquire; all data is derived from public state
// (no hardcoded assumptions about map or opponent behavior).
type OpponentModel struct {
	CurrentNodeID string
	State         string
	Verified      bool
	Delivered     bool
	TeamID        string
	Guards        map[string]*GuardInfo // enemy guards, by nodeID
	PredictedPath []string              // predicted path to gate
	PredictedGate int                   // predicted gate arrival frame (0 if unknown)
}

// GuardInfo is a snapshot of an enemy guard.
type GuardInfo struct {
	NodeID        string
	Defense       int
	MaxDefense    int
	AgeRound      int
	CompleteRound int
}

// updateOpponent refreshes the opponent model from the current inquire.
func (s *Strategy) updateOpponent(state *game.State, gameMap *game.GameMap) {
	if state.Opponent == nil {
		return
	}
	opp := state.Opponent
	if s.opponent == nil {
		s.opponent = &OpponentModel{}
	}
	s.opponent.CurrentNodeID = opp.CurrentNodeID
	s.opponent.State = opp.State
	s.opponent.Verified = opp.Verified
	s.opponent.Delivered = opp.Delivered
	s.opponent.TeamID = opp.TeamID

	// Scan for enemy guards (owned by opponent team, active, defense > 0).
	s.opponent.Guards = make(map[string]*GuardInfo)
	for nodeID, ns := range state.Nodes {
		if ns.Guard == nil || !ns.Guard.Active || ns.Guard.Defense <= 0 {
			continue
		}
		if ns.Guard.OwnerTeamID != opp.TeamID {
			continue
		}
		s.opponent.Guards[nodeID] = &GuardInfo{
			NodeID:        nodeID,
			Defense:       ns.Guard.Defense,
			MaxDefense:    ns.Guard.MaxDefense,
			AgeRound:      ns.Guard.AgeRound,
			CompleteRound: ns.Guard.CompleteRound,
		}
	}

	// Predict opponent path to gate (shortest path, ignoring obstacles/guards
	// since we can't see the opponent's exact planned route).
	if gameMap != nil && gameMap.GateID != "" && opp.CurrentNodeID != "" {
		path, _ := gameMap.ShortestPath(opp.CurrentNodeID, gameMap.GateID)
		s.opponent.PredictedPath = path
		if path != nil {
			s.opponent.PredictedGate = s.estimateOpponentGateFrame(state, gameMap, path)
		}
	}

	if len(s.opponent.Guards) > 0 {
		log.Debugf("round %d: opponent at %s state=%s, %d enemy guards at %v",
			state.Round, opp.CurrentNodeID, opp.State, len(s.opponent.Guards), guardNodeIDs(s.opponent.Guards))
	}
}

func guardNodeIDs(guards map[string]*GuardInfo) []string {
	out := make([]string, 0, len(guards))
	for id := range guards {
		out = append(out, id)
	}
	return out
}

// estimateOpponentGateFrame estimates the opponent's arrival frame at the gate.
// Uses the same speed model as self prediction (mock-calibrated, speed=10).
func (s *Strategy) estimateOpponentGateFrame(state *game.State, gameMap *game.GameMap, path []string) int {
	if len(path) == 0 {
		return 0
	}
	frame := state.Round
	current := path[0]
	speed := 10
	for i := 1; i < len(path); i++ {
		edge := findEdge(gameMap, current, path[i])
		if edge == nil {
			return 0
		}
		frame += (edge.Distance + speed - 1) / speed
		node := gameMap.Node(path[i])
		if node != nil && node.ProcessType != "" && node.ProcessType != "VERIFY" {
			frame += node.ProcessRound
		}
		current = path[i]
	}
	return frame
}

// weakenTargets returns enemy guard node IDs sorted by priority:
// 1. Guards blocking our planned path (highest — they directly block us)
// 2. Guards on key passes (strategic chokepoints)
// 3. Other guards (lowest — may be on opponent's own path, less urgent)
func (s *Strategy) weakenTargets(state *game.State, gameMap *game.GameMap) []string {
	if s.opponent == nil || len(s.opponent.Guards) == 0 {
		return nil
	}
	var blocking, keyPass, other []string
	if s.lastPlan != nil {
		for _, b := range s.lastPlan.Blockers {
			if b.Type == "enemy_guard" {
				blocking = append(blocking, b.NodeID)
			}
		}
	}
	for nodeID := range s.opponent.Guards {
		if contains(blocking, nodeID) {
			continue
		}
		node := gameMap.Node(nodeID)
		if node != nil && node.NodeType == "KEY_PASS" {
			keyPass = append(keyPass, nodeID)
		} else {
			other = append(other, nodeID)
		}
	}
	return append(append(blocking, keyPass...), other...)
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
