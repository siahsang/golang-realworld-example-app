package stringutils

import (
	"fmt"
	"github.com/siahsang/blog/internal/utils"
	"strconv"
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
	return utils.Map(v, func(item T) string { return ToString(item) })
}

func INCluse[T any](list []T) (placeholders []string, args []any) {
	placeholders = make([]string, len(list))
	args = make([]any, len(list))
	for i, id := range list {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	return placeholders, args
}
