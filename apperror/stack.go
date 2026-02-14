package apperror

import (
	"bytes"
	"fmt"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

func Stack() string {
	var b strings.Builder
	pcs := make([]uintptr, 64)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&b, "  %s\n    %s:%d\n",
			frame.Function,
			filepath.Base(frame.File),
			frame.Line,
		)
		if !more {
			break
		}
	}
	return b.String()
}

func Pretty() string {
	raw := debug.Stack()
	lines := bytes.Split(raw, []byte("\n"))

	var out []string
	for i := range lines {
		line := string(lines[i])

		// skip runtime & empty noise
		if line == "" ||
			strings.Contains(line, "runtime/debug.Stack") ||
			strings.Contains(line, "runtime/panic.go") {
			continue
		}

		// function line
		if !strings.HasPrefix(line, "\t") && strings.Contains(line, "(") {
			out = append(out, "  "+line)
			continue
		}

		// file:line
		if strings.HasPrefix(line, "\t") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				file := filepath.Base(parts[0])
				out = append(out, "    "+file)
			}
		}
	}

	return strings.Join(out, "\n")
}

func PrettyStack(skipPkgs []string, existingStack string) string {
	const depth = 32

	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs) // skip runtime + wrapper awal
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
Outer:
	for {
		f, more := frames.Next()

		// skip runtime
		if strings.HasPrefix(f.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		// skip internal package
		for _, v := range skipPkgs {
			if strings.Contains(f.Function, v) {
				if !more {
					break Outer
				}
				continue Outer
			}
		}

		fmt.Fprintf(&b, "  %s\n    %s:%d\n",
			f.Function,
			filepath.Base(f.File),
			f.Line,
		)

		if !more {
			break
		}
	}

	newStack := strings.TrimSpace(b.String())
	oldStack := strings.TrimSpace(existingStack)
	if newStack == "" {
		return oldStack
	}
	if oldStack == "" {
		return newStack + "\n"
	}

	// dedup: jangan ulang stack yang sama
	if newStack == oldStack {
		return existingStack
	}

	return newStack + "\n"
}

func PrettyExistingStack(skipPkgs []string, existingStack string) string {
	existingStack = strings.TrimSpace(existingStack)
	if existingStack == "" {
		return ""
	}

	// normalize: split, trim, buang empty
	raw := strings.Split(existingStack, "\n")
	lines := make([]string, 0, len(raw))
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		lines = append(lines, s)
	}

	shouldSkip := func(fn string) bool {
		if fn == "" {
			return true
		}
		if strings.HasPrefix(fn, "runtime.") {
			return true
		}
		for _, p := range skipPkgs {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if strings.Contains(fn, p) {
				return true
			}
		}
		return false
	}

	var b strings.Builder

	// parse in pairs: [func, file]
	for i := 0; i < len(lines); {
		fn := lines[i]
		file := ""
		if i+1 < len(lines) {
			file = lines[i+1]
		}

		// heuristik: kalau "file" ternyata bukan file:line, tetap treat sebagai func saja
		isFileLine := file != "" && (strings.Contains(file, ".go:") || strings.Contains(file, ":"))
		if !isFileLine {
			// single line frame (fallback)
			if !shouldSkip(fn) {
				b.WriteString("  ")
				b.WriteString(fn)
				b.WriteString("\n")
			}
			i++
			continue
		}

		// normal pair
		if !shouldSkip(fn) {
			b.WriteString("  ")
			b.WriteString(fn)
			b.WriteString("\n    ")
			b.WriteString(file)
			b.WriteString("\n")
		}

		i += 2
	}

	return b.String()
}
