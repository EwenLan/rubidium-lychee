package protocol

import "encoding/json"

// StartMessage is the static match info sent by the server after both
// clients have sent registration.
type StartMessage struct {
	MatchID       string            `json:"matchId"`
	RulesVersion  string            `json:"rulesVersion,omitempty"`
	SeedHash      string            `json:"seedHash,omitempty"`
	Round         int               `json:"round"`
	Tick          int               `json:"tick,omitempty"`
	DurationRound int               `json:"durationRound"`
	Map           MapConfig         `json:"map"`
	Players       []PlayerInfo      `json:"players"`
	Nodes         []NodeDef         `json:"nodes"`
	Edges         []EdgeDef         `json:"edges"`
	RoutePaths    []json.RawMessage `json:"routePaths,omitempty"`
	Resources     []ResourceDef     `json:"resources"`
	TaskTemplates []TaskTemplate    `json:"taskTemplates"`
}

// PlayerInfo is a player entry in start.players[].
type PlayerInfo struct {
	PlayerID int    `json:"playerId"`
	Camp     int    `json:"camp"`
	TeamID   string `json:"teamId"`
	Name     string `json:"name,omitempty"`
}

// NodeDef is a static node entry in start.nodes[].
type NodeDef struct {
	NodeID   string `json:"nodeId"`
	Code     string `json:"code,omitempty"`
	Name     string `json:"name,omitempty"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Type     string `json:"type,omitempty"`
	Icon     string `json:"icon,omitempty"`
	NodeType string `json:"nodeType"`
	Start    bool   `json:"start,omitempty"`
	Terminal bool   `json:"terminal,omitempty"`
}

// EdgeDef is a static edge entry in start.edges[].
type EdgeDef struct {
	EdgeID        string `json:"edgeId"`
	FromNode      string `json:"fromNode"`
	ToNode        string `json:"toNode"`
	FromNodeID    string `json:"fromNodeId,omitempty"`
	ToNodeID      string `json:"toNodeId,omitempty"`
	RouteType     string `json:"routeType"`
	Distance      int    `json:"distance"`
	Bidirectional bool   `json:"bidirectional"`
	PathID        string `json:"pathId,omitempty"`
}

// ResourceDef is a static resource placement in start.resources[]
// and start.map.gameplay.resources[].
type ResourceDef struct {
	NodeID       string `json:"nodeId"`
	ResourceType string `json:"resourceType"`
	Count        int    `json:"count"`
	ClaimRound   int    `json:"claimRound"`
}

// TaskTemplate is a static task template in start.taskTemplates[].
type TaskTemplate struct {
	TaskTemplateID        string   `json:"taskTemplateId"`
	Name                  string   `json:"name,omitempty"`
	CandidateNodeIDs      []string `json:"candidateNodeIds,omitempty"`
	ProcessType           string   `json:"processType,omitempty"`
	ProcessRound          int      `json:"processRound"`
	RequiredFreshness     float64  `json:"requiredFreshness,omitempty"`
	RequiredResourceTypes []string `json:"requiredResourceTypes,omitempty"`
	Score                 int      `json:"score"`
}

// MapConfig is start.map.
type MapConfig struct {
	SchemaVersion string    `json:"schemaVersion,omitempty"`
	MapID         string    `json:"mapId,omitempty"`
	MapName       string    `json:"mapName,omitempty"`
	DesignVersion string    `json:"designVersion,omitempty"`
	MapConfigFile string    `json:"mapConfigFile,omitempty"`
	Data          string    `json:"data,omitempty"`
	MaxX          int       `json:"maxX,omitempty"`
	MaxY          int       `json:"maxY,omitempty"`
	Nodes         []NodeDef `json:"nodes,omitempty"`
	Edges         []EdgeDef `json:"edges,omitempty"`
	Gameplay      Gameplay  `json:"gameplay"`
}

// Gameplay is start.map.gameplay.
type Gameplay struct {
	Roles                    Roles               `json:"roles"`
	Resources                []ResourceDef       `json:"resources,omitempty"`
	ProcessNodes             []ProcessNodeDef    `json:"processNodes,omitempty"`
	TaskCandidates           map[string][]string `json:"taskCandidates,omitempty"`
	RouteTaskBuckets         map[string][]string `json:"routeTaskBuckets,omitempty"`
	ObstacleCandidateNodeIDs []string            `json:"obstacleCandidateNodeIds,omitempty"`
}

// Roles is start.map.gameplay.roles.
type Roles struct {
	StartNodeID         string   `json:"startNodeId"`
	TerminalNodeIDs     []string `json:"terminalNodeIds"`
	GateNodeID          string   `json:"gateNodeId"`
	SafeZoneNodeIDs     []string `json:"safeZoneNodeIds,omitempty"`
	ReverifyNodeID      string   `json:"reverifyNodeId,omitempty"`
	RushExcludedNodeIDs []string `json:"rushExcludedNodeIds,omitempty"`
}

// ProcessNodeDef is start.map.gameplay.processNodes[].
type ProcessNodeDef struct {
	NodeID       string `json:"nodeId"`
	ProcessType  string `json:"processType"`
	ProcessRound int    `json:"processRound"`
	CanWindow    bool   `json:"canWindow"`
}
