package vfs

import (
	"strings"
)

func IsAbs(path string) bool {
	return strings.HasPrefix(path, string(PathSeparator))
}

type lazybuf struct {
	path string
	buf  []byte
	w    int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}
	return b.path[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.path) && b.path[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) string() string {
	if b.buf == nil {
		return b.path[:b.w]
	}
	return string(b.buf[:b.w])
}

func Clean(path string) string {
	if path == "" {
		return "."
	}
	rooted := IsPathSeparator(path[0])

	// Invariants:
	//	reading from path; r is index of next byte to process.
	//	writing to buf; w is index of next byte to write.
	//	dotdot is index in buf where .. must stop, either because
	//		it is the leading slash or it is a leading ../../.. prefix.
	n := len(path)
	out := lazybuf{path: path}
	r, dotdot := 0, 0
	if rooted {
		out.append(PathSeparator)
		r, dotdot = 1, 1
	}

	for r < n {
		switch {
		case IsPathSeparator(path[r]):
			// empty path element
			r++
		case path[r] == '.' && (r+1 == n || IsPathSeparator(path[r+1])):
			// . element
			r++
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || IsPathSeparator(path[r+2])):
			// .. element: remove to last separator
			r += 2
			switch {
			case out.w > dotdot:
				// can backtrack
				out.w--
				for out.w > dotdot && !IsPathSeparator(out.index(out.w)) {
					out.w--
				}
			case !rooted:
				// cannot backtrack, but not rooted, so append .. element.
				if out.w > 0 {
					out.append(PathSeparator)
				}
				out.append('.')
				out.append('.')
				dotdot = out.w
			}
		default:
			// real path element.
			// add slash if needed
			if rooted && out.w != 1 || !rooted && out.w != 0 {
				out.append(PathSeparator)
			}
			// copy element
			for ; r < n && !IsPathSeparator(path[r]); r++ {
				out.append(path[r])
			}
		}
	}

	// Turn empty string into "."
	if out.w == 0 {
		out.append('.')
	}

	return out.string()
}

func Split(path string) (dir, file string) {
	i := len(path) - 1
	for i >= 0 && !IsPathSeparator(path[i]) {
		i--
	}
	return path[:i+1], path[i+1:]
}

func Join(elem ...string) string {
	for i, e := range elem {
		if e != "" {
			return Clean(strings.Join(elem[i:], string(PathSeparator)))
		}
	}
	return ""
}

func Ext(path string) string {
	for i := len(path) - 1; i >= 0 && !IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func Base(path string) string {
	if path == "" {
		return "."
	}
	// Strip trailing slashes.
	for len(path) > 0 && IsPathSeparator(path[len(path)-1]) {
		path = path[0 : len(path)-1]
	}
	// Find the last element
	i := len(path) - 1
	for i >= 0 && !IsPathSeparator(path[i]) {
		i--
	}
	if i >= 0 {
		path = path[i+1:]
	}
	// If empty now, it had only slashes.
	if path == "" {
		return string(PathSeparator)
	}
	return path
}

func Dir(path string) string {
	i := len(path) - 1
	for i >= 0 && !IsPathSeparator(path[i]) {
		i--
	}
	dir := Clean(path[:i+1])

	return dir
}
