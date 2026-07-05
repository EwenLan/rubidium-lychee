package game

import (
	"fmt"

	"rubidium-lychee/internal/protocol"
)

// GameMap is the static map built from start.nodes[] and start.edges[].
// It holds the adjacency graph used for pathfinding and the semantic roles
// (start/gate/terminal) used by strategy.
type GameMap struct {
	Nodes      map[string]*Node
	Edges      []*Edge
	Adjacency  map[string][]AdjEntry
	Roles      Roles
	StartID    string
	GateID     string
	TerminalID string // single terminal for this game (S15)
}

// Node is a static map node.
type Node struct {
	NodeID       string
	NodeType     string
	X, Y         int
	Start        bool
	Terminal     bool
	ProcessType  string // from start.map.gameplay.processNodes[]
	ProcessRound int
}

// Edge is a static map edge.
type Edge struct {
	EdgeID        string
	From, To      string
	RouteType     string
	Distance      int
	Bidirectional bool
}

// AdjEntry is one entry in a node's adjacency list.
type AdjEntry struct {
	To        string
	EdgeID    string
	RouteType string
	Distance  int
}

// Roles mirrors start.map.gameplay.roles.
type Roles struct {
	StartNodeID     string
	GateNodeID      string
	TerminalNodeIDs []string
	SafeZoneNodeIDs []string
}

// BuildMap constructs a GameMap from the start message.
func BuildMap(start *protocol.StartMessage) (*GameMap, error) {
	if start == nil {
		return nil, fmt.Errorf("start is nil")
	}
	m := &GameMap{
		Nodes:     make(map[string]*Node),
		Adjacency: make(map[string][]AdjEntry),
	}
	for _, n := range start.Nodes {
		m.Nodes[n.NodeID] = &Node{
			NodeID:   n.NodeID,
			NodeType: n.NodeType,
			X:        n.X,
			Y:        n.Y,
			Start:    n.Start,
			Terminal: n.Terminal,
		}
	}
	for _, pn := range start.Map.Gameplay.ProcessNodes {
		if n, ok := m.Nodes[pn.NodeID]; ok {
			n.ProcessType = pn.ProcessType
			n.ProcessRound = pn.ProcessRound
		}
	}
	for _, e := range start.Edges {
		edge := &Edge{
			EdgeID:        e.EdgeID,
			From:          e.FromNode,
			To:            e.ToNode,
			RouteType:     e.RouteType,
			Distance:      e.Distance,
			Bidirectional: e.Bidirectional,
		}
		m.Edges = append(m.Edges, edge)
		m.Adjacency[edge.From] = append(m.Adjacency[edge.From], AdjEntry{
			To: edge.To, EdgeID: edge.EdgeID, RouteType: edge.RouteType, Distance: edge.Distance,
		})
		if edge.Bidirectional && edge.To != edge.From {
			m.Adjacency[edge.To] = append(m.Adjacency[edge.To], AdjEntry{
				To: edge.From, EdgeID: edge.EdgeID, RouteType: edge.RouteType, Distance: edge.Distance,
			})
		}
	}
	m.Roles = Roles{
		StartNodeID:     start.Map.Gameplay.Roles.StartNodeID,
		GateNodeID:      start.Map.Gameplay.Roles.GateNodeID,
		TerminalNodeIDs: start.Map.Gameplay.Roles.TerminalNodeIDs,
		SafeZoneNodeIDs: start.Map.Gameplay.Roles.SafeZoneNodeIDs,
	}
	m.StartID = m.Roles.StartNodeID
	m.GateID = m.Roles.GateNodeID
	if len(m.Roles.TerminalNodeIDs) > 0 {
		m.TerminalID = m.Roles.TerminalNodeIDs[0]
	}
	return m, nil
}

// HasNode reports whether nodeID exists in the map.
func (m *GameMap) HasNode(id string) bool {
	_, ok := m.Nodes[id]
	return ok
}

// Node returns the node by ID, or nil if not found.
func (m *GameMap) Node(id string) *Node {
	return m.Nodes[id]
}

// Neighbors returns the adjacency entries for nodeID.
func (m *GameMap) Neighbors(id string) []AdjEntry {
	return m.Adjacency[id]
}
