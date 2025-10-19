package core

import (
	"fmt"
	"strings"
)

type RenderError struct {
	Step    string
	Message string
	Details string
}

func (e *RenderError) Error() string {
	msg := fmt.Sprintf("❌ Rendering failed at %s: %s", e.Step, e.Message)
	if e.Details != "" {
		msg += fmt.Sprintf("\n   Details: %s", e.Details)
	}
	return msg
}

func ExtractJSErrorContext(jsErr string) string {
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
