// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"bytes"
	"net/http"
	"sync"
)

type Tree struct {
	handlers       []Handler
	names          []string
	methods, nodes map[string]*Node
	pool           sync.Pool
}

func (t *Tree) naming(name string, n *Node) {
	if len(name) > 0 {
		if t.nodes == nil {
			t.nodes = map[string]*Node{name: n}
		} else if _, ok := t.nodes[name]; ok {
			panic("duplicate route name: " + name)
		} else {
			t.nodes[name] = n
		}
	}
}

func (t *Tree) getParams() []string {
	if t.pool.New == nil {
		t.pool.New = func() interface{} {
			return make([]string, len(t.names))
		}
	}
	return t.pool.Get().([]string)
}

func (t *Tree) putParams(params []string) {
	t.pool.Put(params[:len(t.names)])
}

func (t *Tree) match(method, path string) (n *Node, params []string) {
	params = t.getParams()
	f := func(s string) (ok bool) {
		n, ok = t.methods[s]
		if ok {
			n, ok = n.match(path, params[:0])
			if ok && n != nil && n.handlers != nil {
				params = params[:len(n.names)]
				return true
			}
		}
		return false
	}
	if f(method) {
		return
	}
	if method != "" && f("") {
		return
	}
	n, params = nil, params[:0]
	return
}

func (t *Tree) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	n, params := t.match(r.Method, r.URL.Path)
	ctx := &Context{ResponseWriter: w, Request: r, Params: params, tree: t, node: n}
	ctx.call(ctx.index)
	t.putParams(params)
}

func (t *Tree) merge(r *Router) {
	var i int
	handlers := make([]Handler, 0, 10)
	tags := make([]Tag, 0, 10)
Loop:
	for {
		handlers = append(handlers, r.handlers...)
		for ; i < len(r.routes); i++ {
			rr := r.routes[i]
			handlers = append(handlers, rr.handlers...)
			tags = append(tags, rr.tags...)
			if rr.below != nil {
				r = rr.below
				i = 0
				continue Loop
			}
			n, ok := t.methods[rr.method]
			if !ok {
				n = new(Node)
				t.methods[rr.method] = n
			}
			var names []string
			k := 0
			for _, tag := range tags {
				if tag.Kind > 0 {
					k++
				}
			}
			if k > 0 {
				names = make([]string, k)
				k = 0
			}
			for _, tag := range tags {
				if tag.Kind > 0 {
					n = n.mergeVariable(tag)
					names[k] = tag.Name
					k++
				} else {
					n = n.mergeStatic(tag)
				}
			}
			n.names = names
			if n.handlers != nil {
				panic("duplicate route " + rr.method + " " + n.path())
			} else {
				n.handlers = copyHandlers(handlers)
			}
			t.naming(rr.Name, n)
			if len(names) > len(t.names) {
				t.names = names
			}
			handlers = handlers[:len(handlers)-len(rr.handlers)]
			tags = tags[:len(tags)-len(rr.tags)]
		}
		if r.above == nil {
			break
		} else {
			i = r.above.index
			handlers = handlers[:len(handlers)-len(r.handlers)-len(r.above.handlers)]
			tags = tags[:len(tags)-len(r.above.tags)]
			r = r.above.above
		}
	}
}

type Node struct {
	tag          Tag
	above        *Node
	index        int
	static, back *Static
	variables    []*Node
	handlers     []Handler
	names        []string
}

type Static struct {
	below     *Node
	prefix    string
	indexes   []byte
	constants []*Static
	back      *Static
}

func (x *Static) merge(s string) *Static {
	for {
		i := bytes.IndexByte(x.indexes, s[0])
		if i == -1 {
			break
		}
		x = x.constants[i]
		i = 1
		for i < len(x.prefix) && i < len(s) {
			if x.prefix[i] == s[i] {
				i++
			} else {
				break
			}
		}
		if i == len(x.prefix) {
			if i == len(s) {
				return x
			} else {
				s = s[i:]
			}
		} else {
			y := &Static{
				below:     x.below,
				prefix:    x.prefix[i:],
				constants: x.constants,
				indexes:   x.indexes,
				back:      x,
			}
			x.below = nil
			x.prefix = x.prefix[:i]
			x.constants = []*Static{y}
			x.indexes = []byte{y.prefix[0]}
			if i == len(s) {
				return x
			}
			y = &Static{prefix: s[i:], back: x}
			x.constants = append(x.constants, y)
			x.indexes = append(x.indexes, y.prefix[0])
			return y
		}
	}
	y := &Static{prefix: s, back: x}
	x.constants = append(x.constants, y)
	x.indexes = append(x.indexes, y.prefix[0])
	return y
}

func (n *Node) mergeStatic(t Tag) *Node {
	if len(t.Name) == 0 {
		return n
	}
	if n.static == nil {
		n.static = new(Static)
	}
	x := n.static.merge(t.Name)
	if x.below == nil {
		x.below = &Node{tag: t, above: n, back: x}
	}
	return x.below
}

func (n *Node) mergeVariable(t Tag) *Node {
	var i int
	for ; i < len(n.variables); i++ {
		if v := n.variables[i]; t.Same(v.tag) {
			if t.Name != v.tag.Name {
				panic("inconsistent tag name: " + t.Name)
			}
			return v
		} else if t.Kind < v.tag.Kind {
			break
		}
	}
	n.variables = append(n.variables, nil)
	for j := len(n.variables) - 1; j > i; j-- {
		n.variables[j] = n.variables[j-1]
		n.variables[j].index = j + 1
	}
	v := &Node{tag: t, above: n, index: i + 1}
	n.variables[i] = v
	return v
}

func (n *Node) path() (s string) {
	for n != nil {
		s = n.tag.String() + s
		n = n.above
	}
	return
}

func (n *Node) match(path string, params []string) (*Node, bool) {
	var x *Static
	s, i := path, -1
Loop:
	for {
		if len(s) == 0 {
			return n, true
		}
		if i == -1 {
			x = n.static
			for x != nil {
				i = bytes.IndexByte(x.indexes, s[0])
				if i >= 0 {
					x = x.constants[i]
					i = len(x.prefix)
					if len(s) >= i && s[:i] == x.prefix {
						s = s[i:]
						if len(s) == 0 {
							return x.below, true
						}
						continue
					} else {
						x = x.back
					}
				}
				break
			}
			for x != nil {
				if x.below != nil {
					n = x.below
					i = -1
					continue Loop
				}
				s = path[len(path)-len(s)-len(x.prefix):]
				x = x.back
			}
			i = 0
		}
		for ; i < len(n.variables); i++ {
			if j := n.variables[i].tag.Boundary(s); j > 0 {
				params = append(params, s[:j])
				s = s[j:]
				n = n.variables[i]
				i = -1
				continue Loop
			}
		}
		if n.above == nil {
			break
		}
		if x = n.back; x != nil {
			for {
				s = path[len(path)-len(s)-len(x.prefix):]
				x = x.back
				if x == nil {
					break
				} else if x.below != nil {
					n = x.below
					i = 0
					continue Loop
				}
			}
		} else {
			i = len(params) - 1
			s = path[len(path)-len(s)-len(params[i]):]
			params = params[:i]
		}
		i = n.index
		n = n.above
	}
	return nil, false
}
