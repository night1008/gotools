package removezero

func RemoveMapZeroNumValue(m map[string]interface{}, excludeKeysMap map[string]struct{}) (map[string]interface{}, error) {
	for k, v := range m {
		if _, ok := excludeKeysMap[k]; ok {
			continue
		}
		switch val := v.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			if val == 0 || val == 0.0 {
				delete(m, k)
			}
		}
	}
	return m, nil
}
