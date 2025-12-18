package types

import (
	"fmt"
	"strconv"
)

type RouteParam struct {
	RouteID string
	Name    string
	Type    string
}

var GlobalRouteParams = map[string][]RouteParam{}

func GetParsedParam(id string, name string, val any) any {
	params := GlobalRouteParams[id]
	for _, param := range params {
		if param.Name == name {
			return (parseParam(val, param.Type))
		}
	}
	return fmt.Sprintf("%v", val)
}

func parseParam(val any, dtype string) any {
	strVal := fmt.Sprintf("%v", val)

	switch dtype {
	case "string":
		return strVal
	case "int":
		i, err := strconv.Atoi(strVal)
		if err != nil {
			return 0
		}
		return i
	case "int64":
		i, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return int64(0)
		}
		return i
	case "float64":
		f, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return 0.0
		}
		return f
	case "bool":
		b, err := strconv.ParseBool(strVal)
		if err != nil {
			return false
		}
		return b
	default:
		return strVal
	}
}
