package depgraph_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/gomatic/go-depgraph"
)

func TestSortPlacesDependencyBeforeDependent(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"v_view", "t_table"},
		[]depgraph.Edge[string]{{Dependent: "v_view", Dependency: "t_table"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"t_table", "v_view"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestSortOrdersIndependentNodesLexically(t *testing.T) {
	got, err := depgraph.Sort([]string{"c", "a", "b"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a", "b", "c"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestSortDiamondRespectsDepthThenLexical(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"d", "b", "c", "a"},
		[]depgraph.Edge[string]{
			{Dependent: "d", Dependency: "b"},
			{Dependent: "d", Dependency: "c"},
			{Dependent: "b", Dependency: "a"},
			{Dependent: "c", Dependency: "a"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a", "b", "c", "d"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestSortIgnoresEdgesToUnknownNodes(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"v_view"},
		[]depgraph.Edge[string]{{Dependent: "v_view", Dependency: "t_external"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"v_view"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestSortEmptyInputReturnsEmptyOrder(t *testing.T) {
	got, err := depgraph.Sort[string](nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Sort = %v, want empty", got)
	}
}

func TestSortSingleNodeNoEdges(t *testing.T) {
	got, err := depgraph.Sort([]string{"a"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestSortSelfDependencyIsCycle(t *testing.T) {
	_, err := depgraph.Sort(
		[]string{"a"},
		[]depgraph.Edge[string]{{Dependent: "a", Dependency: "a"}},
	)
	if !errors.Is(err, depgraph.ErrCycle) {
		t.Fatalf("Sort error = %v, want ErrCycle", err)
	}
}

func TestSortSupportsNonStringNodes(t *testing.T) {
	got, err := depgraph.Sort(
		[]int{3, 1, 2},
		[]depgraph.Edge[int]{{Dependent: 1, Dependency: 3}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []int{2, 3, 1}; !slices.Equal([]int(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

func TestErrCycleMessage(t *testing.T) {
	if got := depgraph.ErrCycle.Error(); got != "depgraph: dependency cycle" {
		t.Fatalf("ErrCycle.Error() = %q", got)
	}
}

func TestSortDetectsCycle(t *testing.T) {
	_, err := depgraph.Sort(
		[]string{"a", "b"},
		[]depgraph.Edge[string]{
			{Dependent: "a", Dependency: "b"},
			{Dependent: "b", Dependency: "a"},
		},
	)
	if !errors.Is(err, depgraph.ErrCycle) {
		t.Fatalf("Sort error = %v, want ErrCycle", err)
	}
}

// TestSortDetectsThreeNodeCycle covers a cycle of length N (here 3): a→b→c→a.
// No node ever reaches indegree zero, so the order stays empty and Sort must
// report ErrCycle rather than emit a partial ordering.
func TestSortDetectsThreeNodeCycle(t *testing.T) {
	_, err := depgraph.Sort(
		[]string{"a", "b", "c"},
		[]depgraph.Edge[string]{
			{Dependent: "a", Dependency: "b"},
			{Dependent: "b", Dependency: "c"},
			{Dependent: "c", Dependency: "a"},
		},
	)
	if !errors.Is(err, depgraph.ErrCycle) {
		t.Fatalf("Sort error = %v, want ErrCycle", err)
	}
}

// TestSortPartialCycleReportsErrCycle exercises the len(order) != len(present)
// boundary: an acyclic node "x" can be emitted, yet the b↔c cycle leaves two
// nodes unprocessed. The contract is all-or-nothing — a partial ordering is a
// cycle, asserted via the specific ErrCycle sentinel.
func TestSortPartialCycleReportsErrCycle(t *testing.T) {
	_, err := depgraph.Sort(
		[]string{"x", "b", "c"},
		[]depgraph.Edge[string]{
			{Dependent: "b", Dependency: "c"},
			{Dependent: "c", Dependency: "b"},
		},
	)
	if !errors.Is(err, depgraph.ErrCycle) {
		t.Fatalf("Sort error = %v, want ErrCycle", err)
	}
}

// TestSortDuplicateEdgesTolerated locks the deliberate contract (edge dedup was
// removed in history): a repeated identical edge raises indegree more than once
// but is decremented exactly as many times, so the node still releases. The
// result must be the same single ordering as one edge, with no false cycle and
// no duplicated node.
func TestSortDuplicateEdgesTolerated(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"b", "a"},
		[]depgraph.Edge[string]{
			{Dependent: "b", Dependency: "a"},
			{Dependent: "b", Dependency: "a"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a", "b"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

// TestSortDisconnectedComponents verifies two independent chains (x→y, a→b)
// share a single global priority queue: the lexicographically smallest ready
// node is always chosen next regardless of component, giving [a b x y].
func TestSortDisconnectedComponents(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"x", "a", "y", "b"},
		[]depgraph.Edge[string]{
			{Dependent: "y", Dependency: "x"},
			{Dependent: "b", Dependency: "a"},
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a", "b", "x", "y"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

// TestSortDependencyOverridesNaiveLexical pins the lexicographically-smallest
// topological-order contract, distinguishing it from a naive lexical sort of
// all nodes. "c" sorts after "b" alphabetically, but b depends on c, so c must
// precede b: the only valid output is [a c b], not [a b c].
func TestSortDependencyOverridesNaiveLexical(t *testing.T) {
	got, err := depgraph.Sort(
		[]string{"a", "b", "c"},
		[]depgraph.Edge[string]{{Dependent: "b", Dependency: "c"}},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := []string{"a", "c", "b"}; !slices.Equal([]string(got), want) {
		t.Fatalf("Sort = %v, want %v", got, want)
	}
}

// FuzzSort drives arbitrary node counts and edge lists derived from fuzz bytes
// and asserts the structural invariants of Sort: it never panics; a returned
// error is exactly ErrCycle; and a successful order is a permutation of the
// input in which every relevant edge places the dependency before its
// dependent.
func FuzzSort(f *testing.F) {
	f.Add([]byte(nil))
	f.Add([]byte{1})             // single node
	f.Add([]byte{2, 0, 1})       // 0 before 1
	f.Add([]byte{1, 0, 0})       // self-loop -> cycle
	f.Add([]byte{2, 0, 1, 1, 0}) // two-node cycle
	f.Add([]byte{8, 7, 0, 6, 1, 5, 2, 4, 3, 3, 4})
	f.Fuzz(func(t *testing.T, data []byte) {
		nodes, edges := decodeGraph(data)
		assertValidSort(t, nodes, edges)
	})
}

// decodeGraph maps fuzz bytes onto a node set (1..8 integer labels) and a list
// of edges (each pair of remaining bytes), so arbitrary input becomes a
// well-formed graph over present nodes.
func decodeGraph(data []byte) (depgraph.Nodes[int], []depgraph.Edge[int]) {
	if len(data) == 0 {
		return depgraph.Nodes[int]{}, nil
	}
	n := int(data[0])%8 + 1
	nodes := make(depgraph.Nodes[int], n)
	for i := range nodes {
		nodes[i] = i
	}
	var edges []depgraph.Edge[int]
	for i := 1; i+1 < len(data); i += 2 {
		edges = append(edges, depgraph.Edge[int]{
			Dependent:  int(data[i]) % n,
			Dependency: int(data[i+1]) % n,
		})
	}
	return nodes, edges
}

func assertValidSort(t *testing.T, nodes depgraph.Nodes[int], edges []depgraph.Edge[int]) {
	t.Helper()
	order, err := depgraph.Sort(nodes, edges)
	if err != nil {
		if !errors.Is(err, depgraph.ErrCycle) {
			t.Fatalf("Sort error = %v, want ErrCycle", err)
		}
		return
	}
	pos := assertPermutation(t, nodes, order)
	assertEdgesRespected(t, edges, pos)
}

func assertPermutation(t *testing.T, nodes depgraph.Nodes[int], order depgraph.Order[int]) map[int]int {
	t.Helper()
	if len(order) != len(nodes) {
		t.Fatalf("len(order) = %d, want %d", len(order), len(nodes))
	}
	pos := positionsOf(t, order)
	for _, n := range nodes {
		if _, ok := pos[n]; !ok {
			t.Fatalf("node %v missing from order %v", n, order)
		}
	}
	return pos
}

func positionsOf(t *testing.T, order depgraph.Order[int]) map[int]int {
	t.Helper()
	pos := make(map[int]int, len(order))
	for i, n := range order {
		if _, dup := pos[n]; dup {
			t.Fatalf("node %v duplicated in order %v", n, order)
		}
		pos[n] = i
	}
	return pos
}

func assertEdgesRespected(t *testing.T, edges []depgraph.Edge[int], pos map[int]int) {
	t.Helper()
	for _, e := range edges {
		if pos[e.Dependency] >= pos[e.Dependent] {
			t.Fatalf("edge %v->%v violated: dependency at %d, dependent at %d",
				e.Dependency, e.Dependent, pos[e.Dependency], pos[e.Dependent])
		}
	}
}
