package activity

import "maps"

// MergeMetadata merges src into dst following the activity API rules:
// top-level keys are shallow-assigned from src; the "media" key uses a
// one-level map merge so partial updates can add title/singer without
// wiping the other field.
func MergeMetadata(dst, src map[string]any) {
	if dst == nil || src == nil {
		return
	}
	for k, v := range src {
		if k == "media" {
			mergeMediaField(dst, v)
			continue
		}
		dst[k] = v
	}
}

func mergeMediaField(dst map[string]any, v any) {
	srcMap, ok := v.(map[string]any)
	if !ok {
		dst["media"] = v
		return
	}
	existing, ok := dst["media"].(map[string]any)
	if !ok || existing == nil {
		dst["media"] = cloneAnyMap(srcMap)
		return
	}
	maps.Copy(existing, srcMap)
}

func cloneAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
