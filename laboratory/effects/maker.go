package effects

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/ast"
)

type Maker struct {
	lock sync.Mutex
	// Map from old LHComponent id to new id after instruction (typically 1)
	afterInst map[string][]string
	// Map from old LHComponent id to new id after sample
	afterSample map[string][]string
	// Map from from wtype world to ast world
	byComp map[*wtype.Liquid]*ast.UseComp
	byID   map[string][]*ast.UseComp
}

func NewMaker() *Maker {
	return &Maker{
		afterInst:   make(map[string][]string),
		afterSample: make(map[string][]string),
		byComp:      make(map[*wtype.Liquid]*ast.UseComp),
		byID:        make(map[string][]*ast.UseComp),
	}
}

func (m *Maker) UnmarshalJSON(bs []byte) error {
	m2 := NewMaker()
	*m = *m2

	return json.Unmarshal(bs, &m.afterInst)
}

func (m *Maker) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.afterInst)
}

func (a *Maker) makeComp(c *wtype.Liquid) *ast.UseComp {
	u, ok := a.byComp[c]
	if !ok {
		u = &ast.UseComp{
			Value: c,
		}
		a.byComp[c] = u
	}

	a.byID[c.ID] = append(a.byID[c.ID], u)

	return u
}

func (a *Maker) makeCommand(in *CommandInst) ast.Node {
	for _, arg := range in.Args {
		in.Command.From = append(in.Command.From, a.makeComp(arg))
	}

	out := a.makeComp(in.Result[0]) // MIS --> may need updating
	out.From = append(out.From, in.Command)
	return out
}

// resolveReuses tracks samples of the same component sharing the same id.
func (a *Maker) resolveReuses() {
	for _, uses := range a.byID {
		// HACK: assume that samples are used sequentially; remove when
		// dependencies are tracked individually

		// Make sure we don't introduce any loops
		seen := make(map[*ast.UseComp]bool)
		var us []*ast.UseComp
		for _, u := range uses {
			if seen[u] {
				continue
			}
			seen[u] = true
			us = append(us, u)
		}
		for idx, u := range us {
			if idx == 0 {
				continue
			}
			u.From = append(u.From, uses[idx-1])
		}
	}
}

// Manifest dependencies across opaque blocks
func (a *Maker) resolveUpdates(m map[string][]string) {
	for oldID, newIDs := range m {
		// Uses by id are sequential from resolveReuses, so it is sufficient to
		// match first use to last def
		for _, newID := range newIDs {
			if len(a.byID[newID]) == 0 {
				continue
			}
			if len(a.byID[oldID]) == 0 {
				continue
			}

			new := a.byID[newID][0]
			old := a.byID[oldID][len(a.byID[oldID])-1]
			new.From = append(new.From, old)
		}
	}
}

func (a *Maker) removeMultiEdges() {
	for _, use := range a.byComp {
		var filtered []ast.Node
		seen := make(map[ast.Node]bool)
		for _, from := range use.From {
			if seen[from] {
				continue
			}
			seen[from] = true
			filtered = append(filtered, from)
		}
		use.From = filtered
	}
}

func (a *Maker) UpdateAfterInst(oldID, newID string) {
	a.lock.Lock()
	defer a.lock.Unlock()
	a.afterInst[oldID] = append(a.afterInst[oldID], newID)
}

// Normalize commands into well-formed AST
func (a *Maker) MakeNodes(insts []*CommandInst) ([]ast.Node, error) {
	var nodes []ast.Node
	for _, inst := range insts {
		nodes = append(nodes, a.makeCommand(inst))
	}

	for comp := range a.byComp {
		// Contains all descendents rather then direct ones
		for _, kid := range strings.Split(comp.DaughterID, "_") {
			if comp.ID != kid {
				a.afterSample[comp.ID] = append(a.afterSample[comp.ID], kid)
			}
		}
	}

	a.resolveReuses()
	a.resolveUpdates(a.afterInst)
	a.resolveUpdates(a.afterSample)
	a.removeMultiEdges()

	return nodes, nil
}
