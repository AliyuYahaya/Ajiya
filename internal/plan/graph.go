package plan

import "sort"

// Graph is the dependency graph between tickets. Edges to unknown IDs and
// self-dependencies are left out; checks report those separately.
type Graph struct {
	deps       map[string][]string // ticket -> what it depends on
	dependants map[string][]string // ticket -> what depends on it
	ids        []string            // sorted
}

// NewGraph builds the graph of the given tickets.
func NewGraph(tickets []*Ticket) *Graph {
	g := &Graph{deps: map[string][]string{}, dependants: map[string][]string{}}
	for _, t := range tickets {
		if _, dup := g.deps[t.ID]; !dup {
			g.ids = append(g.ids, t.ID)
			g.deps[t.ID] = nil
		}
	}
	SortIDs(g.ids)
	for _, t := range tickets {
		for _, d := range t.Depends {
			if _, ok := g.deps[d]; ok && d != t.ID {
				g.deps[t.ID] = append(g.deps[t.ID], d)
				g.dependants[d] = append(g.dependants[d], t.ID)
			}
		}
	}
	for _, m := range []map[string][]string{g.deps, g.dependants} {
		for id := range m {
			m[id] = uniqueSorted(m[id])
		}
	}
	return g
}

func uniqueSorted(ids []string) []string {
	SortIDs(ids)
	out := ids[:0]
	for i, id := range ids {
		if i == 0 || id != ids[i-1] {
			out = append(out, id)
		}
	}
	return out
}

// Dependants returns the tickets that depend directly on id, sorted.
func (g *Graph) Dependants(id string) []string { return g.dependants[id] }

// Downstream returns every ticket that depends on id, directly or not, sorted.
func (g *Graph) Downstream(id string) []string {
	seen := map[string]bool{}
	queue := []string{id}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range g.dependants[cur] {
			if !seen[d] && d != id {
				seen[d] = true
				queue = append(queue, d)
			}
		}
	}
	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}
	SortIDs(out)
	return out
}

// Cycles returns one loop per group of tickets that depend on each other in a
// circle, each as a path that ends where it starts: A, B, A. Deterministic.
func (g *Graph) Cycles() [][]string {
	var cycles [][]string
	for _, scc := range g.components() {
		if len(scc) < 2 {
			continue
		}
		cycles = append(cycles, g.loopWithin(scc))
	}
	sort.Slice(cycles, func(i, j int) bool { return LessID(cycles[i][0], cycles[j][0]) })
	return cycles
}

// components returns the strongly connected components (Tarjan).
func (g *Graph) components() [][]string {
	index := map[string]int{}
	low := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var out [][]string
	n := 0
	var visit func(v string)
	visit = func(v string) {
		index[v], low[v] = n, n
		n++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range g.deps[v] {
			if _, seen := index[w]; !seen {
				visit(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] == index[v] {
			var comp []string
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				comp = append(comp, w)
				if w == v {
					break
				}
			}
			SortIDs(comp)
			out = append(out, comp)
		}
	}
	for _, id := range g.ids {
		if _, seen := index[id]; !seen {
			visit(id)
		}
	}
	return out
}

// loopWithin finds the shortest loop through the lowest ID of a component.
func (g *Graph) loopWithin(comp []string) []string {
	in := map[string]bool{}
	for _, id := range comp {
		in[id] = true
	}
	start := comp[0]
	prev := map[string]string{}
	queue := []string{start}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range g.deps[cur] {
			if !in[d] {
				continue
			}
			if d == start {
				path := []string{start}
				for x := cur; x != start; x = prev[x] {
					path = append(path, x)
				}
				// path is start <- ... <- cur in reverse; flip to follow dependencies.
				loop := []string{start}
				for i := len(path) - 1; i >= 1; i-- {
					loop = append(loop, path[i])
				}
				return append(loop, start)
			}
			if _, seen := prev[d]; !seen {
				prev[d] = cur
				queue = append(queue, d)
			}
		}
	}
	return append(comp, comp[0]) // not reached for a real component
}

// LoopWith returns a loop that giving ticket id the dependencies deps would
// create, or nil if there would be none.
func LoopWith(tickets []*Ticket, id string, deps []string) []string {
	changed := make([]*Ticket, 0, len(tickets))
	for _, t := range tickets {
		if t.ID == id {
			c := *t
			c.Depends = deps
			t = &c
		}
		changed = append(changed, t)
	}
	for _, c := range NewGraph(changed).Cycles() {
		for _, x := range c {
			if x == id {
				return c
			}
		}
	}
	return nil
}
