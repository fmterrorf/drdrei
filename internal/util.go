package internal

func uniqString(strs []string) []string {
	kv := make(map[string]bool)
	for _, v := range strs {
		kv[v] = true
	}
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	return keys
}
