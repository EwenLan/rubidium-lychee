package game

import (
	"fmt"

	"rubidium-lychee/internal/protocol"
)

// State is the per-frame game state derived from inquire. It identifies
// self vs opponent and indexes nodes by ID for quick lookup.
type State struct {
	Round          int
	Phase          string
	Self           *protocol.PlayerState
	Opponent       *protocol.PlayerState
	Nodes          map[string]*protocol.NodeState
	Weather        *protocol.Weather
	Tasks          []protocol.Task
	Bounties       []protocol.Bounty
	Contests       []protocol.Contest
	Events         []protocol.Event
	ActionResults  []protocol.ActionResult
}

// BuildState parses an inquire message into a State, splitting players into
// self and opponent by playerID.
func BuildState(inq *protocol.InquireMessage, selfPlayerID int) (*State, error) {
	if inq == nil {
		return nil, fmt.Errorf("inquire is nil")
	}
	s := &State{
		Round:         inq.Round,
		Phase:         inq.Phase,
		Nodes:         make(map[string]*protocol.NodeState),
		Weather:       &inq.Weather,
		Tasks:         inq.Tasks,
		Bounties:      inq.Bounties,
		Contests:      inq.Contests,
		Events:        inq.Events,
		ActionResults: inq.ActionResults,
	}
	for i := range inq.Players {
		p := &inq.Players[i]
		if p.PlayerID == selfPlayerID {
			s.Self = p
		} else {
			s.Opponent = p
		}
	}
	for i := range inq.Nodes {
		n := &inq.Nodes[i]
		s.Nodes[n.NodeID] = n
	}
	if s.Self == nil {
		return nil, fmt.Errorf("self player %d not found in inquire.players[]", selfPlayerID)
	}
	return s, nil
}

// SelfAt returns true if self's current node equals nodeID.
func (s *State) SelfAt(nodeID string) bool {
	return s.Self != nil && s.Self.CurrentNodeID == nodeID
}

// Node returns the per-frame node state by ID, or nil if not found.
func (s *State) Node(id string) *protocol.NodeState {
	return s.Nodes[id]
}
