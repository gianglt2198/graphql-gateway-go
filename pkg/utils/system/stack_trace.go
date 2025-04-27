package system

import (
	"fmt"
	"runtime"
)

func GetStackTrace() []string {
	stackTrace := []string{}
	maximumCallerDepth := 64
	minimumCallerDepth := 2

	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	f, ok := frames.Next()
	for ok {
		stackTrace = append(stackTrace, fmt.Sprintf("%s:%d", f.File, f.Line))
		f, ok = frames.Next()
	}
	return stackTrace
}
