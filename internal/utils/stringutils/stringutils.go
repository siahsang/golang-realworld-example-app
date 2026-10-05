package stringutils

import (
	"fmt"
	"strconv"

	"github.com/siahsang/blog/internal/utils/functional"
)

type StringNumber interface {
	~int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

func ToString[T StringNumber](v T) string {

	switch a := any(v).(type) {
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(int64(v), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(uint64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", a)
	}
}

func ToListString[T StringNumber](v []T) []string {
	return functional.Map(v, func(item T) string { return ToString(item) })
}

func INCluseOld[T any](list []T) (placeholders []string, args []any) {
	return INClause(list, 1)
}

func INClause[T any](list []T, startIndex int) (placeholders []string, args []any) {
	placeholders = make([]string, len(list))
	args = make([]any, len(list))
	for i, id := range list {
		placeholders[i] = fmt.Sprintf("$%d", startIndex)
		args[i] = id
		startIndex++
	}

	return placeholders, args
}
