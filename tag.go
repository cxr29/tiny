// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"regexp"
	"strings"
)

type Tag struct {
	Kind byte
	Name string
	Rule *regexp.Regexp
}

func (t Tag) Same(x Tag) bool {
	if t.Kind == x.Kind {
		switch t.Kind {
		case kRegexp:
			return t.Rule.String() == x.Rule.String()
		case 0:
			return t.Name == x.Name
		default:
			return true
		}
	}
	return false
}

func (t Tag) Boundary(s string) (i int) {
	switch t.Kind {
	case kBoolean:
		i = booleanBoundary(s)
	case kInteger:
		i = integerBoundary(s)
	case kNumber:
		i = numberBoundary(s)
	case kPart:
		i = partBoundary(s)
	case kRegexp:
		if a := t.Rule.FindStringIndex(s); a != nil && a[0] == 0 {
			i = a[1]
		}
	case kString:
		i = len(s)
	case 0:
		i = len(t.Name)
		if len(s) < i || s[:i] != t.Name {
			i = 0
		}
	}
	return
}

func (t Tag) String() (s string) {
	switch t.Kind {
	case 0:
		return t.Name
	case kRegexp:
		s = t.Name + t.Rule.String()
	case kPart:
		s = t.Name
	default:
		s = t.Name + string(delimiter[2]) + string(t.Kind)
	}
	return string(delimiter[0]) + s + string(delimiter[1])
}

func mustSplitPath(s string) (a []Tag) {
	a = splitPath(s)
	if len(a) == 0 && len(s) > 0 {
		panic(s)
	}
	return
}

const delimiter = "<>:^"

func splitPath(s string) (a []Tag) {
	if len(s) > 0 {
		for {
			i := strings.IndexByte(s, delimiter[0])
			j := strings.IndexByte(s, delimiter[1])
			if i == -1 && j == -1 {
				a = append(a, Tag{Name: s})
				return
			} else if i >= 0 && j >= 0 && i < j {
				if i > 0 {
					a = append(a, Tag{Name: s[:i]})
				}
				if t := splitTag(s[i+1 : j]); t.Kind > 0 {
					a = append(a, t)
					s = s[j+1:]
					if len(s) > 0 {
						continue
					}
					return
				}
			}
			break
		}
	}
	return nil
}

func splitTag(s string) (t Tag) {
	if strings.ContainsAny(s, delimiter[:2]) {
		return
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case delimiter[2]:
			switch s[i+1:] {
			case "b", "bool", "boolean":
				t.Kind = kBoolean
			case "i", "int", "integer":
				t.Kind = kInteger
			case "n", "num", "number":
				t.Kind = kNumber
			case "s", "str", "string":
				t.Kind = kString
			default:
				return
			}
			t.Name = s[:i]
			return
		case delimiter[3]:
			if r, err := regexp.Compile(s[i:]); err == nil {
				return Tag{kRegexp, s[:i], r}
			}
			return
		}
	}
	t.Kind = kPart
	t.Name = s
	return
}
