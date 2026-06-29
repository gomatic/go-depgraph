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
