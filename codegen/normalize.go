package codegen

import (
	"github.com/antha-lang/antha/graph"
	"sort"
)

type counts struct {
	Counts []int
	Nodes  []AstNode
}

func (a *counts) Len() int {
	return len(a.Counts)
}

func (a *counts) Less(i, j int) bool {
	return a.Counts[i] < a.Counts[j]
}

func (a *counts) Swap(i, j int) {
	a.Counts[i], a.Counts[j] = a.Counts[j], a.Counts[i]
	a.Nodes[i], a.Nodes[j] = a.Nodes[j], a.Nodes[i]
}

// Prune out root expressions that appear as subexpressions of other
// expressions.
func pruneRoots(g *AstGraph, root *BundleExpr) (*BundleExpr, error) {
	// For an expression to contain another, it must contain at least as many
	// nodes. Process roots in descending size order.
	var c counts
	for _, n := range root.From {
		count := 0
		graph.Visit(graph.VisitOpt{
			Root:  n,
			Graph: g,
			Visitor: func(graph.Node) error {
				count += 1
				return nil
			},
		})
		c.Counts = append(c.Counts, count)
		c.Nodes = append(c.Nodes, n)
	}

	sort.Sort(sort.Reverse(&c))

	r := &BundleExpr{}
	seen := make(map[graph.Node]bool)
	for _, n := range c.Nodes {
		if seen[n] {
			continue
		}

		graph.Visit(graph.VisitOpt{
			Root:  n,
			Graph: g,
			Visitor: func(n graph.Node) error {
				seen[n] = true
				return nil
			},
		})
		r.From = append(r.From, n)
	}

	return r, nil
}

// Cleanup client input
func normalize(root AstNode) (AstNode, error) {
	g := toGraph(toGraphOpt{
		Root: root,
	})
	if r, ok := root.(*BundleExpr); ok {
		return pruneRoots(g, r)
	} else if err := graph.IsDag(g); err != nil {
		return nil, err
	} else {
		return &BundleExpr{From: []AstNode{root}}, nil
	}
}