package game

import "container/heap"

// ShortestPath returns the sequence of node IDs from src to dst with minimal
// total edge distance, plus that total distance. If no path exists, returns
// nil and -1.
func (m *GameMap) ShortestPath(src, dst string) ([]string, int) {
	return m.ShortestPathExcluding(src, dst, nil)
}

// ShortestPathExcluding finds the shortest path from src to dst, skipping
// nodes in the blocked set. Returns nil path and -1 distance if no path exists.
func (m *GameMap) ShortestPathExcluding(src, dst string, blocked map[string]bool) ([]string, int) {
	if !m.HasNode(src) || !m.HasNode(dst) {
		return nil, -1
	}
	if src != dst && blocked[dst] {
		return nil, -1
	}
	if src == dst {
		return []string{src}, 0
	}
	dist := make(map[string]int)
	prev := make(map[string]string)
	dist[src] = 0
	pq := &pathPQ{}
	heap.Init(pq)
	heap.Push(pq, &pathItem{node: src, dist: 0})
	for pq.Len() > 0 {
		it := heap.Pop(pq).(*pathItem)
		if it.node == dst {
			break
		}
		if it.dist > dist[it.node] {
			continue
		}
		for _, e := range m.Adjacency[it.node] {
			if blocked[e.To] {
				continue
			}
			nd := it.dist + e.Distance
			if d, ok := dist[e.To]; !ok || nd < d {
				dist[e.To] = nd
				prev[e.To] = it.node
				heap.Push(pq, &pathItem{node: e.To, dist: nd})
			}
		}
	}
	if _, ok := dist[dst]; !ok {
		return nil, -1
	}
	var path []string
	cur := dst
	for {
		path = append([]string{cur}, path...)
		if cur == src {
			break
		}
		next, ok := prev[cur]
		if !ok {
			return nil, -1
		}
		cur = next
	}
	if len(path) == 0 || path[0] != src {
		return nil, -1
	}
	return path, dist[dst]
}

type pathItem struct {
	node string
	dist int
}

type pathPQ []*pathItem

func (p pathPQ) Len() int           { return len(p) }
func (p pathPQ) Less(i, j int) bool { return p[i].dist < p[j].dist }
func (p pathPQ) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p *pathPQ) Push(x any) {
	*p = append(*p, x.(*pathItem))
}

func (p *pathPQ) Pop() any {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[:n-1]
	return x
}
