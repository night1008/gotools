package compare

const (
	CompareSymbolEq    = "="
	CompareSymbolNotEq = "!="
	CompareSymbolLt    = "<"
	CompareSymbolLte   = "<="
	CompareSymbolGt    = ">"
	CompareSymbolGte   = ">="
)

func Compare[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64](a, b T, op string) bool {
	switch op {
	case CompareSymbolEq:
		return a == b
	case CompareSymbolNotEq:
		return a != b
	case CompareSymbolGt:
		return a > b
	case CompareSymbolLt:
		return a < b
	case CompareSymbolGte:
		return a >= b
	case CompareSymbolLte:
		return a <= b
	}
	return false
}
