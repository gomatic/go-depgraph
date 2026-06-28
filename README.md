# go-depgraph

A tiny, dependency-free topological sort for Go: `depgraph.Sort` orders a set of nodes so that every node comes after the nodes it depends on. Ties are broken lexically for deterministic, reviewable output, and a dependency cycle is reported as `ErrCycle`.

## Install

```sh
go get github.com/gomatic/go-depgraph
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/gomatic/go-depgraph"
)

func main() {
	order, err := depgraph.Sort(
		[]string{"v_view", "t_table"},
		[]depgraph.Edge[string]{{Dependent: "v_view", Dependency: "t_table"}},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(order) // [t_table v_view]
}
```

`Sort` is generic over any [`cmp.Ordered`](https://pkg.go.dev/cmp#Ordered) node type. Edges referencing nodes outside the input set are ignored (assumed already present), and `errors.Is(err, depgraph.ErrCycle)` distinguishes the cycle case.

## Maintenance

The shared build config (`Makefile`, `.golangci.yaml`, `.editorconfig`, `.gitignore`, `.github/`) is owned and distributed by [`nicerobot/tools.repository`](https://github.com/nicerobot/tools.repository) — do not edit it in-tree; per-repo divergence belongs in a `Makefile.local`.
