package harness

import (
	"sort"
)

// NormalizeValue recursively normalizes maps/slices for stable comparison.
func NormalizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make(map[string]interface{}, len(x))
		for _, k := range keys {
			out[k] = NormalizeValue(x[k])
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i := range x {
			out[i] = NormalizeValue(x[i])
		}
		return out
	case []map[string]interface{}:
		out := make([]map[string]interface{}, len(x))
		for i := range x {
			norm := NormalizeValue(x[i])
			if m, ok := norm.(map[string]interface{}); ok {
				out[i] = m
			} else {
				out[i] = x[i]
			}
		}
		return out
	default:
		return v
	}
}

// NormalizeMapSlice normalizes a slice of output items.
func NormalizeMapSlice(in []map[string]interface{}) []map[string]interface{} {
	out := make([]map[string]interface{}, len(in))
	for i := range in {
		norm := NormalizeValue(in[i])
		if m, ok := norm.(map[string]interface{}); ok {
			out[i] = m
		} else {
			out[i] = in[i]
		}
	}
	return out
}
