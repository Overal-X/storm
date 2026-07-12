package storm

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
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

// contextNameRegex matches a bare context name, using the same character
// class as the ${{ name.key }} expression regex. It is used to disambiguate
// the two --context-file forms: "name:path" vs a whole-contexts "path".
var contextNameRegex = regexp.MustCompile(`^\w+$`)

// LoadContextFiles reads context from one or more YAML/JSON files into the
// nested context map. Because YAML is a superset of JSON, a single YAML
// parser handles both formats.
//
// Each entry may take one of two forms:
//   - "path": the file's top-level keys are context names, each mapping to
//     that context's key/value map (a whole-contexts file).
//   - "name:path": the file holds a single context's key/value map, assigned
//     under "name".
//
// The forms are disambiguated by splitting on the first ":" — if the left
// part is a bare word (\w+) it is treated as "name:path"; otherwise the whole
// entry is treated as a "path". Files are merged in order, with later files
// overriding earlier ones at the key level.
func LoadContextFiles(files []string) (map[string]map[string]any, error) {
	result := map[string]map[string]any{}

	for _, entry := range files {
		name, path := "", entry
		if parts := strings.SplitN(entry, ":", 2); len(parts) == 2 && contextNameRegex.MatchString(parts[0]) {
			name, path = parts[0], parts[1]
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read context file %q: %w", path, err)
		}

		if name != "" {
			var parsed map[string]any
			if err := yaml.Unmarshal(raw, &parsed); err != nil {
				return nil, fmt.Errorf("invalid context file %q for context %q: %w", path, name, err)
			}
			MergeContexts(result, map[string]map[string]any{name: parsed})
			continue
		}

		var parsed map[string]map[string]any
		if err := yaml.Unmarshal(raw, &parsed); err != nil {
			return nil, fmt.Errorf("invalid context file %q: %w", path, err)
		}
		MergeContexts(result, parsed)
	}

	return result, nil
}

// MergeContexts merges src into dst at the key level: for each context name in
// src, keys are copied into the matching context map in dst (created if
// absent), with src values overriding existing ones.
func MergeContexts(dst, src map[string]map[string]any) {
	for name, ctx := range src {
		if dst[name] == nil {
			dst[name] = map[string]any{}
		}
		for key, val := range ctx {
			dst[name][key] = val
		}
	}
}

// BuildContexts combines context files and inline context flags into a single
// context map. Files are applied first (providing the base), then inline flags
// override individual keys.
func BuildContexts(files, flags []string, format string) (map[string]map[string]any, error) {
	contexts, err := LoadContextFiles(files)
	if err != nil {
		return nil, err
	}

	inline, err := ParseContextFlags(flags, format)
	if err != nil {
		return nil, err
	}

	MergeContexts(contexts, inline)

	return contexts, nil
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
