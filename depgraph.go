// Package depgraph topologically orders a set of nodes so that every node is
// placed after the nodes it depends on. Ties are broken lexically for
// deterministic, reviewable output, and Sort reports a cycle when no valid
// ordering exists.
//
// The package is generic over any ordered node type and has no dependencies
// beyond the standard library.
package depgraph

import (
	"cmp"
	"slices"
)

// Error is the sentinel error type for the depgraph package.
type Error string

// Error returns the error message.
func (e Error) Error() string { return string(e) }

// ErrCycle is returned when the dependencies form a cycle and no valid ordering
// exists.
const ErrCycle Error = "depgraph: dependency cycle"

// Edge records that Dependent must be ordered after Dependency.
type Edge[T cmp.Ordered] struct {
	Dependent  T
	Dependency T
}

// Order is a topologically sorted sequence of nodes.
type Order[T cmp.Ordered] []T

// Sort returns the nodes in dependency order: every Dependency precedes its
// Dependent. Ties are broken lexically for deterministic, reviewable output.
// Edges referencing nodes outside the set are ignored (assumed already
// present). It returns ErrCycle when no valid ordering exists.
func Sort[T cmp.Ordered](nodes []T, edges []Edge[T]) (Order[T], error) {
	present := toSet(nodes)
	indegree, dependents := buildAdjacency(present, edges)
	order := kahn(present, indegree, dependents)
	if len(order) != len(present) {
		return nil, ErrCycle
	}
	return order, nil
}

func toSet[T cmp.Ordered](nodes []T) map[T]struct{} {
	set := make(map[T]struct{}, len(nodes))
	for _, n := range nodes {
		set[n] = struct{}{}
	}
	return set
}

func buildAdjacency[T cmp.Ordered](present map[T]struct{}, edges []Edge[T]) (map[T]int, map[T][]T) {
	indegree := make(map[T]int, len(present))
	for n := range present {
		indegree[n] = 0
	}
	dependents := make(map[T][]T)
	seen := make(map[Edge[T]]struct{}, len(edges))
	for _, e := range edges {
		if relevant(present, e) && !duplicate(seen, e) {
			indegree[e.Dependent]++
			dependents[e.Dependency] = append(dependents[e.Dependency], e.Dependent)
		}
	}
	return indegree, dependents
}

func relevant[T cmp.Ordered](present map[T]struct{}, e Edge[T]) bool {
	_, dependent := present[e.Dependent]
	_, dependency := present[e.Dependency]
	return dependent && dependency
}

func duplicate[T cmp.Ordered](seen map[Edge[T]]struct{}, e Edge[T]) bool {
	if _, ok := seen[e]; ok {
		return true
	}
	seen[e] = struct{}{}
	return false
}

func kahn[T cmp.Ordered](present map[T]struct{}, indegree map[T]int, dependents map[T][]T) Order[T] {
	ready := readyNodes(indegree)
	order := make(Order[T], 0, len(present))
	for len(ready) > 0 {
		n := ready[0]
		ready = ready[1:]
		order = append(order, n)
		ready = release(ready, indegree, dependents[n])
	}
	return order
}

func readyNodes[T cmp.Ordered](indegree map[T]int) []T {
	var ready []T
	for n, d := range indegree {
		if d == 0 {
			ready = append(ready, n)
		}
	}
	slices.Sort(ready)
	return ready
}

func release[T cmp.Ordered](ready []T, indegree map[T]int, dependents []T) []T {
	for _, d := range dependents {
		indegree[d]--
		if indegree[d] == 0 {
			i, _ := slices.BinarySearch(ready, d)
			ready = slices.Insert(ready, i, d)
		}
	}
	return ready
}
