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
	methods, names map[string]*Node
	node           *Node
	pool           sync.Pool
}

func (t *Tree) naming(name string, n *Node) {
	if len(name) > 0 {
		if t.names == nil {
			t.names = make(map[string]*Node, 10)
		} else if _, ok := t.names[name]; ok {
			panic("duplicate route name: " + name)
		}
		t.names[name] = n
	}
}

func (t *Tree) getParams() []string {
	if t.pool.New == nil {
		t.pool.New = func() interface{} {
			var a []string
			if t.node != nil && len(t.node.params) > 0 {
				a = make([]string, 0, len(t.node.params))
			}
			return a
		}
	}
	return t.pool.Get().([]string)
}

func (t *Tree) putParams(params []string) {
	if t.node == nil || cap(params) >= len(t.node.params) {
		params = params[:0]
		t.pool.Put(params)
	}
}

func (t *Tree) match(method, path string) (n *Node, params []string) {
	params = t.getParams()
	f := func(s string) (ok bool) {
		if n, ok = t.methods[s]; ok {
			n, ok = n.match(path, params[:0])
			if ok && n != nil && len(n.handlers) > 0 {
				params = params[:len(n.params)]
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

func newTree(r *Router) *Tree {
	if r.above != nil {
		panic("subrouter not allowed")
	}
	t := &Tree{
		methods:  make(map[string]*Node, 10),
		handlers: copyHandlers(r.handlers),
	}
	var i int
	handlers := make([]Handler, 0, 10)
	tags := make([]Tag, 0, 10)
Loop:
	for {
		if r.above != nil && i == 0 {
			handlers = append(handlers, r.handlers...)
		}
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
			var params []string
			{
				n := 0
				for _, t := range tags {
					if t.Kind > 0 {
						n++
					}
				}
				if n > 0 {
					params = make([]string, 0, n)
				}
			}
			for i := 0; i < len(tags); i++ {
				if tags[i].Kind > 0 {
					n = n.mergeVariable(tags[i])
					params = append(params, tags[i].Name)
				} else {
					s := tags[i].Name
					for i++; i < len(tags); i++ {
						if tags[i].Kind > 0 {
							break
						} else {
							s += tags[i].Name
						}
					}
					i--
					n = n.mergeStatic(s)
				}
			}
			n.params = params
			if len(n.handlers) > 0 {
				var s string
				for _, t := range tags {
					s += t.String()
				}
				panic("duplicate route " + rr.method + " " + s)
			} else {
				n.handlers = copyHandlers(handlers)
			}
			if len(n.params) > 0 {
				if t.node == nil || len(n.params) > len(t.node.params) {
					t.node = n
				}
			}
			t.naming(rr.Name, n)
			handlers = handlers[:len(handlers)-len(rr.handlers)]
			tags = tags[:len(tags)-len(rr.tags)]
		}
		if r.above != nil {
			i = r.above.index
			handlers = handlers[:len(handlers)-len(r.handlers)-len(r.above.handlers)]
			tags = tags[:len(tags)-len(r.above.tags)]
			r = r.above.above
		} else {
			break
		}
	}
	return t
}

type Node struct {
	tag          Tag
	above        *Node
	static, back *Static
	variables    []*Node
	index        int
	handlers     []Handler
	params       []string
}

type Static struct {
	below     *Node
	prefix    string
	indexes   []byte
	constants []*Static
	back      *Static
}

func (x *Static) merge(s string) *Static {
	if x.back != nil || len(s) == 0 {
		panic(false)
	}
	for i, ok := 0, len(x.prefix) > 0; ; {
		if ok {
			ok = false
		} else {
			i = bytes.IndexByte(x.indexes, s[0])
			if i == -1 {
				break
			}
			x = x.constants[i]
			i = 1
		}
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
				prefix:    x.prefix[i:],
				constants: x.constants,
				indexes:   x.indexes,
				back:      x,
			}
			if x.below != nil {
				y.below = x.below
				y.below.back = y
				x.below = nil
			}
			x.prefix = x.prefix[:i]
			x.constants = []*Static{y}
			x.indexes = []byte{y.prefix[0]}
			if i == len(s) {
				return x
			}
			s = s[i:]
			break
		}
	}
	y := &Static{prefix: s, back: x}
	x.constants = append(x.constants, y)
	x.indexes = append(x.indexes, y.prefix[0])
	return y
}

func (n *Node) mergeStatic(s string) *Node {
	if len(s) == 0 {
		return n
	}
	var x *Static
	if n.static == nil {
		x = &Static{prefix: s}
		n.static = x
	} else {
		x = n.static.merge(s)
	}
	if x.below == nil {
		x.below = &Node{above: n, back: x}
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

func (n *Node) match(path string, params []string) (*Node, bool) {
	var x *Static
	s, i := path, -1
Loop:
	for {
		if len(s) == 0 {
			return n, true
		}
		if i == -1 {
			if x = n.static; x != nil {
				if i = len(x.prefix); i > 0 {
					if len(s) >= i && s[:i] == x.prefix {
						s = s[i:]
						if len(s) == 0 {
							return x.below, true
						}
					} else {
						x = nil
					}
				}
			}
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
