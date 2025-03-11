package configs

import "strings"

// unflattenMap converts a flat map with dot-separated keys into a nested map.
func unflattenMap(flat map[string]any) map[string]any {
	nested := make(map[string]any)
	for key, value := range flat {
		parts := strings.Split(key, ".")
		curr := nested
		for i, part := range parts {
			if i == len(parts)-1 {
				curr[part] = value
			} else {
				if next, ok := curr[part]; ok {
					// If it already exists, assert it is a map.
					if nextMap, ok := next.(map[string]any); ok {
						curr = nextMap
					} else {
						// If not, create a new map.
						newMap := make(map[string]any)
						curr[part] = newMap
						curr = newMap
					}
				} else {
					// Create a new nested map.
					newMap := make(map[string]any)
					curr[part] = newMap
					curr = newMap
				}
			}
		}
	}
	return nested
}
