// Copyright (c) 2016 CHEN Xianren. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package tiny

import (
	"strings"
)

const (
	kBoolean = 'b'
	kInteger = 'i'
	kNumber  = 'n'
	kPart    = 'p'
	kRegexp  = 'r'
	kString  = 's'
)

func is0to9(b byte) bool {
	return '0' <= b && b <= '9'
}

func is1to9(b byte) bool {
	return '1' <= b && b <= '9'
}

func booleanBoundary(s string) (i int) { // ^(true|false)
	if strings.HasPrefix(s, "true") {
		i = 4
	} else if strings.HasPrefix(s, "false") {
		i = 5
	}
	return
}

func integerBoundary(s string) (i int) { // ^(0|-?[1-9][0-9]*)(?:$|[^0-9])
	switch len(s) {
	case 0:
		return 0
	case 1:
		if is0to9(s[0]) {
			return 1
		} else {
			return 0
		}
	}
	switch s[0] {
	case '0':
		if is0to9(s[1]) {
			return 0
		} else {
			return 1
		}
	case '-':
		if is1to9(s[1]) {
			i = 2
		} else {
			return 0
		}
	default:
		if is1to9(s[0]) {
			i = 1
		} else {
			return 0
		}
	}
	for i < len(s) {
		if is0to9(s[i]) {
			i++
		} else {
			break
		}
	}
	return
}

func numberBoundary(s string) (i int) { // ^(0|-?0\.[0-9]*[1-9]|-?[1-9][0-9]*(?:\.[0-9]*[1-9])?)(?:$|[^0-9])
	var j int
	for ; i < len(s); i++ {
		if is0to9(s[i]) {
			continue
		}
		switch s[i] {
		case '-':
			if i == 0 {
				continue
			}
		case '.':
			if i == 0 || (i == 1 && s[0] == '-') {
				return 0
			}
			if j == 0 {
				j = i
				continue
			}
		}
		break
	}
	if i > 0 &&
		((s[0] == '-' && (i == 1 || (s[1] == '0' && j != 2))) ||
			(s[0] == '0' && (i != 1 && j != 1)) ||
			(j > 0 && (i-1 == j || s[i-1] == '0'))) {
		return 0
	}
	return
}

func partBoundary(s string) (i int) { // ^([^/]+)
	i = strings.IndexByte(s, '/')
	if i == -1 {
		i = len(s)
	}
	return
}
