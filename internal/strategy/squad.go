package strategy

import (
	"sort"

	"rubidium-lychee/internal/game"
	"rubidium-lychee/internal/log"
	"rubidium-lychee/internal/protocol"
)

// decideSquad evaluates squad intents and returns the highest-priority
// actionable squad action, or nil if none. Called every frame (separate
// category from main cart — both can be submitted in the same frame).
//
// No new squads in RUSH phase (per 任务书 6.5).
func (s *Strategy) decideSquad(state *game.State, gameMap *game.GameMap) *protocol.Action {
	if state.Phase == protocol.PhaseRush {
		return nil
	}
	if state.Self == nil || state.Self.SquadAvailable < 1 {
		return nil
	}

	intents := s.evaluateIntents(state, gameMap)
	// Sort by priority (highest first).
	sort.Slice(intents, func(i, j int) bool {
		return intents[i].Priority > intents[j].Priority
	})

	for _, intent := range intents {
		cost := squadCost(intent.Action)
		if state.Self.SquadAvailable < cost {
			continue
		}
		if s.hasInFlight(intent.Action, intent.Target) {
			continue
		}
		delay := estimateSquadDelay(state, gameMap, intent.Target)
		s.inFlightSquads = append(s.inFlightSquads, SquadDispatch{
			SubmitFrame:  state.Round,
			ArrivalFrame: state.Round + delay,
			Action:       intent.Action,
			TargetNodeID: intent.Target,
			Cost:         cost,
		})
		log.Infof("round %d: squad %s → %s (P%d, delay=%d, cost=%d, remaining=%d→%d)",
			state.Round, intent.Action, intent.Target, intent.Priority, delay, cost,
			state.Self.SquadAvailable, state.Self.SquadAvailable-cost)
		act := &protocol.Action{Action: intent.Action, TargetNodeID: intent.Target}
		return act
	}
	return nil
}

// evaluateIntents returns all triggered squad intents, unsorted.
// Each intent is re-evaluated every frame from current state (adaptive).
func (s *Strategy) evaluateIntents(state *game.State, gameMap *game.GameMap) []SquadIntent {
	var intents []SquadIntent
	self := state.Self

	// --- Skeleton intents (proactive, timing-critical) ---

	// scout_gate: main cart within 15 frames of gate, no marker, not verified.
	gateID := s.gateID
	if gateID == "" {
		gateID = gameMap.GateID
	}
	if !self.Verified && s.lastPred != nil && s.lastPred.GateFrame > 0 && gateID != "" {
		framesToGate := s.lastPred.GateFrame - state.Round
		if framesToGate > 0 && framesToGate <= 15 {
			if !hasOwnScoutMarker(state, gateID) && !s.hasInFlight(protocol.ActionSquadScout, gateID) {
				intents = append(intents, SquadIntent{
					Name: "scout_gate", Priority: 100,
					Action: protocol.ActionSquadScout, Target: gateID,
				})
			}
		}
	}

	// scout_process: main cart within 10 frames of a process node on path.
	if s.lastPlan != nil && s.lastPred != nil {
		for _, nodeID := range s.lastPlan.Nodes {
			node := gameMap.Node(nodeID)
			if node == nil || node.ProcessType == "" || node.ProcessType == "VERIFY" {
				continue
			}
			arrival, ok := s.lastPred.NodeArrival[nodeID]
			if !ok {
				continue
			}
			framesToArrival := arrival - state.Round
			if framesToArrival > 0 && framesToArrival <= 10 {
				if !hasOwnScoutMarker(state, nodeID) && !s.hasInFlight(protocol.ActionSquadScout, nodeID) {
					intents = append(intents, SquadIntent{
						Name: "scout_process", Priority: 80,
						Action: protocol.ActionSquadScout, Target: nodeID,
					})
				}
			}
		}
	}

	// clear_obstacle: obstacle blocking our path (no clear detour found).
	if s.lastPlan != nil {
		for _, b := range s.lastPlan.Blockers {
			if b.Type == "obstacle" && !s.hasInFlight(protocol.ActionSquadClear, b.NodeID) {
				intents = append(intents, SquadIntent{
					Name: "clear_obstacle", Priority: 60,
					Action: protocol.ActionSquadClear, Target: b.NodeID,
				})
			}
		}
	}

	// --- Reactive fill intents ---

	// weaken_blocking: enemy guard on our path.
	if s.lastPlan != nil {
		for _, b := range s.lastPlan.Blockers {
			if b.Type == "enemy_guard" && !s.hasInFlight(protocol.ActionSquadWeaken, b.NodeID) {
				intents = append(intents, SquadIntent{
					Name: "weaken_blocking", Priority: 50,
					Action: protocol.ActionSquadWeaken, Target: b.NodeID,
				})
			}
		}
	}

	// weaken_keypass: enemy guard on a key pass not already on our path.
	if s.opponent != nil {
		for nodeID := range s.opponent.Guards {
			onPath := false
			if s.lastPlan != nil {
				for _, b := range s.lastPlan.Blockers {
					if b.NodeID == nodeID {
						onPath = true
						break
					}
				}
			}
			if onPath {
				continue
			}
			node := gameMap.Node(nodeID)
			if node != nil && node.NodeType == "KEY_PASS" && !s.hasInFlight(protocol.ActionSquadWeaken, nodeID) {
				intents = append(intents, SquadIntent{
					Name: "weaken_keypass", Priority: 40,
					Action: protocol.ActionSquadWeaken, Target: nodeID,
				})
			}
		}
	}

	return intents
}

// squadCost returns the 人手 cost of a squad action.
func squadCost(action string) int {
	if action == protocol.ActionSquadScout {
		return 1
	}
	return 2
}

// estimateSquadDelay computes the Chebyshev-distance delay for a squad action:
// min(15, max(3, ceil(D/3))), D = max(|dx|, |dy|) from main cart's current node.
func estimateSquadDelay(state *game.State, gameMap *game.GameMap, targetNodeID string) int {
	from := gameMap.Node(state.Self.CurrentNodeID)
	to := gameMap.Node(targetNodeID)
	if from == nil || to == nil {
		return 15
	}
	dx := absInt(from.X - to.X)
	dy := absInt(from.Y - to.Y)
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

func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// hasInFlight reports whether a squad action is already in-flight to target.
func (s *Strategy) hasInFlight(action, target string) bool {
	for _, sd := range s.inFlightSquads {
		if sd.Action == action && sd.TargetNodeID == target {
			return true
		}
	}
	return false
}

// updateInFlight removes squads that should have landed by now.
func (s *Strategy) updateInFlight(currentRound int) {
	remaining := s.inFlightSquads[:0]
	for _, sd := range s.inFlightSquads {
		if sd.ArrivalFrame > currentRound {
			remaining = append(remaining, sd)
		}
	}
	s.inFlightSquads = remaining
}
