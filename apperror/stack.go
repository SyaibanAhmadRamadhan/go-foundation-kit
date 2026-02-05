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
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
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

func PrettyStack(skipPkg string, existingStack string) string {
	const depth = 32

	pcs := make([]uintptr, depth)
	n := runtime.Callers(3, pcs) // skip runtime + wrapper awal
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
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
		if strings.Contains(f.Function, skipPkg) {
			if !more {
				break
			}
			continue
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

	fmt.Println("debug")
	return newStack + "\n"
}
