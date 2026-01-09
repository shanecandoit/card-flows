package graph

import (
	"fmt"
	"sort"
)

// Node is a lightweight representation of a card for graph algorithms
type Node struct {
	ID string
	X  float64
	Y  float64
}

// Arrow represents a directed connection between nodes
type Arrow struct {
	FromID string
	ToID   string
}

// TopologicalSort performs a topological sort using Kahn's algorithm.
// It accepts a slice of nodes and arrows and returns an ordered slice of IDs or an error if a cycle is detected.
func TopologicalSort(nodes []Node, arrows []Arrow) ([]string, error) {
	// Build maps
	inDegree := make(map[string]int)
	outs := make(map[string][]string)
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		inDegree[n.ID] = 0
		ids = append(ids, n.ID)
	}

	for _, a := range arrows {
		outs[a.FromID] = append(outs[a.FromID], a.ToID)
		inDegree[a.ToID]++
	}

	// Initialize queue with nodes with inDegree 0
	queue := []string{}
	for _, id := range ids {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}

	// For deterministic ordering, sort initial queue by appearance (IDs) - caller may ensure deterministic mapping
	sort.Strings(queue)

	result := []string{}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		result = append(result, u)

		for _, v := range outs[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	if len(result) != len(nodes) {
		return nil, fmt.Errorf("cycle detected in graph")
	}
	return result, nil
}
