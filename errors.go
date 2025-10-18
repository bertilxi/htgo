package htgo

import (
	"fmt"
	"strings"
)

type renderError struct {
	step    string
	message string
	details string
}

func (e *renderError) Error() string {
	msg := fmt.Sprintf("❌ Rendering failed at %s: %s", e.step, e.message)
	if e.details != "" {
		msg += fmt.Sprintf("\n   Details: %s", e.details)
	}
	return msg
}

func extractJSErrorContext(jsErr string) string {
	jsErr = strings.TrimSpace(jsErr)
	if strings.Contains(jsErr, "ReferenceError") {
		return "Undefined variable or function - check imports and component exports"
	}
	if strings.Contains(jsErr, "TypeError") {
		return "Type error in component - check that props match expected types"
	}
	if strings.Contains(jsErr, "SyntaxError") {
		return "Syntax error in component - check TSX/JSX syntax"
	}
	if strings.Contains(jsErr, "Cannot read") {
		return "Trying to access property on null/undefined - check prop values"
	}
	if len(jsErr) > 200 {
		return jsErr[:200] + "..."
	}
	return jsErr
}
