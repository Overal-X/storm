package storm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	ContextFormatJSON   = "json"
	ContextFormatBase64 = "base64"
)

var exprRegex = regexp.MustCompile(`\$\{\{\s*(\w+)\.(\w+)\s*\}\}`)

// ParseContextFlags parses CLI --context flags in "name:value" format
// into a nested map. Each flag is split on the first ":" only, so JSON
// values containing colons are handled correctly.
//
// The format parameter controls how the value portion is interpreted:
//   - "json" (default): value is raw JSON
//   - "base64": value is a base64-encoded JSON string
func ParseContextFlags(flags []string, format string) (map[string]map[string]any, error) {
	if format == "" {
		format = ContextFormatJSON
	}

	result := map[string]map[string]any{}

	for _, flag := range flags {
		parts := strings.SplitN(flag, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid context flag %q: expected name:value", flag)
		}

		name, raw := parts[0], parts[1]

		jsonBytes := []byte(raw)
		if format == ContextFormatBase64 {
			decoded, err := base64.StdEncoding.DecodeString(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid base64 for context %q: %w", name, err)
			}
			jsonBytes = decoded
		}

		var parsed map[string]any
		if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
			return nil, fmt.Errorf("invalid json for context %q: %w", name, err)
		}

		result[name] = parsed
	}

	return result, nil
}

// RenderTemplate resolves all ${{ context.key }} expressions in content
// using the provided context maps. Rendering is in-memory only.
func RenderTemplate(content string, contexts map[string]map[string]any) (string, error) {
	var renderErr error

	rendered := exprRegex.ReplaceAllStringFunc(content, func(match string) string {
		if renderErr != nil {
			return match
		}

		sub := exprRegex.FindStringSubmatch(match)
		if len(sub) != 3 {
			return match
		}

		contextName, key := sub[1], sub[2]

		ctx, ok := contexts[contextName]
		if !ok {
			renderErr = fmt.Errorf("unknown context %q in expression %s", contextName, match)
			return match
		}

		val, ok := ctx[key]
		if !ok {
			renderErr = fmt.Errorf("key %q not found in context %q", key, contextName)
			return match
		}

		switch v := val.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(v)
		default:
			b, err := json.Marshal(v)
			if err != nil {
				renderErr = err
				return match
			}
			return string(b)
		}
	})

	return rendered, renderErr
}
