package removezero

import "reflect"

func RemoveMapZeroNumValue(m map[string]interface{}, excludeKeysMap map[string]struct{}) map[string]interface{} {
	for k, v := range m {
		if _, ok := excludeKeysMap[k]; ok {
			continue
		}
		switch val := v.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			// 不能使用这种方式，int != int32
			// if val == 0 || val == 0.0 {
			if reflect.ValueOf(val).IsZero() {
				delete(m, k)
			}
		}
	}
	return m
}

func IsNumericZero(v interface{}) (bool, bool) {
	switch n := v.(type) {
	case int:
		return true, n == 0
	case int8:
		return true, n == 0
	case int16:
		return true, n == 0
	case int32:
		return true, n == 0
	case int64:
		return true, n == 0
	case uint:
		return true, n == 0
	case uint8:
		return true, n == 0
	case uint16:
		return true, n == 0
	case uint32:
		return true, n == 0
	case uint64:
		return true, n == 0
	case float32:
		return true, n == 0
	case float64:
		return true, n == 0
	}
	return false, false
}
